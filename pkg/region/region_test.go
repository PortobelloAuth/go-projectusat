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
		{"Wyoming", "WY"},
		{"WY", "WY"},

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

func TestScoreAndFullName(t *testing.T) {
	if sc, _ := region.Score("IL"); sc != 100 {
		t.Fatalf("Score(IL)=%d want 100", sc)
	}
	if sc, _ := region.Score("ILLINOIS"); sc != 90 {
		t.Fatalf("Score(ILLINOIS)=%d want 90", sc)
	}
	if sc, _ := region.Score("MAIN"); sc != 0 {
		t.Fatalf("Score(MAIN)=%d want 0", sc)
	}
	full, ok := region.FullName("CT")
	if !ok || full != "CONNECTICUT" {
		t.Fatalf("FullName(CT)=%q,%v", full, ok)
	}
	code, n, ok := region.LeadingStateMatch([]string{"SOUTH", "CAROLINA", "COUNTY", "ROAD", "22"}, true)
	if !ok || code != "SC" || n != 2 {
		t.Fatalf("LeadingStateMatch = %s %d %v", code, n, ok)
	}
}
