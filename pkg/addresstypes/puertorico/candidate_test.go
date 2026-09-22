package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

// candidates runs the vocabularies an address is assembled from and returns
// every Puerto Rico candidate, over every reading of the last line.
//
// streetsuffixes and directionals are among them because they are the
// vocabularies whose reading of a Spanish street line is the wrong one — see
// #71 — so the tests see whatever they contribute rather than a pool tidied
// for the occasion.
func candidates(source string) []*address.CandidateAddress {
	tokens := token.Tokenize(source)

	var claims []claim.Claim
	claims = append(claims, region.Claims(tokens)...)
	claims = append(claims, postalcode.Claims(tokens)...)
	claims = append(claims, country.Claims(tokens)...)
	claims = append(claims, streetsuffixes.Claims(tokens)...)
	claims = append(claims, directionals.Claims(tokens)...)
	claims = append(claims, secondaryunit.Claims(tokens)...)
	claims = append(claims, puertorico.Claims(tokens)...)

	var found []*address.CandidateAddress
	for _, line := range lastline.LineClaims(tokens, claims) {
		found = append(found, puertorico.Candidates(tokens, claims, line)...)
	}

	return found
}

// street returns the street line of the strongest candidate, and whether there
// is a candidate at all.
func street(t *testing.T, source string) (string, bool) {
	t.Helper()

	var top *address.CandidateAddress
	for _, c := range candidates(source) {
		if top == nil || c.Confidence > top.Confidence {
			top = c
		}
	}
	if top == nil {
		return "", false
	}

	return top.Address.FormatStreetLine(), true
}

// The standard's own Puerto Rico address, p. 24. The secondary identifier line
// is not read yet, so only the street line below it is asserted here.
func TestTheStandardsPuertoRicoAddress(t *testing.T) {
	got, ok := street(t, "URB HIGHLAND GDNS\nCOND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926")
	if !ok {
		t.Fatal("the standard's own example is not read")
	}
	if got != "123 CALLE MAIN" {
		t.Errorf("street line = %q, want %q", got, "123 CALLE MAIN")
	}
}

// The street line ends where the last line begins, with or without a line
// break to say so. A single-line address is the case that shows it: nothing in
// the tokens marks the boundary, and the last line reading is what supplies it.
func TestTheLastLineBoundsTheStreetName(t *testing.T) {
	for _, source := range []string{
		"150 CALLE A\nSAN JUAN PR 00926",
		"150 CALLE A SAN JUAN PR 00926",
		"150 CALLE A, SAN JUAN, PR 00926",
	} {
		t.Run(source, func(t *testing.T) {
			got, ok := street(t, source)
			if !ok {
				t.Fatal("no reading")
			}
			if got != "150 CALLE A" {
				t.Errorf("street line = %q, want %q", got, "150 CALLE A")
			}
		})
	}
}

// A mainland address is never offered a Spanish reading, whatever its words.
//
// "1000 AVE E" is the standard's own Florida example (p. 19) and it satisfies
// the Spanish grammar exactly — a number, a street type, a root name — because
// AVE belongs to both vocabularies. Only the last line separates them, which
// is why the dialect is decided here and not in Claims. This is the answer to
// #71 in the direction nobody asked about.
func TestAMainlandAddressIsNotOfferedASpanishReading(t *testing.T) {
	for _, source := range []string{
		"1000 AVE E\nTAMPA FL 33602",
		"1234 AVE ASHFORD\nTAMPA FL 33602",
		"123 MAIN ST\nTAMPA FL 33602",
	} {
		t.Run(source, func(t *testing.T) {
			if got, ok := street(t, source); ok {
				t.Errorf("a mainland address was read as Puerto Rican: %q", got)
			}
		})
	}
}

// Either half of the last line engages the dialect on its own, so an address
// that arrives without a region is still read in Spanish. See UsePRDialect.
func TestTheZIPCodeEngagesTheDialectWithoutTheRegion(t *testing.T) {
	got, ok := street(t, "1234 AVE ASHFORD\nSAN JUAN 00907")
	if !ok {
		t.Fatal("a Puerto Rico ZIP Code did not engage the dialect")
	}
	if got != "1234 AVE ASHFORD" {
		t.Errorf("street line = %q, want %q", got, "1234 AVE ASHFORD")
	}
}

// Where an urbanization sits above the street line, a reading carrying it is
// offered alongside the one that does not, so that stranding it is a reading
// the parser ranks rather than the only one available.
func TestTheUrbanizationIsOfferedWithTheStreetLine(t *testing.T) {
	var withArea, withoutArea bool
	for _, c := range candidates("URB LAS GLADIOLAS\n150 CALLE A\nSAN JUAN PR 00926") {
		switch c.Address.Area {
		case "URB LAS GLADIOLAS":
			withArea = true
		case "":
			withoutArea = true
		default:
			t.Errorf("area = %q", c.Address.Area)
		}
	}

	if !withArea {
		t.Error("no reading carries the urbanization")
	}
	if !withoutArea {
		t.Error("no reading leaves the urbanization out")
	}
}

// Every candidate names this package as the address type, which is how the
// parser tells one type's reading from another's.
func TestEveryCandidateNamesThisAddressType(t *testing.T) {
	found := candidates("A17 CALLE AMAPOLA\nSAN JUAN PR 00907")
	if len(found) == 0 {
		t.Fatal("no reading")
	}
	for _, c := range found {
		if _, ok := c.Address.Type.(*puertorico.PuertoRicoAddress); !ok {
			t.Errorf("address type = %T, want *puertorico.PuertoRicoAddress", c.Address.Type)
		}
	}
}
