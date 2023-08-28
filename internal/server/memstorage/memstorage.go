package memstorage

import (
	"errors"

	"github.com/erupshis/metrics/internal/server/memstorage/storagemanager"
)

// TODO: make independent package with custom types?
type gauge = float64
type counter = int64

type MemStorage struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
	manager        storagemanager.StorageManager
}

func Create(manager storagemanager.StorageManager) MemStorage {
	return MemStorage{
		make(map[string]gauge),
		make(map[string]counter),
		manager,
	}
}

func (m *MemStorage) RestoreData() {
	gauges, counters := m.manager.RestoreDataFromStorage()

	for key, val := range gauges {
		m.AddGauge(key, val)
	}

	for key, val := range counters {
		m.AddCounter(key, val)
	}
}

func (m *MemStorage) SaveData() {
	m.manager.SaveMetricsInStorage(m.GetAllGauges(), m.GetAllCounters())
}

func (m *MemStorage) AddCounter(name string, value counter) {
	m.counterMetrics[name] += value
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	if value, inMap := m.counterMetrics[name]; inMap {
		return value, nil
	}
	return -1, errors.New("invalid counter name")
}

func (m *MemStorage) GetAllCounters() map[string]counter {
	return copyMap(m.counterMetrics)
}

func (m *MemStorage) AddGauge(name string, value gauge) {
	m.gaugeMetrics[name] = value
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	if value, inMap := m.gaugeMetrics[name]; inMap {
		return value, nil
	}
	return -1.0, errors.New("invalid counter name")
}

func (m *MemStorage) GetAllGauges() map[string]gauge {
	return copyMap(m.gaugeMetrics)
}

func copyMap[V comparable](m map[string]V) map[string]V {
	result := make(map[string]V)
	for k, v := range m {
		result[k] = v
	}
	return result
}
