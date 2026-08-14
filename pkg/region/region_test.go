package region_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/region"
)

func TestNormalizeRegion(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// US states and abbreviations
		{"California", "CA"},
		{"CA", "CA"},
		{"ca", "CA"},
		{"New York", "NY"},
		{"NY", "NY"},
		{"Texas", "TX"},
		{"Illinois", "IL"},

		// Correct spelling and common misspelling of Delaware
		{"Delaware", "DE"},
		{"DELAWARE", "DE"},
		{"DELEWARE", "DE"},
		{"DE", "DE"},

		// District of Columbia must map to DC, not DE
		{"District of Columbia", "DC"},
		{"DISTRICT OF COLUMBIA", "DC"},
		{"DC", "DC"},

		// Possessions and territories
		{"Puerto Rico", "PR"},
		{"Guam", "GU"},
		{"Virgin Islands", "VI"},
		{"American Samoa", "AS"},

		// Canadian provinces
		{"Ontario", "ON"},
		{"British Columbia", "BC"},
		{"Quebec", "QC"},
		{"Newfoundland and Labrador", "NL"},

		// Military "states"
		{"AE", "AE"},
		{"Armed Forces Pacific", "AP"},
		{"Armed Forces Americas", "AA"},

		// Punctuation stripped before lookup
		{"N. Carolina", "NC"},
		{"S. Dakota", "SD"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := region.NormalizeRegion(tc.in)
			if err != nil {
				t.Fatalf("NormalizeRegion(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeRegion(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeRegionUnknown(t *testing.T) {
	got, err := region.NormalizeRegion("Narnia")
	if err == nil {
		t.Fatal("expected error for unrecognized region")
	}
	if got != "" {
		t.Fatalf("got %q, want empty string on error", got)
	}
}

// Cleaning removes punctuation, not digits. An ordinal street name is not a
// region, and it stops being one only if the digit survives long enough to make
// the lookup fail.
func TestNormalizeRegionKeepsDigits(t *testing.T) {
	for _, in := range []string{"2ND", "22ND", "42ND", "2CO", "1UT"} {
		t.Run(in, func(t *testing.T) {
			got, err := region.NormalizeRegion(in)
			if err == nil {
				t.Fatalf("NormalizeRegion(%q) = %q, want an error", in, got)
			}
			if got != "" {
				t.Fatalf("NormalizeRegion(%q) = %q, want empty string on error", in, got)
			}
		})
	}
}

// Punctuation must still be cleaned, which is what the pattern is for.
func TestNormalizeRegionStripsPunctuation(t *testing.T) {
	for _, in := range []string{"U.T.", "UT.", "(UT)"} {
		t.Run(in, func(t *testing.T) {
			got, err := region.NormalizeRegion(in)
			if err != nil {
				t.Fatalf("NormalizeRegion(%q) unexpected error: %v", in, err)
			}
			if got != "UT" {
				t.Fatalf("NormalizeRegion(%q) = %q, want UT", in, got)
			}
		})
	}
}

func TestFuzzyNormalizeRegion(t *testing.T) {
	// Mild typo in a long region name should still resolve when fuzzy is enabled.
	got, err := region.FuzzyNormalizeRegion("Californa")
	if err != nil {
		t.Fatalf("FuzzyNormalizeRegion: %v", err)
	}
	if got != "CA" {
		t.Fatalf("FuzzyNormalizeRegion(Californa) = %q, want CA", got)
	}
}
