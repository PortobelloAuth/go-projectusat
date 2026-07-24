package normalizer_test

import (
	"strings"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

func TestContentNormalizerBasicStreet(t *testing.T) {
	// 123 Main Street, Apt 4, Springfield IL 62701
	in := &address.Address{
		PrimaryNumber:       "123",
		StreetName:          "Main",
		StreetSuffix:        "Street",
		SecondaryDesignator: "Apartment",
		SecondaryNumber:     "4",
		City:                "Springfield",
		Region:              "Illinois",
		Postal:              "62701",
	}
	n := normalizer.NewContentNomalizer()
	got, err := n.Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	want := &address.Address{
		PrimaryNumber:       "123",
		StreetName:          "MAIN",
		StreetSuffix:        "ST",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "4",
		City:                "SPRINGFIELD",
		Region:              "IL",
		Postal:              "62701",
	}

	if *got != *want {
		t.Fatalf("Normalize = %+v, want %+v", got, want)
	}
}

func TestContentNormalizerDirectionalsAndHighway(t *testing.T) {
	in := &address.Address{
		PrimaryNumber:   "100",
		Predirectional:  "North",
		StreetName:      "US Hwy 41",
		StreetSuffix:    "",
		Postdirectional: "Southwest",
		City:            "Miami",
		Region:          "FL",
		Postal:          "33101-1234",
	}
	n := normalizer.NewContentNomalizer()
	got, err := n.Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.Predirectional != "N" {
		t.Errorf("Predirectional = %q, want N", got.Predirectional)
	}
	if got.Postdirectional != "SW" {
		t.Errorf("Postdirectional = %q, want SW", got.Postdirectional)
	}
	if got.StreetName != "US HIGHWAY 41" {
		t.Errorf("StreetName = %q, want US HIGHWAY 41", got.StreetName)
	}
	if got.Postal != "33101-1234" {
		t.Errorf("Postal = %q, want 33101-1234", got.Postal)
	}
	if got.Region != "FL" {
		t.Errorf("Region = %q, want FL", got.Region)
	}
}

func TestContentNormalizerUnknownAndEmpty(t *testing.T) {
	in := &address.Address{
		PrimaryNumber:       "UNKNOWN",
		Predirectional:      "",
		StreetName:          "Main",
		StreetSuffix:        "Ave",
		SecondaryDesignator: "unknown",
		City:                "Springfield",
		Region:              "IL",
		Postal:              "",
		Country:             "Unknown",
	}
	n := normalizer.NewContentNomalizer()
	got, err := n.Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.PrimaryNumber != "" {
		t.Errorf("PrimaryNumber UNKNOWN = %q, want blank", got.PrimaryNumber)
	}
	if got.SecondaryDesignator != "" {
		t.Errorf("SecondaryDesignator UNKNOWN = %q, want blank", got.SecondaryDesignator)
	}
	if got.Country != "" {
		t.Errorf("Country UNKNOWN = %q, want blank", got.Country)
	}
	if got.Predirectional != "" || got.Postdirectional != "" {
		t.Errorf("empty directionals should stay blank, got pre=%q post=%q",
			got.Predirectional, got.Postdirectional)
	}
	if got.StreetSuffix != "AVE" {
		t.Errorf("StreetSuffix = %q, want AVE", got.StreetSuffix)
	}

	// CollapseSpace must run before Upper/NormalizeUnknown so padded UNKNOWN blanks.
	padded, err := n.Normalize(&address.Address{
		PrimaryNumber: " UNKNOWN ",
		StreetName:    "Main",
		StreetSuffix:  "ST",
		Region:        " UNKNOWN ",
		City:          "Springfield",
	})
	if err != nil {
		t.Fatalf("padded UNKNOWN should not error (region blanks, skips NormalizeRegion): %v", err)
	}
	if padded.PrimaryNumber != "" {
		t.Errorf("PrimaryNumber %q → %q, want blank", " UNKNOWN ", padded.PrimaryNumber)
	}
	if padded.Region != "" {
		t.Errorf("Region %q → %q, want blank", " UNKNOWN ", padded.Region)
	}
}

// TODO: move or replicate this test in the postalcode package
func TestContentNormalizerPostalVariants(t *testing.T) {
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
		{"", ""},
		{"unknown", ""},
	}
	n := normalizer.NewContentNomalizer()
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := n.Normalize(&address.Address{Postal: tc.in, Region: "IL", City: "X", StreetName: "Main", StreetSuffix: "ST"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Postal != tc.want {
				t.Fatalf("Postal %q → %q, want %q", tc.in, got.Postal, tc.want)
			}
		})
	}
}

func TestContentNormalizerPreservesDiacritics(t *testing.T) {
	in := &address.Address{
		StreetName:   "José",
		StreetSuffix: "Street",
		City:         "San José",
		Region:       "CA",
		Postal:       "95112",
	}
	n := normalizer.NewContentNomalizer()
	got, err := n.Normalize(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.StreetName != "JOSÉ" {
		t.Errorf("StreetName = %q, want JOSÉ (diacritics preserved)", got.StreetName)
	}
	if got.City != "SAN JOSÉ" {
		t.Errorf("City = %q, want SAN JOSÉ", got.City)
	}
}

func TestContentNormalizerBusinessNameAndWhitespace(t *testing.T) {
	in := &address.Address{
		BusinessName:  "  Acme   Corp.  ",
		PrimaryNumber: "  112-10 ",
		StreetName:    "  Bronx  ",
		StreetSuffix:  " Road ",
		City:          " Bronx ",
		Region:        " ny ",
		Postal:        " 10475 ",
	}
	n := normalizer.NewContentNomalizer()
	got, err := n.Normalize(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.BusinessName != "ACME CORP." {
		// baseField only collapses + upper; does not strip period from free text
		t.Errorf("BusinessName = %q, want ACME CORP.", got.BusinessName)
	}
	if got.PrimaryNumber != "112-10" {
		t.Errorf("PrimaryNumber = %q, want 112-10", got.PrimaryNumber)
	}
	if got.StreetName != "BRONX" {
		t.Errorf("StreetName = %q, want BRONX", got.StreetName)
	}
	if got.StreetSuffix != "RD" {
		t.Errorf("StreetSuffix = %q, want RD", got.StreetSuffix)
	}
	if got.Region != "NY" {
		t.Errorf("Region = %q, want NY", got.Region)
	}
}

func TestContentNormalizerErrors(t *testing.T) {
	n := normalizer.NewContentNomalizer()
	t.Run("bad region", func(t *testing.T) {
		_, err := n.Normalize(&address.Address{Region: "Narnia", City: "X", StreetName: "Main", StreetSuffix: "ST"})
		if err == nil {
			t.Fatal("expected error for unrecognized region")
		}
		if !strings.Contains(err.Error(), "region") {
			t.Errorf("error %q should mention region", err)
		}
	})
	t.Run("bad predirectional", func(t *testing.T) {
		_, err := n.Normalize(&address.Address{Predirectional: "Sideways", StreetName: "Main", StreetSuffix: "ST", Region: "IL"})
		if err == nil {
			t.Fatal("expected error for unrecognized predirectional")
		}
	})
	t.Run("bad street suffix", func(t *testing.T) {
		_, err := n.Normalize(&address.Address{StreetName: "Main", StreetSuffix: "NotASuffix", Region: "IL"})
		if err == nil {
			t.Fatal("expected error for unrecognized street suffix")
		}
	})
	t.Run("bad secondary designator", func(t *testing.T) {
		_, err := n.Normalize(&address.Address{
			StreetName: "Main", StreetSuffix: "ST", Region: "IL",
			SecondaryDesignator: "Wing",
		})
		if err == nil {
			t.Fatal("expected error for unrecognized secondary designator")
		}
	})
}

func TestContentNormalizerEmptyAddress(t *testing.T) {
	n := normalizer.NewContentNomalizer()
	got, err := n.Normalize(&address.Address{})
	if err != nil {
		t.Fatalf("empty Address should not error: %v", err)
	}
	if *got != (address.Address{}) {
		t.Fatalf("got %+v, want empty Address", got)
	}
}

func TestNormalizerWithOptionsSecondaryAsHash(t *testing.T) {
	in := &address.Address{
		PrimaryNumber:       "123",
		StreetName:          "Main",
		StreetSuffix:        "Street",
		SecondaryDesignator: "Apartment",
		SecondaryNumber:     "4",
		City:                "Springfield",
		Region:              "Illinois",
		Postal:              "62701",
	}
	// Content form keeps APT.
	cn := normalizer.NewContentNomalizer()
	content, err := cn.Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if content.SecondaryDesignator != "APT" {
		t.Fatalf("content SecondaryDesignator = %q, want APT", content.SecondaryDesignator)
	}

	// Exchange/matching form rewrites to #.
	n := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{SecondaryAsHash: true})
	got, err := n.Normalize(in)
	if err != nil {
		t.Fatalf("NormalizeWithOptions: %v", err)
	}
	if got.SecondaryDesignator != "#" {
		t.Fatalf("SecondaryAsHash SecondaryDesignator = %q, want #", got.SecondaryDesignator)
	}
	if got.SecondaryNumber != "4" {
		t.Errorf("SecondaryNumber = %q, want 4", got.SecondaryNumber)
	}
	// Suite also becomes #.
	suite, err := n.Normalize(&address.Address{
		StreetName: "Main", StreetSuffix: "ST", Region: "IL",
		SecondaryDesignator: "Suite", SecondaryNumber: "100",
	})
	if err != nil {
		t.Fatalf("Suite: %v", err)
	}
	if suite.SecondaryDesignator != "#" {
		t.Fatalf("Suite SecondaryAsHash = %q, want #", suite.SecondaryDesignator)
	}

	// Input already "#" (SecondaryAsHash output) re-normalizes cleanly.
	hashIn, err := n.Normalize(&address.Address{
		StreetName: "Main", StreetSuffix: "ST", Region: "IL",
		SecondaryDesignator: "#", SecondaryNumber: "4",
	})
	if err != nil {
		t.Fatalf("SecondaryDesignator \"#\": %v", err)
	}
	if hashIn.SecondaryDesignator != "#" {
		t.Fatalf("hash input SecondaryDesignator = %q, want #", hashIn.SecondaryDesignator)
	}

	// Round-trip: SecondaryAsHash then normalize again succeeds with "#".
	again, err := n.Normalize(got)
	if err != nil {
		t.Fatalf("round-trip SecondaryAsHash: %v", err)
	}
	if again.SecondaryDesignator != "#" {
		t.Fatalf("round-trip SecondaryDesignator = %q, want #", again.SecondaryDesignator)
	}
	// Content form also accepts "#".
	contentHash, err := cn.Normalize(&address.Address{
		StreetName: "Main", StreetSuffix: "ST", Region: "IL",
		SecondaryDesignator: "#", SecondaryNumber: "4",
	})
	if err != nil {
		t.Fatalf("Normalize with \"#\": %v", err)
	}
	if contentHash.SecondaryDesignator != "#" {
		t.Fatalf("content hash SecondaryDesignator = %q, want #", contentHash.SecondaryDesignator)
	}
}

func TestNormalizerWithOptionsFuzzy(t *testing.T) {
	// Mild typos: Californa → CA, Aveneu → AVE (Fuzzy* threshold 0.7).
	// "Aveneu" is a real typo (not an alt form); "Avenu"/"AVENU" is a listed alt.
	in := &address.Address{
		PrimaryNumber: "10",
		StreetName:    "Oak",
		StreetSuffix:  "Aveneu",
		City:          "Sacramento",
		Region:        "Californa",
		Postal:        "95814",
	}
	// Exact mode fails.
	cn := normalizer.NewContentNomalizer()
	if _, err := cn.Normalize(in); err == nil {
		t.Fatal("expected error without Fuzzy for mild typos")
	}
	zn := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{})
	if _, err := zn.Normalize(in); err == nil {
		t.Fatal("expected error with zero Options for mild typos")
	}

	fn := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{Fuzzy: true})
	got, err := fn.Normalize(in)
	if err != nil {
		t.Fatalf("Fuzzy: unexpected error: %v", err)
	}
	if got.Region != "CA" {
		t.Errorf("Region = %q, want CA", got.Region)
	}
	if got.StreetSuffix != "AVE" {
		t.Errorf("StreetSuffix = %q, want AVE", got.StreetSuffix)
	}
}

func TestNormalizerWithOptionsDiacriticMode(t *testing.T) {
	in := &address.Address{
		StreetName:   "José",
		StreetSuffix: "Street",
		City:         "San José",
		BusinessName: "Café",
		Region:       "CA",
		Postal:       "95112",
	}

	// Default / empty: preserve diacritics (content form).
	zn := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{})
	preserved, err := zn.Normalize(in)
	if err != nil {
		t.Fatalf("empty DiacriticMode: %v", err)
	}
	if preserved.StreetName != "JOSÉ" || preserved.City != "SAN JOSÉ" || preserved.BusinessName != "CAFÉ" {
		t.Fatalf("preserve: StreetName=%q City=%q BusinessName=%q",
			preserved.StreetName, preserved.City, preserved.BusinessName)
	}

	// substitute: strip Project US@ diacritics then re-upper (Substitute returns lower).
	dsn := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{DiacriticMode: diacritics.SubstituteDiacritics})
	sub, err := dsn.Normalize(in)
	if err != nil {
		t.Fatalf("substitute: %v", err)
	}
	if sub.StreetName != "JOSE" {
		t.Errorf("substitute StreetName = %q, want JOSE", sub.StreetName)
	}
	if sub.City != "SAN JOSE" {
		t.Errorf("substitute City = %q, want SAN JOSE", sub.City)
	}
	if sub.BusinessName != "CAFE" {
		t.Errorf("substitute BusinessName = %q, want CAFE", sub.BusinessName)
	}

	// transliterate: anyascii path then upper.
	dtn := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{DiacriticMode: diacritics.TransliterateDiacritics})
	tr, err := dtn.Normalize(in)
	if err != nil {
		t.Fatalf("transliterate: %v", err)
	}
	if tr.StreetName != "JOSE" {
		t.Errorf("transliterate StreetName = %q, want JOSE", tr.StreetName)
	}
	if tr.City != "SAN JOSE" {
		t.Errorf("transliterate City = %q, want SAN JOSE", tr.City)
	}
	if tr.BusinessName != "CAFE" {
		t.Errorf("transliterate BusinessName = %q, want CAFE", tr.BusinessName)
	}
}

func TestContentNormalizerIsZeroOptions(t *testing.T) {
	// Normalize is content form: equivalent to NormalizeWithOptions(..., Options{}).
	in := &address.Address{
		PrimaryNumber:       "123",
		StreetName:          "Main",
		StreetSuffix:        "Street",
		SecondaryDesignator: "Apartment",
		SecondaryNumber:     "4",
		City:                "Springfield",
		Region:              "Illinois",
		Postal:              "62701",
	}
	cn := normalizer.NewContentNomalizer()
	a, err1 := cn.Normalize(in)
	zn := normalizer.NewNomalizer(normalizer.AddressNormalizationOptions{})
	b, err2 := zn.Normalize(in)
	if err1 != nil || err2 != nil {
		t.Fatalf("errors: Normalize=%v WithOptions=%v", err1, err2)
	}
	if *a != *b {
		t.Fatalf("Normalize = %+v, NormalizeWithOptions(zero) = %+v", a, b)
	}
}
