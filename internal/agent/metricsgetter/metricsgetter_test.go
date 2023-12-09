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
	assert.Equal(t, float64(stats.BuckHashSys), GaugeMetricsGetter["BuckHashSys"](&stats))
	assert.Equal(t, float64(stats.Frees), GaugeMetricsGetter["Frees"](&stats))
	assert.Equal(t, stats.GCCPUFraction, GaugeMetricsGetter["GCCPUFraction"](&stats))
	assert.Equal(t, float64(stats.GCSys), GaugeMetricsGetter["GCSys"](&stats))
	assert.Equal(t, float64(stats.HeapAlloc), GaugeMetricsGetter["HeapAlloc"](&stats))
	assert.Equal(t, float64(stats.HeapIdle), GaugeMetricsGetter["HeapIdle"](&stats))
	assert.Equal(t, float64(stats.HeapInuse), GaugeMetricsGetter["HeapInuse"](&stats))
	assert.Equal(t, float64(stats.HeapObjects), GaugeMetricsGetter["HeapObjects"](&stats))
	assert.Equal(t, float64(stats.HeapReleased), GaugeMetricsGetter["HeapReleased"](&stats))
	assert.Equal(t, float64(stats.HeapSys), GaugeMetricsGetter["HeapSys"](&stats))
	assert.Equal(t, float64(stats.LastGC), GaugeMetricsGetter["LastGC"](&stats))
	assert.Equal(t, float64(stats.Lookups), GaugeMetricsGetter["Lookups"](&stats))
	assert.Equal(t, float64(stats.MCacheInuse), GaugeMetricsGetter["MCacheInuse"](&stats))
	assert.Equal(t, float64(stats.MCacheSys), GaugeMetricsGetter["MCacheSys"](&stats))
	assert.Equal(t, float64(stats.MSpanInuse), GaugeMetricsGetter["MSpanInuse"](&stats))
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
