package networkmsg

import (
	"encoding/json"

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
		return Metric{}, err
	}

	return out, nil
}

func ParsePostBatchValueMessage(message []byte) ([]Metric, error) {
	var out []Metric
	if err := json.Unmarshal(message, &out); err != nil {
		return []Metric{}, err
	}

	return out, nil
}
