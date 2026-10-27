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
		{"HteCases", parsertest.HteCases},
		{"UsAddressCases", parsertest.UsAddressCases},
	}

	for _, s := range suites {
		results, err := parsertest.Run(p, s.cases)
		if err != nil {
			panic(fmt.Errorf("Run() error: %s", err))
		}
		pass, settled, total := parsertest.CountPassAll(results)
		fmt.Printf("%s: %d/%d/%d (pass/settled/total)\n", s.name, pass, settled, total)
		for _, r := range results {
			switch {
			case r.Err != nil:
				fmt.Printf("  ERR  [%s] %q: %v\n", r.Case.Source, r.Case.Input, r.Err)
			case (r.Settled() || r.SettledFields()) && !(r.Pass() || r.PassFields()):
				want := fmt.Sprintf("%s", r.Case.Want)
				if !r.Settled() && r.SettledFields() {
					want = fmt.Sprintf("%q", r.Case.WantFields)
				}
				fmt.Printf("  FAIL [%s] %q\n     want: %q\n      got: %q\n   parsed: %q\n", r.Case.Source, r.Case.Input, want, r.Got, r.GotFields)
			}
		}
	}
}
