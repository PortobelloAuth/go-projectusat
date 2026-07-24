package parse

import (
	"strings"
	"testing"
)

func TestMergeDirectionTokens(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"SOUTH WEST MAIN", "SW MAIN"},
		{"N E MAIN", "NE MAIN"},
		{"NORTH EAST", "NE"},
		{"S W", "SW"},
		{"NORTHEAST MAIN", "NORTHEAST MAIN"}, // already compound, no-op
		{"NE MAIN", "NE MAIN"},
		{"NORTH SOUTH MAIN", "NORTH SOUTH MAIN"}, // opposite: do not merge
		{"EAST WEST", "EAST WEST"},
		{"MAIN STREET", "MAIN STREET"},
	}
	for _, tc := range cases {
		got := strings.Join(mergeDirectionTokens(strings.Fields(tc.in)), " ")
		if got != tc.want {
			t.Errorf("mergeDirectionTokens(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSplitPreStreet(t *testing.T) {
	cases := []struct {
		in, wantBiz, wantRest string
	}{
		{"123 MAIN ST", "", "123 MAIN ST"},
		{"WILLIAMSON MEDICAL CENTER 3000 EDWARD CURD LANE", "WILLIAMSON MEDICAL CENTER", "3000 EDWARD CURD LANE"},
		{"LIVES IN THE TENT NEAR 155 NORTH MAIN STREET", "LIVES IN THE TENT NEAR", "155 NORTH MAIN STREET"},
		{"RD 5A", "", "RD 5A"}, // primary-looking last token with no following body
		{"MAIN", "", "MAIN"},
		{"", "", ""},
	}
	for _, tc := range cases {
		var in []string
		if tc.in != "" {
			in = strings.Fields(tc.in)
		}
		biz, rest := splitPreStreet(in)
		gotRest := strings.Join(rest, " ")
		if biz != tc.wantBiz || gotRest != tc.wantRest {
			t.Errorf("splitPreStreet(%q) = (%q, %q), want (%q, %q)",
				tc.in, biz, gotRest, tc.wantBiz, tc.wantRest)
		}
	}
}
