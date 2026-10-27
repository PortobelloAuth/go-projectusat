package parser_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ordinarystreet"
)

// Distinguish E St from East St
func TestParse(t *testing.T) {
	cases := []struct {
		In   string
		Want address.Address
	}{
		// Post-directional followed by a City with a directional prefix
		{
			In: "43 E 200 N, NORTH SALT LAKE, UT",
			Want: address.Address{
				Type:                &ordinarystreet.OrdinaryStreetAddress{},
				PrimaryNumber:       "43",
				Predirectional:      "E",
				StreetName:          "200",
				StreetSuffix:        "",
				Postdirectional:     "N",
				SecondaryDesignator: "",
				SecondaryNumber:     "",
				City:                "NORTH SALT LAKE",
				Region:              "UT",
				Postal:              "",
				Country:             "",
			},
		},
		// 3253 W 9200 S, West Jordan, UT 84088
		{
			In: "3253 W 9200 S, West Jordan, UT 84088",
			Want: address.Address{
				Type:                &ordinarystreet.OrdinaryStreetAddress{},
				PrimaryNumber:       "3253",
				Predirectional:      "W",
				StreetName:          "9200",
				StreetSuffix:        "",
				Postdirectional:     "S",
				SecondaryDesignator: "",
				SecondaryNumber:     "",
				City:                "WEST JORDAN",
				Region:              "UT",
				Postal:              "84088",
				Country:             "",
			},
		},
	}

	p, err := parser.New()
	if err != nil {
		t.Fatalf("Error instantiating parser: %s", err)
	}
	for _, tc := range cases {
		got, err := p.Parse(tc.In)
		if err != nil {
			t.Fatalf("Error parsing '%s': %s", tc.In, err)
		}

		if !got.Equals(&tc.Want) {
			t.Errorf("Unexpected result parsing '%s': %s expected: %s", tc.In, *got, tc.Want)
		}
	}
}

// TestParseAppliesVerifierToCustomParse verifies that Parse runs
// Options.Verifier on the address returned by Options.CustomParser, rather
// than returning the custom parser's result untouched.
func TestParseAppliesVerifierToCustomParse(t *testing.T) {
	parsed := &address.Address{PrimaryNumber: "123", StreetName: "MAIN", StreetSuffix: "ST"}
	wantErr := errors.New("verification failed")

	p, err := parser.New(parser.AddressParsingOptions{
		CustomParser: parser.ParsingFn(func(source string) (*address.Address, error) {
			return parsed, nil
		}),
		Verifier: func(a *address.Address) (*address.Address, error) {
			return nil, wantErr
		},
	})
	if err != nil {
		t.Fatalf("Error instantiating parser: %s", err)
	}

	_, err = p.Parse("123 MAIN ST")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Parse did not return the verifier's error: got %v, want %v", err, wantErr)
	}
}

// TestParseReturnsVerifierResult verifies that Parse returns whatever
// address the Verifier produces, not the custom parser's original address,
// and that a Parser created without an explicit Verifier defaults to
// IdentityVerifier and leaves the custom parser's address unchanged.
func TestParseReturnsVerifierResult(t *testing.T) {
	parsed := &address.Address{PrimaryNumber: "123", StreetName: "MAIN", StreetSuffix: "ST"}
	verified := &address.Address{PrimaryNumber: "123", StreetName: "MAIN", StreetSuffix: "ST", City: "VERIFIED"}

	p, err := parser.New(parser.AddressParsingOptions{
		CustomParser: parser.ParsingFn(func(source string) (*address.Address, error) {
			return parsed, nil
		}),
		Verifier: func(a *address.Address) (*address.Address, error) {
			return verified, nil
		},
	})
	if err != nil {
		t.Fatalf("Error instantiating parser: %s", err)
	}

	got, err := p.Parse("123 MAIN ST")
	if err != nil {
		t.Fatalf("Unexpected error parsing '123 MAIN ST': %s", err)
	}
	if !got.Equals(verified) {
		t.Errorf("Parse returned the custom parser's address instead of the verifier's result: got %s, want %s", *got, *verified)
	}

	identityParser, err := parser.New(parser.AddressParsingOptions{
		CustomParser: parser.ParsingFn(func(source string) (*address.Address, error) {
			return parsed, nil
		}),
	})
	if err != nil {
		t.Fatalf("Error instantiating parser: %s", err)
	}

	got, err = identityParser.Parse("123 MAIN ST")
	if err != nil {
		t.Fatalf("Unexpected error parsing '123 MAIN ST' with default verifier: %s", err)
	}
	if !got.Equals(parsed) {
		t.Errorf("Parse with default Verifier changed the custom parser's address: got %s, want %s", *got, *parsed)
	}
}

func TestDefaultParserWithSpecSuite(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("could not build zipcity parser: %v", err)
	}

	suites := []struct {
		name  string
		cases []parsertest.Case
	}{
		{"SpecCases", parsertest.SpecCases},
		{"HistoricalCases", parsertest.HistoricalCases},
		{"OursCases", parsertest.OursCases},
		{"ScourgifyCases", parsertest.ScourgifyCases},
		// NOTE: we will enable the HTE and usaddress suites after they have settled
		// and are usable as regression test suites
		// {"HteCases", parsertest.HteCases},
		// {"UsAddressCases", parsertest.UsAddressCases},
	}

	for _, s := range suites {
		results, err := parsertest.Run(p, s.cases)
		if err != nil {
			t.Fatalf("Run() error: %s", err)
		}
		pass, settled, total := parsertest.CountPassAll(results)
		if !(pass == settled && settled == total) {
			t.Errorf("%s: %d/%d/%d (pass/settled/total)\n", s.name, pass, settled, total)
		}
		for _, r := range results {
			switch {
			case r.Err != nil:
				t.Errorf("  ERR  [%s] %q: %v\n", r.Case.Source, r.Case.Input, r.Err)
			case (r.Settled() || r.SettledFields()) && !(r.Pass() || r.PassFields()):
				want := fmt.Sprintf("%s", r.Case.Want)
				if !r.Settled() && r.SettledFields() {
					want = fmt.Sprintf("%q", r.Case.WantFields)
				}
				t.Errorf("  FAIL [%s] %q\n     want: %q\n      got: %q\n   parsed: %q\n", r.Case.Source, r.Case.Input, want, r.Got, r.GotFields)
			}
		}
	}
}
