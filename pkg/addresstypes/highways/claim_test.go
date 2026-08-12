package highways_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/highways"
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
			name: "abbreviated highway is spelled out",
			in:   "HWY 64",
			want: []reading{
				{"HWY 64", claim.PartStreetName, claim.ConfidenceStrong, "HIGHWAY 64"},
			},
		},
		{
			// The glued interstate form is a single token.
			name: "glued interstate",
			in:   "I10",
			want: []reading{
				{"I10", claim.PartStreetName, claim.ConfidenceStrong, "INTERSTATE 10"},
			},
		},
		{
			name: "farm to market keeps its abbreviation",
			in:   "FM 187",
			want: []reading{
				{"FM 187", claim.PartStreetName, claim.ConfidenceStrong, "FM 187"},
			},
		},
		{
			// An ordinary street name is not a highway. NormalizeStreetName
			// would return MAIN STREET here; that is formatting, not evidence.
			name: "ordinary street name is not claimed",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, highways.Claims(tokens))

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

// A state name in front of a highway is abbreviated, per the standard's rule
// that a state used as a portion of the primary street name takes the
// two-letter form. The shorter readings inside it are real highway names too.
func TestClaimsSpansIncludeAndExcludeTheStatePrefix(t *testing.T) {
	tokens := token.Tokenize("CALIFORNIA COUNTY ROAD 555")
	claims := highways.Claims(tokens)

	if len(claims) < 2 {
		t.Fatalf("expected the full span and the span without the state, got %+v", claims)
	}

	full := claims[0]
	if full.Length != 4 || full.Value != "CA COUNTY ROAD 555" {
		t.Errorf("got %+v, want the full span abbreviated to CA COUNTY ROAD 555", full)
	}

	withoutState := claims[1]
	if withoutState.Value != "COUNTY ROAD 555" {
		t.Errorf("got %+v, want COUNTY ROAD 555", withoutState)
	}
	if !full.Overlaps(withoutState) {
		t.Error("nested spans must overlap so the parser can choose between them")
	}
}

// Longest first, so a caller walking the slice sees the most complete reading
// of a given start position before the shorter ones.
func TestClaimsAreOrderedLongestFirst(t *testing.T) {
	tokens := token.Tokenize("HWY 11 BYPASS")
	claims := highways.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected two readings, got %+v", claims)
	}
	if claims[0].Length <= claims[1].Length {
		t.Errorf("expected longest first, got lengths %d then %d", claims[0].Length, claims[1].Length)
	}
	if claims[0].Value != "HIGHWAY 11 BYP" {
		t.Errorf("got %q, want the bypass abbreviated as a suffix", claims[0].Value)
	}
}

// A highway name sits in street-name position inside a full address, not at
// index 0, and must not swallow the primary number in front of it.
func TestClaimsWithinAFullStreetLine(t *testing.T) {
	tokens := token.Tokenize("1234 HWY 64")
	claims := highways.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected only the highway claimed, got %+v", claims)
	}
	if claims[0].Start != 1 {
		t.Errorf("claim %+v should start after the primary number", claims[0])
	}
}
