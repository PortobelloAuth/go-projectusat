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

		// Digits are not punctuation. A token with one in it is passed through
		// rather than cleaned into a country it never named.
		{"2US", "2US"},
		{"1CA", "1CA"},
		{"MX9", "MX9"},
		{"3MEXICO", "3MEXICO"},
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

// Canadian postal codes are letters and digits together, so stripping the
// digits left a two letter residue that this vocabulary has an entry for. These
// are real forward sortation areas: M5X is in Toronto and became MEXICO, and
// C1A is in Charlottetown and became CANADA for reasons that had nothing to do
// with it being Canadian.
func TestNormalizeCountryDoesNotReadPostalCodes(t *testing.T) {
	for _, in := range []string{"C1A", "M5X", "V6C", "C1A 1A1"} {
		t.Run(in, func(t *testing.T) {
			got, err := country.NormalizeCountry(in)
			if err != nil {
				t.Fatalf("NormalizeCountry(%q) unexpected error: %v", in, err)
			}
			if got != in {
				t.Fatalf("NormalizeCountry(%q) = %q, want it passed through unchanged", in, got)
			}
		})
	}
}
