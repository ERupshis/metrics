package memstorage

var Storage = MemStorage{make(map[string]float64), make(map[string]int64)}

type MemStorage struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func (m MemStorage) AddCounter(name string, value int64) {
	m.counterMetrics[name] += value
}

func (m MemStorage) GetCounter(name string) (int64, error) {
	//TODO: need implementation.
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
	return 0, nil
	//if value, inMap := m.gaugeMetrics[name]; inMap {
	//	return value, nil
	//}
	//return -1.0, errors.New("invalid counter name")
}
