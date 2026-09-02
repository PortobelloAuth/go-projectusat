package directionals_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
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
			name: "abbreviation is exact, claimed on both sides",
			in:   "W",
			want: []reading{
				{"W", claim.PartPredirectional, claim.ConfidenceExact, "W"},
				{"W", claim.PartPostdirectional, claim.ConfidenceExact, "W"},
			},
		},
		{
			// NORTH SALT LAKE and SOUTH BEND are why a spelled out direction
			// is rated below its abbreviation.
			name: "spelled out direction is strong, not exact",
			in:   "NORTH",
			want: []reading{
				{"NORTH", claim.PartPredirectional, claim.ConfidenceStrong, "N"},
				{"NORTH", claim.PartPostdirectional, claim.ConfidenceStrong, "N"},
			},
		},
		{
			name: "compound direction",
			in:   "SOUTHWEST",
			want: []reading{
				{"SOUTHWEST", claim.PartPredirectional, claim.ConfidenceStrong, "SW"},
				{"SOUTHWEST", claim.PartPostdirectional, claim.ConfidenceStrong, "SW"},
			},
		},
		{
			// The compound reading and the two separate readings are both
			// offered. The compound ranks lower: the standard spells it as one
			// word, so two tokens is the less expected form.
			name: "compound spelled as two tokens",
			in:   "NORTH EAST",
			want: []reading{
				{"NORTH EAST", claim.PartPredirectional, claim.ConfidenceLikely, "NE"},
				{"NORTH EAST", claim.PartPostdirectional, claim.ConfidenceLikely, "NE"},
				{"NORTH", claim.PartPredirectional, claim.ConfidenceStrong, "N"},
				{"NORTH", claim.PartPostdirectional, claim.ConfidenceStrong, "N"},
				{"EAST", claim.PartPredirectional, claim.ConfidenceStrong, "E"},
				{"EAST", claim.PartPostdirectional, claim.ConfidenceStrong, "E"},
			},
		},
		{
			// The standard calls these out as invalid compounds, and the rule
			// here reaches the same answer without listing them: EW is not a
			// direction, so there is nothing to claim over the pair.
			name: "opposed directions are not a compound",
			in:   "EAST WEST",
			want: []reading{
				{"EAST", claim.PartPredirectional, claim.ConfidenceStrong, "E"},
				{"EAST", claim.PartPostdirectional, claim.ConfidenceStrong, "E"},
				{"WEST", claim.PartPredirectional, claim.ConfidenceStrong, "W"},
				{"WEST", claim.PartPostdirectional, claim.ConfidenceStrong, "W"},
			},
		},
		{
			// NS is not a direction either.
			name: "north south is not a compound",
			in:   "NORTH SOUTH",
			want: []reading{
				{"NORTH", claim.PartPredirectional, claim.ConfidenceStrong, "N"},
				{"NORTH", claim.PartPostdirectional, claim.ConfidenceStrong, "N"},
				{"SOUTH", claim.PartPredirectional, claim.ConfidenceStrong, "S"},
				{"SOUTH", claim.PartPostdirectional, claim.ConfidenceStrong, "S"},
			},
		},
		{
			// The pair is only a compound in the order the standard writes it.
			// EN is not a direction; NE is.
			name: "compound in the wrong order is not a compound",
			in:   "EAST NORTH",
			want: []reading{
				{"EAST", claim.PartPredirectional, claim.ConfidenceStrong, "E"},
				{"EAST", claim.PartPostdirectional, claim.ConfidenceStrong, "E"},
				{"NORTH", claim.PartPredirectional, claim.ConfidenceStrong, "N"},
				{"NORTH", claim.PartPostdirectional, claim.ConfidenceStrong, "N"},
			},
		},
		{
			// Abbreviations concatenate the same way, and the compound reading
			// of two of them stays below a single-token abbreviation.
			name: "compound spelled as two abbreviations",
			in:   "N W",
			want: []reading{
				{"N W", claim.PartPredirectional, claim.ConfidenceStrong, "NW"},
				{"N W", claim.PartPostdirectional, claim.ConfidenceStrong, "NW"},
				{"N", claim.PartPredirectional, claim.ConfidenceExact, "N"},
				{"N", claim.PartPostdirectional, claim.ConfidenceExact, "N"},
				{"W", claim.PartPredirectional, claim.ConfidenceExact, "W"},
				{"W", claim.PartPostdirectional, claim.ConfidenceExact, "W"},
			},
		},
		{
			name: "no directional present",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, directionals.Claims(tokens))

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

// The grid address Aaron wrote into parser_test.go. Both W and S are genuinely
// ambiguous in isolation, and this package must not pretend otherwise: the W is
// a predirectional and the S is a postdirectional, but nothing here can tell.
func TestClaimsDoesNotResolveGridAddress(t *testing.T) {
	tokens := token.Tokenize("3253 W 9200 S")
	claims := directionals.Claims(tokens)

	if len(claims) != 4 {
		t.Fatalf("expected both directionals claimed on both sides, got %+v", claims)
	}

	for i := 0; i < len(claims); i += 2 {
		pre, post := claims[i], claims[i+1]
		if pre.Parts[0].Part != claim.PartPredirectional || post.Parts[0].Part != claim.PartPostdirectional {
			t.Errorf("claims %d,%d are not a pre/post pair: %+v %+v", i, i+1, pre, post)
		}
		if !pre.Overlaps(post) {
			t.Errorf("competing readings of the same token must overlap: %+v %+v", pre, post)
		}
		if pre.Confidence != post.Confidence {
			t.Errorf("neither side should be favored in isolation: %d vs %d", pre.Confidence, post.Confidence)
		}
	}
}

// The W Palm Beach case: the directional reading is real and this package is
// right to offer it. It is wrong, but only something that knows about cities
// can say so.
func TestClaimsOffersTheWrongReadingWhenItCannotKnow(t *testing.T) {
	tokens := token.Tokenize("123 MAIN ST W PALM BEACH FL")
	claims := directionals.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected the W claimed on both sides and nothing else, got %+v", claims)
	}
	if got := token.Join(tokens[claims[0].Start():claims[0].End()]); got != "W" {
		t.Errorf("claimed %q, want %q", got, "W")
	}
}

// A compound directional hard wrapped across two lines is not one compound.
// abbreviateSpan reads a span without regard to where the lines fall, so
// without a bound the last token of the delivery address line and the first of
// the last line combine: "123 FAKE ST NORTH\nEAST ORANGE NJ 07017" claims
// NORTH EAST as NE and takes the first word of the city with it. See
// token.LineEnd.
func TestNoClaimSpansALineBreak(t *testing.T) {
	for _, source := range []string{
		"123 FAKE ST NORTH\nEAST ORANGE NJ 07017",
		"123 NORTH\nWEST ST DENVER CO 80201",
	} {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range directionals.Claims(tokens) {
				if tokens[c.End()-1].Line != tokens[c.Start()].Line {
					t.Errorf("claim over %q spans a line break",
						token.Join(tokens[c.Start():c.End()]))
				}
			}
		})
	}
}

// The bound must not cost the reading it was added to protect: the same
// compound on one line is still claimed as one directional.
func TestACompoundOnOneLineIsStillClaimed(t *testing.T) {
	tokens := token.Tokenize("123 NORTH EAST ST\nDENVER CO 80201")

	var found bool
	for _, c := range directionals.Claims(tokens) {
		if token.Join(tokens[c.Start():c.End()]) == "NORTH EAST" {
			found = true
		}
	}
	if !found {
		t.Error("expected NORTH EAST on one line to still be claimed as a compound")
	}
}
