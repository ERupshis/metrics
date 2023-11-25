package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestExitChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./...")
}
