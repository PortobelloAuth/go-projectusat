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
	if got.StreetName != "123 MAIN STREET" {
		t.Fatalf("StreetName = %q (Task 3 stores full street line)", got.StreetName)
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
	if got.StreetName != "123 MAIN STREET" {
		t.Errorf("StreetName = %q", got.StreetName)
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
}
