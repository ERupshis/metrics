// Package main implements analyzer checks according list:
// - the standard set of analyzers from the golang.org/x/tools/go/analysis/passes package;
//
// To install checker use `go install`.
//
// Usage: staticlint [package]. To run recursively in all subfolders use `staticlint ./...`
//
// By default all analyzers will be used.

package main
