package memstorage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateStorage(t *testing.T) {
	storage := Create()

	require.NotNil(t, storage)
	require.NotNil(t, storage.counterMetrics)
	require.NotNil(t, storage.gaugeMetrics)
}

func TestMemStorage_AddCounter(t *testing.T) {
	storage := Create()
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name   string
		args   args
		result int64
	}{
		{"add valid counter", args{"testCounter", 123}, 123},
		{"add another valid counter", args{"testAnotherCounter", 123}, 123},
		{"add similar valid counter", args{"testAnotherCounter", 123}, 246},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.AddCounter(tt.args.name, tt.args.value)
			assert.Contains(t, storage.counterMetrics, tt.args.name)
			assert.Equal(t, storage.counterMetrics[tt.args.name], tt.result)
		})
	}
}

func TestMemStorage_AddGauge(t *testing.T) {
	storage := Create()
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		args   args
		result float64
	}{
		{"add valid gauge", args{"testGauge", 123.0}, 123.0},
		{"add another valid gauge", args{"testAnotherGauge", 123.0}, 123.0},
		{"add similar valid gauge", args{"testAnotherGauge", 123.0}, 123.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.AddGauge(tt.args.name, tt.args.value)
			assert.Contains(t, storage.gaugeMetrics, tt.args.name)
			assert.Equal(t, storage.gaugeMetrics[tt.args.name], tt.result)
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	storage := Create()
	storage.AddCounter("metric1", 1)

	tests := []struct {
		name    string
		req     string
		want    int64
		wantErr bool
	}{
		{"valid name", "metric1", 1, false},
		{"invalid name", "metric2", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetCounter(tt.req)
			if err != nil && !tt.wantErr {
				assert.NoError(t, err, "GetCounter(%v) missing name", tt.req)
				return
			}

			assert.Equalf(t, tt.want, got, "GetCounter(%v)", tt.req)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	storage := Create()
	storage.AddGauge("metric1", 1.2)

	tests := []struct {
		name    string
		req     string
		want    float64
		wantErr bool
	}{
		{"valid name", "metric1", 1.2, false},
		{"invalid name", "metric2", -1.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetGauge(tt.req)
			if err != nil && !tt.wantErr {
				assert.NoError(t, err, "GetCounter(%v) missing name", tt.req)
				return
			}

			assert.Equalf(t, tt.want, got, "GetCounter(%v)", tt.req)
		})
	}
}
