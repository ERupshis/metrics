package staticio

import (
	"encoding/json"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"
)

// ChecksSA returns slice of static check with SA in short name.
func ChecksSA() []*analysis.Analyzer {
	res := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			res = append(res, v.Analyzer)
		}
	}

	return res
}

// configData describes the structure of the configuration file.
type configData struct {
	Staticcheck []string
}

// ChecksFromConfig returns a slice of static checks with names from the specified config file.
func ChecksFromConfig(configPath string) ([]*analysis.Analyzer, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg configData
	if err = json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}

	res := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			res = append(res, v.Analyzer)
		}
	}

	return res, nil
}
