package pobox_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
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
			name: "standard designator",
			in:   "PO BOX 11890",
			want: []reading{
				{"PO BOX", claim.PartStreetName, claim.ConfidenceExact, "PO BOX"},
				{"11890", claim.PartPrimaryNumber, claim.ConfidenceExact, "11890"},
			},
		},
		{
			name: "spelled out designator",
			in:   "POST OFFICE BOX 11890",
			want: []reading{
				{"POST OFFICE BOX", claim.PartStreetName, claim.ConfidenceExact, "PO BOX"},
				{"11890", claim.PartPrimaryNumber, claim.ConfidenceExact, "11890"},
			},
		},
		{
			name: "single token abbreviation",
			in:   "POB 11890",
			want: []reading{
				{"POB", claim.PartStreetName, claim.ConfidenceExact, "PO BOX"},
				{"11890", claim.PartPrimaryNumber, claim.ConfidenceExact, "11890"},
			},
		},
		{
			// The standard tells developers to rewrite these, but they are
			// ordinary words that mean other things in an address. With a box
			// number after them the match is good; it is not exclusive.
			name: "synonym is a good match but not an exclusive one",
			in:   "DRAWER 214",
			want: []reading{
				{"DRAWER", claim.PartStreetName, claim.ConfidenceStrong, "PO BOX"},
				{"214", claim.PartPrimaryNumber, claim.ConfidenceStrong, "214"},
			},
		},
		{
			// LOCKBOX is built from the same words as the reserved forms and is
			// not one of them.
			name: "lockbox is a synonym",
			in:   "LOCKBOX 214",
			want: []reading{
				{"LOCKBOX", claim.PartStreetName, claim.ConfidenceStrong, "PO BOX"},
				{"214", claim.PartPrimaryNumber, claim.ConfidenceStrong, "214"},
			},
		},
		{
			// A designator without a box number is not an address, so it is not
			// a claim either.
			name: "designator alone is not claimed",
			in:   "PO BOX",
			want: []reading{},
		},
		{
			name: "synonym alone is not claimed",
			in:   "DRAWER",
			want: []reading{},
		},
		{
			name: "not a post office box designator",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, pobox.Claims(tokens))

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

// FIRM CALLER contains CALLER, and both are recognized designators, so the
// same tokens read two ways. Both are offered and they overlap; the parser
// picks.
func TestClaimsOffersNestedDesignator(t *testing.T) {
	tokens := token.Tokenize("FIRM CALLER 214")
	claims := pobox.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected the long and short designator readings, got %+v", claims)
	}
	if claims[0].Start() != 0 || claims[1].Start() != 1 {
		t.Errorf("expected readings starting at FIRM and at CALLER, got %+v", claims)
	}
	if !claims[0].Overlaps(claims[1]) {
		t.Error("competing readings of the same tokens must overlap")
	}
}

// The designator and the box number are one reading. A parser that kept the
// number without the designator would hold a reading this package never
// offered.
func TestClaimIsIndivisible(t *testing.T) {
	tokens := token.Tokenize("PO BOX 11890")
	claims := pobox.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %+v", claims)
	}
	if claims[0].Start() != 0 || claims[0].End() != 3 {
		t.Errorf("claim covers [%d,%d), want [0,3)", claims[0].Start(), claims[0].End())
	}
	if len(claims[0].Parts) != 2 {
		t.Fatalf("expected a street name and a primary number, got %+v", claims[0].Parts)
	}
}

// Every recognized spelling normalizes to PO BOX, which the standard requires
// developers to rewrite the synonyms to.
func TestClaimsAlwaysValuePOBox(t *testing.T) {
	for _, designator := range []string{"PO BOX", "POST OFFICE BOX", "POB", "CALLER", "BIN", "LOCKBOX", "DRAWER"} {
		t.Run(designator, func(t *testing.T) {
			tokens := token.Tokenize(designator + " 214")
			claims := pobox.Claims(tokens)

			if len(claims) == 0 {
				t.Fatalf("expected %q to be claimed with a box number", designator)
			}
			if got := claims[0].Parts[0].Value; got != "PO BOX" {
				t.Errorf("street name value = %q, want PO BOX", got)
			}
		})
	}
}

// A post office box is a delivery address line, so the designator and the box
// number are on one line by definition. Nothing in Normalize can tell: Join
// flattens the span with spaces, so "PO BOX\nDENVER CO 80201" reaches it as
// "PO BOX DENVER" and matches. Left unbounded the city becomes the box number,
// at the highest confidence this package can give anything.
func TestNoClaimSpansALineBreak(t *testing.T) {
	for _, source := range []string{
		"PO BOX\nDENVER CO 80201",
		"POB\nSPRINGFIELD IL 62701",
		"POST OFFICE BOX\nDENVER CO 80201",
		"DRAWER\nDENVER CO 80201",
	} {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range pobox.Claims(tokens) {
				for i := c.Start(); i < c.End(); i++ {
					if tokens[i].Line != tokens[c.Start()].Line {
						t.Errorf("claim %v spans a line break", flatten(tokens, []claim.Claim{c}))
						break
					}
				}
			}
		})
	}
}
