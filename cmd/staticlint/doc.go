// Package main implements analyzer checks according list:
// - the standard set of analyzers from the golang.org/x/tools/go/analysis/passes package;
// - all analyzers of the SA class from the staticcheck.io package;
// - analyzers S1000, S1001, ST1000, ST1005, QF1003, QF1010 from the staticcheck.io package;
// - an analyzer from the public package github.com/kisielk/errcheck/errcheck;
// - an analyzer from the public package github.com/fatih/errwrap/errwrap;
// - a custom analyzer ExitAnalyzer that tracks the usage of direct os.Exit calls in the main function of the main package.
//
// To install checker use `go install`.
//
// Usage: staticlint [package]. To run recursively in all subfolders use `staticlint ./...`
//
// By default all analyzers will be used.

package main
