package base_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/ipvalidator"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/controllers/base"
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
		KeyRSA:        "../../../../rsa/key.pem",
	}
}

// ExampleBaseController_ListHandler demonstrates how to use the ListHandler for displaying metrics in HTML format.
func ExampleBaseController_ListHandler() {
	// Create a sample configuration.
	cfg := createExampleConfig()

	// Create a log, memstorage, and hasher.
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// RSA message decoder.
	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(context.Background(), cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	storage.AddGauge("example", 42.0)
	storage.AddCounter("example", 10)

	// RSA message encoder.
	encoder, err := rsa.CreateEncoder("../../../../rsa/cert.pem")
	if err != nil {
		log.Info("rsa encoder: %v", err)
	}

	// Create a test request for the "/list" endpoint.
	encryptedBody, _ := encoder.Encode(nil)
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer(encryptedBody))
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
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// RSA message decoder.
	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(context.Background(), cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	// RSA message encoder.
	encoder, err := rsa.CreateEncoder("../../../../rsa/cert.pem")
	if err != nil {
		log.Info("rsa encoder: %v", err)
	}

	encryptedBody, _ := encoder.Encode(nil)
	// Create a test request for the "/ping" endpoint.
	req := httptest.NewRequest(http.MethodGet, "/ping", bytes.NewBuffer(encryptedBody))
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
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// RSA message decoder.
	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(context.Background(), cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	// Create an array of test JSON requests for different request types.
	requests := []string{"update", "value", "updates"}

	// RSA message encoder.
	encoder, err := rsa.CreateEncoder("../../../../rsa/cert.pem")
	if err != nil {
		log.Info("rsa encoder: %v", err)
	}

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

		encryptedBody, _ := encoder.Encode(jsonBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s", requestType), bytes.NewBuffer(encryptedBody))
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
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// RSA message decoder.
	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(context.Background(), cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	// RSA message encoder.
	encoder, err := rsa.CreateEncoder("../../../../rsa/cert.pem")
	if err != nil {
		log.Info("rsa encoder: %v", err)
	}

	// Create an array of test requests for different metric types.
	types := []string{"gauge", "counter"}

	for _, metricType := range types {
		// Customize the request based on the metric type.
		encryptedBody, _ := encoder.Encode(nil)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/update/%s", metricType), bytes.NewBuffer(encryptedBody))
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
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// RSA message decoder.
	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(context.Background(), cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	// Add some sample data to the storage for testing.
	storage.AddGauge("example", 42.0)
	storage.AddCounter("example", 10)

	// RSA message encoder.
	encoder, err := rsa.CreateEncoder("../../../../rsa/cert.pem")
	if err != nil {
		log.Info("rsa encoder: %v", err)
	}

	// Create an array of test requests for different metric types.
	types := []string{"gauge", "counter"}

	for _, metricType := range types {
		// Customize the request based on the metric type.
		encryptedBody, _ := encoder.Encode(nil)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/example", metricType), bytes.NewBuffer(encryptedBody))
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
	log := logger.CreateMock()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)
	hashManager := hasher.CreateHasher(cfg.Key, hasher.SHA256, log)

	// RSA message decoder.
	decoder, err := rsa.CreateDecoder(cfg.KeyRSA)
	if err != nil {
		log.Info("rsa decoder: %v", err)
	}
	// Create a HTTPController instance.
	baseController := base.Create(context.Background(), cfg, log, storage, hashManager, decoder, ipvalidator.Create(nil))

	// Add some sample data to the storage for testing.
	storage.AddGauge("example", 42.0)
	storage.AddCounter("example", 10)

	// Create an array of test requests for different metric types.
	names := []string{"gauge_metric", "counter_metric"}
	types := []string{"gauge", "counter"}
	values := []string{"42.0", "10"}

	// RSA message encoder.
	encoder, err := rsa.CreateEncoder("../../../../rsa/cert.pem")
	if err != nil {
		log.Info("rsa encoder: %v", err)
	}

	for i := 0; i < len(names); i++ {
		// Customize the request based on the metric type.
		encryptedBody, _ := encoder.Encode(nil)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/%s/%s/%s", types[i], names[i], values[i]), bytes.NewBuffer(encryptedBody))
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
