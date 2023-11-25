package main

import (
	"github.com/erupshis/metrics/cmd/staticlint/passes"
	"github.com/erupshis/metrics/cmd/staticlint/staticio"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	var mychecks []*analysis.Analyzer

	addPassesChecks(&mychecks)
	addStaticChecksIO(&mychecks)

	multichecker.Main(
		mychecks...,
	)
}

func addPassesChecks(checks *[]*analysis.Analyzer) {
	*checks = append(*checks, passes.Checks()...)
}

func addStaticChecksIO(checks *[]*analysis.Analyzer) {
	*checks = append(*checks, staticio.ChecksSA()...)
}
