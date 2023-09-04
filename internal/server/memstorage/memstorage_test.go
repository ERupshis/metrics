package memstorage

import (
	"testing"

	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testConfig = config.Config{
	Host:          "",
	Restore:       true,
	StoragePath:   "/tmp/metrics-db.json",
	StoreInterval: 300,
}

func TestCreateStorage(t *testing.T) {
	storage := Create(nil)

	require.NotNil(t, storage)
	require.NotNil(t, storage.counterMetrics)
	require.NotNil(t, storage.gaugeMetrics)
	require.Nil(t, storage.manager)
}

func TestMemStorage_AddCounter(t *testing.T) {
	storage := Create(nil)
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
	storage := Create(nil)
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
	storage := Create(nil)
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
	storage := Create(nil)
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
				assert.NoError(t, err, "GetGauge(%v) missing name", tt.req)
				return
			}

			assert.Equalf(t, tt.want, got, "GetGauge(%v)", tt.req)
		})
	}
}

func TestMemStorage_GetAllCounters(t *testing.T) {
	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "valid",
			fields: fields{
				gaugeMetrics:   map[string]gauge{"metric1": 1.1},
				counterMetrics: map[string]counter{"metric2": 1},
				manager:        nil,
			},
			want: map[string]interface{}{"metric2": int64(1)},
		},
		{
			name: "valid empty",
			fields: fields{
				gaugeMetrics:   map[string]gauge{"metric1": 1.1},
				counterMetrics: map[string]counter{},
				manager:        nil,
			},
			want: map[string]interface{}{},
		},
		{
			name: "valid with 2 values",
			fields: fields{
				gaugeMetrics:   map[string]gauge{"metric1": 1.1},
				counterMetrics: map[string]counter{"metric2": 1, "metric3": 2},
				manager:        nil,
			},
			want: map[string]interface{}{"metric2": int64(1), "metric3": int64(2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
				manager:        tt.fields.manager,
			}
			assert.Equalf(t, tt.want, m.GetAllCounters(), "GetAllCounters()")
		})
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "valid",
			fields: fields{
				gaugeMetrics:   map[string]gauge{"metric1": 1.1},
				counterMetrics: map[string]counter{"metric2": 1},
				manager:        nil,
			},
			want: map[string]interface{}{"metric1": 1.1},
		},
		{
			name: "valid empty",
			fields: fields{
				gaugeMetrics:   map[string]gauge{},
				counterMetrics: map[string]counter{},
				manager:        nil,
			},
			want: map[string]interface{}{},
		},
		{
			name: "valid with 2 values",
			fields: fields{
				gaugeMetrics:   map[string]gauge{"metric1": 1.1, "metric3": 2.2},
				counterMetrics: map[string]counter{"metric2": 1},
				manager:        nil,
			},
			want: map[string]interface{}{"metric1": 1.1, "metric3": 2.2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
				manager:        tt.fields.manager,
			}
			assert.Equalf(t, tt.want, m.GetAllGauges(), "GetAllGauges()")
		})
	}
}

func TestMemStorage_IsAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorageManager(ctrl)
	gomock.InOrder(
		m.EXPECT().CheckConnection().Return(true),
		m.EXPECT().CheckConnection().Return(false),
	)

	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "valid positive",
			fields: fields{
				gaugeMetrics:   nil,
				counterMetrics: nil,
				manager:        m,
			},
			want: true,
		},
		{
			name: "valid negative",
			fields: fields{
				gaugeMetrics:   nil,
				counterMetrics: nil,
				manager:        m,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
				manager:        tt.fields.manager,
			}
			assert.Equalf(t, tt.want, m.IsAvailable(), "IsAvailable()")
		})
	}
}

func TestMemStorage_SaveData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorageManager(ctrl)
	gomock.InOrder(
		m.EXPECT().SaveMetricsInStorage(gomock.Any(), gomock.Any()).Return(),
	)

	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "valid",
			fields: fields{
				gaugeMetrics:   nil,
				counterMetrics: nil,
				manager:        m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
				manager:        tt.fields.manager,
			}
			m.SaveData()
		})
	}
}

func TestMemStorage_RestoreData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorageManager(ctrl)

	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	type want struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "valid",
			fields: fields{
				gaugeMetrics:   map[string]gauge{},
				counterMetrics: map[string]counter{},
				manager:        m,
			},
			want: want{
				gaugeMetrics:   map[string]float64{"gauge1": 1.1, "gauge2": 2.2},
				counterMetrics: map[string]int64{"counter1": 1, "counter3": 3},
			},
		},
	}

	m.EXPECT().RestoreDataFromStorage().Return(map[string]float64{"gauge1": 1.1, "gauge2": 2.2}, map[string]int64{"counter1": 1, "counter3": 3})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
				manager:        tt.fields.manager,
			}

			m.RestoreData()

			assert.Equal(t, tt.want.gaugeMetrics, m.gaugeMetrics)
			assert.Equal(t, tt.want.counterMetrics, m.counterMetrics)
		})
	}
}
