package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
)

// area returns the value the best reading assigns to the Area, and whether
// there is a reading at all.
func area(source string) (string, bool) {
	claims := puertorico.Claims(token.Tokenize(source))
	if len(claims) == 0 {
		return "", false
	}
	if len(claims) > 1 {
		return "", false
	}

	return claims[0].Parts[0].Value, true
}

// The standard's own Puerto Rico example. The urbanization occupies the first
// line of the street address block, above the secondary and primary lines.
func TestTheStandardsUrbanizationExample(t *testing.T) {
	got, ok := area("URB HIGHLAND GDNS\nCOND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926")
	if !ok {
		t.Fatal("the standard's own example is not claimed")
	}
	if got != "URB HIGHLAND GDNS" {
		t.Errorf("area = %q, want %q", got, "URB HIGHLAND GDNS")
	}
}

func TestEverySpellingOfTheDesignatorClaimsTheSameArea(t *testing.T) {
	for _, designator := range []string{"URB", "Urb", "URBANIZACION", "Urbanización", "URBANIZATION"} {
		t.Run(designator, func(t *testing.T) {
			got, ok := area(designator + " LAS GLADIOLAS\n150 CALLE A\nSAN JUAN PR 00926")
			if !ok {
				t.Fatalf("%q does not open an urbanization", designator)
			}
			if got != "URB LAS GLADIOLAS" {
				t.Errorf("area = %q, want %q", got, "URB LAS GLADIOLAS")
			}
		})
	}
}

// The name is free text of whatever length its developer chose, and the line
// is the only thing that says where it ends.
func TestTheWholeLineAfterTheDesignatorIsTheName(t *testing.T) {
	got, ok := area("URB JARDINES DE COUNTRY CLUB\n123 CALLE MAIN\nSAN JUAN PR 00926")
	if !ok {
		t.Fatal("a multi word development name is not claimed")
	}
	if got != "URB JARDINES DE COUNTRY CLUB" {
		t.Errorf("area = %q, want the whole line", got)
	}
}

// A claim never crosses a line break, so the street line that follows stays
// available to whatever reads it.
func TestTheClaimStopsAtTheLineBreak(t *testing.T) {
	tokens := token.Tokenize("URB HIGHLAND GDNS\n123 CALLE MAIN\nSAN JUAN PR 00926")
	claims := puertorico.Claims(tokens)
	if len(claims) != 1 {
		t.Fatalf("got %d claims, want 1", len(claims))
	}
	if claims[0].Start() != 0 || claims[0].End() != 3 {
		t.Errorf("extent = [%d,%d), want [0,3): the claim runs past its line",
			claims[0].Start(), claims[0].End())
	}
}

func TestNoUrbanizationIsNoClaims(t *testing.T) {
	for _, source := range []string{
		"123 MAIN ST\nDENVER CO 80202",
		"COND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926",
	} {
		t.Run(source, func(t *testing.T) {
			if claims := puertorico.Claims(token.Tokenize(source)); len(claims) != 0 {
				t.Errorf("got %d claims, want none", len(claims))
			}
		})
	}
}

// A designator with no name after it is a fragment of a pattern that did not
// match, not weak evidence of one.
func TestADesignatorAloneIsNotAnUrbanization(t *testing.T) {
	if claims := puertorico.Claims(token.Tokenize("URB\n123 CALLE MAIN\nSAN JUAN PR 00926")); len(claims) != 0 {
		t.Errorf("got %d claims for a bare designator, want none", len(claims))
	}
}

// The designator has to open its line. Claiming from partway along one would
// swallow whatever preceded it into a component that cannot contain it.
func TestADesignatorPartwayAlongALineIsNotClaimed(t *testing.T) {
	if claims := puertorico.Claims(token.Tokenize("ACME CORP URB HIGHLAND GDNS\n123 CALLE MAIN\nSAN JUAN PR 00926")); len(claims) != 0 {
		t.Errorf("got %d claims, want none: the designator does not open the line", len(claims))
	}
}

// An urbanization sits above the street line, so something has to follow it.
// On a single line there is nothing to say where the name stops, and the
// reading is declined rather than offered weakly.
func TestAnUrbanizationIsNeverTheLastLine(t *testing.T) {
	for _, source := range []string{
		"URB LAS GLADIOLAS 150 CALLE A SAN JUAN PR 00926",
		"123 CALLE MAIN\nURB LAS GLADIOLAS",
	} {
		t.Run(source, func(t *testing.T) {
			if claims := puertorico.Claims(token.Tokenize(source)); len(claims) != 0 {
				t.Errorf("got %d claims, want none", len(claims))
			}
		})
	}
}

// Every claim has to be usable by a parser: one part, in range, naming the
// Area, and never assigning a token twice.
func TestEveryClaimIsWellFormed(t *testing.T) {
	sources := []string{
		"URB HIGHLAND GDNS\nCOND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926",
		"URB LAS GLADIOLAS\n150 CALLE A\nSAN JUAN PR 00926",
		"URB SANTA MARIA\nURB LAS FLORES\nSAN JUAN PR 00926",
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range puertorico.Claims(tokens) {
				if len(c.Parts) != 1 {
					t.Fatalf("claim has %d parts, want 1", len(c.Parts))
				}
				p := c.Parts[0]
				if p.Part != claim.PartArea {
					t.Errorf("part = %q, want %q", p.Part, claim.PartArea)
				}
				if p.Start < 0 || p.End() > len(tokens) {
					t.Errorf("part [%d,%d) is outside the %d tokens", p.Start, p.End(), len(tokens))
				}
				if p.Length < 2 {
					t.Errorf("part length = %d, want at least a designator and a name", p.Length)
				}
				if c.Confidence != claim.ConfidenceExact {
					t.Errorf("confidence = %d, want exact", c.Confidence)
				}
			}
		})
	}
}
