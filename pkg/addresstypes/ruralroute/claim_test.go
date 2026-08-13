package ruralroute_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
)

// reading is a claim part flattened to the token text it covers, so cases read
// as "these words, claimed as this part, this strongly".
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
			name: "standard form",
			in:   "RR 4 BOX 125",
			want: []reading{
				{"RR 4", claim.PartStreetName, claim.ConfidenceExact, "RR 4"},
				{"BOX 125", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 125"},
			},
		},
		{
			// The standard's own examples of the spellings that normalize to RR.
			name: "spelled out designator",
			in:   "RURAL ROUTE 91 BOX A7",
			want: []reading{
				{"RURAL ROUTE 91", claim.PartStreetName, claim.ConfidenceExact, "RR 91"},
				{"BOX A7", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX A7"},
			},
		},
		{
			name: "rfd designator",
			in:   "RFD 82 BOX 12",
			want: []reading{
				{"RFD 82", claim.PartStreetName, claim.ConfidenceExact, "RR 82"},
				{"BOX 12", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 12"},
			},
		},
		{
			// RD borrows a spelling that belongs to ROAD. Requiring the whole
			// pattern is what makes it safe to claim at all: ROAD does not
			// appear as "RD 51 # 25".
			name: "rd designator with a number sign",
			in:   "RD 51 # 25",
			want: []reading{
				{"RD 51", claim.PartStreetName, claim.ConfidenceExact, "RR 51"},
				{"# 25", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 25"},
			},
		},
		{
			// The designator and the route number arrive as one token, so the
			// street name part is one token wide and the value is not.
			name: "glued designator and route number",
			in:   "RR03 BOX 98D",
			want: []reading{
				{"RR03", claim.PartStreetName, claim.ConfidenceExact, "RR 3"},
				{"BOX 98D", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 98D"},
			},
		},
		{
			// The box marker and its number arrive glued instead.
			name: "glued box marker and number",
			in:   "RFD ROUTE 4 #87A",
			want: []reading{
				{"RFD ROUTE 4", claim.PartStreetName, claim.ConfidenceExact, "RR 4"},
				{"#87A", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 87A"},
			},
		},
		{
			// A designator alone is a fragment of a pattern that did not match.
			name: "designator alone is not claimed",
			in:   "RR",
			want: []reading{},
		},
		{
			// The case that made RD dangerous when designators were claimed on
			// their own.
			name: "road is not a rural route",
			in:   "RD",
			want: []reading{},
		},
		{
			name: "box alone is not claimed",
			in:   "BOX 125",
			want: []reading{},
		},
		{
			name: "not a rural route",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, ruralroute.Claims(tokens))

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

// Normalize drops whatever follows the pattern, so a claim built from the first
// span that normalizes would swallow the street name after it. The standard's
// own example is the one that shows it.
func TestClaimsStopsAtTheEndOfThePattern(t *testing.T) {
	tokens := token.Tokenize("RR 2 BOX 18 BRYAN DAIRY RD")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if claims[0].Start() != 0 || claims[0].End() != 4 {
		t.Errorf("claim covers [%d,%d), want [0,4) — the pattern and nothing after it",
			claims[0].Start(), claims[0].End())
	}
}

// The route and the box are one reading. A parser that kept the box without
// the route would hold a reading this package never offered.
func TestClaimIsIndivisible(t *testing.T) {
	tokens := token.Tokenize("RR 4 BOX 125")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if len(claims[0].Parts) != 2 {
		t.Fatalf("expected a street name and a primary number, got %+v", claims[0].Parts)
	}
	if claims[0].Parts[0].Part != claim.PartStreetName ||
		claims[0].Parts[1].Part != claim.PartPrimaryNumber {
		t.Errorf("parts are %+v, want a street name then a primary number", claims[0].Parts)
	}
}

// A rural route does not have to start the line.
func TestClaimsWithinALongerLine(t *testing.T) {
	tokens := token.Tokenize("ATTN SHIPPING RR 4 BOX 125")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if claims[0].Start() != 2 {
		t.Errorf("claim starts at %d, want 2", claims[0].Start())
	}
}
