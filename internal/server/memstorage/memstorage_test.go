package memstorage

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/erupshis/metrics/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateStorage(t *testing.T) {
	storage := Create(context.Background(), &config.Default, nil, logger.CreateMock())

	require.NotNil(t, storage)
	require.NotNil(t, storage.counterMetrics)
	require.NotNil(t, storage.gaugeMetrics)
	require.Nil(t, storage.manager)
}

func TestMemStorage_AddCounter(t *testing.T) {
	storage := Create(context.Background(), &config.Default, nil, logger.CreateMock())
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
	storage := Create(context.Background(), &config.Default, nil, logger.CreateMock())
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
	storage := Create(context.Background(), &config.Default, nil, logger.CreateMock())
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
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	storage := Create(context.Background(), &config.Default, nil, logger.CreateMock())
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
	int1 := int64(1)
	int2 := int64(2)

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
			want: map[string]interface{}{"metric2": &int1},
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
			want: map[string]interface{}{"metric2": &int1, "metric3": &int2},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	float1 := 1.1
	float2 := 2.2

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
			want: map[string]interface{}{"metric1": &float1},
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
			want: map[string]interface{}{"metric1": &float1, "metric3": &float2},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
		m.EXPECT().CheckConnection(gomock.Any()).Return(true, nil),
		m.EXPECT().CheckConnection(gomock.Any()).Return(false, nil),
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
			ok, err := m.IsAvailable(context.Background())
			require.NoError(t, err)
			assert.Equalf(t, tt.want, ok, "IsAvailable()")
		})
	}
}

func TestMemStorage_SaveData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := mocks.NewMockStorageManager(ctrl)
	gomock.InOrder(
		manager.EXPECT().SaveMetricsInStorage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		manager.EXPECT().SaveMetricsInStorage(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("manager err")),
	)

	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	type want struct {
		wantErr bool
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "valid",
			fields: fields{
				gaugeMetrics:   nil,
				counterMetrics: nil,
				manager:        manager,
			},
			want: want{
				wantErr: false,
			},
		},
		{
			name: "error from manager",
			fields: fields{
				gaugeMetrics:   nil,
				counterMetrics: nil,
				manager:        manager,
			},
			want: want{
				wantErr: true,
			},
		},
		{
			name: "manager is not init",
			fields: fields{
				gaugeMetrics:   nil,
				counterMetrics: nil,
				manager:        nil,
			},
			want: want{
				wantErr: true,
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

			err := m.SaveData(context.Background())
			if (err != nil) != tt.want.wantErr {
				t.Errorf("SaveData() error = %v, wantErr %v", err, tt.want.wantErr)
				return
			}
		})
	}
}

func TestMemStorage_RestoreData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := mocks.NewMockStorageManager(ctrl)
	gomock.InOrder(
		manager.EXPECT().RestoreDataFromStorage(context.Background()).Return(
			map[string]float64{"gauge1": 1.1, "gauge2": 2.2},
			map[string]int64{"counter1": 1, "counter3": 3},
			nil),
		manager.EXPECT().RestoreDataFromStorage(context.Background()).Return(
			nil,
			nil,
			fmt.Errorf("manager err")),
	)

	type fields struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		manager        storagemngr.StorageManager
	}
	type want struct {
		gaugeMetrics   map[string]gauge
		counterMetrics map[string]counter
		wantErr        bool
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
				manager:        manager,
			},
			want: want{
				gaugeMetrics:   map[string]float64{"gauge1": 1.1, "gauge2": 2.2},
				counterMetrics: map[string]int64{"counter1": 1, "counter3": 3},
				wantErr:        false,
			},
		},
		{
			name: "valid",
			fields: fields{
				gaugeMetrics:   map[string]gauge{},
				counterMetrics: map[string]counter{},
				manager:        manager,
			},
			want: want{
				gaugeMetrics:   map[string]gauge{},
				counterMetrics: map[string]counter{},
				wantErr:        true,
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

			err := m.RestoreData(context.Background())
			if (err != nil) != tt.want.wantErr {
				t.Errorf("SaveData() error = %v, wantErr %v", err, tt.want.wantErr)
				return
			}

			assert.Equal(t, tt.want.gaugeMetrics, m.gaugeMetrics)
			assert.Equal(t, tt.want.counterMetrics, m.counterMetrics)
		})
	}
}

func BenchmarkMemstorage_copyMapFloat(b *testing.B) {
	size := 1000
	testMap := generateRandomMapFloat(size)

	b.Run("copy values", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMap(testMap)
		}
	})
	b.Run("copy values predefined size", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMapPredefinedSize(testMap)
		}
	})
	b.Run("copy pointers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMapPointers(testMap)
		}
	})
	b.Run("copy values predefined size pointers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMapPredefinedSizePointers(testMap)
		}
	})
}

func generateRandomMapFloat(size int) map[string]float64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	randomMap := make(map[string]float64, size)

	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key%d", i)
		value := rand.Float64() * 100 // Adjust the multiplier based on your needs
		randomMap[key] = value
	}

	return randomMap
}

func BenchmarkMemstorage_copyMapInt64(b *testing.B) {
	size := 1000
	testMap := generateRandomMapInt64(size)

	b.Run("copy values", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMap(testMap)
		}
	})
	b.Run("copy values predefined size", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMapPredefinedSize(testMap)
		}
	})
	b.Run("copy pointers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMapPointers(testMap)
		}
	})
	b.Run("copy values predefined size pointers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			copyMapPredefinedSizePointers(testMap)
		}
	})
}

func generateRandomMapInt64(size int) map[string]int64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	randomMap := make(map[string]int64, size)

	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key%d", i)
		value := rand.Float64() * 100 // Adjust the multiplier based on your needs
		randomMap[key] = int64(value)
	}

	return randomMap
}

func Test_copyMap(t *testing.T) {
	type args[V any] struct {
		m map[string]V
	}
	type testCase[V any] struct {
		name string
		args args[V]
		want map[string]interface{}
	}
	tests := []testCase[int64]{
		{
			name: "int64 valid",
			args: args[int64]{
				m: map[string]int64{"one": 1, "two": 2},
			},
			want: map[string]interface{}{"one": int64(1), "two": int64(2)},
		},
		{
			name: "int64 empty",
			args: args[int64]{
				m: map[string]int64{},
			},
			want: map[string]interface{}{},
		},
	}
	tests2 := []testCase[float64]{
		{
			name: "float64 valid",
			args: args[float64]{
				m: map[string]float64{"one": 1, "two": 2},
			},
			want: map[string]interface{}{"one": float64(1), "two": float64(2)},
		},
		{
			name: "float64 empty",
			args: args[float64]{
				m: map[string]float64{},
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		ttSh := tt
		t.Run(ttSh.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, ttSh.want, copyMap(ttSh.args.m), "copyMap(%v)", ttSh.args.m)
			assert.Equalf(t, ttSh.want, copyMapPredefinedSize(ttSh.args.m), "copyMapPredefinedSize(%v)", ttSh.args.m)
		})
	}
	for _, ttCommon := range tests2 {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, copyMap(tt.args.m), "copyMap(%v)", tt.args.m)
			assert.Equalf(t, tt.want, copyMapPredefinedSize(tt.args.m), "copyMapPredefinedSize(%v)", tt.args.m)
		})
	}
}

func Test_copyMapWithPointers(t *testing.T) {
	type args[V any] struct {
		m map[string]V
	}
	type testCase[V any] struct {
		name string
		args args[V]
		want map[string]interface{}
	}
	tests := []testCase[int64]{
		{
			name: "int64 valid",
			args: args[int64]{
				m: map[string]int64{"one": 1, "two": 2},
			},
			want: map[string]interface{}{"one": int64(1), "two": int64(2)},
		},
		{
			name: "int64 empty",
			args: args[int64]{
				m: map[string]int64{},
			},
			want: map[string]interface{}{},
		},
	}
	tests2 := []testCase[float64]{
		{
			name: "float64 valid",
			args: args[float64]{
				m: map[string]float64{"one": 1, "two": 2},
			},
			want: map[string]interface{}{"one": float64(1), "two": float64(2)},
		},
		{
			name: "float64 empty",
			args: args[float64]{
				m: map[string]float64{},
			},
			want: map[string]interface{}{},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for k, v := range copyMapPredefinedSizePointers(tt.args.m) {
				assert.Equal(t, tt.want[k], *v.(*int64))
			}
			for k, v := range copyMapPointers(tt.args.m) {
				assert.Equal(t, tt.want[k], *v.(*int64))
			}
		})
	}
	for _, ttCommon := range tests2 {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for k, v := range copyMapPredefinedSizePointers(tt.args.m) {
				assert.Equal(t, tt.want[k], *v.(*float64))
			}
			for k, v := range copyMapPointers(tt.args.m) {
				assert.Equal(t, tt.want[k], *v.(*float64))
			}
		})
	}
}
