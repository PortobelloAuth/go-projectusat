// corpusmeasure runs the parsertest corpora — spec examples, historical
// regressions, and ours — through a parser and reports how many of each
// suite it settles and passes.
//
// It defaults to the embedded zipcity-backed parser (zipcityembedded), the
// parser go-projectusat uses by default. Point it at another implementation
// by changing newParserUnderTest below; anything satisfying
// parser.ParsingFunc works, which is how addressparsers scores itself
// against this same corpus (see poetic-systems/addressparsers
// internal/probe/refcorpusmeasure.go).
//
// Run with:
//
//	go run ./test/cmd/corpusmeasure
package main

import (
	"fmt"
	"log"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/zipcityembedded"
)

func newParserUnderTest() (parser.ParsingFunc, error) {
	return zipcityembedded.New()
}

func main() {
	p, err := newParserUnderTest()
	if err != nil {
		log.Fatalf("could not build parser: %v", err)
	}

	suites := []struct {
		name  string
		cases []parsertest.Case
	}{
		{"SpecCases", parsertest.SpecCases},
		{"HistoricalCases", parsertest.HistoricalCases},
		{"OursCases", parsertest.OursCases},
	}

	for _, s := range suites {
		results := parsertest.Run(p, s.cases)
		pass, settled, total := parsertest.CountPass(results)
		fmt.Printf("%s: %d/%d/%d (pass/settled/total)\n", s.name, pass, settled, total)
		for _, r := range results {
			if r.Settled() && !r.Pass() {
				fmt.Printf("  FAIL [%s] %q\n    want: %q\n    got:  %q (err=%v)\n", r.Case.Source, r.Case.Input, r.Case.Want, r.Got, r.Err)
			}
		}
	}
}
