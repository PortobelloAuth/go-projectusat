package region_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
)

// reading is a claim part flattened to the token text it covers, so cases read
// as "these words, claimed as this part, this strongly".
type reading struct {
	text       string
	part       claim.Part
	confidence claim.Confidence
	value      string
}

// flatten expands each claim into one reading per part. Every region claim
// assigns exactly one part, so the two are one to one here; confidence is read
// off the claim, which is where it lives.
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
			name: "two letter code is exact",
			in:   "UT",
			want: []reading{
				{"UT", claim.PartRegion, claim.ConfidenceExact, "UT"},
				{"UT", claim.PartStreetName, claim.ConfidenceLikely, "UTAH"},
			},
		},
		{
			name: "full name is strong",
			in:   "WYOMING",
			want: []reading{
				{"WYOMING", claim.PartRegion, claim.ConfidenceStrong, "WY"},
				{"WYOMING", claim.PartStreetName, claim.ConfidenceLikely, "WYOMING"},
			},
		},
		{
			name: "multi token name is a single span",
			in:   "SOUTH CAROLINA",
			want: []reading{
				{"SOUTH CAROLINA", claim.PartRegion, claim.ConfidenceStrong, "SC"},
				{"SOUTH CAROLINA", claim.PartStreetName, claim.ConfidenceLikely, "SOUTH CAROLINA"},
			},
		},
		{
			// The vocabulary lists MICRONESIA as an alias of the full name, so
			// the shorter reading is claimed too. Both are true readings of
			// these tokens and they overlap; the parser discards one.
			name: "longest span in the vocabulary, and the alias nested in it",
			in:   "FEDERATED STATES OF MICRONESIA",
			want: []reading{
				{"FEDERATED STATES OF MICRONESIA", claim.PartRegion, claim.ConfidenceStrong, "FM"},
				{"MICRONESIA", claim.PartRegion, claim.ConfidenceStrong, "FM"},
			},
		},
		{
			name: "military region is not a street name",
			in:   "AE",
			want: []reading{
				{"AE", claim.PartRegion, claim.ConfidenceExact, "AE"},
			},
		},
		{
			name: "no region present",
			in:   "MAIN STREET",
			want: []reading{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			got := flatten(tokens, region.Claims(tokens))

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

// The state-as-street-name cases in goprojectusat_test.go depend on a region
// token being readable two ways at once. This is the evidence the parser needs
// to produce "8011 WY WY" -> "8011 WYOMING WAY".
func TestClaimsOffersCompetingReadings(t *testing.T) {
	tokens := token.Tokenize("8011 WY WY")
	claims := region.Claims(tokens)

	if len(claims) != 4 {
		t.Fatalf("expected both WY tokens claimed as region and as street name, got %+v", claims)
	}

	for _, c := range claims {
		if !c.Overlaps(claims[0]) && !c.Overlaps(claims[2]) {
			t.Errorf("reading %+v covers neither WY token", c)
		}
	}

	region0, street0 := claims[0], claims[1]
	if !region0.Overlaps(street0) {
		t.Fatalf("competing readings of the same token must overlap: %+v and %+v", region0, street0)
	}
	if region0.Confidence <= street0.Confidence {
		t.Errorf("region reading should outrank the street name reading in isolation: %d vs %d",
			region0.Confidence, street0.Confidence)
	}
	if got := street0.Parts[0].Value; got != "WYOMING" {
		t.Errorf("street name reading should spell the state out, got %q", got)
	}
}

// An ordinal street name must produce no region claim at all. 500 W 2ND AVE is
// an ordinary address, and a region claim on 2ND puts North Dakota in the middle
// of the street address line, where the parser has to argue it back out.
func TestClaimsIgnoresOrdinalStreetNames(t *testing.T) {
	tokens := token.Tokenize("500 W 2ND AVE")
	if claims := region.Claims(tokens); len(claims) != 0 {
		t.Fatalf("expected no region claims in %q, got %+v", "500 W 2ND AVE", claims)
	}
}

// A region name that runs past the end of the token slice must not be read as
// a shorter match by accident, and must not panic.
func TestClaimsNearEndOfInput(t *testing.T) {
	tokens := token.Tokenize("SPRINGFIELD SOUTH")
	for _, c := range region.Claims(tokens) {
		if c.End() > len(tokens) {
			t.Fatalf("reading %+v runs past %d tokens", c, len(tokens))
		}
	}
}

// A region name hard wrapped across two lines is not a region name. Join
// flattens a span with spaces, so without a bound "NEW\nYORK" reaches the
// lookup as "NEW YORK" and a street name on one line and a city on the next
// become a state. See token.LineEnd.
func TestNoClaimSpansALineBreak(t *testing.T) {
	for _, source := range []string{
		"123 MAIN ST NEW\nYORK NY 10001",
		"123 NEW\nYORK ST DENVER CO 80201",
		"123 MAIN ST NORTH\nCAROLINA 27601",
	} {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range region.Claims(tokens) {
				if tokens[c.End()-1].Line != tokens[c.Start()].Line {
					t.Errorf("claim over %q spans a line break",
						token.Join(tokens[c.Start():c.End()]))
				}
			}
		})
	}
}

// The bound must not cost the reading it was added to protect: the same phrase
// on one line is still a region.
func TestAPhraseOnOneLineIsStillClaimed(t *testing.T) {
	tokens := token.Tokenize("123 MAIN ST\nNEW YORK NY 10001")

	var found bool
	for _, c := range region.Claims(tokens) {
		if token.Join(tokens[c.Start():c.End()]) == "NEW YORK" {
			found = true
		}
	}
	if !found {
		t.Error("expected NEW YORK on its own line to still be claimed")
	}
}
