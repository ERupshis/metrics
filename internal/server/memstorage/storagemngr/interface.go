package storagemngr

import "context"

// MetricData represents the structure of metric data, including the metric name, value type, and value.
type MetricData struct {
	Name      string `json:"name"`
	ValueType string `json:"type"`
	Value     string `json:"value"`
}

// StorageManager is an interface that defines methods for managing the storage of metric data.
// Implementations of this interface handle tasks such as saving metrics, restoring data,
// checking connection status, and closing the storage.
//
//go:generate mockgen -destination=../../../../mocks/mock_StorageManager.go -package=mocks github.com/erupshis/metrics/internal/server/memstorage/storagemngr StorageManager
type StorageManager interface {
	// SaveMetricsInStorage saves gauge and counter metric values in the storage.
	// The provided context is used for cancellation and timeout.
	SaveMetricsInStorage(ctx context.Context, gaugeValues map[string]interface{}, counterValues map[string]interface{}) error

	// RestoreDataFromStorage retrieves and restores stored metric data from the storage.
	// The provided context is used for cancellation and timeout.
	// It returns two maps containing gauge and counter metric values respectively.
	RestoreDataFromStorage(ctx context.Context) (map[string]float64, map[string]int64, error)

	// CheckConnection checks the connection status to the storage.
	// The provided context is used for cancellation and timeout.
	// It returns a boolean indicating whether the connection is available and an error if any.
	CheckConnection(ctx context.Context) (bool, error)

	// Close closes the storage, releasing any resources associated with it.
	// It returns an error if the closure process encounters any issues.
	Close() error
}
