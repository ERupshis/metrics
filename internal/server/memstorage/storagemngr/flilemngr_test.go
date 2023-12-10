package storagemngr

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testFolder = "/test"

var testConfig = config.Config{
	Host:          "",
	Restore:       true,
	StoragePath:   "/tmp/metrics-db.json",
	StoreInterval: 300,
}

func createFileManagerTest(dataPath string, logger logger.BaseLogger) *FileManager {
	return &FileManager{path: dataPath, logger: logger}
}

func TestFileManager_IsFileOpen(t *testing.T) {
	cfg, _ := config.Parse()
	log := logger.CreateMock()

	fm := createFileManagerTest(cfg.StoragePath, log)
	if fm.IsFileOpen() {
		t.Errorf("IsFileOpen() file wasn't opened before")
	}

	if err := fm.OpenFile(testFolder+"/ddf/dd", false); err != nil {
		t.Errorf("OpenFile() without Trunc error = %v", err)
	}

	if !fm.IsFileOpen() {
		t.Errorf("IsFileOpen() file was opened before, but manager thinks opposite")
	}

	if err := fm.CloseFile(); err != nil {
		t.Errorf("CloseFile() file wasn't opened before, error = %v", err)
	}

	if fm.IsFileOpen() {
		t.Errorf("IsFileOpen() file was closed before, but manager thinks its still open")
	}
}

func TestFileManager_OpenFile(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "valid open file",
			fields: fields{path: testFolder + "/aa"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.CreateMock()

			fm := createFileManagerTest(testConfig.StoragePath, log)

			if err := fm.OpenFile(tt.fields.path, false); err != nil {
				t.Errorf("OpenFile() without Trunc error = %v", err)
			}
		})
	}
}

func TestFileManager_CloseFile(t *testing.T) {
	// test close on missing file.
	log := logger.CreateMock()

	fm := createFileManagerTest(testConfig.StoragePath, log)
	if err := fm.CloseFile(); err != nil {
		t.Errorf("CloseFile() file wasn't opened before, error = %v", err)
	}

	type fields struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid close file",
			fields: fields{path: testFolder + "/aa11"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fm.OpenFile(tt.fields.path, false); err != nil {
				t.Errorf("OpenFile() without Trunc error = %v", err)
			}

			if err := fm.CloseFile(); err != nil {
				t.Errorf("CloseFile() error = %v", err)
			}

			if fm.IsFileOpen() {
				t.Errorf("IsFileOpen() file wasn't closed properly")
			}
		})
	}
}

func TestFileManager_initWriterAndScanner(t *testing.T) {
	var path = testFolder + "/dd"
	log := logger.CreateMock()

	fm := createFileManagerTest(path, log)
	_ = fm.OpenFile(path, false)
	_ = fm.CloseFile()

	type fields struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "valid with existing file",
			fields:  fields{path: path},
			wantErr: false,
		},
		{
			name:    "valid with missing file",
			fields:  fields{path: testFolder + "/sdf"},
			wantErr: false,
		},
		{
			name:    "invalid with subfolder missing file",
			fields:  fields{path: testFolder + "/asd/sdf"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.path = tt.fields.path
			if err := fm.initWriter(false); (err != nil) != tt.wantErr {
				t.Errorf("initWriter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := fm.initScanner(); (err != nil) != tt.wantErr {
				t.Errorf("initScanner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileManager_WriteAndReadMetric(t *testing.T) {
	numFloat := 123.0
	numInt64 := int64(64)
	_ = os.RemoveAll(testFolder)
	log := logger.CreateMock()

	fm := createFileManagerTest(testConfig.StoragePath, log)

	type fields struct {
		path string
	}
	type args struct {
		name  string
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "float valid",
			fields:  fields{path: testFolder + "/asd"},
			args:    args{name: "someM", value: &numFloat},
			wantErr: false,
		},
		{
			name:    "int64 valid",
			fields:  fields{path: testFolder + "/asd"},
			args:    args{name: "someMInt", value: &numInt64},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = fm.OpenFile(tt.fields.path, true)
			if err := fm.WriteMetric(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("WriteMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			metric, err := fm.ScanMetric()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, metric.Name, tt.args.name)

			_ = fm.CloseFile()
		})
	}
}

func TestFileManager_WriteAndScanMetricOnClosed(t *testing.T) {
	_ = os.RemoveAll(testFolder)

	log := logger.CreateMock()
	defer log.Sync()
	fm := createFileManagerTest(testConfig.StoragePath, log)
	tests := []struct {
		name    string
		want    *MetricData
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "closed file",
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fm.ScanMetric()
			if !tt.wantErr(t, err, "ScanMetric()") {
				return
			}

			err = fm.WriteMetric("asd", 123)
			if !tt.wantErr(t, err, "WriteMetric()") {
				return
			}

			assert.Equalf(t, tt.want, got, "ScanMetric()")
		})
	}
}

func TestFileManager_parseMetric(t *testing.T) {
	_ = os.RemoveAll(testFolder)

	log := logger.CreateMock()
	defer log.Sync()
	fm := createFileManagerTest(testConfig.StoragePath, log)

	type args struct {
		metric   *MetricData
		gauges   map[string]float64
		counters map[string]int64
	}
	type want struct {
		gauges   map[string]float64
		counters map[string]int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "int64 valid",
			args: args{
				metric: &MetricData{
					Name:      "int64 metric",
					ValueType: counterType,
					Value:     "123",
				},
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
			want: want{
				gauges:   map[string]float64{},
				counters: map[string]int64{"int64 metric": 123},
			},
		},
		{
			name: "float64 valid",
			args: args{
				metric: &MetricData{
					Name:      "float64 metric",
					ValueType: gaugeType,
					Value:     "123",
				},
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
			want: want{
				gauges:   map[string]float64{"float64 metric": 123},
				counters: map[string]int64{},
			},
		},
		{
			name: "int64 incorrect metric value",
			args: args{
				metric: &MetricData{
					Name:      "int64 metric",
					ValueType: counterType,
					Value:     "asd",
				},
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
			want: want{
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
		},
		{
			name: "float64 incorrect metric value",
			args: args{
				metric: &MetricData{
					Name:      "float64 metric",
					ValueType: gaugeType,
					Value:     "asd",
				},
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
			want: want{
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm.parseMetric(tt.args.metric, &tt.args.gauges, &tt.args.counters)

			assert.True(t, reflect.DeepEqual(tt.args.gauges, tt.want.gauges))
			assert.True(t, reflect.DeepEqual(tt.args.counters, tt.want.counters))
		})
	}
}

func TestFileManager_SaveAndReadFile(t *testing.T) {
	_ = os.RemoveAll(testFolder)
	gauge := float64(123)
	counter := int64(12)

	var path = testFolder + "/dd1"

	log := logger.CreateMock()

	fm := createFileManagerTest(path, log)

	type args struct {
		gauges   map[string]interface{}
		counters map[string]interface{}
	}
	type want struct {
		gauges   map[string]float64
		counters map[string]int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "float64 valid",
			args: args{

				gauges:   map[string]interface{}{"float64 metric": &gauge},
				counters: map[string]interface{}{},
			},
			want: want{
				gauges:   map[string]float64{"float64 metric": 123},
				counters: map[string]int64{},
			},
		},
		{
			name: "int64 valid",
			args: args{
				gauges:   map[string]interface{}{},
				counters: map[string]interface{}{"counter metric": &counter},
			},
			want: want{
				gauges:   map[string]float64{},
				counters: map[string]int64{"counter metric": int64(12)},
			},
		},
		{
			name: "int64 and float64 valid",
			args: args{
				gauges:   map[string]interface{}{"float64 metric": &gauge},
				counters: map[string]interface{}{"counter metric": &counter},
			},
			want: want{
				gauges:   map[string]float64{"float64 metric": 123},
				counters: map[string]int64{"counter metric": int64(12)},
			},
		},
		{
			name: "nothing to save",
			args: args{
				gauges:   map[string]interface{}{},
				counters: map[string]interface{}{},
			},
			want: want{
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
		},
		{
			name: "incorrect metric type",
			args: args{
				gauges:   map[string]interface{}{"float64 metric": 123},
				counters: map[string]interface{}{},
			},
			want: want{
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.RemoveAll(path)
			err := fm.SaveMetricsInStorage(context.Background(), tt.args.gauges, tt.args.counters)
			require.NoError(t, err)
			require.NoError(t, fm.Close())

			gauges, counters, err := fm.RestoreDataFromStorage(context.Background())
			require.NoError(t, err)
			assert.True(t, reflect.DeepEqual(gauges, tt.want.gauges))
			assert.True(t, reflect.DeepEqual(counters, tt.want.counters))

		})
	}
}
