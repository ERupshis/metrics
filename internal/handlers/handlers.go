package handlers

import (
	"errors"
	"github.com/ERupshis/metrics/internal/memstorage"
	"net/http"
	"strconv"
	"strings"
)

var storage = memstorage.CreateStorage()

func splitCounterAndGaugeRequest(req string) ([]string, error) {
	request := strings.Split(req, "/")
	if len(request) != 5 || request[4] == "" {
		return []string{}, errors.New("incorrect request size")
	}

	return request, nil
}

func Invalid(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func Counter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, err := splitCounterAndGaugeRequest(r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if val, err := strconv.ParseInt(request[4], 10, 64); err == nil {
		storage.AddCounter(request[3], val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func Gauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, err := splitCounterAndGaugeRequest(r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if val, err := strconv.ParseFloat(request[4], 64); err == nil {
		storage.AddGauge(request[3], val)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
