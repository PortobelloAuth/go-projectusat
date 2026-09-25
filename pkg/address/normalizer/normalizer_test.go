package normalizer_test

import (
	"strings"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
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

// A directional street name is spelled out, whether it arrives abbreviated or
// not: Project US@ p.17 gives NORTH AVE as the correct form.
// TestContentNormalizerKeepsAnAlphabetLetterAndSpellsOutADirectionalName
// checks the p.17/p.18 rule this normalizer implements: a single letter is
// left as written because it may be an alphabet indicator (1000 AVENUE E),
// and directional letters SHOULD NOT be combined with alphabet indicators
// (p.17); anything longer than one letter is a direction, not a letter of
// the alphabet, and is spelled out (p.18: BAY WEST DRIVE, NORTH AVE).
func TestContentNormalizerKeepsAnAlphabetLetterAndSpellsOutADirectionalName(t *testing.T) {
	n := normalizer.NewContentNomalizer()
	for _, tc := range []struct {
		name   string
		suffix string
		want   string
	}{
		{"N", "Avenue", "123 N AVE"},
		{"North", "Avenue", "123 NORTH AVE"},
		{"SE", "Fwy", "123 SOUTHEAST FWY"},
		{"E", "St", "123 E ST"},
		{"AVE E", "", "123 AVENUE E"},
		{"BAY W", "Drive", "123 BAY WEST DR"},
	} {
		got, err := n.Normalize(&address.Address{PrimaryNumber: "123", StreetName: tc.name, StreetSuffix: tc.suffix})
		if err != nil {
			t.Fatalf("Normalize(name=%q, suffix=%q): unexpected error: %v", tc.name, tc.suffix, err)
		}
		if got.FormatStreetLine() != tc.want {
			t.Errorf("street line for name %q suffix %q = %q, want %q", tc.name, tc.suffix, got.FormatStreetLine(), tc.want)
		}
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

func TestNormalizerKeepsTypeAreaAndDetail(t *testing.T) {
	// PO Box 159753 PMB 3571 — the standard's own CMRA example, with an
	// urbanization above it so every field the street line does not own is
	// exercised at once.
	in := &address.Address{
		Type:          &pobox.POBoxAddress{},
		Area:          "Urb  Highland Gdns",
		PrimaryNumber: "159753",
		StreetName:    "PO Box",
		Detail:        "pmb 3571",
		City:          "San Juan",
		Region:        "PR",
		Postal:        "00926",
	}
	got, err := normalizer.NewContentNomalizer().Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.Type != in.Type {
		t.Fatalf("Normalize dropped Type: got %v, want %T", got.Type, in.Type)
	}
	if got.Area != "URB HIGHLAND GDNS" || got.Detail != "PMB 3571" {
		t.Fatalf("Normalize Area = %q, Detail = %q; want URB HIGHLAND GDNS, PMB 3571", got.Area, got.Detail)
	}
	if want := "PO BOX 159753 PMB 3571"; got.FormatStreetLine() != want {
		t.Fatalf("FormatStreetLine = %q, want %q", got.FormatStreetLine(), want)
	}
}

func TestNormalizerNormalizesPOBoxStreetName(t *testing.T) {
	in := &address.Address{
		Type:          &pobox.POBoxAddress{},
		PrimaryNumber: "8755",
		StreetName:    "Post Office Box",
		City:          "Provo",
		Region:        "UT",
		Postal:        "84604",
	}
	got, err := normalizer.NewContentNomalizer().Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.Type != in.Type {
		t.Fatalf("Normalize dropped Type: got %v, want %T", got.Type, in.Type)
	}
	if got.StreetName != "PO BOX" {
		t.Fatalf("PO Box StreetName not normalized = %q", got.StreetName)
	}
	if got.PrimaryNumber != "8755" {
		t.Fatalf("PO Box PrimaryNumber not normalized = %q", got.PrimaryNumber)
	}
	if got.Area != "" || got.Detail != "" {
		t.Fatalf("Non-empty values in Area = %q, Detail = %q", got.Area, got.Detail)
	}
	if want := "PO BOX 8755"; got.FormatStreetLine() != want {
		t.Fatalf("FormatStreetLine = %q, want %q", got.FormatStreetLine(), want)
	}
}

// go-projectusat#115: Publication 28 §223 and Project US@ (p.20) both
// require a city name spelled out in its entirety, so ST/STE/MT/FT heading a
// city name must expand. A lone or trailing abbreviation, with nothing
// following it, is left as written per the shared table's position rule.
func TestContentNormalizerExpandsCityAbbreviations(t *testing.T) {
	n := normalizer.NewContentNomalizer()
	for _, tc := range []struct {
		city string
		want string
	}{
		{"ST CLOUD", "SAINT CLOUD"},
		{"SAINT CLOUD", "SAINT CLOUD"},
		{"MT VERNON", "MOUNT VERNON"},
		{"FT WORTH", "FORT WORTH"},
		{"STE GENEVIEVE", "SAINTE GENEVIEVE"},
		{"ST", "ST"}, // lone abbreviation, nothing follows: left as written
		{"SPRINGFIELD", "SPRINGFIELD"},
	} {
		got, err := n.Normalize(&address.Address{PrimaryNumber: "1", StreetName: "Main", StreetSuffix: "St", City: tc.city, Region: "MN", Postal: "56301"})
		if err != nil {
			t.Fatalf("Normalize(city=%q): unexpected error: %v", tc.city, err)
		}
		if got.City != tc.want {
			t.Errorf("City for %q = %q, want %q", tc.city, got.City, tc.want)
		}
	}
}

// go-projectusat#95: a Puerto Rico address must use only its own Spanish
// street-type vocabulary (pkg/addresstypes/puertorico), never the shared
// English suffix table. AVE and BLVD collide between the two tables, so
// without a dialect check a PR address silently mistranslates: the spec's
// own example (p.25), 1234 AVE ASHFORD, must stay Spanish (AVENIDA) rather
// than becoming the English AVENUE.
func TestContentNormalizerPuertoRicoStreetTypeStaysSpanish(t *testing.T) {
	n := normalizer.NewContentNomalizer()

	// Spec p.25 example: abbreviated Spanish street type expands to its
	// Spanish primary form, not the colliding English one.
	got, err := n.Normalize(&address.Address{
		PrimaryNumber: "1234",
		StreetName:    "AVE Ashford",
		City:          "San Juan",
		Region:        "PR",
		Postal:        "00907",
	})
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.StreetName != "AVENIDA ASHFORD" {
		t.Errorf("StreetName = %q, want AVENIDA ASHFORD", got.StreetName)
	}

	// Already-full Spanish form is left unchanged.
	got, err = n.Normalize(&address.Address{
		PrimaryNumber: "1234",
		StreetName:    "Avenida Ashford",
		City:          "San Juan",
		Region:        "PR",
		Postal:        "00907",
	})
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.StreetName != "AVENIDA ASHFORD" {
		t.Errorf("StreetName = %q, want AVENIDA ASHFORD (already full form)", got.StreetName)
	}

	// Non-PR address: English behavior is unchanged, AVE still expands to
	// AVENUE inside a multi-word street name.
	got, err = n.Normalize(&address.Address{
		PrimaryNumber: "1234",
		StreetName:    "AVE Ashford",
		City:          "Miami",
		Region:        "FL",
		Postal:        "33101",
	})
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.StreetName != "AVENUE ASHFORD" {
		t.Errorf("StreetName = %q, want AVENUE ASHFORD (English unaffected)", got.StreetName)
	}
}

// go-projectusat#114: a head-position ST/STE/MT/FT in a street *name*, with
// another word following it, is read from the city table as
// SAINT/SAINTE/MOUNT/FORT rather than reaching the street suffix table and
// becoming STREET/ROUTE. A trailing suffix-position ST (MAIN ST) and a lone
// ST street name are unaffected: they never entered this loop, or entered it
// as the single-word case that #108 already settled.
func TestContentNormalizerSpellsOutCityAbbreviationHeadingAStreetName(t *testing.T) {
	n := normalizer.NewContentNomalizer()
	for _, tc := range []struct {
		streetName string
		want       string
	}{
		{"ST CLAIR", "SAINT CLAIR"},
		{"FT MYERS", "FORT MYERS"},
		{"STE GENEVIEVE", "SAINTE GENEVIEVE"},
		{"MAIN", "MAIN"}, // unrelated, unaffected
		// A name of nothing but the abbreviation and a direction is a suffix
		// a reading absorbed, not a saint: SAINT NORTHWEST is not a street.
		// The suffix table gets it instead. addressparsers#25 reads
		// 100 EAST ST NW that way while it is weighing the alternatives.
		{"ST NW", "STREET NORTHWEST"},
		{"ST NORTH EAST", "STREET NORTH EAST"},
	} {
		got, err := n.Normalize(&address.Address{PrimaryNumber: "435", Predirectional: "S", StreetName: tc.streetName, StreetSuffix: "St", City: "Toledo", Region: "OH", Postal: "43601"})
		if err != nil {
			t.Fatalf("Normalize(streetName=%q): unexpected error: %v", tc.streetName, err)
		}
		if got.StreetName != tc.want {
			t.Errorf("StreetName for %q = %q, want %q", tc.streetName, got.StreetName, tc.want)
		}
	}

	// A trailing ST in a street name (MAIN ST as the name itself, not the
	// suffix field) must not expand: nothing follows it.
	got, err := n.Normalize(&address.Address{PrimaryNumber: "1", StreetName: "Main St"})
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.StreetName != "MAIN ST" {
		t.Errorf("StreetName = %q, want MAIN ST (trailing ST must not expand)", got.StreetName)
	}

	// And a ST inside the name with words after it is still the suffix word,
	// not a saint. The table's "spelled out when another word follows" rule
	// is about city names, where ST can only be SAINT; a street name has a
	// word after its suffix whenever it carries a trailing direction.
	got, err = n.Normalize(&address.Address{PrimaryNumber: "1011", StreetName: "Main Thing St North East"})
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if strings.Contains(got.StreetName, "SAINT") {
		t.Errorf("StreetName = %q, want no SAINT in it (ST here is the suffix word)", got.StreetName)
	}
}
