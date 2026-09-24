package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
)

// The standard's own Incorrect Form column for numbered streets, pp. 26-27.
// Every row here is a line the standard prints and tells a developer to
// rewrite, so reading one has to produce the Correct Form beside it.
func TestTheStandardsNumberedStreetLines(t *testing.T) {
	for _, c := range []struct {
		source string
		number string
		name   string
	}{
		{"CALLE 1 A17", "A17", "CALLE 1"},
		// p. 26 prints "13 CALLE 191" as the Correct Form for this row,
		// which drops the B1 of B113. The rules on the same two pages are
		// "MUST place the house number before the street name" and "MUST NOT
		// use hyphens to separate the letter from the number", and neither
		// licenses discarding two characters of a house number, so the rule
		// is followed rather than the printed example. See
		// NormalizeNumberedStreetLine.
		{"CALLE 191 B113", "B113", "CALLE 191"},
		{"CALLE 125 C-19", "C19", "CALLE 125"},
		{"CALLE 19 BLQ 199 Casa 31", "199-31", "CALLE 19"},
		{"CALLE 117 Bloque 23 Núm.18", "23-18", "CALLE 117"},
	} {
		t.Run(c.source, func(t *testing.T) {
			number, name, err := puertorico.NormalizeNumberedStreetLine(c.source)
			if err != nil {
				t.Fatalf("NormalizeNumberedStreetLine(%q) = %v", c.source, err)
			}
			if number != c.number || name != c.name {
				t.Errorf("NormalizeNumberedStreetLine(%q) = %q, %q; want %q, %q",
					c.source, number, name, c.number, c.name)
			}
		})
	}
}

// p. 27 names the identifiers that separate the numbers of a Puerto Rico
// address and says they "MUST NOT be included in patient addresses", so each
// one is recognized in order to be dropped. The numbers it introduced are
// what survives.
func TestTheIdentifiersThatMustNotBeIncluded(t *testing.T) {
	for _, c := range []struct {
		source string
		number string
	}{
		{"CALLE 19 BLOQUE 199 CASA 31", "199-31"},
		{"CALLE 19 BLQ 199 LOTE 31", "199-31"},
		{"CALLE 117 NUM 18", "18"},
		{"CALLE 117 NÚM. 18", "18"},
		{"CALLE 117 CASA 18", "18"},
		{"CALLE 117 LOTE 18", "18"},
		{"CALLE 117 NO 18", "18"},
		{"CALLE 117 # 18", "18"},
		{"CALLE 117 #18", "18"},
	} {
		t.Run(c.source, func(t *testing.T) {
			number, _, err := puertorico.NormalizeNumberedStreetLine(c.source)
			if err != nil {
				t.Fatalf("NormalizeNumberedStreetLine(%q) = %v", c.source, err)
			}
			if number != c.number {
				t.Errorf("primary number = %q, want %q", number, c.number)
			}
		})
	}
}

// An error means "not mine", per CONTRIBUTING §1.6.
//
// The two refusals worth naming: "CALLE 3 NO" is p. 26's own line, where NO is
// the Spanish directional for Northwest and not the number identifier it is
// spelled the same as — nothing follows it for it to introduce. And "CALLE
// AMAPOLA A17" is a named street, not a numbered one, so the trailing A17 is
// as likely to be part of the name as a house number and this pattern is not
// evidence either way.
func TestWhatIsNotAPuertoRicoNumberedStreetLine(t *testing.T) {
	for _, source := range []string{
		"CALLE 3 NO",
		"CALLE AMAPOLA A17",
		"CALLE 1",
		"CALLE 19 ACME CORP",
		"CALLE 19 BLQ 199 CASA 31 LOTE 7",
		"A17 CALLE 1",
		"123 CALLE MAIN",
		"URB LAS GLADIOLAS",
		"",
	} {
		t.Run(source, func(t *testing.T) {
			if _, _, err := puertorico.NormalizeNumberedStreetLine(source); err == nil {
				t.Errorf("NormalizeNumberedStreetLine(%q) read a numbered street line", source)
			}
		})
	}
}
