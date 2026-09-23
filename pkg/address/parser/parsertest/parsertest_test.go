package parsertest_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
)

// TestRunReportsPerCase checks Run and Result.Pass against a stub parser
// with a known right answer and a known wrong one, rather than against a
// real parser: this package ships the corpus and the runner, not a
// linguistic implementation, so its own tests must not need one either.
func TestRunReportsPerCase(t *testing.T) {
	stub := parser.ParsingFn(func(source string) (*address.Address, error) {
		if source == "100 MAIN ST" {
			return &address.Address{PrimaryNumber: "100", StreetName: "MAIN", StreetSuffix: "ST"}, nil
		}
		return &address.Address{StreetName: "WRONG"}, nil
	})

	cases := []parsertest.Case{
		{Source: "ours", Note: "right answer", Input: "100 MAIN ST", Want: "100 MAIN ST"},
		{Source: "ours", Note: "wrong answer", Input: "200 OAK AVE", Want: "200 OAK AVE"},
		{Source: "ours", Note: "no ground truth yet", Input: "300 ELM PARK"},
	}

	results := parsertest.Run(stub, cases)
	if len(results) != len(cases) {
		t.Fatalf("Run returned %d results for %d cases", len(results), len(cases))
	}
	// An unsettled case must not read as a parser failure; see Case.Want.
	if results[2].Settled() || results[2].Pass() {
		t.Errorf("case %q: Settled() = %v, Pass() = %v; want false, false",
			cases[2].Note, results[2].Settled(), results[2].Pass())
	}
	if !results[0].Pass() {
		t.Errorf("case %q: got %q, want %q, Pass() = false", cases[0].Note, results[0].Got, cases[0].Want)
	}
	if results[1].Pass() {
		t.Errorf("case %q: Pass() = true for a parser that returns the wrong address", cases[1].Note)
	}

	pass, settled, total := parsertest.CountPass(results)
	if pass != 1 || settled != 2 || total != 3 {
		t.Errorf("CountPass = %d, %d, %d; want 1, 2, 3", pass, settled, total)
	}
}

// TestCasesComposition pins how many cases each set of the corpus carries.
// It is a ratchet on the corpus's shape, not on any parser's score: it exists
// so a case added or dropped from spec.go, historical.go, or ours.go is a
// deliberate, reviewed number change here, not a silent drift.
func TestCasesComposition(t *testing.T) {
	const (
		wantSpec       = 46 * 2     // one fixed-point and one reaches case per Incorrect/Correct pair
		wantHistorical = 55 + 3 + 4 // csharpParity + gridHistorical + saintHistorical
		wantOurs       = 11
	)

	if got := len(parsertest.SpecCases); got != wantSpec {
		t.Errorf("len(SpecCases) = %d, want %d", got, wantSpec)
	}
	if got := len(parsertest.HistoricalCases); got != wantHistorical {
		t.Errorf("len(HistoricalCases) = %d, want %d", got, wantHistorical)
	}
	if got := len(parsertest.OursCases); got != wantOurs {
		t.Errorf("len(OursCases) = %d, want %d", got, wantOurs)
	}
	if want := wantSpec + wantHistorical + wantOurs; len(parsertest.Cases) != want {
		t.Errorf("len(Cases) = %d, want %d", len(parsertest.Cases), want)
	}
}
