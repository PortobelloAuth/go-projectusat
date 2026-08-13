package secondaryunit_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
)

// reading is a claim part flattened to the token text it covers, so cases read
// as "these words, claimed as this part, this strongly".
type reading struct {
	text       string
	part       claim.Part
	confidence claim.Confidence
	value      string
}

// flatten expands each claim into one reading per part, so a numbered
// designator and its number appear as two consecutive rows. That they belong
// to one claim is what TestNumberedDesignatorIsOneClaim pins.
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
			name: "standard abbreviation is exact",
			in:   "APT 4B",
			want: []reading{
				{"APT", claim.PartSecondaryDesignator, claim.ConfidenceExact, "APT"},
				{"4B", claim.PartSecondaryNumber, claim.ConfidenceExact, "4B"},
			},
		},
		{
			name: "spelled out designator is strong",
			in:   "APARTMENT 4B",
			want: []reading{
				{"APARTMENT", claim.PartSecondaryDesignator, claim.ConfidenceStrong, "APT"},
				{"4B", claim.PartSecondaryNumber, claim.ConfidenceStrong, "4B"},
			},
		},
		{
			// A unit number need not contain a digit at all.
			name: "single letter unit number",
			in:   "UNIT A",
			want: []reading{
				{"UNIT", claim.PartSecondaryDesignator, claim.ConfidenceExact, "UNIT"},
				{"A", claim.PartSecondaryNumber, claim.ConfidenceExact, "A"},
			},
		},
		{
			// The standard's designator for a numbered unit of unknown type.
			name: "hash designator",
			in:   "# 4B",
			want: []reading{
				{"#", claim.PartSecondaryDesignator, claim.ConfidenceExact, "#"},
				{"4B", claim.PartSecondaryNumber, claim.ConfidenceExact, "4B"},
			},
		},
		{
			// An unnumbered designator standing alone is the whole pattern.
			name: "unnumbered designator",
			in:   "BSMT",
			want: []reading{
				{"BSMT", claim.PartSecondaryDesignator, claim.ConfidenceExact, "BSMT"},
			},
		},
		{
			// A numbered designator without a number is a fragment of a pattern
			// that did not match, not weak evidence of a secondary unit.
			name: "numbered designator alone is not claimed",
			in:   "APT",
			want: []reading{},
		},
		{
			// Nothing rules out a unit named WEST, so the reading is offered.
			// It is contested rather than refused: KEY WEST is a city far more
			// often than it is unit WEST of a key.
			name: "designator followed by an ordinary word is contested",
			in:   "KEY WEST",
			want: []reading{
				{"KEY", claim.PartSecondaryDesignator, claim.ConfidenceLikely, "KEY"},
				{"WEST", claim.PartSecondaryNumber, claim.ConfidenceLikely, "WEST"},
			},
		},
		{
			// A unit number may be several letters. PENTHOUSE is also an
			// unnumbered designator in its own right, so the same token is
			// claimed twice over and the parser is left both readings.
			name: "multi letter unit number is contested",
			in:   "APT PENTHOUSE",
			want: []reading{
				{"APT", claim.PartSecondaryDesignator, claim.ConfidenceLikely, "APT"},
				{"PENTHOUSE", claim.PartSecondaryNumber, claim.ConfidenceLikely, "PENTHOUSE"},
				{"PENTHOUSE", claim.PartSecondaryDesignator, claim.ConfidenceStrong, "PH"},
			},
		},
		{
			name: "not a designator",
			in:   "MAIN",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, secondaryunit.Claims(tokens))

			if len(got) != len(tc.want) {
				t.Fatalf("Claims(%q) returned %d readings, want %d: %+v", tc.in, len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("Claims(%q)[%d] = %+v, want %+v", tc.in, i, got[i], w)
				}
			}
		})
	}
}

// The designator and its number are one reading. A parser that could accept
// APT without 4B would hold a reading this package never offered, which is
// what grouping the parts prevents.
func TestNumberedDesignatorIsOneClaim(t *testing.T) {
	tokens := token.Tokenize("APT 4B")
	claims := secondaryunit.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim covering both tokens, got %+v", claims)
	}
	if len(claims[0].Parts) != 2 {
		t.Fatalf("expected a designator and a number, got %+v", claims[0].Parts)
	}
	if claims[0].Start() != 0 || claims[0].End() != 2 {
		t.Errorf("claim covers [%d,%d), want [0,2)", claims[0].Start(), claims[0].End())
	}
}

// The unnumbered designators are the ones that are ordinary English words —
// FRONT, REAR, SIDE, LOWER — so the pattern that would otherwise vouch for
// them is exactly the one they do not have. Nothing here fixes that; it is a
// reminder that this package's unnumbered claims carry the most risk.
func TestUnnumberedDesignatorStandsAlone(t *testing.T) {
	for _, in := range []string{"BSMT", "LBBY", "REAR", "PH"} {
		t.Run(in, func(t *testing.T) {
			tokens := token.Tokenize(in)
			claims := secondaryunit.Claims(tokens)

			if len(claims) != 1 || len(claims[0].Parts) != 1 {
				t.Fatalf("expected one single part claim, got %+v", claims)
			}
			if claims[0].Parts[0].Part != claim.PartSecondaryDesignator {
				t.Errorf("claimed %v, want a secondary designator", claims[0].Parts[0].Part)
			}
		})
	}
}
