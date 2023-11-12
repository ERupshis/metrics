package storagemngr

import (
	"os"
	"testing"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/stretchr/testify/assert"
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
	cfg := config.Parse()
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
	fm.OpenFile(path, false)
	fm.CloseFile()

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
	os.RemoveAll(testFolder)
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
			args:    args{name: "someM", value: float64(123)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.OpenFile(tt.fields.path, true)
			if err := fm.WriteMetric(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("WriteMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			metric, err := fm.ScanMetric()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, metric.Name, tt.args.name)

			fm.CloseFile()
		})
	}

	// check trunc
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.OpenFile(tt.fields.path, false)

			metric, err := fm.ScanMetric()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, metric.Name, tt.args.name)

			fm.CloseFile()
		})
	}
}
