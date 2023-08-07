package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage *memstorage.MemStorage
}

func Create(storage *memstorage.MemStorage) *Handler {
	return &Handler{storage}
}

func (h *Handler) Invalid(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (h *Handler) MissingName(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) ListHandler(w http.ResponseWriter, _ *http.Request) {
	if _, err := io.WriteString(w, "<html><body>"); err != nil {
		panic(err)
	}

	if _, err := io.WriteString(w, "<caption>GAUGES</caption><table border = 2>"); err != nil {
		panic(err)
	}
	for name, value := range h.storage.GetAllGauges() {
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
	for name, value := range h.storage.GetAllCounters() {
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

func (h *Handler) PostCounter(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		h.storage.AddCounter(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PostGauge(w http.ResponseWriter, r *http.Request) {
	name, value := chi.URLParam(r, "name"), chi.URLParam(r, "value")

	if val, err := strconv.ParseFloat(value, 64); err == nil {
		h.storage.AddGauge(name, val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//TODO: still actual?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := chi.URLParam(r, "name")

	if value, err := h.storage.GetCounter(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, fmt.Sprintf("%d", value)); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) GetGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//TODO: still actual?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := chi.URLParam(r, "name")

	if value, err := h.storage.GetGauge(name); err == nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
