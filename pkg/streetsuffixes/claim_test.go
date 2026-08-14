package streetsuffixes_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
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
			name: "standard abbreviation is exact",
			in:   "AVE",
			want: []reading{
				{"AVE", claim.PartStreetSuffix, claim.ConfidenceExact, "AVE"},
			},
		},
		{
			name: "spelled out name is strong",
			in:   "AVENUE",
			want: []reading{
				{"AVENUE", claim.PartStreetSuffix, claim.ConfidenceStrong, "AVE"},
			},
		},
		{
			name: "alternate spelling is strong",
			in:   "AVEN",
			want: []reading{
				{"AVEN", claim.PartStreetSuffix, claim.ConfidenceStrong, "AVE"},
			},
		},
		{
			// The abbreviation and the spelled out name are the same word, so
			// the exact rule applies.
			name: "suffix whose abbreviation is its own name",
			in:   "PARK",
			want: []reading{
				{"PARK", claim.PartStreetSuffix, claim.ConfidenceExact, "PARK"},
			},
		},
		{
			name: "trailing punctuation does not prevent a claim",
			in:   "ST.",
			want: []reading{
				{"ST.", claim.PartStreetSuffix, claim.ConfidenceExact, "ST"},
			},
		},
		{
			name: "not a suffix",
			in:   "MAIN",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, streetsuffixes.Claims(tokens))

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

// A misspelling is not evidence of the word that was meant. Info can fuzzy
// match on request; Claims deliberately does not use it.
func TestClaimsDoesNotFuzzyMatch(t *testing.T) {
	tokens := token.Tokenize("CRESENT")

	if claims := streetsuffixes.Claims(tokens); len(claims) != 0 {
		t.Errorf("expected no claim for a misspelling, got %+v", claims)
	}
}

// PARK AVE has two suffixes in it, and only the second one is the suffix. This
// package cannot tell them apart and must offer both.
func TestClaimsOffersEverySuffixInTheInput(t *testing.T) {
	tokens := token.Tokenize("PARK AVE")
	claims := streetsuffixes.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected both tokens claimed, got %+v", claims)
	}
	if claims[0].Overlaps(claims[1]) {
		t.Error("claims over different tokens should not overlap")
	}
}

// An ordinal street name must produce no suffix claim. Before this was fixed,
// 500 1ST AVE offered two claims that the parser could not tell apart: both
// single tokens, both ConfidenceExact, and the wrong one first.
func TestClaimsIgnoresOrdinalStreetNames(t *testing.T) {
	tokens := token.Tokenize("500 1ST AVE")
	claims := streetsuffixes.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected only AVE claimed, got %+v", claims)
	}
	if claims[0].Parts[0].Start != 2 {
		t.Errorf("expected the claim on AVE at index 2, got %+v", claims[0])
	}
}

// ST is STREET here and SAINT elsewhere. Both readings are real; this package
// owns only one of them and states it without hedging.
func TestClaimsStreetAbbreviationDespiteSaintAmbiguity(t *testing.T) {
	tokens := token.Tokenize("ST")
	claims := streetsuffixes.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if claims[0].Parts[0].Value != "ST" || claims[0].Confidence != claim.ConfidenceExact {
		t.Errorf("got %+v, want the STREET reading at exact confidence", claims[0])
	}
}
