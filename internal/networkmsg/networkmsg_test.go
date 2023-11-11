package networkmsg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricSerialization(t *testing.T) {
	counterMetric := CreateCounterMetrics("counter_metric", 42)
	gaugeMetric := CreateGaugeMetrics("gauge_metric", 3.14)

	// Test CreateCounterMetrics serialization
	counterJSON, err := json.Marshal(counterMetric)
	if err != nil {
		t.Errorf("Error marshalling counter metric: %v", err)
	}

	// Test CreateGaugeMetrics serialization
	gaugeJSON, err := json.Marshal(gaugeMetric)
	if err != nil {
		t.Errorf("Error marshalling gauge metric: %v", err)
	}

	// Test CreatePostUpdateMessage
	counterPostMessage := CreatePostUpdateMessage(counterMetric)
	assert.Equal(t, counterJSON, counterPostMessage)

	gaugePostMessage := CreatePostUpdateMessage(gaugeMetric)
	assert.Equal(t, gaugeJSON, gaugePostMessage)

	// Test ParsePostValueMessage
	parsedCounterMetric, err := ParsePostValueMessage(counterPostMessage)
	if err != nil {
		t.Errorf("Error parsing counter post message: %v", err)
	}

	parsedGaugeMetric, err := ParsePostValueMessage(gaugePostMessage)
	if err != nil {
		t.Errorf("Error parsing gauge post message: %v", err)
	}

	// Compare original and parsed metrics
	if !isMetricsEqual(&counterMetric, &parsedCounterMetric) {
		t.Errorf("Original and parsed counter metrics do not match")
	}

	if !isMetricsEqual(&gaugeMetric, &parsedGaugeMetric) {
		t.Errorf("Original and parsed gauge metrics do not match")
	}

}

func isMetricsEqual(m1 *Metric, m2 *Metric) bool {
	if m1.ID != m1.ID {
		return false
	}

	if m1.MType != m1.MType {
		return false
	}

	if m1.Value != nil && m2.Value == nil ||
		m1.Value == nil && m2.Value != nil ||
		m1.Value != nil && m2.Value != nil && *m1.Value != *m2.Value {
		return false
	}

	if m1.Delta != nil && m2.Delta == nil ||
		m1.Delta == nil && m2.Delta != nil ||
		m1.Delta != nil && m2.Delta != nil && *m1.Delta != *m2.Delta {
		return false
	}

	return true
}
