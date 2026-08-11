package postalcode_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
)

func TestNormalizePostal(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"62701", "62701"},
		{"62701-1234", "62701-1234"},
		{"627011234", "62701-1234"},
		{"62701 1234", "62701-1234"},
		{"k1a 0b1", "K1A 0B1"},
		{"K1A  0B1", "K1A 0B1"},
		{"K1A0B1", "K1A 0B1"},
		{"k1a0b1", "K1A 0B1"},
		{"K1A-0B1", "K1A 0B1"},
		{"", ""},
		{"unknown", ""},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := postalcode.Normalize(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Postal %q → %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
