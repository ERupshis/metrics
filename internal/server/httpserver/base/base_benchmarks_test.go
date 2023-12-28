package base_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/ipvalidator"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/server/httpserver/base"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
)

func BenchmarkBaseController_postMetrics(b *testing.B) {
	size := 100
	testSlice := generateRandomMetricsSlice(size)

	b.Run("post via URI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			postURI(testSlice)
		}
	})

	b.Run("post via postJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			postJSON(testSlice)
		}
	})

	b.Run("post via postBatchJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			postBatchJSON(testSlice)
		}
	})
}

func generateRandomMetricsSlice(size int) []networkmsg.Metric {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	types := []string{"gauge", "counter"}

	var res []networkmsg.Metric
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key_%d", i)
		typeNum := rand.Intn(2)
		tmpMetric := networkmsg.Metric{
			ID:    key,
			MType: types[typeNum],
		}

		if typeNum == 0 {
			valueFloat := float64(rand.Intn(200))
			tmpMetric.Value = &valueFloat
		} else {
			valueInt64 := int64(rand.Intn(200))
			tmpMetric.Delta = &valueInt64
		}

		res = append(res, tmpMetric)
	}

	return res
}

func postURI(metrics []networkmsg.Metric) {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(context.Background(), &cfg, storageManager, log)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(&cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	for i := 0; i < len(metrics); i++ {
		// Customize the request based on the metric type.
		var req *http.Request
		if metrics[i].MType == "gauge" {
			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/%s/%s/%f", metrics[i].MType, metrics[i].ID, *metrics[i].Value), nil)
		} else {
			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/%s/%s/%d", metrics[i].MType, metrics[i].ID, *metrics[i].Delta), nil)
		}

		w := httptest.NewRecorder()

		// Use the router to handle the request.
		baseController.Route().ServeHTTP(w, req)
	}
}

func postJSON(metrics []networkmsg.Metric) {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(context.Background(), &cfg, storageManager, log)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(&cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	for i := 0; i < len(metrics); i++ {
		// Customize the request based on the metric type.
		var req *http.Request
		if metrics[i].MType == "gauge" {
			buf := bytes.NewReader(networkmsg.CreatePostUpdateMessage(metrics[i]))
			req = httptest.NewRequest(http.MethodPost, "/update/", buf)
		} else {
			buf := bytes.NewReader(networkmsg.CreatePostUpdateMessage(metrics[i]))
			req = httptest.NewRequest(http.MethodPost, "/update/", buf)
		}

		w := httptest.NewRecorder()

		// Use the router to handle the request.
		baseController.Route().ServeHTTP(w, req)
	}
}

func postBatchJSON(metrics []networkmsg.Metric) {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(context.Background(), &cfg, storageManager, log)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(&cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	var req *http.Request
	body, _ := json.Marshal(&metrics)
	req = httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))

	w := httptest.NewRecorder()

	// Use the router to handle the request.
	baseController.Route().ServeHTTP(w, req)
}

func BenchmarkBaseController_getMetrics(b *testing.B) {
	size := 100
	testSlice := generateRandomMetricsSlice(size)

	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(context.Background(), &cfg, storageManager, log)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(&cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	var req *http.Request
	body, _ := json.Marshal(testSlice)
	req = httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	baseController.Route().ServeHTTP(w, req)

	b.Run("get via URI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getURI(baseController, testSlice)
		}
	})

	b.Run("get via postJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getJSON(baseController, testSlice)
		}
	})
}

func getURI(controller *base.HTTPController, metrics []networkmsg.Metric) {
	for i := 0; i < len(metrics); i++ {
		// Customize the request based on the metric type.
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/value/%s/%s", metrics[i].MType, metrics[i].ID), nil)
		w := httptest.NewRecorder()

		// Use the router to handle the request.
		controller.Route().ServeHTTP(w, req)
	}
}

func getJSON(controller *base.HTTPController, metrics []networkmsg.Metric) {
	for i := 0; i < len(metrics); i++ {
		// Customize the request based on the metric type.
		var req *http.Request
		body, _ := json.Marshal(metrics[i])
		req = httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		// Use the router to handle the request.
		controller.Route().ServeHTTP(w, req)
	}
}
