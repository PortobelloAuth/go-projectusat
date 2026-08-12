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
			name: "facility designator opens the street line",
			in:   "PSC",
			want: []reading{
				{"PSC", claim.PartStreetName, claim.ConfidenceExact, "PSC"},
			},
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

// The standard's own example, claimed end to end. The numbers are not claimed:
// 3 is an assigned number and 4120 a box number only because of the tokens in
// front of them, which is positional knowledge this package does not have.
func TestClaimsFullMilitaryAddress(t *testing.T) {
	tokens := token.Tokenize("PSC 3 BOX 4120\nAPO AE 09021-0002")
	got := flatten(tokens, military.Claims(tokens))

	want := []reading{
		{"PSC", claim.PartStreetName, claim.ConfidenceExact, "PSC"},
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

// BOX is deliberately unclaimed. Rural route addresses use the same word for
// the same purpose, and it should not be owned by two packages at once.
func TestClaimsDoesNotClaimBox(t *testing.T) {
	tokens := token.Tokenize("BOX")

	if claims := military.Claims(tokens); len(claims) != 0 {
		t.Errorf("expected BOX to be left to a shared vocabulary, got %+v", claims)
	}
}

// UNIT is a facility designator here and a secondary unit designator in
// pkg/secondaryunit. Both readings are real; this package states its own
// without hedging.
func TestClaimsUnitDespiteSecondaryUnitAmbiguity(t *testing.T) {
	tokens := token.Tokenize("UNIT")
	claims := military.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if claims[0].Part != claim.PartStreetName || claims[0].Confidence != claim.ConfidenceExact {
		t.Errorf("got %+v, want the street name reading at exact confidence", claims[0])
	}
}
