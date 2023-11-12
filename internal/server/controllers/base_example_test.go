package controllers_test

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
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
)

func createExampleConfig() config.Config {
	return config.Config{
		Host:          "localhost:8080",
		LogLevel:      "Info",
		Restore:       true,
		StoreInterval: 5,
		StoragePath:   "/tmp/metrics-db.json",
		DataBaseDSN:   "postgres://postgres:postgres@localhost:5432/metrics_db?sslmode=disable",
		Key:           "",
	}
}

// ExampleBaseController_ListHandler demonstrates how to use the ListHandler for displaying metrics in HTML format.
func ExampleBaseController_ListHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateLogger("info")
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	storage.AddGauge("example", 42.0)
	storage.AddCounter("example", 10)

	// Create a test request for the "/list" endpoint.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Use the router to handle the request.
	baseController.Route().ServeHTTP(w, req)

	// Output the response status code.
	fmt.Println("Response Status Code:", w.Code)

	// Output:
	// Response Status Code: 200
}

// ExampleBaseController_checkStorageHandler demonstrates how to use the checkStorageHandler for the "/ping" endpoint.
func ExampleBaseController_checkStorageHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateLogger("info")
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	// Create a test request for the "/ping" endpoint.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	// Use the router to handle the request.
	baseController.Route().ServeHTTP(w, req)

	// Output the response status code.
	fmt.Println("Response Status Code:", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	// Output:
	// Response Status Code: 200
	// Response Body:
}

// ExampleBaseController_jsonHandler demonstrates how to use the jsonHandler for different JSON request types.
func ExampleBaseController_jsonHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateLogger("info")
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	// Create an array of test JSON requests for different request types.
	requests := []string{"update", "value", "updates"}

	for _, requestType := range requests {
		// Customize the JSON body based on the request type.
		var jsonBody []byte

		switch requestType {
		case "update":
			jsonBody = []byte(`{"id": "example", "type": "gauge", "value": 42.0}`)
		case "value":
			jsonBody = []byte(`{"id": "example", "type": "gauge", "value": 42.0}`)
		case "updates":
			jsonBody = []byte(`[{"id": "example1", "type": "gauge", "value": 42.0}, {"id": "example2", "type": "counter", "delta": 10}]`)
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s", requestType), bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		// Use the router to handle the request.
		baseController.Route().ServeHTTP(w, req)

		// Output the response status code.
		fmt.Printf("Response Status Code for %s: %d\n", requestType, w.Code)
		fmt.Printf("Response Body: %s\n", w.Body.String())
	}

	// Output:
	// Response Status Code for update: 200
	// Response Body: {"id":"example","type":"gauge","value":42}
	// Response Status Code for value: 200
	// Response Body: {"id":"example","type":"gauge","value":42}
	// Response Status Code for updates: 200
	// Response Body: {}

}

// ExampleBaseController_missingNameHandler demonstrates how to use the missingNameHandler for different metric types.
func ExampleBaseController_missingNameHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateLogger("info")
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	// Create an array of test requests for different metric types.
	types := []string{"gauge", "counter"}

	for _, metricType := range types {
		// Customize the request based on the metric type.
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/update/%s", metricType), nil)
		w := httptest.NewRecorder()

		// Use the router to handle the request.
		baseController.Route().ServeHTTP(w, req)

		// Output the response status code.
		fmt.Printf("Response Status Code for missingNameHandler with type %s: %d\n", metricType, w.Code)
	}

	// Output:
	// Response Status Code for missingNameHandler with type gauge: 404
	// Response Status Code for missingNameHandler with type counter: 404
}

// ExampleBaseController_getHandler demonstrates how to use the getHandler for different metric types.
func ExampleBaseController_getHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateLogger("info")
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	// Add some sample data to the storage for testing.
	storage.AddGauge("example", 42.0)
	storage.AddCounter("example", 10)

	// Create an array of test requests for different metric types.
	types := []string{"gauge", "counter"}

	for _, metricType := range types {
		// Customize the request based on the metric type.
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/example", metricType), nil)
		w := httptest.NewRecorder()

		// Use the router to handle the request.
		baseController.Route().ServeHTTP(w, req)

		// Output the response status code and body.
		fmt.Printf("Response Status Code for getHandler with type %s: %d\n", metricType, w.Code)
		fmt.Printf("Response Body for getHandler with type %s: %s\n", metricType, w.Body.String())
	}

	// Output:
	// Response Status Code for getHandler with type gauge: 200
	// Response Body for getHandler with type gauge: 42
	// Response Status Code for getHandler with type counter: 200
	// Response Body for getHandler with type counter: 10
}

func ExampleBaseController_postHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateLogger("info")
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	// Add some sample data to the storage for testing.
	storage.AddGauge("example", 42.0)
	storage.AddCounter("example", 10)

	// Create an array of test requests for different metric types.
	names := []string{"gauge_metric", "counter_metric"}
	types := []string{"gauge", "counter"}
	values := []string{"42.0", "10"}

	for i := 0; i < len(names); i++ {
		// Customize the request based on the metric type.
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/%s/%s/%s", types[i], names[i], values[i]), nil)
		w := httptest.NewRecorder()

		// Use the router to handle the request.
		baseController.Route().ServeHTTP(w, req)

		// Output the response status code and body.
		fmt.Printf("Response Status Code for postHandler: %d\n", w.Code)
	}
	// Output:
	// Response Status Code for postHandler: 200
	// Response Status Code for postHandler: 200
}

func BenchmarkBaseController_postData(b *testing.B) {
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
	log, _ := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

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
	log, _ := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	for i := 0; i < len(metrics); i++ {
		// Customize the request based on the metric type.
		var req *http.Request
		if metrics[i].MType == "gauge" {
			buf := bytes.NewReader(networkmsg.CreatePostUpdateMessage(metrics[i]))
			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/"), buf)
		} else {
			buf := bytes.NewReader(networkmsg.CreatePostUpdateMessage(metrics[i]))
			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/"), buf)
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
	log, _ := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// Create a BaseController instance.
	baseController := controllers.CreateBase(context.Background(), cfg, log, storage, hashManager)

	var req *http.Request
	body, _ := json.Marshal(&metrics)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/updates/"), bytes.NewBuffer(body))

	w := httptest.NewRecorder()

	// Use the router to handle the request.
	baseController.Route().ServeHTTP(w, req)
}
