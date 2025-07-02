// Package main запускает multichecker со множеством анализаторов.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"

	"github.com/gostaticanalysis/nilerr"
	myAnalyzer "github.com/sinfirst/URL-Cutter/cmd/staticlint/myAnalyzer"
)

func main() {
	var analyzers []*analysis.Analyzer

	for _, a := range staticcheck.Analyzers {
		if a.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	analyzers = append(analyzers, nilerr.Analyzer)
	analyzers = append(analyzers, myAnalyzer.Analyzer)

	multichecker.Main(analyzers...)
}
