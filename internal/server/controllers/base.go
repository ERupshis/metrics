package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/erupshis/metrics/internal/agent/ticker"
	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/filemngr"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/go-chi/chi/v5"
)

type BaseController struct {
	config      config.Config
	storage     memstorage.MemStorage
	logger      logger.BaseLogger
	compressor  compressor.GzipHandler
	fileManager *filemngr.FileManager
}

func CreateBase(config config.Config, logger logger.BaseLogger) *BaseController {
	controller := &BaseController{
		config:      config,
		storage:     memstorage.Create(),
		logger:      logger,
		compressor:  compressor.GzipHandler{},
		fileManager: filemngr.Create(),
	}

	controller.restoreDataFromFileIfNeed()
	return controller
}

func (c *BaseController) GetConfig() *config.Config {
	return &c.config
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Use(c.logger.LogHandler)
	r.Use(c.compressor.GzipHandle)

	r.Get("/", c.ListHandler)

	r.Route("/{request}", func(r chi.Router) {
		r.Post("/", c.jsonHandler)
		r.Route("/{type}", func(r chi.Router) {
			r.HandleFunc("/", c.missingNameHandler)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", c.getHandler)
				r.Post("/{value}", c.postHandler)
			})
		})
	})
	r.NotFound(c.badRequestHandler)
	return r
}

func (c *BaseController) badRequestHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (c *BaseController) missingNameHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

const (
	postRequest = "update"
	getRequest  = "value"

	gaugeType   = "gauge"
	counterType = "counter"
)

// JSON HANDLER
func (c *BaseController) jsonHandler(w http.ResponseWriter, r *http.Request) {
	request := chi.URLParam(r, "request")

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	c.logger.Info("[BaseController::jsonHandler] Handle JSON request with body: %s", buf.String())
	data, err := networkmsg.ParsePostValueMessage(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch request {
	case postRequest:
		c.jsonPostHandler(w, &data)
	case getRequest:
		c.jsonGetHandler(w, &data)
	}
}

func (c *BaseController) jsonPostHandler(w http.ResponseWriter, data *networkmsg.Metrics) {
	if data.MType == gaugeType {
		valueIn := new(float64)
		if data.Value != nil {
			valueIn = data.Value
		}
		c.storage.AddGauge(data.ID, *valueIn)
		valueOut, _ := c.storage.GetGauge(data.ID)
		data.Value = &valueOut
	} else if data.MType == counterType {
		valueIn := new(int64)
		if data.Delta != nil {
			valueIn = data.Delta
		}
		c.storage.AddCounter(data.ID, *valueIn)
		value, _ := c.storage.GetCounter(data.ID)
		data.Delta = &value
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(networkmsg.CreatePostUpdateMessage(*data))
}

func (c *BaseController) jsonGetHandler(w http.ResponseWriter, data *networkmsg.Metrics) {
	if data.MType == gaugeType {
		value, err := c.storage.GetGauge(data.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		data.Value = &value
	} else if data.MType == counterType {
		value, err := c.storage.GetCounter(data.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		data.Delta = &value
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(networkmsg.CreatePostUpdateMessage(*data))
}

// postHandler POST HTTP REQUEST HANDLING.
func (c *BaseController) postHandler(w http.ResponseWriter, r *http.Request) {
	request := chi.URLParam(r, "request")
	valueType := chi.URLParam(r, "type")

	if request != postRequest {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if valueType == gaugeType {
		c.postGaugeHandler(w, r)
		return
	} else if valueType == counterType {
		c.postCounterHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (c *BaseController) postCounterHandler(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	c.logger.Info("[BaseController::postCounterHandler] Handle url post request for: '%s'(%s) value", name, value)
	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		c.storage.AddCounter(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (c *BaseController) postGaugeHandler(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	c.logger.Info("[BaseController::postGaugeHandler] Handle url post request for: '%s'(%s) value", name, value)
	if val, err := strconv.ParseFloat(value, 64); err == nil {
		c.storage.AddGauge(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// getHandler GET HTTP REQUEST HANDLING.
func (c *BaseController) getHandler(w http.ResponseWriter, r *http.Request) {
	request := chi.URLParam(r, "request")
	valueType := chi.URLParam(r, "type")

	if request != getRequest {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if valueType == gaugeType {
		c.getGaugeHandler(w, r)
		return
	} else if valueType == counterType {
		c.getCounterHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (c *BaseController) getCounterHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	c.logger.Info("[BaseController::getCounterHandler] Handle url get request for: '%s' value", name)
	if value, err := c.storage.GetCounter(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, fmt.Sprintf("%d", value)); err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (c *BaseController) getGaugeHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	c.logger.Info("[BaseController::getGaugeHandler] Handle url get request for: '%s' value", name)
	if value, err := c.storage.GetGauge(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// HTML METRICS LIST PROCESSING.

const tmplMap = `
<html><body>
<caption>GAUGES</caption>
<table border = 2>
{{- range $key, $value := .Gauges}}
<tr><td>{{ $key }}</td><td>{{ $value }}</td></tr>
{{- end}}
</table>

<caption>COUNTERS</caption>
<table border = 2>
{{- range $key, $value := .Counters}}
<tr><td>{{ $key }}</td><td>{{ $value }}</td></tr>
{{- end}}
</table>
</body></html>
`

type tmplData struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func (c *BaseController) ListHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.New("mapTemplate").Parse(tmplMap)
	if err != nil {
		fmt.Println("Error parsing gauge template:", err)
		return
	}

	gaugesMap := c.storage.GetAllGauges()
	countersMap := c.storage.GetAllCounters()

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.Execute(w, tmplData{gaugesMap, countersMap}); err != nil {
		panic(err)
	}
}

// FILE METRICS MANAGING.

func (c *BaseController) ScheduleDataStoringInFile() *time.Ticker {
	var interval int64 = 1
	if c.config.StoreInterval > 1 {
		interval = c.config.StoreInterval
	}

	interval = 10
	storeTicker := ticker.CreateWithSecondsInterval(interval)
	go ticker.Run(storeTicker, func() { c.saveMetricsInFile() })
	return storeTicker
}

func (c *BaseController) saveMetricsInFile() {
	if !c.fileManager.IsFileOpen() {
		if err := c.fileManager.OpenFile(c.config.StoragePath, true); err != nil {
			c.logger.Info("[BaseController::SameMetricsInFile] cannot save metrics data in file. Failed to open '%s' file.", c.config.StoragePath)
			return
		}
		defer c.fileManager.CloseFile()
	}

	for name, val := range c.storage.GetAllGauges() {
		c.fileManager.WriteMetric(name, val)
	}

	for name, val := range c.storage.GetAllCounters() {
		c.fileManager.WriteMetric(name, val)
	}
}

func (c *BaseController) restoreDataFromFileIfNeed() {
	if !c.config.Restore {
		return
	}

	if !c.fileManager.IsFileOpen() {
		if err := c.fileManager.OpenFile(c.config.StoragePath, false); err != nil {
			c.logger.Info("[BaseController::SameMetricsInFile] cannot read metrics from file. Failed to open '%s' file.", c.config.StoragePath)
			return
		}
		defer c.fileManager.CloseFile()
	}

	metric, _ := c.fileManager.ScanMetric()
	for metric != nil {
		switch metric.ValueType {
		case "gauge":
			value, err := strconv.ParseFloat(metric.Value, 64)
			if err != nil {
				c.logger.Info("[BaseController::restoreDataFromFileIfNeed] failed to parse float64 value for '%s'", metric.Name)
			}
			c.storage.AddGauge(metric.Name, value)
		case "counter":
			value, err := strconv.ParseInt(metric.Value, 10, 64)
			if err != nil {
				c.logger.Info("[BaseController::restoreDataFromFileIfNeed] failed to parse int64 value for '%s'", metric.Name)
			}
			c.storage.AddCounter(metric.Name, value)
		}

		metric, _ = c.fileManager.ScanMetric()
	}
}
