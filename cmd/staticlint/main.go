package main

import (
	"fmt"

	"github.com/erupshis/metrics/cmd/staticlint/passes"
	"github.com/erupshis/metrics/cmd/staticlint/staticio"
	"github.com/erupshis/metrics/internal/logger"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

const config = "cmd/staticlint/staticio/config.json"

func main() {
	log := logger.CreateLogger("info")
	defer log.Sync()

	var checks []*analysis.Analyzer

	if err := addPassesChecks(&checks); err != nil {
		log.Info("failed to add passes checks")
	}

	if err := addStaticChecksIO(&checks); err != nil {
		log.Info("failed to add static checks")
	}

	multichecker.Main(
		checks...,
	)
}

func addPassesChecks(checks *[]*analysis.Analyzer) error {
	*checks = append(*checks, passes.Checks()...)
	return nil
}

func addStaticChecksIO(checks *[]*analysis.Analyzer) error {
	*checks = append(*checks, staticio.ChecksSA()...)

	configChecks, err := staticio.ChecksFromConfig(config)
	if err != nil {
		return fmt.Errorf("add checks mentioned in config.json: %w", err)
	}

	*checks = append(*checks, configChecks...)
	return nil
}
