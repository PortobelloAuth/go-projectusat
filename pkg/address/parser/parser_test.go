package parser_test

import (
	"errors"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
)

// Distinguish E St from East St
func TestParse(t *testing.T) {
	t.Skip("parser.Parse returns \"Not implemented\" until go-projectusat#61 lands")
	cases := []struct {
		In   string
		Want address.Address
	}{
		// Post-directional followed by a City with a directional prefix
		{
			In: "43 E 200 N, NORTH SALT LAKE, UT",
			Want: address.Address{
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
				PrimaryNumber:       "3253",
				Predirectional:      "W",
				StreetName:          "9200",
				StreetSuffix:        "",
				Postdirectional:     "S",
				SecondaryDesignator: "",
				SecondaryNumber:     "",
				City:                "West Jordan",
				Region:              "UT",
				Postal:              "84088",
				Country:             "",
			},
		},
	}

	p := parser.New()
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

	p := parser.New(parser.AddressParsingOptions{
		CustomParser: parser.ParsingFn(func(source string) (*address.Address, error) {
			return parsed, nil
		}),
		Verifier: func(a *address.Address) (*address.Address, error) {
			return nil, wantErr
		},
	})

	_, err := p.Parse("123 MAIN ST")
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

	p := parser.New(parser.AddressParsingOptions{
		CustomParser: parser.ParsingFn(func(source string) (*address.Address, error) {
			return parsed, nil
		}),
		Verifier: func(a *address.Address) (*address.Address, error) {
			return verified, nil
		},
	})

	got, err := p.Parse("123 MAIN ST")
	if err != nil {
		t.Fatalf("Unexpected error parsing '123 MAIN ST': %s", err)
	}
	if !got.Equals(verified) {
		t.Errorf("Parse returned the custom parser's address instead of the verifier's result: got %s, want %s", *got, *verified)
	}

	identityParser := parser.New(parser.AddressParsingOptions{
		CustomParser: parser.ParsingFn(func(source string) (*address.Address, error) {
			return parsed, nil
		}),
	})

	got, err = identityParser.Parse("123 MAIN ST")
	if err != nil {
		t.Fatalf("Unexpected error parsing '123 MAIN ST' with default verifier: %s", err)
	}
	if !got.Equals(parsed) {
		t.Errorf("Parse with default Verifier changed the custom parser's address: got %s, want %s", *got, *parsed)
	}
}
