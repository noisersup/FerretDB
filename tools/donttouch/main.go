package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var analyzer = &analysis.Analyzer{
	Name: "checkswitch",
	Doc:  "check the preferred order of types in the switch",
	Run:  run,
}

func main() {
	singlechecker.Main(analyzer)
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			var id *ast.Ident
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				id = fun
			case *ast.SelectorExpr:
				id = fun.Sel
			}

			if id != nil && !pass.TypesInfo.Types[id].IsType() {
			}
			return true
		})
	}
	return nil, nil
}
