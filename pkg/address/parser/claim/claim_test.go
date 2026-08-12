package claim_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
)

func TestEnd(t *testing.T) {
	cases := []struct {
		name  string
		claim claim.Claim
		want  int
	}{
		{"single token", claim.Claim{Start: 0, Length: 1}, 1},
		{"multi token span", claim.Claim{Start: 2, Length: 4}, 6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.claim.End(); got != tc.want {
				t.Errorf("End() = %d, want %d", got, tc.want)
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
			a:    claim.Claim{Start: 1, Length: 1},
			b:    claim.Claim{Start: 1, Length: 1},
			want: true,
		},
		{
			name: "shorter span nested in a longer one",
			a:    claim.Claim{Start: 0, Length: 4},
			b:    claim.Claim{Start: 3, Length: 1},
			want: true,
		},
		{
			name: "spans that share their middle tokens",
			a:    claim.Claim{Start: 0, Length: 3},
			b:    claim.Claim{Start: 2, Length: 3},
			want: true,
		},
		{
			name: "adjacent spans do not overlap",
			a:    claim.Claim{Start: 0, Length: 2},
			b:    claim.Claim{Start: 2, Length: 2},
			want: false,
		},
		{
			name: "disjoint spans",
			a:    claim.Claim{Start: 0, Length: 1},
			b:    claim.Claim{Start: 5, Length: 1},
			want: false,
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

// A claim covering no tokens would silently overlap nothing, which would make
// it invisible to a parser resolving competing readings. Length is documented
// as always at least 1; this pins the boundary that documentation describes.
func TestZeroLengthClaimCoversNothing(t *testing.T) {
	empty := claim.Claim{Start: 2, Length: 0}

	if empty.End() != empty.Start {
		t.Errorf("End() = %d, want %d for a zero length claim", empty.End(), empty.Start)
	}
	if empty.Overlaps(claim.Claim{Start: 0, Length: 5}) {
		t.Error("a zero length claim should not overlap anything")
	}
}
