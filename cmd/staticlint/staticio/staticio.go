package staticio

import (
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
