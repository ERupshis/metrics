package main

import (
	"github.com/erupshis/metrics/cmd/staticlint/passes"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	var mychecks []*analysis.Analyzer

	addPassesChecks(&mychecks)

	multichecker.Main(
		mychecks...,
	)
}

func addPassesChecks(checks *[]*analysis.Analyzer) {
	*checks = append(*checks, passes.PassesChecks()...)
}
