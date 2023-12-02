// Package networkmsg implements struct for REST message.
// Provides support messages to parse incoming messages in JSON and create them.
package networkmsg

import (
	"encoding/json"
	"fmt"

	"github.com/mailru/easyjson"
)

// Metric definition of transferred data.
//
//go:generate easyjson -all networkmsg.go
type Metric struct {
	ID    string   `json:"id"`              // name
	MType string   `json:"type"`            // defines type of metric (gauge/counter)
	Delta *int64   `json:"delta,omitempty"` // value for counter type
	Value *float64 `json:"value,omitempty"` // value for gauge type
}

// CreateCounterMetrics creates counter Metric entity.
func CreateCounterMetrics(name string, value int64) Metric {
	return Metric{
		ID:    name,
		MType: "counter",
		Delta: &value,
		Value: nil,
	}
}

// CreateGaugeMetrics creates gauge Metric entity.
func CreateGaugeMetrics(name string, value float64) Metric {
	return Metric{
		ID:    name,
		MType: "gauge",
		Delta: nil,
		Value: &value,
	}
}

// CreatePostUpdateMessage converts Metric entity into JSON format.
func CreatePostUpdateMessage(data Metric) []byte {
	out, err := easyjson.Marshal(data)
	if err != nil {
		panic(err)
	}

	return out
}

// ParsePostValueMessage converts JSON definition into Metric entity.
func ParsePostValueMessage(message []byte) (Metric, error) {
	var out Metric
	if err := easyjson.Unmarshal(message, &out); err != nil {
		return out, fmt.Errorf("parse metric: %w", err)
	}

	_, err := isMetricValid(out)
	return out, err
}

// ParsePostBatchValueMessage converts JSON definition of several Metric entities into []Metric.
func ParsePostBatchValueMessage(message []byte) ([]Metric, error) {
	var out []Metric
	if err := json.Unmarshal(message, &out); err != nil {
		return []Metric{}, fmt.Errorf("parse metrics batch: %w", err)
	}

	err := fmt.Errorf("found invalid metrics in message: ")
	for i, metric := range out {
		if _, errMetric := isMetricValid(metric); errMetric != nil {
			err = fmt.Errorf("%v %d: %w |", err, i, errMetric)
		}
	}

	if err.Error() != "found invalid metrics in message: " {
		return out, err
	}

	return out, nil
}

// isMetricValid validates that Metric data is correct.
// Checks name and type on empty, check two delta and value fields filling at the same time.
func isMetricValid(m Metric) (bool, error) {
	errMsg := ""
	if m.ID == "" {
		errMsg += "missing name "
	}

	if m.MType == "" {
		errMsg += "missing type "
	}

	if (m.Delta != nil) && (m.Value != nil) {
		errMsg += " both values exists"
	}

	if len(errMsg) != 0 {
		return false, fmt.Errorf("%s", errMsg)
	}

	return true, nil
}
