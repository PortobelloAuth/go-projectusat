package military_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/military"
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
			name: "post office designation stands where a city stands",
			in:   "APO",
			want: []reading{
				{"APO", claim.PartCity, claim.ConfidenceExact, "APO"},
			},
		},
		{
			name: "armed forces region",
			in:   "AE",
			want: []reading{
				{"AE", claim.PartRegion, claim.ConfidenceExact, "AE"},
			},
		},
		{
			// The facility and its number are the street name; the box and its
			// number are the primary address on it.
			name: "facility designator opens the street line",
			in:   "PSC 3 BOX 4120",
			want: []reading{
				{"PSC 3", claim.PartStreetName, claim.ConfidenceExact, "PSC 3"},
				{"BOX 4120", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 4120"},
			},
		},
		{
			// A designator on its own is a fragment of a pattern that did not
			// match, not weak evidence of a military address.
			name: "facility designator alone is not claimed",
			in:   "PSC",
			want: []reading{},
		},
		{
			name: "not military vocabulary",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, military.Claims(tokens))

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

// The standard's own example, claimed end to end.
func TestClaimsFullMilitaryAddress(t *testing.T) {
	tokens := token.Tokenize("PSC 3 BOX 4120\nAPO AE 09021-0002")
	got := flatten(tokens, military.Claims(tokens))

	want := []reading{
		{"PSC 3", claim.PartStreetName, claim.ConfidenceExact, "PSC 3"},
		{"BOX 4120", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 4120"},
		{"APO", claim.PartCity, claim.ConfidenceExact, "APO"},
		{"AE", claim.PartRegion, claim.ConfidenceExact, "AE"},
	}

	if len(got) != len(want) {
		t.Fatalf("Claims returned %d claims, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("claim %d = %+v, want %+v", i, got[i], w)
		}
	}
}

// BOX is never claimed on its own. Here it opens the primary number, in a
// rural route it does the same, and in a PO box it belongs to the designator.
// The word means nothing without the pattern around it.
func TestClaimsDoesNotClaimBoxAlone(t *testing.T) {
	for _, in := range []string{"BOX", "BOX 4120"} {
		t.Run(in, func(t *testing.T) {
			tokens := token.Tokenize(in)

			if claims := military.Claims(tokens); len(claims) != 0 {
				t.Errorf("expected nothing claimed without the facility in front, got %+v", claims)
			}
		})
	}
}

// UNIT is a facility designator here and a secondary unit designator in
// pkg/secondaryunit. Requiring the whole pattern is what keeps the two apart:
// a bare UNIT is left to pkg/secondaryunit, and only UNIT 2050 BOX 4190 is a
// military street line.
func TestClaimsUnitOnlyWithinTheFullPattern(t *testing.T) {
	if claims := military.Claims(token.Tokenize("UNIT 4B")); len(claims) != 0 {
		t.Errorf("a bare UNIT belongs to pkg/secondaryunit, got %+v", claims)
	}

	tokens := token.Tokenize("UNIT 2050 BOX 4190")
	claims := military.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim over the whole street line, got %+v", claims)
	}
	if claims[0].Start() != 0 || claims[0].End() != 4 {
		t.Errorf("claim covers [%d,%d), want [0,4)", claims[0].Start(), claims[0].End())
	}
	if len(claims[0].Parts) != 2 {
		t.Errorf("expected a street name and a primary number, got %+v", claims[0].Parts)
	}
}

// The facility and its box are one delivery address line. Join flattens the
// span with spaces, so without a bound "PSC 3\nBOX 4120 APO AE 09021" reaches
// NormalizeStreetLine as "PSC 3 BOX 4120" and matches across the break. The
// single token city and region claims are unaffected; those are one token each
// and cannot span anything.
func TestNoStreetLineClaimSpansALineBreak(t *testing.T) {
	for _, source := range []string{
		"PSC 3\nBOX 4120 APO AE 09021",
		"UNIT 2050\nBOX 4190 APO AP 96278",
		"CMR\n1234 BOX 5678",
	} {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range military.Claims(tokens) {
				for i := c.Start(); i < c.End(); i++ {
					if tokens[i].Line != tokens[c.Start()].Line {
						t.Errorf("claim over [%d,%d) spans a line break", c.Start(), c.End())
						break
					}
				}
			}
		})
	}
}
