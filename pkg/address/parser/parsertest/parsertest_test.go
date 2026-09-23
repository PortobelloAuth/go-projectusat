package parsertest_test

import (
	"fmt"
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

// TestRunFieldsReportsPerCase checks RunFields and FieldsResult.Pass against
// a stub parser, the same way TestRunReportsPerCase checks Run: a matching
// decomposition, a mismatching one, an unsettled case (nil WantFields), and
// — the case this type exists for — a decomposition that matches everywhere
// except a field WantFields expects to be empty. That last one is the case
// Want could never catch: "123 NORTH PARK" renders the same whether NORTH is
// split out as a predirectional or not, so only comparing the empty
// Predirectional field against a wrongly non-empty one catches the bad
// split (go-projectusat#123).
func TestRunFieldsReportsPerCase(t *testing.T) {
	stub := parser.ParsingFn(func(source string) (*address.Address, error) {
		switch source {
		case "999 UNASSERTED ST":
			return &address.Address{PrimaryNumber: "999", StreetName: "UNASSERTED", StreetSuffix: "ST"}, nil
		case "2 OAK AVE":
			return &address.Address{PrimaryNumber: "2", StreetName: "OAK", StreetSuffix: "AVE"}, nil
		case "3 MAPLE DR":
			// Wrong: StreetName should be MAPLE, not MAPLEWOOD.
			return &address.Address{PrimaryNumber: "3", StreetName: "MAPLEWOOD", StreetSuffix: "DR"}, nil
		case "123 NORTH PARK":
			// Wrong in a way Want cannot see: this reads NORTH as a
			// predirectional and PARK as the name, rendering "123 N PARK".
			// WantFields for this case asserts NORTH PARK as one unsplit
			// street name instead, so Predirectional must come back empty.
			return &address.Address{PrimaryNumber: "123", Predirectional: "N", StreetName: "PARK"}, nil
		}
		return nil, fmt.Errorf("stub: unexpected input %q", source)
	})

	cases := []parsertest.Case{
		{Source: "ours", Note: "no decomposition asserted yet", Input: "999 UNASSERTED ST"},
		{
			Source:     "ours",
			Note:       "matching decomposition",
			Input:      "2 OAK AVE",
			WantFields: &address.Address{PrimaryNumber: "2", StreetName: "OAK", StreetSuffix: "AVE"},
		},
		{
			Source:     "ours",
			Note:       "mismatching decomposition",
			Input:      "3 MAPLE DR",
			WantFields: &address.Address{PrimaryNumber: "3", StreetName: "MAPLE", StreetSuffix: "DR"},
		},
		{
			Source:     "ours",
			Note:       "NORTH PARK is one unsplit street name, not a predirectional plus PARK",
			Input:      "123 NORTH PARK",
			WantFields: &address.Address{PrimaryNumber: "123", StreetName: "NORTH PARK"},
		},
	}

	results := parsertest.RunFields(stub, cases)
	if len(results) != len(cases) {
		t.Fatalf("RunFields returned %d results for %d cases", len(results), len(cases))
	}

	// A nil WantFields must not read as a parser failure; see Case.WantFields.
	if results[0].Settled() || results[0].Pass() {
		t.Errorf("case %q: Settled() = %v, Pass() = %v; want false, false",
			cases[0].Note, results[0].Settled(), results[0].Pass())
	}
	if fields := results[0].Fields(); fields != nil {
		t.Errorf("case %q: Fields() = %+v for an unsettled case; want nil", cases[0].Note, fields)
	}

	if !results[1].Pass() {
		t.Errorf("case %q: Pass() = false, Fields() = %+v", cases[1].Note, results[1].Fields())
	}

	if results[2].Pass() {
		t.Errorf("case %q: Pass() = true for a mismatching decomposition", cases[2].Note)
	}
	if !fieldMismatch(results[2].Fields(), "StreetName", "MAPLE", "MAPLEWOOD") {
		t.Errorf("case %q: Fields() = %+v, want a StreetName mismatch of \"MAPLE\" vs \"MAPLEWOOD\"",
			cases[2].Note, results[2].Fields())
	}

	if results[3].Pass() {
		t.Errorf("case %q: Pass() = true for a decomposition with a wrongly non-empty Predirectional", cases[3].Note)
	}
	if !fieldMismatch(results[3].Fields(), "Predirectional", "", "N") {
		t.Errorf("case %q: Fields() = %+v, want a Predirectional mismatch of \"\" vs \"N\"",
			cases[3].Note, results[3].Fields())
	}

	pass, settled, total := parsertest.CountPassFields(results)
	if pass != 1 || settled != 3 || total != 4 {
		t.Errorf("CountPassFields = %d, %d, %d; want 1, 3, 4", pass, settled, total)
	}
}

// fieldMismatch reports whether fields carries name with exactly want and
// got, and that it disagrees — the shape a caller uses to isolate which
// field a case is failing on.
func fieldMismatch(fields []parsertest.FieldResult, name, want, got string) bool {
	for _, f := range fields {
		if f.Field == name {
			return !f.Match() && f.Want == want && f.Got == got
		}
	}
	return false
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
