package filemngr

import "testing"

func TestFileManager_CloseFile(t *testing.T) {
	type fields struct {
		path    string
		writer  *fileWriter
		scanner *fileScanner
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := &FileManager{
				path:    tt.fields.path,
				writer:  tt.fields.writer,
				scanner: tt.fields.scanner,
			}
			if err := fm.CloseFile(); (err != nil) != tt.wantErr {
				//t.Errorf("CloseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
