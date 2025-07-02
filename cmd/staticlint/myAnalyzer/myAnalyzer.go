// Package analyzer пакет с описание анализатора на предмет os.Exit
package analyzer

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer анализатор на предмет os.Exit
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "запрещает использование os.Exit в функции main пакета main",
	Run:  run,
}

// run фунция для проверки кода
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok || sel.Sel.Name != "Exit" || pkgIdent.Name != "os" {
					return true
				}
				obj := pass.TypesInfo.Uses[pkgIdent]
				if obj == nil || !strings.HasPrefix(obj.Type().String(), "package") {
					return true
				}
				pkgName, ok := obj.(*types.PkgName)
				if ok && pkgName.Imported().Path() == "os" {
					pass.Reportf(call.Lparen, "запрещено использовать os.Exit в функции main пакета main")
				}
				return true
			})
		}
	}
	return nil, nil
}
