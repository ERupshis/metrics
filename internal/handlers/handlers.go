package handlers

import (
	"fmt"
	"github.com/ERupshis/metrics/internal/memstorage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
)

var storage = memstorage.CreateStorage()

func Invalid(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func MissingName(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func ListHandler(w http.ResponseWriter, _ *http.Request) {
	if _, err := io.WriteString(w, "<html><body>"); err != nil {
		panic(err)
	}

	if _, err := io.WriteString(w, "<caption>GAUGES</caption><table border = 2>"); err != nil {
		panic(err)
	}
	for name, value := range storage.GetAllGauges() {
		if _, err := io.WriteString(w, fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", name, strconv.FormatFloat(value, 'f', -1, 64))); err != nil {
			panic(err)
		}
	}
	if _, err := io.WriteString(w, "</table>"); err != nil {
		panic(err)
	}

	if _, err := io.WriteString(w, "<caption>COUNTERS</caption><table border = 2>"); err != nil {
		panic(err)
	}
	for name, value := range storage.GetAllCounters() {
		if _, err := io.WriteString(w, fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", name, value)); err != nil {
			panic(err)
		}
	}

	if _, err := io.WriteString(w, "</body></html>"); err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func PostCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//TODO: still actual?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		storage.AddCounter(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func PostGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//TODO: still actual?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	if val, err := strconv.ParseFloat(value, 64); err == nil {
		storage.AddGauge(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func GetCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//TODO: still actual?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := chi.URLParam(r, "name")

	if value, err := storage.GetCounter(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, fmt.Sprintf("%d", value)); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func GetGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//TODO: still actual?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := chi.URLParam(r, "name")

	if value, err := storage.GetGauge(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
