// Package base provides HTTP request handling functionality for metrics management.
// It includes a HTTPController struct that serves as the foundation for handling various HTTP endpoints
// related to metrics, such as adding, retrieving, and displaying gauge and counter values.
// The package also contains middleware and utility functions for configuring and managing
// logging, data compression, hashing, and storage operations.
package base

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/ipvalidator"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/go-chi/chi/v5"
)

// HTTPController represents the base controller for handling HTTP requests and managing metrics.
type HTTPController struct {
	config      config.Config
	storage     memstorage.MemStorage
	logger      logger.BaseLogger
	compressor  compressor.GzipHandler
	hash        *hasher.Hasher
	decoder     *rsa.Decoder
	validatorIP *ipvalidator.ValidatorIP
}

// Create initializes and returns a new instance of HTTPController.
// It takes a context, configuration, logger, MemStorage, and Hasher as parameters.
// If data restoration is enabled, it attempts to restore data from a file.
func Create(ctx context.Context, config config.Config, logger logger.BaseLogger, storage *memstorage.MemStorage, hash *hasher.Hasher, decoder *rsa.Decoder, validatorIP *ipvalidator.ValidatorIP) *HTTPController {
	controller := &HTTPController{
		config:      config,
		storage:     *storage,
		logger:      logger,
		compressor:  compressor.GzipHandler{},
		hash:        hash,
		decoder:     decoder,
		validatorIP: validatorIP,
	}

	if !controller.config.Restore {
		controller.logger.Info("[HTTPController::Create] data restoring from file switched off.")
	} else {
		err := controller.storage.RestoreData(ctx)
		if err != nil {
			controller.logger.Info("[HTTPController::Create] data restoring: %v", err)
		}
	}

	return controller
}

// GetConfig returns a pointer to the configuration of the HTTPController.
func (c *HTTPController) GetConfig() *config.Config {
	return &c.config
}

// Route returns a new chi.Mux router configured with middleware and handlers for HTTPController.
func (c *HTTPController) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Use(c.logger.LogHandler)
	r.Use(c.validatorIP.ValidateIPHandler)
	r.Use(c.decoder.DecodeRSAHandler)
	r.Use(c.hash.Handler)
	r.Use(c.compressor.GzipHandle)

	r.Get("/", c.ListHandler)
	r.Get("/ping", c.checkStorageHandler)
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

// badRequestHandler handles HTTP requests with a status of BadRequest (400).
func (c *HTTPController) badRequestHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

// missingNameHandler handles HTTP requests with a status of NotFound (404).
func (c *HTTPController) missingNameHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// checkStorageHandler handles the "/ping" endpoint to check the availability of storage.
func (c *HTTPController) checkStorageHandler(w http.ResponseWriter, r *http.Request) {
	c.hash.WriteHashHeaderInResponseIfNeed(w, []byte{})

	if _, err := c.storage.IsAvailable(r.Context()); err != nil {
		c.logger.Info("[HTTPController:checkStorageHandler] storage is not available, error: %v")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

const (
	postBatchRequest = "updates"
	postRequest      = "update"
	getRequest       = "value"

	gaugeType   = "gauge"
	counterType = "counter"
)

// jsonHandler handles JSON requests and delegates to specific handlers based on the request type.
func (c *HTTPController) jsonHandler(w http.ResponseWriter, r *http.Request) {
	request := chi.URLParam(r, "request")

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	// c.logger.Info("[HTTPController::jsonHandler] Handle JSON request with body: %s", buf.String())

	var responseBody []byte
	switch request {
	case postRequest:
		metric, err := networkmsg.ParsePostValueMessage(buf.Bytes())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		responseBody = c.jsonPostHandler(w, &metric)

	case postBatchRequest:
		data, err := networkmsg.ParsePostBatchValueMessage(buf.Bytes())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		responseBody = c.jsonPostBatchHandler(w, data)

	case getRequest:
		metric, err := networkmsg.ParsePostValueMessage(buf.Bytes())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		responseBody = c.jsonGetHandler(w, &metric)
	}

	if responseBody == nil {
		return
	}

	c.hash.WriteHashHeaderInResponseIfNeed(w, responseBody)
	_, _ = w.Write(responseBody)
}

// jsonPostBatchHandler handles batch JSON requests and adds metrics to storage.
func (c *HTTPController) jsonPostBatchHandler(w http.ResponseWriter, metrics []networkmsg.Metric) []byte {
	for _, metric := range metrics {
		c.addMetricFromMessage(&metric)
	}

	w.Header().Add("Content-Type", "application/json")
	return []byte("{}")
}

// jsonPostHandler handles single JSON requests and adds a metric to storage.
func (c *HTTPController) jsonPostHandler(w http.ResponseWriter, data *networkmsg.Metric) []byte {
	c.addMetricFromMessage(data)
	w.Header().Add("Content-Type", "application/json")
	return networkmsg.CreatePostUpdateMessage(*data)
}

// addMetricFromMessage adds a metric to storage based on the metric type.
func (c *HTTPController) addMetricFromMessage(data *networkmsg.Metric) {
	switch data.MType {
	case gaugeType:
		valueIn := new(float64)
		if data.Value != nil {
			valueIn = data.Value
		}
		c.storage.AddGauge(data.ID, *valueIn)
		valueOut, _ := c.storage.GetGauge(data.ID)
		data.Value = &valueOut
	case counterType:
		valueIn := new(int64)
		if data.Delta != nil {
			valueIn = data.Delta
		}
		c.storage.AddCounter(data.ID, *valueIn)
		value, _ := c.storage.GetCounter(data.ID)
		data.Delta = &value
	}
}

// jsonGetHandler handles JSON GET requests and retrieves metrics from storage.
func (c *HTTPController) jsonGetHandler(w http.ResponseWriter, data *networkmsg.Metric) []byte {
	switch data.MType {
	case gaugeType:
		value, err := c.storage.GetGauge(data.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil
		}
		data.Value = &value
	case counterType:
		value, err := c.storage.GetCounter(data.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil
		}
		data.Delta = &value
	}

	w.Header().Add("Content-Type", "application/json")
	return networkmsg.CreatePostUpdateMessage(*data)
}

// postHandler handles HTTP POST requests based on the request and type.
func (c *HTTPController) postHandler(w http.ResponseWriter, r *http.Request) {
	request := chi.URLParam(r, "request")
	valueType := chi.URLParam(r, "type")

	if request != postRequest {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch valueType {
	case gaugeType:
		c.postGaugeHandler(w, r)
		return
	case counterType:
		c.postCounterHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

// postCounterHandler handles HTTP POST requests for counter metrics.
func (c *HTTPController) postCounterHandler(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	c.logger.Info("[HTTPController::postCounterHandler] handle url post request for: '%s'(%s) value", name, value)
	c.hash.WriteHashHeaderInResponseIfNeed(w, []byte{})

	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		c.storage.AddCounter(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// postGaugeHandler handles HTTP POST requests for gauge metrics.
func (c *HTTPController) postGaugeHandler(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	c.logger.Info("[HTTPController::postGaugeHandler] handle url post request for: '%s'(%s) value", name, value)
	c.hash.WriteHashHeaderInResponseIfNeed(w, []byte{})

	if val, err := strconv.ParseFloat(value, 64); err == nil {
		c.storage.AddGauge(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// getHandler handles HTTP GET requests based on the request and type.
func (c *HTTPController) getHandler(w http.ResponseWriter, r *http.Request) {
	request := chi.URLParam(r, "request")
	valueType := chi.URLParam(r, "type")

	if request != getRequest {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch valueType {
	case gaugeType:
		c.getGaugeHandler(w, r)
		return
	case counterType:
		c.getCounterHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

// getCounterHandler handles HTTP GET requests for counter metrics.
func (c *HTTPController) getCounterHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	c.logger.Info("[HTTPController::getCounterHandler] handle url get request for: '%s' value", name)
	if value, err := c.storage.GetCounter(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		responseBody := fmt.Sprintf("%d", value)
		c.hash.WriteHashHeaderInResponseIfNeed(w, []byte(responseBody))
		if _, err = io.WriteString(w, responseBody); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		c.logger.Info("[HTTPController::getGaugeHandler] counter metric not found error: %v", err)
		c.hash.WriteHashHeaderInResponseIfNeed(w, []byte{})
		w.WriteHeader(http.StatusNotFound)
	}
}

// getGaugeHandler handles HTTP GET requests for gauge metrics.
func (c *HTTPController) getGaugeHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	c.logger.Info("[HTTPController::getGaugeHandler] handle url get request for: '%s' value", name)
	if value, err := c.storage.GetGauge(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		responseBody := strconv.FormatFloat(value, 'f', -1, 64)
		c.hash.WriteHashHeaderInResponseIfNeed(w, []byte(responseBody))
		if _, err = io.WriteString(w, responseBody); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		c.logger.Info("[HTTPController::getGaugeHandler] gauge metric not found error: %v", err)
		c.hash.WriteHashHeaderInResponseIfNeed(w, []byte{})
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
	Gauges   map[string]interface{}
	Counters map[string]interface{}
}

// ListHandler handles HTTP requests to display a list of gauges and counters in HTML format.
func (c *HTTPController) ListHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.New("mapTemplate").Parse(tmplMap)
	if err != nil {
		c.logger.Info("[HTTPController:ListHandler] error parsing gauge template: %v", err)
		return
	}

	gaugesMap := c.storage.GetAllGauges()
	countersMap := c.storage.GetAllCounters()

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	buf := bytes.Buffer{}
	writer := bufio.NewWriter(&buf)

	if err = tmpl.Execute(writer, tmplData{gaugesMap, countersMap}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	err = writer.Flush()
	if err != nil {
		c.logger.Info("[HTTPController:ListHandler] flush writer failed: %v", err)
	}

	c.hash.WriteHashHeaderInResponseIfNeed(w, buf.Bytes())
	_, err = w.Write(buf.Bytes())
	if err != nil {
		c.logger.Info("[HTTPController:ListHandler] failed to write body: %v", err)
	}
}
