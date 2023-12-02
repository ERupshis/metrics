// Package memstorage provides an in-memory storage implementation for metrics,
// including gauges and counters, along with functionality for managing data persistence.
// The package defines a MemStorage type, which is an in-memory storage structure.
// It supports adding, retrieving, and managing gauge and counter metrics.
package memstorage

import (
	"context"
	"fmt"

	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
)

// gauge represents a floating-point metric value.
type gauge = float64

// counter represents an integer metric value.
type counter = int64

// MemStorage is an in-memory storage structure for gauge and counter metrics.
// It also includes a StorageManager for handling data persistence.
type MemStorage struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
	manager        storagemngr.StorageManager
}

// Create initializes and returns a new instance of MemStorage with the provided StorageManager.
func Create(manager storagemngr.StorageManager) *MemStorage {
	return &MemStorage{
		make(map[string]gauge),
		make(map[string]counter),
		manager,
	}
}

// RestoreData retrieves and restores stored metrics data from the associated StorageManager.
// It populates the in-memory storage with the retrieved data.
func (m *MemStorage) RestoreData(ctx context.Context) error {
	if m.manager == nil {
		return fmt.Errorf("manager is not initialized")
	}

	gauges, counters, err := m.manager.RestoreDataFromStorage(ctx)
	if err != nil {
		return fmt.Errorf("restore data: %w", err)
	}

	for key, val := range gauges {
		m.AddGauge(key, val)
	}

	for key, val := range counters {
		m.AddCounter(key, val)
	}

	return nil
}

// IsAvailable checks the availability of the associated StorageManager.
func (m *MemStorage) IsAvailable(ctx context.Context) (bool, error) {
	if m.manager == nil {
		return false, fmt.Errorf("storage manager is not initialized")
	}
	return m.manager.CheckConnection(ctx)
}

// SaveData saves the current in-memory metrics data using the associated StorageManager.
func (m *MemStorage) SaveData(ctx context.Context) error {
	if m.manager == nil {
		return fmt.Errorf("storage manager is not initialized")
	}

	if err := m.manager.SaveMetricsInStorage(ctx, m.GetAllGauges(), m.GetAllCounters()); err != nil {
		return fmt.Errorf("save data: %w", err)
	}
	return nil
}

// AddCounter adds the specified value to the counter metric with the given name.
func (m *MemStorage) AddCounter(name string, value counter) {
	m.counterMetrics[name] += value
}

// GetCounter retrieves the value of the counter metric with the given name.
func (m *MemStorage) GetCounter(name string) (int64, error) {
	if value, inMap := m.counterMetrics[name]; inMap {
		return value, nil
	}
	return -1, fmt.Errorf("invalid counter name '%s'", name)
}

// GetAllCounters returns a copy of the map containing all counter metrics.
func (m *MemStorage) GetAllCounters() map[string]interface{} {
	return copyMapPredefinedSizePointers(m.counterMetrics)
}

// AddGauge adds the specified value to the gauge metric with the given name.
func (m *MemStorage) AddGauge(name string, value gauge) {
	m.gaugeMetrics[name] = value
}

// GetGauge retrieves the value of the gauge metric with the given name.
func (m *MemStorage) GetGauge(name string) (float64, error) {
	if value, inMap := m.gaugeMetrics[name]; inMap {
		return value, nil
	}
	return -1.0, fmt.Errorf("invalid gauge name '%s'", name)
}

// GetAllGauges returns a copy of the map containing all gauge metrics.
func (m *MemStorage) GetAllGauges() map[string]interface{} {
	return copyMapPredefinedSizePointers(m.gaugeMetrics)
}

// The following functions create copies of maps with specific pointer handling.

// copyMap creates and returns a new map with values copied from the provided map.
func copyMap[V any](m map[string]V) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// copyMapPredefinedSize creates and returns a new map with a predefined size
// and values copied from the provided map.
func copyMapPredefinedSize[V any](m map[string]V) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// copyMapPredefinedSizePointers creates and returns a new map with a predefined size
// and values copied from the provided map, with each value as a pointer.
func copyMapPredefinedSizePointers[V any](m map[string]V) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		v := v
		result[k] = &v
	}
	return result
}

// copyMapPointers creates and returns a new map with values copied from the provided map,
// with each value as a pointer.
func copyMapPointers[V any](m map[string]V) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		v := v
		result[k] = &v
	}
	return result
}
