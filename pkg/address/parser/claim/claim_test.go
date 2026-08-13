package claim_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
)

// span builds a claim covering length tokens from start as a single part, for
// the cases that are about extent rather than about what is assigned.
func span(start, length int) claim.Claim {
	return claim.Claim{
		Parts: []claim.ClaimPart{{Start: start, Length: length}},
	}
}

func TestClaimPartEnd(t *testing.T) {
	cases := []struct {
		name string
		part claim.ClaimPart
		want int
	}{
		{"single token", claim.ClaimPart{Start: 0, Length: 1}, 1},
		{"multi token span", claim.ClaimPart{Start: 2, Length: 4}, 6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.part.End(); got != tc.want {
				t.Errorf("End() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestClaimExtent(t *testing.T) {
	cases := []struct {
		name                     string
		claim                    claim.Claim
		wantStart, wantEnd, want int
	}{
		{
			name:      "single part",
			claim:     span(0, 1),
			wantStart: 0, wantEnd: 1, want: 1,
		},
		{
			// RR 4 BOX 125: the rural route is a numbered street and the box is
			// a primary address on it. Two parts, four tokens, one reading.
			name: "adjacent parts",
			claim: claim.Claim{
				Confidence: claim.ConfidenceExact,
				Parts: []claim.ClaimPart{
					{Start: 0, Length: 2, Part: claim.PartStreetName, Value: "RR 4"},
					{Start: 2, Length: 2, Part: claim.PartPrimaryNumber, Value: "BOX 125"},
				},
			},
			wantStart: 0, wantEnd: 4, want: 4,
		},
		{
			// The extent is derived, so the parts need not arrive in order.
			name: "parts out of token order",
			claim: claim.Claim{
				Parts: []claim.ClaimPart{
					{Start: 3, Length: 2},
					{Start: 1, Length: 1},
				},
			},
			wantStart: 1, wantEnd: 5, want: 4,
		},
		{
			// No pattern in the library does this today. The extent spans the
			// gap, which is what Overlaps is documented to treat as competing.
			name: "parts with a gap between them",
			claim: claim.Claim{
				Parts: []claim.ClaimPart{
					{Start: 0, Length: 1},
					{Start: 4, Length: 1},
				},
			},
			wantStart: 0, wantEnd: 5, want: 5,
		},
		{
			name:      "no parts",
			claim:     claim.Claim{},
			wantStart: 0, wantEnd: 0, want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.claim.Start(); got != tc.wantStart {
				t.Errorf("Start() = %d, want %d", got, tc.wantStart)
			}
			if got := tc.claim.End(); got != tc.wantEnd {
				t.Errorf("End() = %d, want %d", got, tc.wantEnd)
			}
			if got := tc.claim.Length(); got != tc.want {
				t.Errorf("Length() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOverlaps(t *testing.T) {
	cases := []struct {
		name string
		a, b claim.Claim
		want bool
	}{
		{
			// Two readings of the same token: a region claim and a street name
			// claim over "WY". This is the case the parser has to arbitrate.
			name: "identical span",
			a:    span(1, 1),
			b:    span(1, 1),
			want: true,
		},
		{
			name: "shorter span nested in a longer one",
			a:    span(0, 4),
			b:    span(3, 1),
			want: true,
		},
		{
			name: "spans that share their middle tokens",
			a:    span(0, 3),
			b:    span(2, 3),
			want: true,
		},
		{
			name: "adjacent spans do not overlap",
			a:    span(0, 2),
			b:    span(2, 2),
			want: false,
		},
		{
			name: "disjoint spans",
			a:    span(0, 1),
			b:    span(5, 1),
			want: false,
		},
		{
			// PO BOX 11890 read as a post office box, against 11890 read as a
			// ZIP on its shape alone. Both are real, they differ in length,
			// and one of them has to lose — which is why a claim is a span
			// rather than a per-token score.
			name: "multi part claim against a shorter claim inside it",
			a: claim.Claim{
				Confidence: claim.ConfidenceExact,
				Parts: []claim.ClaimPart{
					{Start: 0, Length: 2, Part: claim.PartStreetName, Value: "PO BOX"},
					{Start: 2, Length: 1, Part: claim.PartPrimaryNumber, Value: "11890"},
				},
			},
			b:    span(2, 1),
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Overlaps(tc.b); got != tc.want {
				t.Errorf("Overlaps() = %v, want %v", got, tc.want)
			}
			if got := tc.b.Overlaps(tc.a); got != tc.want {
				t.Errorf("Overlaps() is not symmetric: reversed = %v, want %v", got, tc.want)
			}
		})
	}
}

// A claim covering no tokens would silently overlap everything containing its
// start, which would make a claim that asserts nothing look like a competing
// reading of real tokens.
func TestClaimWithNoPartsCoversNothing(t *testing.T) {
	empty := claim.Claim{Confidence: claim.ConfidenceExact}

	if empty.Length() != 0 {
		t.Errorf("Length() = %d, want 0 for a claim with no parts", empty.Length())
	}
	if empty.Overlaps(span(0, 5)) {
		t.Error("a claim with no parts should not overlap anything")
	}
	if empty.Overlaps(empty) {
		t.Error("a claim with no parts should not overlap itself")
	}
}

// A claim that assigns a gap-free run of parts covers exactly those tokens and
// no others, which is what lets a parser trust the extent it arbitrates on.
func TestExtentIsDerivedFromParts(t *testing.T) {
	// PSC 3 BOX 4120, starting at token 2 of a longer address.
	c := claim.Claim{
		Confidence: claim.ConfidenceExact,
		Parts: []claim.ClaimPart{
			{Start: 2, Length: 2, Part: claim.PartStreetName, Value: "PSC 3"},
			{Start: 4, Length: 2, Part: claim.PartPrimaryNumber, Value: "BOX 4120"},
		},
	}

	if c.Start() != 2 || c.End() != 6 {
		t.Errorf("extent = [%d,%d), want [2,6)", c.Start(), c.End())
	}
	if c.Overlaps(span(1, 1)) {
		t.Error("claim should not reach the token before its first part")
	}
	if c.Overlaps(span(6, 1)) {
		t.Error("claim should not reach the token after its last part")
	}
}
