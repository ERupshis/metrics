package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/go-chi/chi/v5"
)

type BaseController struct {
	config  config.Config
	storage *memstorage.MemStorage
}

func CreateBase() *BaseController {
	return &BaseController{config.Parse(), memstorage.Create()}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", c.ListHandler)
	r.Route("/update", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Get("/", c.MissingName)
			r.Post("/", c.MissingName)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/{value}", c.PostCounter)
			})

		})
		r.Route("/gauge", func(r chi.Router) {
			r.Get("/", c.MissingName)
			r.Post("/", c.MissingName)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/{value}", c.PostGauge)
			})
		})
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Get("/", c.MissingName)
			r.Post("/", c.MissingName)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", c.GetCounter)
			})

		})
		r.Route("/gauge", func(r chi.Router) {
			r.Get("/", c.MissingName)
			r.Post("/", c.MissingName)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", c.GetGauge)
			})
		})
	})

	r.NotFound(c.Invalid)

	return r
}

func (c *BaseController) GetConfig() *config.Config {
	return &c.config
}

func (c *BaseController) Invalid(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (c *BaseController) MissingName(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *BaseController) PostCounter(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		c.storage.AddCounter(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (c *BaseController) PostGauge(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	if val, err := strconv.ParseFloat(value, 64); err == nil {
		c.storage.AddGauge(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (c *BaseController) GetCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if value, err := c.storage.GetCounter(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, fmt.Sprintf("%d", value)); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (c *BaseController) GetGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if value, err := c.storage.GetGauge(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
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
	if err := tmpl.Execute(w, tmplData{gaugesMap, countersMap}); err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
}
