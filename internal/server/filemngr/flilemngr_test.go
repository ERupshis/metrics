package filemngr

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFolder = "/test"

func TestFileManager_IsFileOpen(t *testing.T) {
	os.RemoveAll(testFolder)

	fm := Create()
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
	os.RemoveAll(testFolder)
	type fields struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "valid open file",
			fields: fields{path: testFolder + "/asd/aa"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := Create()

			if err := fm.OpenFile(tt.fields.path, false); err != nil {
				t.Errorf("OpenFile() without Trunc error = %v", err)
			}
		})
	}
}

func TestFileManager_CloseFile(t *testing.T) {
	os.RemoveAll(testFolder)
	//test close on missing file.
	fm := Create()
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
			fields: fields{path: testFolder + "/asd/aa11"},
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
	os.RemoveAll(testFolder)
	var path = testFolder + "/dd"
	fm := Create()
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
	fm := Create()

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

	//check trunc
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
