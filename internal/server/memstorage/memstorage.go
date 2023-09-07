package memstorage

import (
	"context"
	"fmt"

	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
)

// TODO: make independent package with custom types?
type gauge = float64
type counter = int64

type MemStorage struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
	manager        storagemngr.StorageManager
}

func Create(manager storagemngr.StorageManager) *MemStorage {
	return &MemStorage{
		make(map[string]gauge),
		make(map[string]counter),
		manager,
	}
}

func (m *MemStorage) RestoreData(ctx context.Context) {
	if m.manager == nil {
		return
	}

	gauges, counters, err := m.manager.RestoreDataFromStorage(ctx)
	if err != nil {
		//TODO log.
		return
	}

	for key, val := range gauges {
		m.AddGauge(key, val)
	}

	for key, val := range counters {
		m.AddCounter(key, val)
	}
}

func (m *MemStorage) IsAvailable(ctx context.Context) (bool, error) {
	return m.manager.CheckConnection(ctx)
}

func (m *MemStorage) SaveData(ctx context.Context) error {
	if m.manager == nil {
		return fmt.Errorf("storage manager is not initialized")
	}

	if err := m.manager.SaveMetricsInStorage(ctx, m.GetAllGauges(), m.GetAllCounters()); err != nil {
		return fmt.Errorf("save data: %w", err)
	}
	return nil
}

func (m *MemStorage) AddCounter(name string, value counter) {
	m.counterMetrics[name] += value
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	if value, inMap := m.counterMetrics[name]; inMap {
		return value, nil
	}
	return -1, fmt.Errorf("invalid counter name '%s'", name)
}

func (m *MemStorage) GetAllCounters() map[string]interface{} {
	return copyMap(m.counterMetrics)
}

func (m *MemStorage) AddGauge(name string, value gauge) {
	m.gaugeMetrics[name] = value
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	if value, inMap := m.gaugeMetrics[name]; inMap {
		return value, nil
	}
	return -1.0, fmt.Errorf("invalid gauge name '%s'", name)
}

func (m *MemStorage) GetAllGauges() map[string]interface{} {
	return copyMap(m.gaugeMetrics)
}

func copyMap[V any](m map[string]V) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}
