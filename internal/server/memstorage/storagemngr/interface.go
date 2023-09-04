package storagemngr

type MetricData struct {
	Name      string `json:"name"`
	ValueType string `json:"type"`
	Value     string `json:"value"`
}

//go:generate mockgen -destination=./mocks/mock_StorageManager.go -package=mocks github.com/erupshis/metrics/internal/server/memstorage/storagemngr StorageManager
type StorageManager interface {
	SaveMetricsInStorage(gaugeValues map[string]interface{}, counterValues map[string]interface{})
	RestoreDataFromStorage() (map[string]float64, map[string]int64)
	CheckConnection() bool
	Close() error
}
