package metricsgetter

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsGetter(t *testing.T) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	assert.Equal(t, float64(stats.Alloc), GaugeMetricsGetter["Alloc"](&stats))
	assert.Equal(t, float64(stats.Frees), GaugeMetricsGetter["Frees"](&stats))
	assert.Equal(t, float64(stats.Lookups), GaugeMetricsGetter["Lookups"](&stats))
	assert.Equal(t, float64(stats.MSpanSys), GaugeMetricsGetter["MSpanSys"](&stats))
	assert.Equal(t, float64(stats.OtherSys), GaugeMetricsGetter["OtherSys"](&stats))

}
