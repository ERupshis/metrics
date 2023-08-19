package networkmsg

import (
	"github.com/mailru/easyjson"
)

//go:generate easyjson -all networkmsg.go
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func CreateCounterMetrics(name string, value int64) Metrics {
	return Metrics{
		ID:    name,
		MType: "counter",
		Delta: &value,
	}
}

func CreateGaugeMetrics(name string, value float64) Metrics {
	return Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
}

func CreatePostUpdateMessage(data Metrics) []byte {
	out, err := easyjson.Marshal(data)
	if err != nil {
		panic(err)
	}

	return out
}

func ParsePostValueMessage(message []byte) (Metrics, error) {
	var out Metrics
	if err := easyjson.Unmarshal(message, &out); err != nil {
		return Metrics{}, err
	}

	return out, nil
}
