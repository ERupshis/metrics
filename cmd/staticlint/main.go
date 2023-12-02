package main

import (
	"fmt"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/staticlint/exitcheck"
	"github.com/erupshis/metrics/internal/staticlint/passes"
	"github.com/erupshis/metrics/internal/staticlint/staticio"
	"github.com/fatih/errwrap/errwrap"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

const config = "cmd/staticlint/config.json"

func main() {
	log := logger.CreateLogger("info")
	defer log.Sync()

	var checks []*analysis.Analyzer

	if err := addPassesChecks(&checks); err != nil {
		log.Info("failed to add passes checks: %v", err)
	}

	if err := addStaticChecksIO(&checks); err != nil {
		log.Info("failed to add static checks: %v", err)
	}

	if err := addPublicAnalyzers(&checks); err != nil {
		log.Info("failed to add public analyzers: %v", err)
	}

	if err := addExitCheck(&checks); err != nil {
		log.Info("failed to add exit in main func analyzer: %v", err)
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

func addPublicAnalyzers(checks *[]*analysis.Analyzer) error {
	*checks = append(*checks, errcheck.Analyzer)
	*checks = append(*checks, errwrap.Analyzer)
	return nil
}

func addExitCheck(checks *[]*analysis.Analyzer) error {
	*checks = append(*checks, exitcheck.Analyzer)
	return nil
}
