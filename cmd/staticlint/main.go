// Package main запускает multichecker со множеством анализаторов.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"

	"github.com/gostaticanalysis/nilerr"
	"github.com/sinfirst/URL-Cutter/cmd/staticlint/noexit"
)

func main() {
	var analyzers []*analysis.Analyzer

	for _, a := range staticcheck.Analyzers {
		if len(a.Analyzer.Name) > 1 && a.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, a.Analyzer)
		}
	}
	analyzers = append(analyzers, nilerr.Analyzer)
	analyzers = append(analyzers, noexit.Analyzer)

	multichecker.Main(analyzers...)
}
