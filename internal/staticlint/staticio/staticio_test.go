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
	// Create a temporary config file for testing
	tempConfigPath := createTempConfigFile(t, `{"Staticcheck": ["S1000", "S1001", "ST1000", "ST1005", "QF1003", "QF1010"]}`)
	defer func() {
		_ = os.Remove(tempConfigPath)
	}()

	// Call the function with the temporary config file
	result, err := ChecksFromConfig(tempConfigPath)
	if err != nil {
		t.Fatalf("ChecksFromConfig returned an error: %v", err)
	}

	// Add your assertions based on the expected result
	// For example, check the length of the result or specific analyzers
	if len(result) != 6 {
		t.Errorf("Expected 6 analyzers, got %d", len(result))
	}

	// Add more assertions as needed
}

func createTempConfigFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer func() {
		_ = tmpfile.Close()
	}()

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}

	return tmpfile.Name()
}

func TestReadConfigFromFile(t *testing.T) {
	tempConfigPath := createTempConfigFile(t, `{"Staticcheck": ["S1000", "S1001", "ST1000", "ST1005", "QF1003", "QF1010"]}`)
	defer func() {
		_ = os.Remove(tempConfigPath)
	}()

	// Call the function with the temporary config file
	result, err := readConfigFromFile(tempConfigPath)
	if err != nil {
		t.Fatalf("readConfigFromFile returned an error: %v", err)
	}

	assert.Equal(t, 6, len(result.Staticcheck))
}
