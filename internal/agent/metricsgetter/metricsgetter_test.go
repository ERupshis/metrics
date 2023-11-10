package metricsgetter

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsGetter(t *testing.T) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	assert.Equal(t, float64(stats.Alloc), GaugeMetricsGetter["Alloc"](&stats))
	assert.Equal(t, float64(stats.Frees), GaugeMetricsGetter["Frees"](&stats))
	assert.Equal(t, float64(stats.Lookups), GaugeMetricsGetter["Lookups"](&stats))
	assert.Equal(t, float64(stats.MSpanSys), GaugeMetricsGetter["MSpanSys"](&stats))
	assert.Equal(t, float64(stats.OtherSys), GaugeMetricsGetter["OtherSys"](&stats))
	assert.Equal(t, float64(stats.HeapAlloc), GaugeMetricsGetter["HeapAlloc"](&stats))
	assert.Equal(t, float64(stats.LastGC), GaugeMetricsGetter["LastGC"](&stats))

}

func TestExtraMetricsGetter(t *testing.T) {
	for _, valFunc := range AdditionalGaugeMetricsGetter {
		_, err := valFunc()
		require.NoError(t, err)
	}
}
