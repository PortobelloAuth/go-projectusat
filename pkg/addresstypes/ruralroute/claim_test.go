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

// The standard says a rural route line "SHOULD NOT allow additional
// designations, such as town or street names", so the trailing text in its own
// example is text that does not belong on the line rather than a street name.
// Both readings are offered: the pattern alone, and the whole line with the
// extra tokens absorbed into the primary number.
func TestClaimsOfferTheWholeStreetLine(t *testing.T) {
	tokens := token.Tokenize("RR 2 BOX 18 BRYAN DAIRY RD")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected the pattern and the whole line, got %+v", claims)
	}

	pattern, line := claims[0], claims[1]

	if pattern.Start() != 0 || pattern.End() != 4 {
		t.Errorf("pattern claim covers [%d,%d), want [0,4)", pattern.Start(), pattern.End())
	}
	if pattern.Confidence != claim.ConfidenceExact {
		t.Errorf("pattern claim confidence is %d, want %d", pattern.Confidence, claim.ConfidenceExact)
	}

	if line.Start() != 0 || line.End() != 7 {
		t.Errorf("street line claim covers [%d,%d), want [0,7)", line.Start(), line.End())
	}
	if line.Confidence != claim.ConfidenceLikely {
		t.Errorf("street line claim confidence is %d, want %d — where the line ends is the uncertain part",
			line.Confidence, claim.ConfidenceLikely)
	}
}

// Absorbing the trailing tokens does not change what the part says. The value
// is what Normalize produced for the pattern, in both readings.
func TestAbsorbedTokensDoNotChangeTheValue(t *testing.T) {
	tokens := token.Tokenize("RR 2 BOX 18 BRYAN DAIRY RD")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected the pattern and the whole line, got %+v", claims)
	}

	for i, c := range claims {
		if len(c.Parts) != 2 {
			t.Fatalf("claims[%d] has parts %+v, want a street name and a primary number", i, c.Parts)
		}
		if c.Parts[0].Value != "RR 2" || c.Parts[1].Value != "BOX 18" {
			t.Errorf("claims[%d] values are %q and %q, want \"RR 2\" and \"BOX 18\"",
				i, c.Parts[0].Value, c.Parts[1].Value)
		}
	}

	absorbed := claims[1].Parts[1]
	if token.Join(tokens[absorbed.Start:absorbed.End()]) != "BOX 18 BRYAN DAIRY RD" {
		t.Errorf("absorbing part covers %q, want the box and everything after it",
			token.Join(tokens[absorbed.Start:absorbed.End()]))
	}
}

// The street line has to end somewhere, and a token slice only knows which
// line each token is on. A city and region on the next line are not absorbed.
func TestAbsorptionStopsAtTheEndOfTheLine(t *testing.T) {
	tokens := token.Tokenize("RR 2 BOX 18 BRYAN DAIRY RD\nSALT LAKE CITY UT")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected the pattern and the whole line, got %+v", claims)
	}
	if claims[1].End() != 7 {
		t.Errorf("street line claim ends at %d, want 7 — the last token of the first line",
			claims[1].End())
	}
}

// Where the pattern already is the whole line there is nothing to absorb, and
// offering a second identical claim would invent a competing reading.
func TestNoStreetLineClaimWhenThePatternFillsTheLine(t *testing.T) {
	tokens := token.Tokenize("RR 4 BOX 125")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
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

// The route and its box are one delivery address line. Join flattens the span
// with spaces, so without a bound "RR 2\nBOX 18 BRYAN OH 43506" reaches
// Normalize as "RR 2 BOX 18 ..." and matches across the break. Absorption was
// already bounded to the line; this is the other end of the same rule.
func TestNoClaimSpansALineBreak(t *testing.T) {
	for _, source := range []string{
		"RR 2\nBOX 18 BRYAN OH 43506",
		"RFD ROUTE 4\nBOX 125 BRYAN OH 43506",
		"RR\n2 BOX 18",
	} {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range ruralroute.Claims(tokens) {
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

// A highway contract route is claimed exactly as a rural route is: the whole
// pattern, split into the route as a street name and the box as a primary
// address number. Nothing in claim.go names a designator, so this is the test
// that the split is driven by what Normalize emitted and not by the letters RR.
func TestAHighwayContractRouteIsClaimedLikeARuralRoute(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []reading
	}{
		{
			name: "standard form",
			in:   "HC 4 BOX 125",
			want: []reading{
				{"HC 4", claim.PartStreetName, claim.ConfidenceExact, "HC 4"},
				{"BOX 125", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 125"},
			},
		},
		{
			// The longest designator the vocabulary accepts. It is three
			// tokens, which is what maxSpan had to grow to reach.
			name: "spelled out designator",
			in:   "HIGHWAY CONTRACT ROUTE 4 BOX 125",
			want: []reading{
				{"HIGHWAY CONTRACT ROUTE 4", claim.PartStreetName, claim.ConfidenceExact, "HC 4"},
				{"BOX 125", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 125"},
			},
		},
		{
			name: "hcr designator",
			in:   "HCR 4 BOX 125",
			want: []reading{
				{"HCR 4", claim.PartStreetName, claim.ConfidenceExact, "HC 4"},
				{"BOX 125", claim.PartPrimaryNumber, claim.ConfidenceExact, "BOX 125"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, ruralroute.Claims(tokens))

			if len(got) != len(tc.want) {
				t.Fatalf("Claims(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("reading %d = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// The longest span the vocabulary accepts: a three word designator and a number
// marker on each half, eight tokens in all. maxSpan had to grow from seven to
// reach it, and this is what would fail if it shrank back.
//
// Only the extent and the values are asserted. Where the claim divides into its
// street name and primary number parts is wrong here — the marker on the route
// half is mistaken for the box boundary — but that is true of "RR NO 2 BOX 18"
// just as much, so it is a bug of its own and not this vocabulary's. See #73.
func TestTheLongestHighwayContractRouteIsClaimedWhole(t *testing.T) {
	tokens := token.Tokenize("HIGHWAY CONTRACT ROUTE NO 4 BOX NO 125")
	claims := ruralroute.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if claims[0].Start() != 0 || claims[0].End() != 8 {
		t.Errorf("claim covers [%d,%d), want [0,8)", claims[0].Start(), claims[0].End())
	}

	want := map[claim.Part]string{
		claim.PartStreetName:    "HC 4",
		claim.PartPrimaryNumber: "BOX 125",
	}
	for _, p := range claims[0].Parts {
		if p.Value != want[p.Part] {
			t.Errorf("%s = %q, want %q", p.Part, p.Value, want[p.Part])
		}
	}
}
