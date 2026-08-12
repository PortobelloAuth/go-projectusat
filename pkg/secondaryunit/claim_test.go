package secondaryunit_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
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
		out = append(out, reading{
			token.Join(tokens[c.Start:c.End()]),
			c.Part,
			c.Confidence,
			c.Value,
		})
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
			in:   "APT",
			want: []reading{
				{"APT", claim.PartSecondaryDesignator, claim.ConfidenceExact, "APT"},
			},
		},
		{
			name: "spelled out designator is strong",
			in:   "APARTMENT",
			want: []reading{
				{"APARTMENT", claim.PartSecondaryDesignator, claim.ConfidenceStrong, "APT"},
			},
		},
		{
			// The standard's designator for a numbered unit of unknown type.
			name: "hash designator",
			in:   "#",
			want: []reading{
				{"#", claim.PartSecondaryDesignator, claim.ConfidenceExact, "#"},
			},
		},
		{
			// An unnumbered designator is still a designator.
			name: "unnumbered designator",
			in:   "BSMT",
			want: []reading{
				{"BSMT", claim.PartSecondaryDesignator, claim.ConfidenceExact, "BSMT"},
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

// The number is not claimed. 4B is a secondary number only because APT precedes
// it, and that is positional knowledge this package does not have.
func TestClaimsDoesNotClaimTheNumber(t *testing.T) {
	tokens := token.Tokenize("APT 4B")
	claims := secondaryunit.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected the designator alone to be claimed, got %+v", claims)
	}
	if claims[0].Start != 0 || claims[0].Length != 1 {
		t.Errorf("claim %+v should cover only APT", claims[0])
	}
}

// A parser pairing a designator with its number needs to know whether the
// standard expects one. That knowledge stays on SecondaryUnit rather than
// becoming a second lookup function beside Claims.
func TestNumberedIsReachableFromAClaim(t *testing.T) {
	cases := map[string]bool{
		"APT":  true,
		"FL":   true,
		"BSMT": false,
		"PH":   false,
	}

	for designator, want := range cases {
		t.Run(designator, func(t *testing.T) {
			tokens := token.Tokenize(designator)
			claims := secondaryunit.Claims(tokens)
			if len(claims) != 1 {
				t.Fatalf("expected one claim for %q, got %+v", designator, claims)
			}

			info, err := secondaryunit.Info(claims[0].Value)
			if err != nil {
				t.Fatalf("Info(%q) failed: %v", claims[0].Value, err)
			}
			if info.Numbered != want {
				t.Errorf("Numbered = %v, want %v", info.Numbered, want)
			}
		})
	}
}
