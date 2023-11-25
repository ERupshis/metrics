// Package exitcheck defines analyzer based on analysis.Analyzer.
// Detects call os.Exit in main func.
package exitcheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "detect direct os.Exit call in main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	mainFunc := getMainFunction(pass)

	if mainFunc == nil {
		return nil, nil
	}

	inspectFunc := func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
				if _, ok = sel.X.(*ast.Ident); ok {
					if isExitCall(n) {
						pass.Reportf(n.Pos(), "avoid direct os.Exit calls in main")
					}
				}
			}
		}
		return true
	}

	ast.Inspect(mainFunc.Body, inspectFunc)

	return nil, nil
}

func getMainFunction(pass *analysis.Pass) *ast.FuncDecl {
	for _, file := range pass.Files {
		if isTmpFile(pass.Fset.File(file.Pos()).Name()) {
			continue
		}

		if file.Name.Name == "main" {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
					return fn
				}
			}
		}
	}
	return nil
}

func isExitCall(call *ast.CallExpr) bool {
	selectorExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	if ident.Name == "os" && selectorExpr.Sel.Name == "Exit" {
		return true
	}

	return false
}

func isTmpFile(filePath string) bool {
	return strings.Contains(filePath, "go-build")
}
