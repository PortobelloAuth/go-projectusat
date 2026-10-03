package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
)

// reading is a Claim flattened to the token text it covers, mirroring
// pkg/directionals/claim_test.go's own helper of the same name: these cases
// read as "these words, claimed as this part, this strongly."
type reading struct {
	text       string
	part       claim.Part
	confidence claim.Confidence
	value      string
}

func flattenDirectionals(tokens []token.Token, claims []claim.Claim) []reading {
	out := make([]reading, 0, len(claims))
	for _, c := range claims {
		for _, p := range c.Parts {
			if p.Part != claim.PartPredirectional && p.Part != claim.PartPostdirectional {
				continue
			}
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

func hasReading(got []reading, want reading) bool {
	for _, r := range got {
		if r == want {
			return true
		}
	}

	return false
}

// TestSpanishDirectionalClaims checks puertorico.Claims's Spanish-directional
// readings (go-projectusat#154), the same way pkg/directionals/claim_test.go
// checks the English vocabulary's. The data is Spanish-only — addresstables'
// Spanish() rows — so the abbreviation asserted is always the Spanish one
// (NOROESTE -> NO), never the English directional the Spanish row happens to
// point at.
func TestSpanishDirectionalClaims(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []reading
	}{
		{
			// The abbreviation is exact, claimed on both sides — the shape
			// p.26's own worked examples take: "1510 CALLE 3 NO" never
			// spells NOROESTE out.
			name: "abbreviation is exact, claimed on both sides",
			in:   "NO",
			want: []reading{
				{"NO", claim.PartPredirectional, claim.ConfidenceExact, "NO"},
				{"NO", claim.PartPostdirectional, claim.ConfidenceExact, "NO"},
			},
		},
		{
			// This is the actual gap #154 opened: NormalizeStreetLine passes
			// a literal "NO" through untouched, so the fixed-point spec case
			// never exercises recognition at all. A spelled-out word is the
			// only input that does.
			name: "spelled out direction is strong, not exact, and abbreviates to the Spanish form",
			in:   "NOROESTE",
			want: []reading{
				{"NOROESTE", claim.PartPredirectional, claim.ConfidenceStrong, "NO"},
				{"NOROESTE", claim.PartPostdirectional, claim.ConfidenceStrong, "NO"},
			},
		},
		{
			name: "a plain West spelling abbreviates to its own Spanish form, not the English one",
			in:   "OESTE",
			want: []reading{
				{"OESTE", claim.PartPredirectional, claim.ConfidenceStrong, "O"},
				{"OESTE", claim.PartPostdirectional, claim.ConfidenceStrong, "O"},
			},
		},
		{
			// A compound written as two words, both spelled out: NORTE ESTE
			// abbreviates to NE the same way NORTH EAST does in English.
			name: "compound direction, both words spelled out",
			in:   "NORTE ESTE",
			want: []reading{
				{"NORTE ESTE", claim.PartPredirectional, claim.ConfidenceLikely, "NE"},
				{"NORTE ESTE", claim.PartPostdirectional, claim.ConfidenceLikely, "NE"},
			},
		},
		{
			name: "an ordinary word is not a directional",
			in:   "CALLE",
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flattenDirectionals(tokens, puertorico.Claims(tokens))

			if len(tc.want) == 0 {
				if len(got) != 0 {
					t.Fatalf("Claims(%q) = %+v, want no directional readings", tc.in, got)
				}
				return
			}

			for _, w := range tc.want {
				if !hasReading(got, w) {
					t.Errorf("Claims(%q) = %+v, missing reading %+v", tc.in, got, w)
				}
			}
		})
	}
}

// TestSpanishDirectionalClaimsBothParts pins the structural requirement
// go-projectusat#154 asked for explicitly: a Spanish directional, like its
// English counterpart in pkg/directionals, claims both PartPredirectional
// and PartPostdirectional over the same span, rather than committing to a
// position — the parser decides which side of the street name the token
// actually fell on, not this package.
func TestSpanishDirectionalClaimsBothParts(t *testing.T) {
	tokens := token.Tokenize("NOROESTE")
	got := flattenDirectionals(tokens, puertorico.Claims(tokens))

	var pre, post bool
	for _, r := range got {
		if r.part == claim.PartPredirectional {
			pre = true
		}
		if r.part == claim.PartPostdirectional {
			post = true
		}
	}

	if !pre || !post {
		t.Errorf("Claims(%q) = %+v, want both a PartPredirectional and a PartPostdirectional reading", "NOROESTE", got)
	}
}
