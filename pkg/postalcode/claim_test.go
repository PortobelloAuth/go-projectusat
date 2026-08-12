package postalcode_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
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
			// A five digit number is a well formed ZIP and an ordinary primary
			// number. Nothing in this package can tell those apart.
			name: "bare ZIP is weak",
			in:   "84088",
			want: []reading{
				{"84088", claim.PartPostal, claim.ConfidenceWeak, "84088"},
			},
		},
		{
			name: "ZIP+4 written with a hyphen",
			in:   "84088-1234",
			want: []reading{
				{"84088-1234", claim.PartPostal, claim.ConfidenceStrong, "84088-1234"},
			},
		},
		{
			name: "ZIP+4 written solid",
			in:   "840881234",
			want: []reading{
				{"840881234", claim.PartPostal, claim.ConfidenceStrong, "84088-1234"},
			},
		},
		{
			name: "canadian postal code as one token",
			in:   "M5V3L9",
			want: []reading{
				{"M5V3L9", claim.PartPostal, claim.ConfidenceStrong, "M5V 3L9"},
			},
		},
		{
			name: "not a postal code",
			in:   "MAIN STREET",
			want: []reading{},
		},
		{
			name: "four digits are not a ZIP",
			in:   "1234",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, postalcode.Claims(tokens))

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

// Canada Post writes a postal code as two groups, so it arrives as two tokens
// and has to be claimed as a span.
func TestClaimsCanadianAcrossTwoTokens(t *testing.T) {
	tokens := token.Tokenize("M5V 3L9")
	claims := postalcode.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected one claim over both tokens, got %+v", claims)
	}
	if claims[0].Length != 2 || claims[0].Value != "M5V 3L9" {
		t.Errorf("got %+v, want a length 2 claim valued M5V 3L9", claims[0])
	}
}

// A ZIP+4 written with a space contains a bare ZIP inside it. Both readings are
// real and they overlap, so both are offered and the longer one is stronger.
func TestClaimsOffersZIPInsideZIPPlusFour(t *testing.T) {
	tokens := token.Tokenize("84088 1234")
	claims := postalcode.Claims(tokens)

	if len(claims) != 2 {
		t.Fatalf("expected the ZIP+4 span and the ZIP inside it, got %+v", claims)
	}

	plusFour, bare := claims[0], claims[1]
	if plusFour.Length != 2 || bare.Length != 1 {
		t.Fatalf("expected a length 2 and a length 1 claim, got %+v", claims)
	}
	if !plusFour.Overlaps(bare) {
		t.Error("the ZIP is inside the ZIP+4; the claims must overlap")
	}
	if plusFour.Confidence <= bare.Confidence {
		t.Errorf("the longer shape is better evidence: %d vs %d", plusFour.Confidence, bare.Confidence)
	}
}

// A postal code sits at the end of the last line, not at index 0.
func TestClaimsAtEndOfAddress(t *testing.T) {
	tokens := token.Tokenize("3253 W 9200 S, West Jordan, UT 84088")
	claims := postalcode.Claims(tokens)

	if len(claims) != 1 {
		t.Fatalf("expected only the ZIP claimed, got %+v", claims)
	}
	if got := token.Join(tokens[claims[0].Start:claims[0].End()]); got != "84088" {
		t.Errorf("claimed %q, want %q", got, "84088")
	}
}
