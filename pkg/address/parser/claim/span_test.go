package claim_test

import (
	"slices"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
)

func TestSpanEnd(t *testing.T) {
	if got := (claim.Span{Start: 2, Length: 3}).End(); got != 5 {
		t.Errorf("End() = %d, want 5", got)
	}
}

func TestGaps(t *testing.T) {
	cases := []struct {
		name  string
		span  claim.Span
		parts []claim.ClaimPart
		want  []claim.Span
	}{
		{
			name: "nothing covered is one gap",
			span: claim.Span{Start: 0, Length: 4},
			want: []claim.Span{{Start: 0, Length: 4}},
		},
		{
			name:  "fully covered is no gaps",
			span:  claim.Span{Start: 0, Length: 4},
			parts: []claim.ClaimPart{{Start: 0, Length: 4}},
		},
		{
			name:  "a hole in the middle is one run, not one per token",
			span:  claim.Span{Start: 0, Length: 5},
			parts: []claim.ClaimPart{{Start: 0, Length: 1}, {Start: 3, Length: 2}},
			want:  []claim.Span{{Start: 1, Length: 2}},
		},
		{
			name:  "separate holes are separate runs",
			span:  claim.Span{Start: 0, Length: 6},
			parts: []claim.ClaimPart{{Start: 1, Length: 1}, {Start: 3, Length: 1}},
			want:  []claim.Span{{Start: 0, Length: 1}, {Start: 2, Length: 1}, {Start: 4, Length: 2}},
		},
		{
			name:  "overlapping parts cover once",
			span:  claim.Span{Start: 0, Length: 4},
			parts: []claim.ClaimPart{{Start: 0, Length: 3}, {Start: 1, Length: 3}},
		},
		{
			name:  "parts outside the span do not cover it",
			span:  claim.Span{Start: 2, Length: 2},
			parts: []claim.ClaimPart{{Start: 0, Length: 2}, {Start: 4, Length: 2}},
			want:  []claim.Span{{Start: 2, Length: 2}},
		},
		{
			name:  "a part reaching in covers only the overlap",
			span:  claim.Span{Start: 2, Length: 4},
			parts: []claim.ClaimPart{{Start: 0, Length: 4}},
			want:  []claim.Span{{Start: 4, Length: 2}},
		},
		{
			name: "an empty span has no gaps",
			span: claim.Span{Start: 3, Length: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.span.Gaps(tc.parts)
			if !slices.Equal(got, tc.want) {
				t.Errorf("Gaps() = %v, want %v", got, tc.want)
			}
		})
	}
}
