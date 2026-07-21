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
