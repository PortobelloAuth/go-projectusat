package ruralroute_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
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
			name: "standard designator",
			in:   "RR",
			want: []reading{
				{"RR", claim.PartStreetName, claim.ConfidenceExact, "RR"},
			},
		},
		{
			name: "spelled out designator is a two token span",
			in:   "RURAL ROUTE",
			want: []reading{
				{"RURAL ROUTE", claim.PartStreetName, claim.ConfidenceExact, "RR"},
			},
		},
		{
			name: "rural free delivery",
			in:   "RFD",
			want: []reading{
				{"RFD", claim.PartStreetName, claim.ConfidenceExact, "RR"},
			},
		},
		{
			// RD is also the standard abbreviation for ROAD.
			name: "borrowed spelling is rated lower",
			in:   "RD",
			want: []reading{
				{"RD", claim.PartStreetName, claim.ConfidenceLikely, "RR"},
			},
		},
		{
			name: "not a rural route designator",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, ruralroute.Claims(tokens))

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

// The standard's own example. Only the designator is claimed: the route number
// and the box number depend on what precedes them, and BOX belongs to
// pkg/secondaryunit.
func TestClaimsFullRuralRoute(t *testing.T) {
	tokens := token.Tokenize("RR 2 BOX 18")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected the designator alone, got %+v", claims)
	}
	if claims[0].Start != 0 || claims[0].Length != 1 {
		t.Errorf("claim %+v should cover only RR", claims[0])
	}
}

// Every recognized spelling normalizes to RR, which is the only form the
// standard permits in a patient record.
func TestClaimsAlwaysValueRR(t *testing.T) {
	for _, in := range []string{"RR", "RFD", "RD", "RURAL ROUTE", "RFD ROUTE"} {
		t.Run(in, func(t *testing.T) {
			tokens := token.Tokenize(in)
			claims := ruralroute.Claims(tokens)

			if len(claims) == 0 {
				t.Fatalf("expected %q to be claimed", in)
			}
			if claims[0].Value != "RR" {
				t.Errorf("Value = %q, want RR", claims[0].Value)
			}
		})
	}
}
