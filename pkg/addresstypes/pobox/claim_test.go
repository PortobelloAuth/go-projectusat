package pobox_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
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
			name: "standard designator is a two token span",
			in:   "PO BOX",
			want: []reading{
				{"PO BOX", claim.PartStreetName, claim.ConfidenceExact, "PO BOX"},
			},
		},
		{
			name: "spelled out designator is a three token span",
			in:   "POST OFFICE BOX",
			want: []reading{
				{"POST OFFICE BOX", claim.PartStreetName, claim.ConfidenceExact, "PO BOX"},
			},
		},
		{
			name: "single token abbreviation",
			in:   "POB",
			want: []reading{
				{"POB", claim.PartStreetName, claim.ConfidenceExact, "PO BOX"},
			},
		},
		{
			// The standard tells developers to rewrite these, but they are
			// ordinary words that mean other things in an address.
			name: "synonym is real but contested",
			in:   "DRAWER",
			want: []reading{
				{"DRAWER", claim.PartStreetName, claim.ConfidenceStrong, "PO BOX"},
			},
		},
		{
			// LOCKBOX contains BOX but is a synonym, not a reserved spelling.
			name: "lockbox is a synonym",
			in:   "LOCKBOX",
			want: []reading{
				{"LOCKBOX", claim.PartStreetName, claim.ConfidenceStrong, "PO BOX"},
			},
		},
		{
			name: "firm caller is a two token synonym",
			in:   "FIRM CALLER",
			want: []reading{
				{"FIRM CALLER", claim.PartStreetName, claim.ConfidenceStrong, "PO BOX"},
				{"CALLER", claim.PartStreetName, claim.ConfidenceStrong, "PO BOX"},
			},
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

// The standard's own example. The box number is not claimed: it means
// something only because of the designator in front of it.
func TestClaimsFullPOBox(t *testing.T) {
	tokens := token.Tokenize("POST OFFICE BOX 11890")
	claims := pobox.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected the designator alone, got %+v", claims)
	}
	if claims[0].Length != 3 || claims[0].End() != 3 {
		t.Errorf("claim %+v should cover the designator and stop before the number", claims[0])
	}
}

// This package claims the span PO BOX. pkg/secondaryunit separately claims the
// BOX token inside it as a secondary designator, and both readings are real:
// the same tokens could be a PO box or a street address with a box unit. This
// package must claim the longer span and say nothing about the shorter one, or
// the parser has nothing to weigh.
func TestClaimsSpanNotTheInnerBoxToken(t *testing.T) {
	tokens := token.Tokenize("PO BOX 11890")
	claims := pobox.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected exactly one claim, got %+v", claims)
	}
	if claims[0].Start != 0 || claims[0].Length != 2 {
		t.Errorf("claim %+v should cover PO BOX, not BOX alone", claims[0])
	}
}

// Every recognized spelling normalizes to PO BOX, which the standard requires
// developers to rewrite the synonyms to.
func TestClaimsAlwaysValuePOBox(t *testing.T) {
	for _, in := range []string{"PO BOX", "POST OFFICE BOX", "POB", "CALLER", "BIN", "LOCKBOX", "DRAWER"} {
		t.Run(in, func(t *testing.T) {
			tokens := token.Tokenize(in)
			claims := pobox.Claims(tokens)

			if len(claims) == 0 {
				t.Fatalf("expected %q to be claimed", in)
			}
			if claims[0].Value != "PO BOX" {
				t.Errorf("Value = %q, want PO BOX", claims[0].Value)
			}
		})
	}
}
