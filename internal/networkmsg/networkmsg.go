package networkmsg

import (
	"encoding/json"
	"fmt"

	"github.com/mailru/easyjson"
)

//go:generate easyjson -all networkmsg.go
type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func CreateCounterMetrics(name string, value int64) Metric {
	return Metric{
		ID:    name,
		MType: "counter",
		Delta: &value,
		Value: nil,
	}
}

func CreateGaugeMetrics(name string, value float64) Metric {
	return Metric{
		ID:    name,
		MType: "gauge",
		Delta: nil,
		Value: &value,
	}
}

func CreatePostUpdateMessage(data Metric) []byte {
	out, err := easyjson.Marshal(data)
	if err != nil {
		panic(err)
	}

	return out
}

func ParsePostValueMessage(message []byte) (Metric, error) {
	var out Metric
	if err := easyjson.Unmarshal(message, &out); err != nil {
		return out, fmt.Errorf("parse metric: %w", err)
	}

	_, err := isMetricValid(out)
	return out, err
}

// TODO: need add bench.
func ParsePostBatchValueMessage(message []byte) ([]Metric, error) {
	var out []Metric
	if err := json.Unmarshal(message, &out); err != nil {
		return []Metric{}, fmt.Errorf("parse metrics batch: %w", err)
	}

	err := fmt.Errorf("found invalid metrics in message: ")
	for i, metric := range out {
		if _, errMetric := isMetricValid(metric); errMetric != nil {
			err = fmt.Errorf("%w %d: %w |", err, i, errMetric)
		}
	}

	if err.Error() != "found invalid metrics in message: " {
		return out, err
	}

	return out, nil
}

// TODO: need add bench.
func isMetricValid(m Metric) (bool, error) {
	errMsg := ""
	if m.ID == "" {
		errMsg += "missing name "
	}

	if m.MType == "" {
		errMsg += "missing type "
	}

	if (m.Delta != nil) == (m.Value != nil) {
		errMsg += " both values exists or missing "
	}

	if len(errMsg) != 0 {
		return false, fmt.Errorf("%s", errMsg)
	}

	return true, nil
}
