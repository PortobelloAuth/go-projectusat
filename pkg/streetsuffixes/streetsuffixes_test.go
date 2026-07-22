package streetsuffixes_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

func TestNormalizeStreetSuffixPrimary(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"AVE", "AVENUE"},
		{"avenue", "AVENUE"},
		{"AVENUE", "AVENUE"},
		{"AV", "AVENUE"},
		{"ST", "STREET"},
		{"street", "STREET"},
		{"STR", "STREET"},
		{"BLVD", "BOULEVARD"},
		{"boulevard", "BOULEVARD"},
		{"DR", "DRIVE"},
		{"drive", "DRIVE"},
		{"RD", "ROAD"},
		{"road", "ROAD"},
		{"LN", "LANE"},
		{"lane", "LANE"},
		{"CT", "COURT"},
		{"court", "COURT"},
		{"CIR", "CIRCLE"},
		{"circle", "CIRCLE"},
		{"HWY", "HIGHWAY"},
		{"highway", "HIGHWAY"},
		{"PL", "PLACE"},
		{"place", "PLACE"},
		{"PKWY", "PARKWAY"},
		{"parkway", "PARKWAY"},
		{"ALY", "ALLEY"},
		{"alley", "ALLEY"},
		{"BYP", "BYPASS"},
		{"bypass", "BYPASS"},
		{"WAY", "WAY"},
		{"WY", "WAY"},
		{"way", "WAY"},
		{"KEY", "KEY"},
		{"KY", "KEY"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := streetsuffixes.NormalizeStreetSuffix(tc.in)
			if err != nil {
				t.Fatalf("NormalizeStreetSuffix(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeStreetSuffix(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeStreetSuffixAbbreviation(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"AVENUE", "AVE"},
		{"avenue", "AVE"},
		{"AVE", "AVE"},
		{"STREET", "ST"},
		{"ST", "ST"},
		{"BOULEVARD", "BLVD"},
		{"DRIVE", "DR"},
		{"ROAD", "RD"},
		{"LANE", "LN"},
		{"COURT", "CT"},
		{"CIRCLE", "CIR"},
		{"HIGHWAY", "HWY"},
		{"PLACE", "PL"},
		{"PARKWAY", "PKWY"},
		{"ALLEY", "ALY"},
		{"BYPASS", "BYP"},
		{"WAY", "WAY"},
		{"WY", "WAY"},
		{"way", "WAY"},
		{"KEY", "KY"},
		{"KY", "KY"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := streetsuffixes.NormalizeStreetSuffixAbreviation(tc.in)
			if err != nil {
				t.Fatalf("NormalizeStreetSuffixAbreviation(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeStreetSuffixAbreviation(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeStreetSuffixUnknown(t *testing.T) {
	got, err := streetsuffixes.NormalizeStreetSuffix("NOTASUFFIX")
	if err == nil {
		t.Fatal("expected error for unrecognized street suffix")
	}
	if got != "" {
		t.Fatalf("got %q, want empty string on error", got)
	}
}

func TestFuzzyNormalizeStreetSuffix(t *testing.T) {
	// Real typo (not a listed alt form — "AVENU" is an alt; "Aveneu" is not).
	const typo = "Aveneu"
	if _, err := streetsuffixes.NormalizeStreetSuffix(typo); err == nil {
		t.Fatalf("NormalizeStreetSuffix(%q) succeeded; fixture is not a real typo", typo)
	}
	got, err := streetsuffixes.FuzzyNormalizeStreetSuffix(typo)
	if err != nil {
		t.Fatalf("FuzzyNormalizeStreetSuffix: %v", err)
	}
	if got != "AVENUE" {
		t.Fatalf("FuzzyNormalizeStreetSuffix(%q) = %q, want AVENUE", typo, got)
	}
}
