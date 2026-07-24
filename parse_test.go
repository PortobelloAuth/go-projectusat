package goprojectusat

import (
	"strings"
	"testing"
)

func TestParseEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\n"} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}

func TestParseMilitaryMultiline(t *testing.T) {
	raw := "PSC 3 BOX 4120\nAPO AE 09021-0002"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Parse returns structured fields; may not yet Normalize — content may be upper from parse clean.
	if got.StreetName != "PSC 3 BOX 4120" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.City != "APO" || got.Region != "AE" || got.Postal != "09021-0002" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if Format(norm) != "PSC 3 BOX 4120\nAPO AE 09021-0002" {
		t.Errorf("Format = %q", Format(norm))
	}
}

func TestParseMilitaryCommaSeparated(t *testing.T) {
	raw := "UNIT 2050 BOX 4190, APO AP 96278-2050"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "UNIT 2050 BOX 4190" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.City != "APO" || got.Region != "AP" {
		t.Errorf("city/region = %q %q", got.City, got.Region)
	}
}

func TestParseMilitarySingleLineSpaceSeparated(t *testing.T) {
	// Space-only (no newline/comma) must still parse overseas military.
	raw := "PSC 3 BOX 4120 APO AE 09021-0002"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "PSC 3 BOX 4120" {
		t.Errorf("StreetName = %q, want PSC 3 BOX 4120", got.StreetName)
	}
	if got.BusinessName != "" {
		t.Errorf("BusinessName = %q, want empty", got.BusinessName)
	}
	if got.City != "APO" || got.Region != "AE" || got.Postal != "09021-0002" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
	if got.PrimaryNumber != "" || got.SecondaryDesignator != "" {
		t.Errorf("expected no civilian peels, primary=%q sec=%q", got.PrimaryNumber, got.SecondaryDesignator)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if Format(norm) != "PSC 3 BOX 4120\nAPO AE 09021-0002" {
		t.Errorf("Format = %q", Format(norm))
	}
}

func TestParseMilitarySingleLineWithBusiness(t *testing.T) {
	raw := "HQ DET PSC 3 BOX 4120 APO AE 09021-0002"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "PSC 3 BOX 4120" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.BusinessName != "HQ DET" {
		t.Errorf("BusinessName = %q, want HQ DET", got.BusinessName)
	}
	if got.City != "APO" || got.Region != "AE" || got.Postal != "09021-0002" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
}

func TestParseMilitaryStreetCivilianLastLine(t *testing.T) {
	// Military-form street with civilian city must keep street intact.
	raw := "UNIT 2050 BOX 4190\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "UNIT 2050 BOX 4190" {
		t.Errorf("StreetName = %q, want UNIT 2050 BOX 4190", got.StreetName)
	}
	if got.PrimaryNumber != "" || got.SecondaryDesignator != "" {
		t.Errorf("expected no civilian peels, primary=%q sec=%q", got.PrimaryNumber, got.SecondaryDesignator)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" || got.Postal != "62701" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
}

func TestParseSimpleMultiline(t *testing.T) {
	raw := "123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" || got.Postal != "62701" {
		t.Fatalf("last line = %+v", got)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

// PL-004a: comma in last line must not hard-split into a city-less last segment.
func TestParseCommaInLastLine(t *testing.T) {
	raw := "123 Main Street\nSpringfield, IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" || got.Postal != "62701" {
		t.Fatalf("last = city=%q region=%q postal=%q", got.City, got.Region, got.Postal)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

// PL-004b: comma before apartment stays on the street line.
func TestParseCommaBeforeApt(t *testing.T) {
	raw := "123 Main Street, Apt 4\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
	if got.SecondaryDesignator != "APT" || got.SecondaryNumber != "4" {
		t.Fatalf("secondary = %q %q, want APT 4", got.SecondaryDesignator, got.SecondaryNumber)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if FormatStreetLine(norm) != "123 MAIN ST APT 4" {
		t.Fatalf("FormatStreetLine = %q, want 123 MAIN ST APT 4", FormatStreetLine(norm))
	}
}

// PL-005: sole remaining name token that is directional stays as StreetName.
func TestParseDirectionalOnlyStreetName(t *testing.T) {
	raw := "123 South\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" {
		t.Fatalf("PrimaryNumber = %q, want 123", got.PrimaryNumber)
	}
	if got.StreetName != "SOUTH" {
		t.Fatalf("StreetName = %q, want SOUTH", got.StreetName)
	}
	if got.Predirectional != "" {
		t.Fatalf("Predirectional = %q, want empty", got.Predirectional)
	}
	if got.Postdirectional != "" {
		t.Fatalf("Postdirectional = %q, want empty", got.Postdirectional)
	}
}

func TestParseWithBusinessLine(t *testing.T) {
	raw := "Acme Corp\n123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.BusinessName != "ACME CORP" {
		t.Errorf("BusinessName = %q", got.BusinessName)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Errorf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

func TestParseLastLineCanadian(t *testing.T) {
	// "OTTAWA ON K1A 0B1" — postal is two tokens.
	raw := "10 Wellington Street\nOttawa ON K1A 0B1"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Region != "ON" {
		t.Errorf("Region = %q, want ON", got.Region)
	}
	// Postal after Normalize would be K1A 0B1; Parse may store "K1A 0B1"
	if !strings.Contains(strings.ToUpper(got.Postal), "K1A") {
		t.Errorf("Postal = %q", got.Postal)
	}
	if got.PrimaryNumber != "10" || got.StreetName != "WELLINGTON" || got.StreetSuffix != "ST" {
		t.Errorf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

func TestParseStreetComponents(t *testing.T) {
	raw := "123 North Main Street Apt 4\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" {
		t.Errorf("PrimaryNumber = %q", got.PrimaryNumber)
	}
	if got.Predirectional != "N" {
		t.Errorf("Predirectional = %q", got.Predirectional)
	}
	if got.StreetName != "MAIN" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.StreetSuffix != "ST" {
		t.Errorf("StreetSuffix = %q", got.StreetSuffix)
	}
	if got.SecondaryDesignator != "APT" {
		t.Errorf("SecondaryDesignator = %q", got.SecondaryDesignator)
	}
	if got.SecondaryNumber != "4" {
		t.Errorf("SecondaryNumber = %q", got.SecondaryNumber)
	}
}

func TestParseStreetPostdirectionalAndHash(t *testing.T) {
	raw := "100 Main Street Southwest # 12\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Postdirectional != "SW" {
		t.Errorf("Postdirectional = %q", got.Postdirectional)
	}
	if got.SecondaryDesignator != "#" || got.SecondaryNumber != "12" {
		t.Errorf("secondary = %q %q", got.SecondaryDesignator, got.SecondaryNumber)
	}
	if got.PrimaryNumber != "100" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Errorf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

func TestParseStreetHighway(t *testing.T) {
	raw := "3324 TN HIGHWAY 431\nSomewhere KY 40000"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "3324" {
		t.Errorf("PrimaryNumber = %q", got.PrimaryNumber)
	}
	if got.StreetName != "TN HIGHWAY 431" {
		t.Errorf("StreetName = %q, want TN HIGHWAY 431", got.StreetName)
	}
}

func TestParseThenNormalizeRoundTrip(t *testing.T) {
	raw := "123 North Main Street Apt 4\nSpringfield IL 62701"
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	norm, err := Normalize(parsed)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "123 N MAIN ST APT 4\nSPRINGFIELD IL 62701"
	if Format(norm) != want {
		t.Fatalf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParseUnknownTokenInRegionErrors(t *testing.T) {
	_, err := Parse("123 Main St\nSpringfield ZZ 62701")
	if err == nil {
		t.Fatal("expected error for bad region")
	}
}

func TestParseMilitaryRejectsCountryOnLastLine(t *testing.T) {
	// military package rejects extra tokens; Parse should not accept as military.
	// May fall through to civilian and fail region — either error is OK.
	_, err := Parse("PSC 3 BOX 4120\nAPO AE GERMANY 09021-0002")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseDirectionalsNE(t *testing.T) {
	raw := "101 Northeast Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Predirectional != "NE" {
		t.Errorf("Predirectional = %q, want NE", got.Predirectional)
	}
	if got.PrimaryNumber != "101" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Errorf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

func TestParseGluedHashSecondary(t *testing.T) {
	raw := "100 Main Street #12\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.SecondaryDesignator != "#" || got.SecondaryNumber != "12" {
		t.Errorf("secondary = %q %q", got.SecondaryDesignator, got.SecondaryNumber)
	}
}

func TestParseRuralRouteMultiline(t *testing.T) {
	raw := "Rural Route 91 Box A7\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "RR 91 BOX A7" {
		t.Errorf("StreetName = %q, want RR 91 BOX A7", got.StreetName)
	}
	if got.PrimaryNumber != "" || got.StreetSuffix != "" {
		t.Errorf("expected empty primary/suffix for RR, got primary=%q suffix=%q",
			got.PrimaryNumber, got.StreetSuffix)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" || got.Postal != "62701" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "RR 91 BOX A7\nSPRINGFIELD IL 62701"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParseRuralRouteVariants(t *testing.T) {
	cases := []struct {
		street string
		want   string
	}{
		{"RFD 61 #87b", "RR 61 BOX 87B"},
		{"RD 61 # 87b", "RR 61 BOX 87B"},
		{"RR0061 #87b", "RR 61 BOX 87B"},
	}
	for _, tc := range cases {
		raw := tc.street + "\nSpringfield IL 62701"
		got, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.street, err)
		}
		if got.StreetName != tc.want {
			t.Errorf("Parse(%q).StreetName = %q, want %q", tc.street, got.StreetName, tc.want)
		}
		norm, err := Normalize(got)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		wantFmt := tc.want + "\nSPRINGFIELD IL 62701"
		if Format(norm) != wantFmt {
			t.Errorf("Format = %q, want %q", Format(norm), wantFmt)
		}
	}
}

func TestParsePOBoxMultiline(t *testing.T) {
	cases := []struct {
		street string
		want   string
	}{
		{"Post office Box G", "PO BOX G"},
		{"PO Box 11890", "PO BOX 11890"},
	}
	for _, tc := range cases {
		raw := tc.street + "\nSpringfield IL 62701"
		got, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.street, err)
		}
		if got.StreetName != tc.want {
			t.Errorf("StreetName = %q, want %q", got.StreetName, tc.want)
		}
		if got.PrimaryNumber != "" || got.StreetSuffix != "" {
			t.Errorf("expected empty primary/suffix for PO Box, got primary=%q suffix=%q",
				got.PrimaryNumber, got.StreetSuffix)
		}
		norm, err := Normalize(got)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		wantFmt := tc.want + "\nSPRINGFIELD IL 62701"
		if Format(norm) != wantFmt {
			t.Errorf("Format = %q, want %q", Format(norm), wantFmt)
		}
	}
}

func TestParseRDHighwayNotRuralRoute(t *testing.T) {
	// Bare "RD 5A" is a highway-style street name, not rural route (no BOX/#).
	raw := "RD 5A\nSomewhere KY 40000"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName == "RR 5A BOX" || strings.HasPrefix(got.StreetName, "RR ") {
		t.Fatalf("RD 5A must not parse as rural route, got StreetName=%q", got.StreetName)
	}
	// Highway rewrite via parseStreetLine → ROAD 5A (no primary number).
	if got.StreetName != "ROAD 5A" {
		t.Errorf("StreetName = %q, want ROAD 5A", got.StreetName)
	}
}

func TestParseMultiTokenDirectionals(t *testing.T) {
	raw := "1011 South West Main Street North East Apt 12\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "1011" {
		t.Errorf("PrimaryNumber = %q, want 1011", got.PrimaryNumber)
	}
	if got.Predirectional != "SW" {
		t.Errorf("Predirectional = %q, want SW", got.Predirectional)
	}
	if got.StreetName != "MAIN" {
		t.Errorf("StreetName = %q, want MAIN", got.StreetName)
	}
	if got.StreetSuffix != "ST" {
		t.Errorf("StreetSuffix = %q, want ST", got.StreetSuffix)
	}
	if got.Postdirectional != "NE" {
		t.Errorf("Postdirectional = %q, want NE", got.Postdirectional)
	}
	if got.SecondaryDesignator != "APT" || got.SecondaryNumber != "12" {
		t.Errorf("secondary = %q %q, want APT 12", got.SecondaryDesignator, got.SecondaryNumber)
	}
}

func TestParseMultiTokenPredirectionalAbbreviated(t *testing.T) {
	// NORTH E (and N. E. after punctuation strip) merge to NE predirectional.
	for _, street := range []string{
		"3000 NORTH E MAIN STREET",
		"3000 N E MAIN STREET",
		"3000 N. E. MAIN STREET",
		"3000 NORTH EAST MAIN STREET",
	} {
		raw := street + "\nSpringfield IL 62701"
		got, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", street, err)
		}
		if got.Predirectional != "NE" {
			t.Errorf("Parse(%q): Predirectional = %q, want NE", street, got.Predirectional)
		}
		if got.PrimaryNumber != "3000" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
			t.Errorf("Parse(%q): street = primary=%q name=%q suffix=%q",
				street, got.PrimaryNumber, got.StreetName, got.StreetSuffix)
		}
	}
}

func TestParseDoesNotMergeOppositeDirectionals(t *testing.T) {
	// NORTH SOUTH / EAST WEST stay as street-name material (not compound dirs).
	raw := "123 North South Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Predirectional != "N" {
		t.Errorf("Predirectional = %q, want N (only first cardinal)", got.Predirectional)
	}
	if got.StreetName != "SOUTH MAIN" {
		t.Errorf("StreetName = %q, want SOUTH MAIN (unmerged opposite)", got.StreetName)
	}
	if got.StreetSuffix != "ST" {
		t.Errorf("StreetSuffix = %q, want ST", got.StreetSuffix)
	}
	if got.Postdirectional != "" {
		t.Errorf("Postdirectional = %q, want empty", got.Postdirectional)
	}
}

func TestParseLeadingSecondaryApartment(t *testing.T) {
	raw := "Apartment 3200 152 South Tech Drive\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "152" {
		t.Errorf("PrimaryNumber = %q, want 152", got.PrimaryNumber)
	}
	if got.Predirectional != "S" {
		t.Errorf("Predirectional = %q, want S", got.Predirectional)
	}
	if got.StreetName != "TECH" {
		t.Errorf("StreetName = %q, want TECH", got.StreetName)
	}
	if got.StreetSuffix != "DR" {
		t.Errorf("StreetSuffix = %q, want DR", got.StreetSuffix)
	}
	if got.SecondaryDesignator != "APT" || got.SecondaryNumber != "3200" {
		t.Errorf("secondary = %q %q, want APT 3200", got.SecondaryDesignator, got.SecondaryNumber)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "152 S TECH DR APT 3200\nMIAMI FL 33101"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParseLeadingHashNoPrimary(t *testing.T) {
	raw := "#3200 South Tech Drive\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "" {
		t.Errorf("PrimaryNumber = %q, want empty", got.PrimaryNumber)
	}
	if got.Predirectional != "S" {
		t.Errorf("Predirectional = %q, want S", got.Predirectional)
	}
	if got.StreetName != "TECH" {
		t.Errorf("StreetName = %q, want TECH", got.StreetName)
	}
	if got.StreetSuffix != "DR" {
		t.Errorf("StreetSuffix = %q, want DR", got.StreetSuffix)
	}
	if got.SecondaryDesignator != "#" || got.SecondaryNumber != "3200" {
		t.Errorf("secondary = %q %q, want # 3200", got.SecondaryDesignator, got.SecondaryNumber)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "S TECH DR # 3200\nMIAMI FL 33101"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParseLeadingHashWithPrimary(t *testing.T) {
	raw := "#3200 152 South Tech Drive\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "152" {
		t.Errorf("PrimaryNumber = %q, want 152", got.PrimaryNumber)
	}
	if got.Predirectional != "S" {
		t.Errorf("Predirectional = %q, want S", got.Predirectional)
	}
	if got.StreetName != "TECH" {
		t.Errorf("StreetName = %q, want TECH", got.StreetName)
	}
	if got.StreetSuffix != "DR" {
		t.Errorf("StreetSuffix = %q, want DR", got.StreetSuffix)
	}
	if got.SecondaryDesignator != "#" || got.SecondaryNumber != "3200" {
		t.Errorf("secondary = %q %q, want # 3200", got.SecondaryDesignator, got.SecondaryNumber)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "152 S TECH DR # 3200\nMIAMI FL 33101"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParseLeadingUnitWithTrailingUpper(t *testing.T) {
	raw := "Unit 3200 152 Tech Drive Upper\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "152" {
		t.Errorf("PrimaryNumber = %q, want 152", got.PrimaryNumber)
	}
	if got.StreetName != "TECH" {
		t.Errorf("StreetName = %q, want TECH", got.StreetName)
	}
	if got.StreetSuffix != "DR" {
		t.Errorf("StreetSuffix = %q, want DR", got.StreetSuffix)
	}
	if got.SecondaryDesignator != "UNIT" {
		t.Errorf("SecondaryDesignator = %q, want UNIT", got.SecondaryDesignator)
	}
	// Trailing non-numbered UPPER peels after UNIT 3200; appended for Format.
	if got.SecondaryNumber != "3200 UPPR" {
		t.Errorf("SecondaryNumber = %q, want %q", got.SecondaryNumber, "3200 UPPR")
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "152 TECH DR UNIT 3200 UPPR\nMIAMI FL 33101"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParseLeadingSecondaryNoNumberNotReordered(t *testing.T) {
	// Numbered designator without a following unit number must not invent one.
	// "Apartment" alone at the start should not be reordered as secondary+number.
	raw := "Apartment South Tech Drive\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Without a unit number, leading APT is not moved; SOUTH may be predirectional
	// or name tokens absorb APARTMENT.
	if got.SecondaryNumber != "" {
		t.Errorf("SecondaryNumber = %q, want empty (no invented number)", got.SecondaryNumber)
	}
}

func TestParseMilitaryNotBrokenByLeadingUnit(t *testing.T) {
	// Military "UNIT N BOX N" must stay on the military fast path.
	raw := "UNIT 2050 BOX 4190\nAPO AP 96278-2050"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "UNIT 2050 BOX 4190" {
		t.Errorf("StreetName = %q, want military street line intact", got.StreetName)
	}
	if got.SecondaryDesignator != "" || got.PrimaryNumber != "" {
		t.Errorf("expected no civilian secondary/primary peels on military, got sec=%q primary=%q",
			got.SecondaryDesignator, got.PrimaryNumber)
	}
}

func TestParseSameLineBusinessPreStreet(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		business   string
		primary    string
		predir     string
		streetName string
		suffix     string
		wantFmt    string // Format(Normalize) spot check; empty skips
	}{
		{
			name:       "williamson medical center",
			raw:        "Williamson Medical Center 3000 Edward Curd Lane\nSpringfield IL 62701",
			business:   "WILLIAMSON MEDICAL CENTER",
			primary:    "3000",
			streetName: "EDWARD CURD",
			suffix:     "LN",
			wantFmt:    "WILLIAMSON MEDICAL CENTER\n3000 EDWARD CURD LN\nSPRINGFIELD IL 62701",
		},
		{
			name:       "ucent building with predir and ordinal name",
			raw:        "UCENT Building 847 North 49th Street\nSpringfield IL 62701",
			business:   "UCENT BUILDING",
			primary:    "847",
			predir:     "N",
			streetName: "49TH",
			suffix:     "ST",
			wantFmt:    "UCENT BUILDING\n847 N 49TH ST\nSPRINGFIELD IL 62701",
		},
		{
			name:       "narrative lives in the tent near",
			raw:        "Lives in the tent near 155 North Main Street\nSpringfield IL 62701",
			business:   "LIVES IN THE TENT NEAR",
			primary:    "155",
			predir:     "N",
			streetName: "MAIN",
			suffix:     "ST",
			wantFmt:    "LIVES IN THE TENT NEAR\n155 N MAIN ST\nSPRINGFIELD IL 62701",
		},
		{
			name:       "center of hope",
			raw:        "Center of Hope 110 East 7th Street\nSpringfield IL 62701",
			business:   "CENTER OF HOPE",
			primary:    "110",
			predir:     "E",
			streetName: "7TH",
			suffix:     "ST",
			wantFmt:    "CENTER OF HOPE\n110 E 7TH ST\nSPRINGFIELD IL 62701",
		},
		// Ordinary streets that start with a primary must not invent BusinessName.
		{
			name:       "ordinary street no business",
			raw:        "123 Main Street\nSpringfield IL 62701",
			business:   "",
			primary:    "123",
			streetName: "MAIN",
			suffix:     "ST",
			wantFmt:    "123 MAIN ST\nSPRINGFIELD IL 62701",
		},
		{
			name:       "ordinary with predir no business",
			raw:        "847 North 49th Street\nSpringfield IL 62701",
			business:   "",
			primary:    "847",
			predir:     "N",
			streetName: "49TH",
			suffix:     "ST",
		},
		// Leading secondary reorder runs first — still no false pre-street.
		{
			name:       "leading secondary not pre-street",
			raw:        "Apartment 3200 152 South Tech Drive\nMiami FL 33101",
			business:   "",
			primary:    "152",
			predir:     "S",
			streetName: "TECH",
			suffix:     "DR",
		},
		// Multi-line business stays; same-line only fills when multi-line empty.
		{
			name:       "multi-line business preferred alone",
			raw:        "Acme Corp\n123 Main Street\nSpringfield IL 62701",
			business:   "ACME CORP",
			primary:    "123",
			streetName: "MAIN",
			suffix:     "ST",
		},
		// Both multi-line and same-line: same-line prepended to multi-line.
		{
			name:       "same-line prepends multi-line business",
			raw:        "Acme Corp\nBuilding A 123 Main Street\nSpringfield IL 62701",
			business:   "BUILDING A ACME CORP",
			primary:    "123",
			streetName: "MAIN",
			suffix:     "ST",
			wantFmt:    "BUILDING A ACME CORP\n123 MAIN ST\nSPRINGFIELD IL 62701",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.BusinessName != tc.business {
				t.Errorf("BusinessName = %q, want %q", got.BusinessName, tc.business)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.Predirectional != tc.predir {
				t.Errorf("Predirectional = %q, want %q", got.Predirectional, tc.predir)
			}
			if got.StreetName != tc.streetName {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.streetName)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
			if tc.wantFmt == "" {
				return
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFmt {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFmt)
			}
		})
	}
}

func TestParseMilitaryRRPONotBrokenByPreStreet(t *testing.T) {
	// Military / RR / PO rewrite paths run before pre-street and must stay intact.
	cases := []struct {
		raw, street string
	}{
		{"PSC 3 BOX 4120\nAPO AE 09021-0002", "PSC 3 BOX 4120"},
		{"Rural Route 91 Box A7\nSpringfield IL 62701", "RR 91 BOX A7"},
		{"PO Box 11890\nSpringfield IL 62701", "PO BOX 11890"},
	}
	for _, tc := range cases {
		got, err := Parse(tc.raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.raw, err)
		}
		if got.StreetName != tc.street {
			t.Errorf("Parse(%q).StreetName = %q, want %q", tc.raw, got.StreetName, tc.street)
		}
		if got.BusinessName != "" {
			t.Errorf("Parse(%q).BusinessName = %q, want empty", tc.raw, got.BusinessName)
		}
		if got.PrimaryNumber != "" {
			t.Errorf("Parse(%q).PrimaryNumber = %q, want empty", tc.raw, got.PrimaryNumber)
		}
	}
}

func TestParseDoubleSuffixAndStateAsStreetName(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		primary    string
		streetName string
		suffix     string
	}{
		// Double street suffix: rightmost is StreetSuffix; left stays in name as primary form.
		{
			name:       "avenue drive",
			raw:        "2000 Main Avenue Drive\nSpringfield IL 62701",
			primary:    "2000",
			streetName: "MAIN AVENUE",
			suffix:     "DR",
		},
		{
			name:       "pky ave expands parkway",
			raw:        "2002 Main Pky Ave\nSpringfield IL 62701",
			primary:    "2002",
			streetName: "MAIN PARKWAY",
			suffix:     "AVE",
		},
		{
			name:       "street road",
			raw:        "2004 Oak Street Road\nSpringfield IL 62701",
			primary:    "2004",
			streetName: "OAK STREET",
			suffix:     "RD",
		},
		// State as complete street name: spell out full US state when entire name is state.
		{
			name:       "OK avenue",
			raw:        "8000 OK Avenue\nOklahoma City OK 73101",
			primary:    "8000",
			streetName: "OKLAHOMA",
			suffix:     "AVE",
		},
		{
			name:       "CT drive",
			raw:        "8004 CT Drive\nHartford CT 06101",
			primary:    "8004",
			streetName: "CONNECTICUT",
			suffix:     "DR",
		},
		{
			// Last CT is suffix COURT; first CT is state CONNECTICUT.
			name:       "CT CT state then court",
			raw:        "8006 CT CT\nHartford CT 06101",
			primary:    "8006",
			streetName: "CONNECTICUT",
			suffix:     "CT",
		},
		{
			// Last WY is suffix WAY (not WY); first WY is state WYOMING.
			name:       "WY WY state then way",
			raw:        "8011 WY WY\nSpringfield IL 62701",
			primary:    "8011",
			streetName: "WYOMING",
			suffix:     "WAY",
		},
		{
			name:       "full state name already spelled",
			raw:        "8008 Oklahoma Avenue\nOklahoma City OK 73101",
			primary:    "8008",
			streetName: "OKLAHOMA",
			suffix:     "AVE",
		},
		{
			name:       "multi-word state new york",
			raw:        "8010 New York Avenue\nAlbany NY 12201",
			primary:    "8010",
			streetName: "NEW YORK",
			suffix:     "AVE",
		},
		// State as portion of street name / highway path: do not force full spell.
		{
			name:       "state highway abbrev preserved",
			raw:        "8105 TN 431\nSomewhere KY 40000",
			primary:    "8105",
			streetName: "TN HIGHWAY 431",
			suffix:     "",
		},
		{
			name:       "state portion multi-word not only-state",
			raw:        "8106 OK Main Street\nOklahoma City OK 73101",
			primary:    "8106",
			streetName: "OK MAIN",
			suffix:     "ST",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.StreetName != tc.streetName {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.streetName)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
		})
	}
}

func TestParseGridStyleDoubleDirectionals(t *testing.T) {
	// Salt Lake City–style grid addresses: numeric street name, pre+post
	// directionals, no street suffix required.
	tests := []struct {
		name    string
		raw     string
		primary string
		predir  string
		street  string
		postdir string
		wantFmt string
	}{
		{
			name:    "spelled directionals",
			raw:     "1016 East 1700 South\nSalt Lake City UT 84105",
			primary: "1016",
			predir:  "E",
			street:  "1700",
			postdir: "S",
			wantFmt: "1016 E 1700 S\nSALT LAKE CITY UT 84105",
		},
		{
			name:    "abbreviated directionals",
			raw:     "842 E 1700 S\nSalt Lake City UT 84105",
			primary: "842",
			predir:  "E",
			street:  "1700",
			postdir: "S",
			wantFmt: "842 E 1700 S\nSALT LAKE CITY UT 84105",
		},
		{
			name:    "north west grid",
			raw:     "500 North 300 West\nSalt Lake City UT 84103",
			primary: "500",
			predir:  "N",
			street:  "300",
			postdir: "W",
			wantFmt: "500 N 300 W\nSALT LAKE CITY UT 84103",
		},
		{
			name:    "compound predir still merges",
			raw:     "100 Northeast 200 South\nSalt Lake City UT 84111",
			primary: "100",
			predir:  "NE",
			street:  "200",
			postdir: "S",
			wantFmt: "100 NE 200 S\nSALT LAKE CITY UT 84111",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.Predirectional != tc.predir {
				t.Errorf("Predirectional = %q, want %q", got.Predirectional, tc.predir)
			}
			if got.StreetName != tc.street {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.street)
			}
			if got.StreetSuffix != "" {
				t.Errorf("StreetSuffix = %q, want empty (grid has no suffix)", got.StreetSuffix)
			}
			if got.Postdirectional != tc.postdir {
				t.Errorf("Postdirectional = %q, want %q", got.Postdirectional, tc.postdir)
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFmt {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFmt)
			}
		})
	}
}

func TestParseFractionalPrimary(t *testing.T) {
	// Fractional house numbers: PrimaryNumber is the integer portion; the
	// fraction token stays in StreetName with slash retained (KeepSlash).
	// Choice: Primary="123", StreetName="1/2 MAIN" (not Primary="123 1/2").
	tests := []struct {
		name       string
		raw        string
		primary    string
		streetName string
		suffix     string
		wantFmt    string
	}{
		{
			name:       "half main street",
			raw:        "123 1/2 Main Street\nSpringfield IL 62701",
			primary:    "123",
			streetName: "1/2 MAIN",
			suffix:     "ST",
			wantFmt:    "123 1/2 MAIN ST\nSPRINGFIELD IL 62701",
		},
		{
			// Format order is PRIMARY PREDIR STREET, so the fraction (in
			// StreetName) appears after the predirectional field.
			name:       "quarter with predirectional after fraction",
			raw:        "45 1/4 North Oak Avenue\nSpringfield IL 62701",
			primary:    "45",
			streetName: "1/4 OAK",
			suffix:     "AVE",
			wantFmt:    "45 N 1/4 OAK AVE\nSPRINGFIELD IL 62701",
		},
		{
			name:       "half with abbreviated predir",
			raw:        "123 1/2 N Main Street\nSpringfield IL 62701",
			primary:    "123",
			streetName: "1/2 MAIN",
			suffix:     "ST",
			wantFmt:    "123 N 1/2 MAIN ST\nSPRINGFIELD IL 62701",
		},
		{
			name:       "fraction with secondary",
			raw:        "123 1/2 Main Street Apt 4\nSpringfield IL 62701",
			primary:    "123",
			streetName: "1/2 MAIN",
			suffix:     "ST",
			wantFmt:    "123 1/2 MAIN ST APT 4\nSPRINGFIELD IL 62701",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.StreetName != tc.streetName {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.streetName)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
			// Slash must survive Parse (KeepSlash) and Normalize free-text.
			if !strings.Contains(got.StreetName, "/") {
				t.Errorf("StreetName %q lost fractional slash", got.StreetName)
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFmt {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFmt)
			}
		})
	}
}

func TestParseMultiSecondaryTrailingUnits(t *testing.T) {
	// Multiple trailing secondaries peel right-to-left repeatedly and combine
	// into SecondaryDesignator (leftmost) + SecondaryNumber (remainder) so
	// Format yields "BLDG 420 RM 120".
	tests := []struct {
		name    string
		raw     string
		primary string
		street  string
		suffix  string
		secDes  string
		secNum  string
		wantFmt string
	}{
		{
			name:    "building and room",
			raw:     "450 Jane Stanford Way Building 420 Room 120\nSpringfield IL 62701",
			primary: "450",
			street:  "JANE STANFORD",
			suffix:  "WAY", // Way suffix prefers WAY (not WY)
			secDes:  "BLDG",
			secNum:  "420 RM 120",
			wantFmt: "450 JANE STANFORD WAY BLDG 420 RM 120\nSPRINGFIELD IL 62701",
		},
		{
			name:    "suite and floor",
			raw:     "100 Main Street Suite 200 Floor 3\nSpringfield IL 62701",
			primary: "100",
			street:  "MAIN",
			suffix:  "ST",
			secDes:  "STE",
			secNum:  "200 FL 3",
			wantFmt: "100 MAIN ST STE 200 FL 3\nSPRINGFIELD IL 62701",
		},
		{
			name:    "building room upper",
			raw:     "450 Jane Stanford Way Building 420 Room 120 Upper\nSpringfield IL 62701",
			primary: "450",
			street:  "JANE STANFORD",
			suffix:  "WAY",
			secDes:  "BLDG",
			secNum:  "420 RM 120 UPPR",
			wantFmt: "450 JANE STANFORD WAY BLDG 420 RM 120 UPPR\nSPRINGFIELD IL 62701",
		},
		{
			name:    "single secondary still works",
			raw:     "450 Jane Stanford Way Building 420\nSpringfield IL 62701",
			primary: "450",
			street:  "JANE STANFORD",
			suffix:  "WAY",
			secDes:  "BLDG",
			secNum:  "420",
			wantFmt: "450 JANE STANFORD WAY BLDG 420\nSPRINGFIELD IL 62701",
		},
		{
			name:    "alpha unit numbers",
			raw:     "10 Oak Drive Building A Room B\nSpringfield IL 62701",
			primary: "10",
			street:  "OAK",
			suffix:  "DR",
			secDes:  "BLDG",
			secNum:  "A RM B",
			wantFmt: "10 OAK DR BLDG A RM B\nSPRINGFIELD IL 62701",
		},
		{
			name:    "hash after building",
			raw:     "10 Main Street Building 5 # 12\nSpringfield IL 62701",
			primary: "10",
			street:  "MAIN",
			suffix:  "ST",
			secDes:  "BLDG",
			secNum:  "5 # 12",
			wantFmt: "10 MAIN ST BLDG 5 # 12\nSPRINGFIELD IL 62701",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.StreetName != tc.street {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.street)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
			if got.SecondaryDesignator != tc.secDes {
				t.Errorf("SecondaryDesignator = %q, want %q", got.SecondaryDesignator, tc.secDes)
			}
			if got.SecondaryNumber != tc.secNum {
				t.Errorf("SecondaryNumber = %q, want %q", got.SecondaryNumber, tc.secNum)
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFmt {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFmt)
			}
		})
	}
}

func TestParseHyphenatedPrimary(t *testing.T) {
	// NYC-style hyphenated house numbers keep the hyphen (KeepHyphen).
	tests := []struct {
		name    string
		raw     string
		primary string
		street  string
		suffix  string
		wantFmt string
	}{
		{
			name:    "bronx road",
			raw:     "112-10 Bronx Road\nBronx NY 10475",
			primary: "112-10",
			street:  "BRONX",
			suffix:  "RD",
			wantFmt: "112-10 BRONX RD\nBRONX NY 10475",
		},
		{
			name:    "with predirectional",
			raw:     "35-11 35th Avenue\nAstoria NY 11106",
			primary: "35-11",
			street:  "35TH",
			suffix:  "AVE",
			wantFmt: "35-11 35TH AVE\nASTORIA NY 11106",
		},
		{
			name:    "with secondary",
			raw:     "112-10 Bronx Road Apt 3B\nBronx NY 10475",
			primary: "112-10",
			street:  "BRONX",
			suffix:  "RD",
			wantFmt: "112-10 BRONX RD APT 3B\nBRONX NY 10475",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.StreetName != tc.street {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.street)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
			if !strings.Contains(got.PrimaryNumber, "-") {
				t.Errorf("PrimaryNumber %q lost hyphen", got.PrimaryNumber)
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFmt {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFmt)
			}
		})
	}
}

func TestParseSpecialFormsNotBrokenByGridMultiSecondary(t *testing.T) {
	// Military / RR / PO Box / leading secondary must remain intact after
	// multi-secondary and grid parsing changes.
	tests := []struct {
		name    string
		raw     string
		wantFmt string
		check   func(t *testing.T, got Address)
	}{
		{
			name:    "military unit box",
			raw:     "UNIT 2050 BOX 4190\nAPO AP 96278-2050",
			wantFmt: "UNIT 2050 BOX 4190\nAPO AP 96278-2050",
			check: func(t *testing.T, got Address) {
				if got.StreetName != "UNIT 2050 BOX 4190" {
					t.Errorf("StreetName = %q", got.StreetName)
				}
				if got.SecondaryDesignator != "" || got.PrimaryNumber != "" {
					t.Errorf("military must not civilian-peel, sec=%q primary=%q",
						got.SecondaryDesignator, got.PrimaryNumber)
				}
			},
		},
		{
			name:    "rural route",
			raw:     "RR 2 BOX 152\nAnytown KY 40000",
			wantFmt: "RR 2 BOX 152\nANYTOWN KY 40000",
			check: func(t *testing.T, got Address) {
				if got.StreetName != "RR 2 BOX 152" {
					t.Errorf("StreetName = %q", got.StreetName)
				}
			},
		},
		{
			name:    "po box",
			raw:     "PO BOX 123\nAnytown KY 40000",
			wantFmt: "PO BOX 123\nANYTOWN KY 40000",
			check: func(t *testing.T, got Address) {
				if got.StreetName != "PO BOX 123" {
					t.Errorf("StreetName = %q", got.StreetName)
				}
			},
		},
		{
			name:    "leading secondary apartment",
			raw:     "APT 4 123 MAIN ST\nSpringfield IL 62701",
			wantFmt: "123 MAIN ST APT 4\nSPRINGFIELD IL 62701",
			check: func(t *testing.T, got Address) {
				if got.PrimaryNumber != "123" || got.SecondaryDesignator != "APT" || got.SecondaryNumber != "4" {
					t.Errorf("primary=%q sec=%q %q", got.PrimaryNumber, got.SecondaryDesignator, got.SecondaryNumber)
				}
			},
		},
		{
			name:    "unit with trailing upper",
			raw:     "Unit 3200 152 Tech Drive Upper\nMiami FL 33101",
			wantFmt: "152 TECH DR UNIT 3200 UPPR\nMIAMI FL 33101",
			check: func(t *testing.T, got Address) {
				if got.SecondaryDesignator != "UNIT" || got.SecondaryNumber != "3200 UPPR" {
					t.Errorf("secondary = %q %q", got.SecondaryDesignator, got.SecondaryNumber)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFmt {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFmt)
			}
		})
	}
}

func TestParseDirectionalAsStreetName(t *testing.T) {
	// When the only name token after peeling suffix is a directional, keep it as
	// StreetName (spelled out via NormalizeDirectional), not Predirectional.
	// Contrast: "1005 South Main Street" → Predir S, name MAIN, ST.
	tests := []struct {
		name       string
		raw        string
		primary    string
		predir     string
		streetName string
		suffix     string
		postdir    string
	}{
		{
			name:       "south avenue spelled",
			raw:        "1005 South Avenue\nSpringfield IL 62701",
			primary:    "1005",
			streetName: "SOUTH",
			suffix:     "AVE",
		},
		{
			name:       "s avenue expands to south",
			raw:        "1005 S Avenue\nSpringfield IL 62701",
			primary:    "1005",
			streetName: "SOUTH",
			suffix:     "AVE",
		},
		{
			name:       "west boulevard",
			raw:        "1005 West Boulevard\nSpringfield IL 62701",
			primary:    "1005",
			streetName: "WEST",
			suffix:     "BLVD",
		},
		{
			name:       "n street expands to north",
			raw:        "1005 N Street\nSpringfield IL 62701",
			primary:    "1005",
			streetName: "NORTH",
			suffix:     "ST",
		},
		{
			name:       "south main still predirectional",
			raw:        "1005 South Main Street\nSpringfield IL 62701",
			primary:    "1005",
			predir:     "S",
			streetName: "MAIN",
			suffix:     "ST",
		},
		{
			name:       "ne avenue compound expands",
			raw:        "1005 NE Avenue\nSpringfield IL 62701",
			primary:    "1005",
			streetName: "NORTHEAST",
			suffix:     "AVE",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.Predirectional != tc.predir {
				t.Errorf("Predirectional = %q, want %q", got.Predirectional, tc.predir)
			}
			if got.StreetName != tc.streetName {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.streetName)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
			if got.Postdirectional != tc.postdir {
				t.Errorf("Postdirectional = %q, want %q", got.Postdirectional, tc.postdir)
			}
		})
	}
}

func TestParseAvenueLetterPostdir(t *testing.T) {
	// "1001 Avenue E" → Primary 1001, StreetName AVENUE, Postdir E (suffix not
	// peeled when it would leave an empty name). C# Format: "1001 AVENUE E".
	tests := []struct {
		name       string
		raw        string
		primary    string
		streetName string
		suffix     string
		postdir    string
		wantFormat string
	}{
		{
			name:       "avenue e",
			raw:        "1001 Avenue E\nSpringfield IL 62701",
			primary:    "1001",
			streetName: "AVENUE",
			postdir:    "E",
			wantFormat: "1001 AVENUE E\nSPRINGFIELD IL 62701",
		},
		{
			name:       "ave e",
			raw:        "1001 Ave E\nSpringfield IL 62701",
			primary:    "1001",
			streetName: "AVENUE",
			postdir:    "E",
			wantFormat: "1001 AVENUE E\nSPRINGFIELD IL 62701",
		},
		{
			name:       "avenue east",
			raw:        "1001 Avenue East\nSpringfield IL 62701",
			primary:    "1001",
			streetName: "AVENUE",
			postdir:    "E",
			wantFormat: "1001 AVENUE E\nSPRINGFIELD IL 62701",
		},
		{
			name:       "main avenue e keeps suffix",
			raw:        "100 Main Avenue E\nSpringfield IL 62701",
			primary:    "100",
			streetName: "MAIN",
			suffix:     "AVE",
			postdir:    "E",
			wantFormat: "100 MAIN AVE E\nSPRINGFIELD IL 62701",
		},
		{
			name:       "avenue e no primary",
			raw:        "Avenue E\nSpringfield IL 62701",
			streetName: "AVENUE",
			postdir:    "E",
			wantFormat: "AVENUE E\nSPRINGFIELD IL 62701",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.PrimaryNumber != tc.primary {
				t.Errorf("PrimaryNumber = %q, want %q", got.PrimaryNumber, tc.primary)
			}
			if got.StreetName != tc.streetName {
				t.Errorf("StreetName = %q, want %q", got.StreetName, tc.streetName)
			}
			if got.StreetSuffix != tc.suffix {
				t.Errorf("StreetSuffix = %q, want %q", got.StreetSuffix, tc.suffix)
			}
			if got.Postdirectional != tc.postdir {
				t.Errorf("Postdirectional = %q, want %q", got.Postdirectional, tc.postdir)
			}
			norm, err := Normalize(got)
			if err != nil {
				t.Fatalf("Normalize: %v", err)
			}
			if Format(norm) != tc.wantFormat {
				t.Errorf("Format = %q, want %q", Format(norm), tc.wantFormat)
			}
		})
	}
}

func TestParsePoundSecondarySpacing(t *testing.T) {
	// After parse+normalize+format, # secondary appears as "# 12" (space via
	// joinNonEmpty). Both glued and spaced inputs converge.
	for _, street := range []string{
		"100 Main Street #12",
		"100 Main Street # 12",
	} {
		raw := street + "\nMiami FL 33101"
		got, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", street, err)
		}
		if got.SecondaryDesignator != "#" || got.SecondaryNumber != "12" {
			t.Errorf("Parse(%q): secondary = %q %q, want # 12",
				street, got.SecondaryDesignator, got.SecondaryNumber)
		}
		norm, err := Normalize(got)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		want := "100 MAIN ST # 12\nMIAMI FL 33101"
		if Format(norm) != want {
			t.Errorf("Parse(%q): Format = %q, want %q", street, Format(norm), want)
		}
	}
}

func TestParsePuertoRicoCalleTrailingNumber(t *testing.T) {
	// PR order: Spanish type before name, house number trailing.
	raw := "Calle Luna 123\nSan Juan PR 00901"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" {
		t.Errorf("PrimaryNumber = %q, want 123", got.PrimaryNumber)
	}
	if got.StreetName != "CALLE LUNA" {
		t.Errorf("StreetName = %q, want CALLE LUNA", got.StreetName)
	}
	if got.StreetSuffix != "" {
		t.Errorf("StreetSuffix = %q, want empty (type is in StreetName)", got.StreetSuffix)
	}
	if got.City != "SAN JUAN" || got.Region != "PR" || got.Postal != "00901" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "123 CALLE LUNA\nSAN JUAN PR 00901"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParsePuertoRicoCalleUSOrder(t *testing.T) {
	// US order with Spanish type: number first, type before or after name.
	raw := "123 Calle Luna\nSan Juan PR 00901"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "CALLE LUNA" || got.StreetSuffix != "" {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "123 CALLE LUNA\nSAN JUAN PR 00901"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParsePuertoRicoAvenida(t *testing.T) {
	raw := "Avenida Ponce de Leon 456\nSan Juan PR 00907"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "456" {
		t.Errorf("PrimaryNumber = %q, want 456", got.PrimaryNumber)
	}
	if got.StreetName != "AVENIDA PONCE DE LEON" {
		t.Errorf("StreetName = %q, want AVENIDA PONCE DE LEON", got.StreetName)
	}
	if got.StreetSuffix != "" {
		t.Errorf("StreetSuffix = %q, want empty (type is in StreetName)", got.StreetSuffix)
	}
}

func TestParsePuertoRicoSecondaryURBAndApartamento(t *testing.T) {
	raw := "Calle Luna 123 Apartamento 4\nSan Juan PR 00901"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "CALLE LUNA" || got.StreetSuffix != "" {
		t.Errorf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
	if got.SecondaryDesignator != "APT" || got.SecondaryNumber != "4" {
		t.Errorf("secondary = %q %q, want APT 4", got.SecondaryDesignator, got.SecondaryNumber)
	}

	rawURB := "123 Calle Luna URB\nPonce PR 00716"
	gotURB, err := Parse(rawURB)
	if err != nil {
		t.Fatalf("Parse URB: %v", err)
	}
	if gotURB.SecondaryDesignator != "URB" {
		t.Errorf("SecondaryDesignator = %q, want URB", gotURB.SecondaryDesignator)
	}
	if gotURB.PrimaryNumber != "123" || gotURB.StreetName != "CALLE LUNA" || gotURB.StreetSuffix != "" {
		t.Errorf("street = primary=%q name=%q suffix=%q", gotURB.PrimaryNumber, gotURB.StreetName, gotURB.StreetSuffix)
	}
	norm, err := Normalize(gotURB)
	if err != nil {
		t.Fatalf("Normalize URB: %v", err)
	}
	want := "123 CALLE LUNA URB\nPONCE PR 00716"
	if Format(norm) != want {
		t.Errorf("Format = %q, want %q", Format(norm), want)
	}
}

func TestParsePuertoRicoApartado(t *testing.T) {
	raw := "Apartado 11890\nSan Juan PR 00901"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "PO BOX 11890" {
		t.Errorf("StreetName = %q, want PO BOX 11890", got.StreetName)
	}
	if got.PrimaryNumber != "" || got.StreetSuffix != "" {
		t.Errorf("expected empty primary/suffix for Apartado, got primary=%q suffix=%q",
			got.PrimaryNumber, got.StreetSuffix)
	}
}

func TestParseNonPRDoesNotForceSpanishStreetType(t *testing.T) {
	// CLL is PR-specific; without region PR it must not peel as suffix.
	// English street still works on non-PR.
	raw := "123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}

	// A free-text line with Spanish type on a non-PR region leaves the type in the name.
	rawES := "123 Calle Luna\nSpringfield IL 62701"
	gotES, err := Parse(rawES)
	if err != nil {
		t.Fatalf("Parse non-PR Spanish: %v", err)
	}
	if gotES.StreetSuffix == "CLL" {
		t.Errorf("non-PR must not peel Spanish CLL as StreetSuffix, got %+v", gotES)
	}
	// Calle stays in street name (or residual); primary still peels.
	if gotES.PrimaryNumber != "123" {
		t.Errorf("PrimaryNumber = %q, want 123", gotES.PrimaryNumber)
	}
	if !strings.Contains(gotES.StreetName, "CALLE") && !strings.Contains(gotES.StreetName, "LUNA") {
		t.Errorf("StreetName = %q, expected CALLE/LUNA residual", gotES.StreetName)
	}
}

func TestParseSingleLineMultiWordCity(t *testing.T) {
	// PL-002: city may span multiple tokens on a single-line address.
	raw := "123 Main Street New York NY 10005"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.City != "NEW YORK" {
		t.Fatalf("City = %q, want NEW YORK", got.City)
	}
	if got.Region != "NY" || got.Postal != "10005" {
		t.Fatalf("region/postal = %q %q", got.Region, got.Postal)
	}
	if got.PrimaryNumber != "123" || got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if FormatStreetLine(norm) != "123 MAIN ST" {
		t.Fatalf("FormatStreetLine = %q, want 123 MAIN ST", FormatStreetLine(norm))
	}
}

func TestParseSingleLineSingleWordCityStillWorks(t *testing.T) {
	raw := "123 Main Street Springfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" || got.Postal != "62701" {
		t.Fatalf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
	if FormatStreetLine(got) != "123 MAIN ST" && !(got.PrimaryNumber == "123" && got.StreetName == "MAIN" && got.StreetSuffix == "ST") {
		t.Fatalf("street = primary=%q name=%q suffix=%q", got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
}

func TestParseCompactCanadianPostal(t *testing.T) {
	// PL-003a: compact single-token Canadian postal K1A0B1.
	raw := "10 Wellington Street\nOttawa ON K1A0B1"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Region != "ON" {
		t.Fatalf("Region = %q, want ON", got.Region)
	}
	if got.City != "OTTAWA" {
		t.Fatalf("City = %q, want OTTAWA", got.City)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if strings.ReplaceAll(norm.Postal, " ", "") != "K1A0B1" {
		t.Fatalf("Postal = %q, want K1A 0B1 (or compact equivalent)", norm.Postal)
	}
	if norm.Postal != "K1A 0B1" {
		t.Fatalf("Postal = %q, want spaced K1A 0B1", norm.Postal)
	}
}

func TestParseTrailingCountryOnLastLine(t *testing.T) {
	// PL-003b: trailing USA must not pollute ZIP.
	raw := "123 Main Street\nSpringfield IL 62701 USA"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Postal != "62701" {
		t.Fatalf("Postal = %q, want 62701", got.Postal)
	}
	if strings.Contains(got.Postal, "USA") {
		t.Fatalf("Postal still contains USA: %q", got.Postal)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" {
		t.Fatalf("city/region = %q %q", got.City, got.Region)
	}
	// Country may be captured or dropped.
	if got.Country != "" && got.Country != "USA" {
		t.Fatalf("Country = %q, want USA or empty", got.Country)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if norm.Postal != "62701" {
		t.Fatalf("Normalize Postal = %q, want 62701", norm.Postal)
	}
}

func TestParseTrailingUnitedStatesOnLastLine(t *testing.T) {
	raw := "123 Main Street\nSpringfield IL 62701 UNITED STATES"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Postal != "62701" {
		t.Fatalf("Postal = %q, want 62701", got.Postal)
	}
	if got.Country != "" && got.Country != "UNITED STATES" {
		t.Fatalf("Country = %q", got.Country)
	}
}

// Permanent regressions for PL-008 / PL-009 / PL-022 (state-as-name WAY,
// mid-name directional expansion, KEY as street suffix not secondary).
func TestParseWyomingWayStateAsName(t *testing.T) {
	// "WY WY" → state WYOMING + suffix WAY (not WY/WY).
	got, err := Parse("8011 WY WY\nSpringfield IL 62701")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "8011" || got.StreetName != "WYOMING" || got.StreetSuffix != "WAY" {
		t.Fatalf("street = primary=%q name=%q suffix=%q, want 8011/WYOMING/WAY",
			got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
	if FormatStreetLine(got) != "8011 WYOMING WAY" {
		t.Fatalf("FormatStreetLine = %q, want 8011 WYOMING WAY", FormatStreetLine(got))
	}
}

func TestParseMidNameDirectionalExpanded(t *testing.T) {
	// Mid-name single-letter directional expands: BAY W DRIVE → BAY WEST DR.
	got, err := Parse("1014 BAY W DRIVE\nSpringfield IL 62701")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "1014" || got.StreetName != "BAY WEST" || got.StreetSuffix != "DR" {
		t.Fatalf("street = primary=%q name=%q suffix=%q, want 1014/BAY WEST/DR",
			got.PrimaryNumber, got.StreetName, got.StreetSuffix)
	}
	if got.Predirectional != "" || got.Postdirectional != "" {
		t.Fatalf("unexpected pre/post dir: pre=%q post=%q", got.Predirectional, got.Postdirectional)
	}
	if FormatStreetLine(got) != "1014 BAY WEST DR" {
		t.Fatalf("FormatStreetLine = %q, want 1014 BAY WEST DR", FormatStreetLine(got))
	}
}

func TestParseKeyAsStreetSuffixNotSecondary(t *testing.T) {
	// Trailing KEY is street suffix KY, not secondary unit KEY.
	got, err := Parse("8007 EAST KENTUCKY KEY\nSpringfield IL 62701")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "8007" || got.Predirectional != "E" ||
		got.StreetName != "KENTUCKY" || got.StreetSuffix != "KY" {
		t.Fatalf("street = primary=%q pre=%q name=%q suffix=%q, want 8007/E/KENTUCKY/KY",
			got.PrimaryNumber, got.Predirectional, got.StreetName, got.StreetSuffix)
	}
	if got.SecondaryDesignator != "" || got.SecondaryNumber != "" {
		t.Fatalf("unexpected secondary: des=%q num=%q", got.SecondaryDesignator, got.SecondaryNumber)
	}
	if FormatStreetLine(got) != "8007 E KENTUCKY KY" {
		t.Fatalf("FormatStreetLine = %q, want 8007 E KENTUCKY KY", FormatStreetLine(got))
	}
	// Numbered KEY still peels as secondary unit.
	got2, err := Parse("100 Main Street Key 12\nSpringfield IL 62701")
	if err != nil {
		t.Fatalf("Parse Key 12: %v", err)
	}
	if got2.SecondaryDesignator != "KEY" || got2.SecondaryNumber != "12" {
		t.Fatalf("Key 12 secondary = des=%q num=%q, want KEY/12",
			got2.SecondaryDesignator, got2.SecondaryNumber)
	}
	if got2.StreetSuffix != "ST" || got2.StreetName != "MAIN" {
		t.Fatalf("Key 12 street = name=%q suffix=%q", got2.StreetName, got2.StreetSuffix)
	}
}

func TestParseGridRoadDecimalExpandsRD(t *testing.T) {
	// C# StreetLineNormalizer: "rd 39.4" → "ROAD 39.4"
	raw := "9028 rd 39.4\nSpringfield IL 62701"
	p, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	n, err := Normalize(p)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	got := FormatStreetLine(n)
	if got != "9028 ROAD 39.4" {
		t.Fatalf("FormatStreetLine = %q, want 9028 ROAD 39.4", got)
	}
}
