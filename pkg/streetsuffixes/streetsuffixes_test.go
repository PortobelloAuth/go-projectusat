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

// Cleaning removes punctuation, not digits. An ordinal is a street name, and
// it stops being read as a suffix only if the digit survives long enough to
// make the lookup fail.
func TestNormalizeStreetSuffixKeepsDigits(t *testing.T) {
	for _, in := range []string{"1ST", "3RD", "21ST", "23RD", "1AVE"} {
		t.Run(in, func(t *testing.T) {
			got, err := streetsuffixes.NormalizeStreetSuffix(in)
			if err == nil {
				t.Fatalf("NormalizeStreetSuffix(%q) = %q, want an error", in, got)
			}
			if got != "" {
				t.Fatalf("NormalizeStreetSuffix(%q) = %q, want empty string on error", in, got)
			}
		})
	}
}

// Punctuation must still be cleaned, which is what the pattern is for.
func TestNormalizeStreetSuffixStripsPunctuation(t *testing.T) {
	for _, in := range []string{"AVE.", "AVE,", "(AVE)"} {
		t.Run(in, func(t *testing.T) {
			got, err := streetsuffixes.NormalizeStreetSuffix(in)
			if err != nil {
				t.Fatalf("NormalizeStreetSuffix(%q) unexpected error: %v", in, err)
			}
			if got != "AVENUE" {
				t.Fatalf("NormalizeStreetSuffix(%q) = %q, want AVENUE", in, got)
			}
		})
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

// DAM, OVERPASS, and PRAIRIE are the three Publication 28 primary names the
// table did not index under their own spelling. The first two were missing
// from their Alt lists; the third was spelled PRAIRE throughout, so the
// correctly spelled word was absent from the table entirely.
func TestNormalizeStreetSuffixAcceptsPub28PrimaryNames(t *testing.T) {
	cases := []struct {
		in    string
		want  string
		short string
	}{
		{"DAM", "DAM", "DM"},
		{"OVERPASS", "OVERPASS", "OPAS"},
		{"PRAIRIE", "PRAIRIE", "PR"},
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

			abbr, err := streetsuffixes.NormalizeStreetSuffixAbreviation(tc.in)
			if err != nil {
				t.Fatalf("NormalizeStreetSuffixAbreviation(%q) unexpected error: %v", tc.in, err)
			}
			if abbr != tc.short {
				t.Fatalf("NormalizeStreetSuffixAbreviation(%q) = %q, want %q", tc.in, abbr, tc.short)
			}
		})
	}
}

// DALE was listed in DAM's Alt as well as its own, and the table is iterated in
// order when the lookup maps are built, so DAM won and DALE normalized to it.
// This is the case worth a test of its own: the other two failed loudly with an
// error, this one returned a confident wrong answer.
func TestNormalizeStreetSuffixDaleIsNotDam(t *testing.T) {
	got, err := streetsuffixes.NormalizeStreetSuffix("DALE")
	if err != nil {
		t.Fatalf("NormalizeStreetSuffix(\"DALE\") unexpected error: %v", err)
	}
	if got != "DALE" {
		t.Fatalf("NormalizeStreetSuffix(\"DALE\") = %q, want DALE", got)
	}

	abbr, err := streetsuffixes.NormalizeStreetSuffixAbreviation("DALE")
	if err != nil {
		t.Fatalf("NormalizeStreetSuffixAbreviation(\"DALE\") unexpected error: %v", err)
	}
	if abbr != "DL" {
		t.Fatalf("NormalizeStreetSuffixAbreviation(\"DALE\") = %q, want DL", abbr)
	}
}
