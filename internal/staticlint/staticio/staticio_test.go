package staticio

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/stylecheck"
)

func Test_addFilteredAnalyzers(t *testing.T) {
	type args struct {
		from   []*lint.Analyzer
		to     []*analysis.Analyzer
		filter map[string]bool
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{
			name: "valid, style analyzers",
			args: args{
				from:   stylecheck.Analyzers,
				to:     make([]*analysis.Analyzer, 0),
				filter: map[string]bool{"ST1000": true},
			},
			wantCount: 1,
		},
		{
			name: "valid, style analyzers, plus simple in filters",
			args: args{
				from:   stylecheck.Analyzers,
				to:     make([]*analysis.Analyzer, 0),
				filter: map[string]bool{"ST1000": true, "S1000": true},
			},
			wantCount: 1,
		},
		{
			name: "nothing found, filters cannot sort quickfix",
			args: args{
				from:   quickfix.Analyzers,
				to:     make([]*analysis.Analyzer, 0),
				filter: map[string]bool{"ST1000": true, "S1000": true},
			},
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addFilteredAnalyzers(tt.args.from, tt.args.to, tt.args.filter)
			assert.Equal(t, tt.wantCount, len(got))
		})
	}
}

func TestChecksSA(t *testing.T) {
	tests := []struct {
		name      string
		wantCount int
	}{
		{
			name:      "valid",
			wantCount: 90,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCount, len(ChecksSA()))
		})
	}
}

func TestChecksFromConfig(t *testing.T) {
	type args struct {
		createTmpFile bool
		fileData      string
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{
			name: "valid",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "ST1000", "ST1005", "QF1003", "QF1010"]}`,
				createTmpFile: true,
			},
			wantCount: 6,
		},
		{
			name: "valid with incorrect analyzers",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "S000", "ST1005", "QF10", "QF1010"]}`,
				createTmpFile: true,
			},
			wantCount: 4,
		},
		{
			name: "failed to read file",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "S000", "ST1005", "QF10", "QF1010"]}`,
				createTmpFile: false,
			},
			wantCount: 0,
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary config file for testing

			tempConfigPath := ""
			if tt.args.createTmpFile {
				tempConfigPath = createTempConfigFile(t, tt.args.fileData)
				defer func() {
					_ = os.Remove(tempConfigPath)
				}()
			}

			result, err := ChecksFromConfig(tempConfigPath)
			if err != nil && tt.args.createTmpFile {
				t.Fatalf("ChecksFromConfig returned an error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("Expected %d analyzers, got %d", tt.wantCount, len(result))
			}
		})
	}
}

func createTempConfigFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer func() {
		_ = tmpfile.Close()
	}()

	if _, err = tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}

	return tmpfile.Name()
}

func TestReadConfigFromFile(t *testing.T) {
	type args struct {
		createTmpFile bool
		fileData      string
	}
	type want struct {
		wantCount int
		wantErr   bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "ST1000", "ST1005", "QF1003", "QF1010"]}`,
				createTmpFile: true,
			},
			want: want{
				wantCount: 6,
			},
		},
		{
			name: "valid with incorrect analyzers",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "S000", "ST1005", "QF10", "QF1010"]}`,
				createTmpFile: true,
			},
			want: want{
				wantCount: 6,
			},
		},
		{
			name: "failed to read file",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "S000", "ST1005", "QF10", "QF1010"]}`,
				createTmpFile: false,
			},
			want: want{
				wantCount: 6,
			},
		},
		{
			name: "incorrect json body",
			args: args{
				fileData:      `{"Staticcheck": ["S1000", "S1001", "S000", "ST1005", "QF10", "QF1010"`,
				createTmpFile: true,
			},
			want: want{
				wantErr:   true,
				wantCount: 6,
			},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			tempConfigPath := ""
			if tt.args.createTmpFile {
				tempConfigPath = createTempConfigFile(t, tt.args.fileData)
				defer func() {
					_ = os.Remove(tempConfigPath)
				}()
			}

			// Call the function with the temporary config file
			result, err := readConfigFromFile(tempConfigPath)
			if tt.args.createTmpFile && !tt.want.wantErr {
				if err != nil {
					t.Fatalf("readConfigFromFile returned an error: %v", err)
				}

				assert.Equal(t, tt.want.wantCount, len(result.Staticcheck))
			}
		})
	}
}
