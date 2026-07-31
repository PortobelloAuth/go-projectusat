package country_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/country"
)

func TestNormalizeCountry(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// remove the country name for US addresses
		{"UNITED STATES", ""},
		{"USA", ""},
		{"US", ""},
		{"U.S.A.", ""},
		{"U.S.", ""},

		// substitute + and & for "and" in country names
		{"trinidad and tobago", "TRINIDAD AND TOBAGO"},
		{"Trinidad + Tobago", "TRINIDAD AND TOBAGO"},
		{"Trinidad & Tobago", "TRINIDAD AND TOBAGO"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := country.NormalizeCountry(tc.in)
			if err != nil {
				t.Fatalf("NormalizeCountry(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeCountry(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
