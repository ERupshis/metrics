package storagemanager

type StorageManager interface {
	SaveMetricsInStorage(gaugeValues map[string]float64, counterValues map[string]int64)
	RestoreDataFromStorage() (map[string]float64, map[string]int64)
}
