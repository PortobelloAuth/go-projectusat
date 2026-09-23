package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
)

// The standard's own street lines, pp. 24-27. Each is already in its correct
// form, so reading one has to leave it as it is.
func TestTheStandardsStreetLines(t *testing.T) {
	for _, c := range []struct {
		source string
		number string
		name   string
	}{
		{"123 CALLE MAIN", "123", "CALLE MAIN"},
		{"1234 CALLE AURORA", "1234", "CALLE AURORA"},
		{"150 CALLE A", "150", "CALLE A"},
		{"1234 AVE ASHFORD", "1234", "AVENIDA ASHFORD"},
		// The standard's own example, still read correctly, but no longer a
		// fixed point byte for byte: spelling AVE out is the cost of #117's
		// ruling, documented on NormalizeStreetLine.
		{"585 AVE FD ROOSEVELT", "585", "AVENIDA FD ROOSEVELT"},
		{"A17 CALLE AMAPOLA", "A17", "CALLE AMAPOLA"},
		// PARQUE is missing from the street type table, so "1025 PARQUE DEL
		// REY" (p. 25) is not read yet. The standard lists PARQUE among the
		// street prefixes in the body and abbreviates it PARQ as a standalone
		// urbanization, which are two different answers for one word; the
		// table gets a row once that is settled. See #60.
		{"1510 CALLE 3 NO", "1510", "CALLE 3 NO"},
	} {
		t.Run(c.source, func(t *testing.T) {
			number, name, err := puertorico.NormalizeStreetLine(c.source)
			if err != nil {
				t.Fatalf("NormalizeStreetLine(%q) = %v", c.source, err)
			}
			if number != c.number || name != c.name {
				t.Errorf("NormalizeStreetLine(%q) = %q, %q; want %q, %q",
					c.source, number, name, c.number, c.name)
			}
		})
	}
}

// A hyphen after a letter is punctuation inside one number and comes out; a
// hyphen between two numbers carries the block and the house and stays. The
// first three are the standard's Incorrect Form column, p. 27.
func TestTheHyphenMeansTwoDifferentThings(t *testing.T) {
	for _, c := range []struct {
		source string
		number string
	}{
		{"A-17 CALLE AMAPOLA", "A17"},
		{"B-17A CALLE 1", "B17A"},
		{"C-19 CALLE 125", "C19"},
		{"199-31 CALLE 19", "199-31"},
	} {
		t.Run(c.source, func(t *testing.T) {
			number, _, err := puertorico.NormalizeStreetLine(c.source)
			if err != nil {
				t.Fatalf("NormalizeStreetLine(%q) = %v", c.source, err)
			}
			if number != c.number {
				t.Errorf("primary number = %q, want %q", number, c.number)
			}
		})
	}
}

// "Developers MUST NOT abbreviate street names" (p. 26), so an abbreviated
// type is spelled out and never the other way around. AVE is not an
// exception: p. 24 permits AVE alongside AVENIDA, but p. 26's MUST NOT
// abbreviate outranks that MAY over the same text, so AVE is spelled out like
// every other type. See NormalizeStreetLine.
func TestAnAbbreviatedTypeIsSpelledOut(t *testing.T) {
	for _, c := range []struct {
		source string
		name   string
	}{
		{"1234 CLL AURORA", "CALLE AURORA"},
		{"1234 PSO DEL REY", "PASEO DEL REY"},
		{"1234 CAM DEL MAR", "CAMINO DEL MAR"},
		{"1234 AVE ASHFORD", "AVENIDA ASHFORD"},
		{"1234 AVENIDA ASHFORD", "AVENIDA ASHFORD"},
	} {
		t.Run(c.source, func(t *testing.T) {
			_, name, err := puertorico.NormalizeStreetLine(c.source)
			if err != nil {
				t.Fatalf("NormalizeStreetLine(%q) = %v", c.source, err)
			}
			if name != c.name {
				t.Errorf("street name = %q, want %q", name, c.name)
			}
		})
	}
}

// An error means "not mine", per CONTRIBUTING §1.6.
//
// A mainland street line whose words this vocabulary does not share is refused
// here. One that shares them is not, and cannot be: AVE is a Puerto Rico
// street type and an English suffix both, so "1000 AVE E" reads perfectly well
// as a Spanish street line and the tokens say nothing against it. What keeps
// that address out of Spanish is the dialect gate in Candidates, which never
// offers this reading under a Florida last line. See TestAMainlandAddressIsNotOfferedASpanishReading.
func TestWhatIsNotAPuertoRicoStreetLine(t *testing.T) {
	for _, source := range []string{
		"123 MAIN ST",
		"1600 PENNSYLVANIA AVE NW",
		"PO BOX 11890",
		"RR 4 BOX 125",
		"CALLE AMAPOLA",
		"123 CALLE",
		"URB LAS GLADIOLAS",
		"",
	} {
		t.Run(source, func(t *testing.T) {
			if _, _, err := puertorico.NormalizeStreetLine(source); err == nil {
				t.Errorf("NormalizeStreetLine(%q) read a Puerto Rico street line", source)
			}
		})
	}
}
