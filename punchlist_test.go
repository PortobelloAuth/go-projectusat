//go:build punchlist

package goprojectusat

import (
	"strings"
	"testing"
)

// Punchlist acceptance tests — expected to FAIL until the linked PL item is fixed.
// Run: go test -tags punchlist . -run 'TestPunchlist_' -v -count=1

func parseNorm(t *testing.T, raw string) Address {
	t.Helper()
	p, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse(%q): %v", raw, err)
	}
	n, err := Normalize(p)
	if err != nil {
		t.Fatalf("Normalize(%+v): %v", p, err)
	}
	return n
}

func formatStreet(t *testing.T, raw string) string {
	t.Helper()
	return FormatStreetLine(parseNorm(t, raw))
}

// --- Blockers ---

func TestPunchlist_PL001_SpaceSeparatedMilitary(t *testing.T) {
	// Single logical line (spaces only) must still parse overseas military.
	raw := "PSC 3 BOX 4120 APO AE 09021-0002"
	n := parseNorm(t, raw)
	if n.StreetName != "PSC 3 BOX 4120" {
		t.Fatalf("StreetName = %q, want PSC 3 BOX 4120 (not business/pre-street split)", n.StreetName)
	}
	if n.BusinessName != "" {
		t.Fatalf("BusinessName = %q, want empty", n.BusinessName)
	}
	if n.City != "APO" || n.Region != "AE" || n.Postal != "09021-0002" {
		t.Fatalf("last line = %q %q %q", n.City, n.Region, n.Postal)
	}
	if Format(n) != "PSC 3 BOX 4120\nAPO AE 09021-0002" {
		t.Fatalf("Format = %q", Format(n))
	}
}

func TestPunchlist_PL002_SingleLineMultiWordCity(t *testing.T) {
	raw := "123 Main Street New York NY 10005"
	n := parseNorm(t, raw)
	if n.City != "NEW YORK" {
		t.Fatalf("City = %q, want NEW YORK", n.City)
	}
	if n.Region != "NY" || n.Postal != "10005" {
		t.Fatalf("region/postal = %q %q", n.Region, n.Postal)
	}
	if FormatStreetLine(n) != "123 MAIN ST" {
		t.Fatalf("street = %q, want 123 MAIN ST", FormatStreetLine(n))
	}
}

func TestPunchlist_PL003a_CompactCanadianPostal(t *testing.T) {
	raw := "10 Wellington Street\nOttawa ON K1A0B1"
	n := parseNorm(t, raw)
	if n.Region != "ON" {
		t.Fatalf("Region = %q, want ON", n.Region)
	}
	// Content form: spaced Canadian postal.
	got := strings.ReplaceAll(n.Postal, " ", "")
	if got != "K1A0B1" {
		t.Fatalf("Postal = %q, want K1A 0B1 (or compact K1A0B1 normalized)", n.Postal)
	}
	if !strings.Contains(n.Postal, "K1A") {
		t.Fatalf("Postal = %q, missing K1A", n.Postal)
	}
}

func TestPunchlist_PL003b_TrailingCountryOnLastLine(t *testing.T) {
	raw := "123 Main Street\nSpringfield IL 62701 USA"
	n := parseNorm(t, raw)
	if n.Postal != "62701" {
		t.Fatalf("Postal = %q, want 62701 (country must not pollute ZIP)", n.Postal)
	}
	// Country may land in Country field or be dropped; must not stay in postal.
	if strings.Contains(n.Postal, "USA") {
		t.Fatalf("Postal still contains USA: %q", n.Postal)
	}
}

// --- Majors ---

func TestPunchlist_PL004a_CommaInLastLine(t *testing.T) {
	raw := "123 Main Street\nSpringfield, IL 62701"
	n := parseNorm(t, raw)
	if n.City != "SPRINGFIELD" || n.Region != "IL" || n.Postal != "62701" {
		t.Fatalf("last = city=%q region=%q postal=%q", n.City, n.Region, n.Postal)
	}
}

func TestPunchlist_PL004b_CommaBeforeApt(t *testing.T) {
	raw := "123 Main Street, Apt 4\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	if FormatStreetLine(n) != "123 MAIN ST APT 4" {
		t.Fatalf("street = %q, want 123 MAIN ST APT 4", FormatStreetLine(n))
	}
}

func TestPunchlist_PL005_DirectionalOnlyStreetName(t *testing.T) {
	raw := "123 South\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	// Predir empty; SOUTH is the street name (directional-as-name).
	if n.PrimaryNumber != "123" {
		t.Fatalf("PrimaryNumber = %q", n.PrimaryNumber)
	}
	if n.StreetName != "SOUTH" {
		t.Fatalf("StreetName = %q, want SOUTH", n.StreetName)
	}
	if n.Predirectional != "" {
		t.Fatalf("Predirectional = %q, want empty", n.Predirectional)
	}
}

func TestPunchlist_PL006_DigitLeadingBusiness(t *testing.T) {
	raw := "3M Corporation 100 Main Street\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	if n.PrimaryNumber != "100" {
		t.Fatalf("PrimaryNumber = %q, want 100 (not 3M)", n.PrimaryNumber)
	}
	if !strings.Contains(n.BusinessName, "3M") {
		t.Fatalf("BusinessName = %q, want to contain 3M", n.BusinessName)
	}
	if FormatStreetLine(n) != "100 MAIN ST" {
		t.Fatalf("street = %q", FormatStreetLine(n))
	}
}

func TestPunchlist_PL007_GridDecimalPeriod(t *testing.T) {
	raw := "123 Road 39.4\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	// Period in grid designator must survive (C#: ROAD 39.4).
	if !strings.Contains(n.StreetName, "39.4") && !strings.Contains(FormatStreetLine(n), "39.4") {
		t.Fatalf("street lost decimal: name=%q format=%q", n.StreetName, FormatStreetLine(n))
	}
}

func TestPunchlist_PL008_WyomingWay(t *testing.T) {
	// C#: "8011 WY WY" → "8011 WYOMING WAY"
	got := formatStreet(t, "8011 WY WY\nSpringfield IL 62701")
	if got != "8011 WYOMING WAY" {
		t.Fatalf("FormatStreetLine = %q, want 8011 WYOMING WAY", got)
	}
}

func TestPunchlist_PL009_BayWestDrive(t *testing.T) {
	// C#: "1014 BAY W DRIVE" → "1014 BAY WEST DR"
	got := formatStreet(t, "1014 BAY W DRIVE\nSpringfield IL 62701")
	if got != "1014 BAY WEST DR" {
		t.Fatalf("FormatStreetLine = %q, want 1014 BAY WEST DR", got)
	}
}

func TestPunchlist_PL010a_MontanaTreasure(t *testing.T) {
	// C#: state as portion of name → abbreviate
	got := formatStreet(t, "8100 Montana Treasure Avenue\nSpringfield IL 62701")
	if got != "8100 MT TREASURE AVE" {
		t.Fatalf("FormatStreetLine = %q, want 8100 MT TREASURE AVE", got)
	}
}

func TestPunchlist_PL010b_SouthCarolinaCountyRoad(t *testing.T) {
	got := formatStreet(t, "8103 South Carolina county road 22\nSpringfield IL 62701")
	if got != "8103 SC COUNTY ROAD 22" {
		t.Fatalf("FormatStreetLine = %q, want 8103 SC COUNTY ROAD 22", got)
	}
}

func TestPunchlist_PL011_UnitThenRoomOrder(t *testing.T) {
	// C#: Unit 3200 152 Tech Dr Room 12 → 152 TECH DR UNIT 3200 RM 12
	got := formatStreet(t, "Unit 3200 152 Tech Drive Room 12\nSpringfield IL 62701")
	if got != "152 TECH DR UNIT 3200 RM 12" {
		t.Fatalf("FormatStreetLine = %q, want 152 TECH DR UNIT 3200 RM 12", got)
	}
}

func TestPunchlist_PL012_BusinessSuiteThenPrimary(t *testing.T) {
	// C#: UCENT Building Suite 480 411 N Central Ave
	raw := "UCENT Building Suite 480 411 N Central Ave\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	if n.PrimaryNumber != "411" {
		t.Fatalf("PrimaryNumber = %q, want 411", n.PrimaryNumber)
	}
	if n.SecondaryDesignator != "STE" || n.SecondaryNumber != "480" {
		t.Fatalf("secondary = %q %q, want STE 480", n.SecondaryDesignator, n.SecondaryNumber)
	}
	if !strings.Contains(n.BusinessName, "UCENT") {
		t.Fatalf("BusinessName = %q, want UCENT…", n.BusinessName)
	}
}

func TestPunchlist_PL013a_RFDRoutePhrase(t *testing.T) {
	got := formatStreet(t, "RFD Route 61 Box 87b\nSpringfield IL 62701")
	if got != "RR 61 BOX 87B" {
		t.Fatalf("FormatStreetLine = %q, want RR 61 BOX 87B", got)
	}
}

func TestPunchlist_PL013b_GluedRRHash(t *testing.T) {
	got := formatStreet(t, "RR0061#87b\nSpringfield IL 62701")
	if got != "RR 61 BOX 87B" {
		t.Fatalf("FormatStreetLine = %q, want RR 61 BOX 87B", got)
	}
}

func TestPunchlist_PL014_TrailingJunkAfterSuffix(t *testing.T) {
	// Boulevard should still peel when non-RR trailing tokens exist.
	n := parseNorm(t, "100 Oak Boulevard Annex\nSpringfield IL 62701")
	if n.StreetSuffix != "BLVD" {
		t.Fatalf("StreetSuffix = %q, want BLVD (got name=%q)", n.StreetSuffix, n.StreetName)
	}
}

func TestPunchlist_PL015_HighwayParseNormalizeParity(t *testing.T) {
	// Same highway content via Parse vs structured fields should match.
	viaParse := formatStreet(t, "100 HWY 66 FRONTAGE ROAD\nSpringfield IL 62701")
	structured, err := Normalize(Address{
		PrimaryNumber: "100",
		StreetName:    "HWY 66 FRONTAGE ROAD",
		City:          "Springfield",
		Region:        "IL",
		Postal:        "62701",
	})
	if err != nil {
		t.Fatal(err)
	}
	viaStruct := FormatStreetLine(structured)
	if viaParse != viaStruct {
		t.Fatalf("Parse path %q != structured path %q", viaParse, viaStruct)
	}
}

func TestPunchlist_PL016_UrbanizacionSpelling(t *testing.T) {
	// Correct Spanish spelling (one 'z') should map like URBANIZATION → URB.
	raw := "Calle Luna 123 Urbanizacion Las Villas\nSan Juan PR 00901"
	// At minimum, URBANIZACION as secondary designator alone:
	raw2 := "123 Calle Luna URBANIZACION\nSan Juan PR 00901"
	p, err := Parse(raw2)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if p.SecondaryDesignator != "URB" && p.SecondaryDesignator != "URBANIZACION" {
		// After Normalize should be URB
		n, err := Normalize(p)
		if err != nil {
			t.Fatalf("Normalize: %v (designator was %q)", err, p.SecondaryDesignator)
		}
		if n.SecondaryDesignator != "URB" {
			t.Fatalf("SecondaryDesignator = %q, want URB", n.SecondaryDesignator)
		}
	}
	_ = raw
}

func TestPunchlist_PL017_UnitBoxCivilianCity(t *testing.T) {
	// Military-form street with civilian last line should not hard-fail inconsistently.
	raw := "UNIT 2050 BOX 4190\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	if n.StreetName != "UNIT 2050 BOX 4190" {
		t.Fatalf("StreetName = %q, want UNIT 2050 BOX 4190", n.StreetName)
	}
	if n.City != "SPRINGFIELD" || n.Region != "IL" {
		t.Fatalf("last = %q %q", n.City, n.Region)
	}
}

func TestPunchlist_PL018_MidLineHashSecondary(t *testing.T) {
	raw := "100 #12 Main Street\nSpringfield IL 62701"
	n := parseNorm(t, raw)
	if n.PrimaryNumber != "100" {
		t.Fatalf("PrimaryNumber = %q, want 100", n.PrimaryNumber)
	}
	if n.SecondaryDesignator != "#" || n.SecondaryNumber != "12" {
		t.Fatalf("secondary = %q %q, want # 12", n.SecondaryDesignator, n.SecondaryNumber)
	}
	if n.StreetName != "MAIN" || n.StreetSuffix != "ST" {
		t.Fatalf("street = %q %q", n.StreetName, n.StreetSuffix)
	}
}

// --- Minors ---

func TestPunchlist_PL019_HyphenatedDirectional(t *testing.T) {
	got := formatStreet(t, "3009 NORTH-EAST MAIN STREET\nSpringfield IL 62701")
	if got != "3009 NE MAIN ST" {
		t.Fatalf("FormatStreetLine = %q, want 3009 NE MAIN ST", got)
	}
}

func TestPunchlist_PL020_ApostropheStrip(t *testing.T) {
	got := formatStreet(t, "4007 West Main' rd\nSpringfield IL 62701")
	if got != "4007 W MAIN RD" {
		t.Fatalf("FormatStreetLine = %q, want 4007 W MAIN RD", got)
	}
}

func TestPunchlist_PL021_HyphenatedStreetToken(t *testing.T) {
	n := parseNorm(t, "100 Main-Street\nSpringfield IL 62701")
	if n.StreetSuffix != "ST" && !strings.HasSuffix(FormatStreetLine(n), "ST") {
		t.Fatalf("expected STREET→ST peel, got name=%q suffix=%q format=%q",
			n.StreetName, n.StreetSuffix, FormatStreetLine(n))
	}
}

func TestPunchlist_PL022_KeyAndPrairie(t *testing.T) {
	// C#: EAST KENTUCKY KEY → E KENTUCKY KY (KEY is suffix, not secondary)
	got := formatStreet(t, "8007 EAST KENTUCKY KEY\nSpringfield IL 62701")
	if got != "8007 E KENTUCKY KY" {
		t.Fatalf("FormatStreetLine = %q, want 8007 E KENTUCKY KY", got)
	}
}

func TestPunchlist_PL023_AptAsStreetName(t *testing.T) {
	// Rare: "Apt" appearing as part of a non-unit street — must not always hard-error.
	// Accept either successful parse with Apt in name, or a clear documented error.
	// Target: "100 Aptitude Street" never confused; use "100 Suite Dreams Lane" style.
	n := parseNorm(t, "100 Suite Dreams Lane\nSpringfield IL 62701")
	if n.PrimaryNumber != "100" {
		t.Fatalf("PrimaryNumber = %q", n.PrimaryNumber)
	}
	if FormatStreetLine(n) == "" {
		t.Fatal("empty street format")
	}
}

func TestPunchlist_PL024_POBAlias(t *testing.T) {
	got := formatStreet(t, "POB 11890\nSpringfield IL 62701")
	if got != "PO BOX 11890" {
		t.Fatalf("FormatStreetLine = %q, want PO BOX 11890", got)
	}
}
