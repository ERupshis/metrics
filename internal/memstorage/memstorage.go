package memstorage

import "fmt"

type gauge = float64
type counter = int64

var Storage = MemStorage{make(map[string]gauge), make(map[string]counter)}

type MemStorage struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
}

func (m MemStorage) AddCounter(name string, value int64) {
	m.counterMetrics[name] += value
}

func (m MemStorage) GetCounter(name string) (int64, error) {
	//TODO: need implementation.
	fmt.Println(name)
	return 0, nil
	//if value, inMap := m.counterMetrics[name]; inMap {
	//	return value, nil
	//}
	//return -1, errors.New("invalid counter name")
}

func (m MemStorage) AddGauge(name string, value float64) {
	m.gaugeMetrics[name] = value
}

func (m MemStorage) GetGauge(name string) (float64, error) {
	//TODO: need implementation.
	fmt.Println(name)
	return 0, nil
	//if value, inMap := m.gaugeMetrics[name]; inMap {
	//	return value, nil
	//}
	//return -1.0, errors.New("invalid counter name")
}
