package country_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
)

// reading is a Claim flattened to the token text it covers, so cases read as
// "these words, claimed as this part, this strongly".
type reading struct {
	text       string
	part       claim.Part
	confidence claim.Confidence
	value      string
}

func flatten(tokens []token.Token, claims []claim.Claim) []reading {
	out := make([]reading, 0, len(claims))
	for _, c := range claims {
		for _, p := range c.Parts {
			out = append(out, reading{
				token.Join(tokens[p.Start:p.End()]),
				p.Part,
				c.Confidence,
				p.Value,
			})
		}
	}

	return out
}

func TestClaims(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []reading
	}{
		{
			name: "spelled out name",
			in:   "CANADA",
			want: []reading{
				{"CANADA", claim.PartCountry, claim.ConfidenceStrong, "CANADA"},
			},
		},
		{
			name: "abbreviation is exact within this vocabulary",
			in:   "MX",
			want: []reading{
				{"MX", claim.PartCountry, claim.ConfidenceExact, "MEXICO"},
			},
		},
		{
			// Project US@ omits the country on a domestic address, so the
			// normalized value is empty. The claim still stands: these tokens
			// are the country component.
			name: "domestic name normalizes to nothing",
			in:   "UNITED STATES",
			want: []reading{
				{"UNITED STATES", claim.PartCountry, claim.ConfidenceStrong, ""},
			},
		},
		{
			name: "domestic abbreviations",
			in:   "USA",
			want: []reading{
				{"USA", claim.PartCountry, claim.ConfidenceExact, ""},
			},
		},
		{
			// NormalizeCountry would happily return "FREEDONIA" here. That is
			// formatting, not evidence, and is not a claim.
			name: "unknown country is not claimed",
			in:   "FREEDONIA",
			want: []reading{},
		},
		{
			name: "no country present",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, country.Claims(tokens))

			if len(got) != len(tc.want) {
				t.Fatalf("Claims(%q) returned %d claims, want %d: %+v", tc.in, len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("Claims(%q)[%d] = %+v, want %+v", tc.in, i, got[i], w)
				}
			}
		})
	}
}

// The country is the last thing on the address, so it has to be found after a
// postal code rather than at the start of the input.
func TestClaimsAtEndOfAddress(t *testing.T) {
	tokens := token.Tokenize("123 MAIN ST, TORONTO, ON M5V 3L9, CANADA")
	claims := country.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one country claim, got %+v", claims)
	}
	if got := token.Join(tokens[claims[0].Start():claims[0].End()]); got != "CANADA" {
		t.Errorf("claimed %q, want %q", got, "CANADA")
	}
}

// A country name that runs past the end of the token slice must not be read as
// a shorter match by accident, and must not panic.
func TestClaimsNearEndOfInput(t *testing.T) {
	tokens := token.Tokenize("SOUTH BEND UNITED")
	for _, c := range country.Claims(tokens) {
		if c.End() > len(tokens) {
			t.Fatalf("claim %+v runs past %d tokens", c, len(tokens))
		}
	}
}

// A country name hard wrapped across two lines is not a country name. Join
// flattens a span with spaces, so without a bound "UNITED\nSTATES" reaches the
// lookup as "UNITED STATES". See token.LineEnd.
func TestNoClaimSpansALineBreak(t *testing.T) {
	tokens := token.Tokenize("123 MAIN ST\nUNITED\nSTATES")
	for _, c := range country.Claims(tokens) {
		if tokens[c.End()-1].Line != tokens[c.Start()].Line {
			t.Errorf("claim over %q spans a line break",
				token.Join(tokens[c.Start():c.End()]))
		}
	}
}

// The bound must not cost the reading it was added to protect: the same phrase
// on one line is still a country. It normalizes to the empty string, which is
// what a domestic country claim says, so the span is what identifies it.
func TestAPhraseOnOneLineIsStillClaimed(t *testing.T) {
	tokens := token.Tokenize("123 MAIN ST\nDENVER CO 80201\nUNITED STATES")

	var found bool
	for _, c := range country.Claims(tokens) {
		if token.Join(tokens[c.Start():c.End()]) == "UNITED STATES" {
			found = true
		}
	}
	if !found {
		t.Error("expected UNITED STATES on its own line to still be claimed")
	}
}
