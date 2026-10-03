package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
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
//
// AVE on input comes back AVENIDA: every street type is spelled out, AVE
// included, per NormalizeStreetLine.
func TestTheZIPCodeEngagesTheDialectWithoutTheRegion(t *testing.T) {
	got, ok := street(t, "1234 AVE ASHFORD\nSAN JUAN 00907")
	if !ok {
		t.Fatal("a Puerto Rico ZIP Code did not engage the dialect")
	}
	if got != "1234 AVENIDA ASHFORD" {
		t.Errorf("street line = %q, want %q", got, "1234 AVENIDA ASHFORD")
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

// pp. 26-27's numbered streets put the house number after the street instead
// of in front of it, and the reading has to come out in the standard's order.
// These are the Incorrect Form column of both tables, read through Candidates
// so that the claim's parts are exercised and not only the line reader.
//
// "CALLE 191 B113" keeps its B1: see NormalizeNumberedStreetLine for why the
// rule is followed here rather than the Correct Form the standard prints.
func TestANumberedStreetPutsItsHouseNumberFirst(t *testing.T) {
	for _, c := range []struct {
		source string
		street string
	}{
		{"CALLE 1 A17\nSAN JUAN PR 00907", "A17 CALLE 1"},
		{"CALLE 191 B113\nSAN JUAN PR 00907", "B113 CALLE 191"},
		{"CALLE 125 C-19\nSAN JUAN PR 00907", "C19 CALLE 125"},
		{"CALLE 19 BLQ 199 Casa 31\nSAN JUAN PR 00907", "199-31 CALLE 19"},
		{"CALLE 117 Bloque 23 Núm.18\nSAN JUAN PR 00907", "23-18 CALLE 117"},
	} {
		t.Run(c.source, func(t *testing.T) {
			got, ok := street(t, c.source)
			if !ok {
				t.Fatal("no reading")
			}
			if got != c.street {
				t.Errorf("street line = %q, want %q", got, c.street)
			}
		})
	}
}

// p. 25's fallback for a condominium line with no primary address number at
// all: the primary number defaults to "1" (#158), or, where the building name
// ends in a roman numeral building number, that number is promoted to the
// primary number and converted to arabic (#159). Both are the issues' own
// examples.
func TestCondominiumLineSynthesizesAPrimaryNumber(t *testing.T) {
	for _, c := range []struct {
		name   string
		source string
		street string
	}{
		{
			"no building number defaults to 1",
			"COND VERDE APT 1120\nSAN JUAN PR 00907",
			"1 COND VERDE APT 1120",
		},
		{
			"trailing roman numeral becomes the primary number",
			"VISTA SUITES III APT 104\nSAN JUAN PR 00907",
			"3 VISTA SUITES APT 104",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, ok := street(t, c.source)
			if !ok {
				t.Fatal("no reading")
			}
			if got != c.street {
				t.Errorf("street line = %q, want %q", got, c.street)
			}
		})
	}
}

// Without a secondary-unit claim sitting flush at the end of the line, the
// condominium fallback must not fire: an unrecognized line is still just
// unrecognized, not an invitation to default a primary number onto any random
// text. This is the same shape TestTheStandardsPuertoRicoAddress and
// TestANumberedStreetPutsItsHouseNumberFirst already read correctly; this
// checks the negative space around #158/#159 specifically.
func TestNoSecondaryUnitMeansNoCondominiumFallback(t *testing.T) {
	if _, ok := street(t, "SOME UNRECOGNIZABLE WORDS HERE\nSAN JUAN PR 00907"); ok {
		t.Error("a line with no secondary unit and no primary number was read anyway")
	}
}

// pp.28-29: where an urbanization line is the entire street-equivalent
// content — there is no ordinary street line below it, only the last line —
// the urbanization itself has to stand in for the street line in the
// formatted address. go-projectusat#162's five FAIL cases are pinned here
// directly against Address.Format, since that is what the issue's "want"
// column actually asserts: the two-line shape for a bare urbanization name,
// and the single combined line where a primary number precedes the
// designator (p.28's own "A17 URB JARDINES FAGOTA" example).
func TestUrbanizationStandsAloneAsTheStreetLine(t *testing.T) {
	for _, c := range []struct {
		name   string
		source string
		want   string
	}{
		{
			"ordinary designator with no standalone exception",
			"URBANIZATION GOLDEN GATE\nSAN JUAN PR 00907",
			"URB GOLDEN GATE\nSAN JUAN PR 00907",
		},
		{
			"primary number precedes an ordinary designator",
			"A17 URB JARDINES FAGOTA\nPONCE PR 00731",
			"A17 JARD FAGOTA\nPONCE PR 00731",
		},
		{
			"standalone exception with no URB prefix",
			"EXT VISTA BELLA\nSAN JUAN PR 00907",
			"EXT VISTA BELLA\nSAN JUAN PR 00907",
		},
		{
			"standalone exception strips a leading URB",
			"URB EXT VISTA BELLA\nSAN JUAN PR 00907",
			"EXT VISTA BELLA\nSAN JUAN PR 00907",
		},
		{
			"standalone exception folds diacritics",
			"URB ALTS DE CANÁ\nSAN JUAN PR 00907",
			"ALTS DE CANA\nSAN JUAN PR 00907",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			var top *address.CandidateAddress
			for _, cand := range candidates(c.source) {
				if top == nil || cand.Confidence > top.Confidence {
					top = cand
				}
			}
			if top == nil {
				t.Fatal("no reading")
			}
			if got := top.Address.Format(); got != c.want {
				t.Errorf("Format() = %q, want %q", got, c.want)
			}
		})
	}
}

// The VISTA regression #162's fix nearly introduced: "VISTA" is itself a
// pp.28-29 standalone-exception word, so a condominium building name that
// happens to start with it ("VISTA SUITES III APT 104") must still be read
// as a condominium line, not mistaken for a standalone urbanization line —
// the trailing APT secondary unit is what tells them apart, see
// condominiumStreetLine and the guard in Candidates ahead of it.
func TestUrbanizationDoesNotShadowACondominiumLine(t *testing.T) {
	got, ok := street(t, "VISTA SUITES III APT 104\nSAN JUAN PR 00907")
	if !ok {
		t.Fatal("no reading")
	}
	if got != "3 VISTA SUITES APT 104" {
		t.Errorf("street line = %q, want %q", got, "3 VISTA SUITES APT 104")
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

// normalizedStreet is street, run through the address type's own Normalize
// first: the street line as the parser finally prints it.
func normalizedStreet(t *testing.T, source string) (*address.Address, string) {
	t.Helper()

	var top *address.CandidateAddress
	for _, c := range candidates(source) {
		if top == nil || c.Confidence > top.Confidence {
			top = c
		}
	}
	if top == nil {
		t.Fatalf("%q: no reading", source)
	}

	out, err := (&puertorico.PuertoRicoAddress{}).Normalize(top.Address, normalizer.AddressNormalizationOptions{})
	if err != nil {
		t.Fatalf("%q: Normalize: %v", source, err)
	}

	return out, out.FormatStreetLine()
}

// go-projectusat#154 / PR #175 review: a Spanish directional that is a pre- or
// postdirectional goes in Predirectional/Postdirectional and is ABBREVIATED;
// one inside the street name stays in StreetName and is expanded there. The
// first two rows are p. 26's own examples.
func TestADirectionalIsPlacedInItsOwnFieldAndAbbreviated(t *testing.T) {
	for _, tc := range []struct {
		source, wantPre, wantName, wantPost, wantLine string
	}{
		{"1510 CALLE 3 NO\nSAN JUAN PR 00926", "", "CALLE 3", "NO", "1510 CALLE 3 NO"},
		{"1620 CALLE 17 SO\nSAN JUAN PR 00926", "", "CALLE 17", "SO", "1620 CALLE 17 SO"},
		{"1510 CALLE 3 NOROESTE\nSAN JUAN PR 00926", "", "CALLE 3", "NO", "1510 CALLE 3 NO"},
		{"1620 CALLE 17 SUDOESTE\nSAN JUAN PR 00926", "", "CALLE 17", "SO", "1620 CALLE 17 SO"},
		{"1620 CALLE 17 SUR\nSAN JUAN PR 00926", "", "CALLE 17", "S", "1620 CALLE 17 S"},
		{"1620 CALLE 17 NORTE ESTE\nSAN JUAN PR 00926", "", "CALLE 17", "NE", "1620 CALLE 17 NE"},
		// The English analogue of p.35's "12 E BUSINESS LN": a predirectional
		// stays abbreviated rather than being expanded inside the name.
		{"12 E CALLE LUNA\nSAN JUAN PR 00926", "E", "CALLE LUNA", "", "12 E CALLE LUNA"},
		{"12 EAST CALLE LUNA\nSAN JUAN PR 00926", "E", "CALLE LUNA", "", "12 E CALLE LUNA"},
		{"12 ESTE CALLE LUNA\nSAN JUAN PR 00926", "E", "CALLE LUNA", "", "12 E CALLE LUNA"},
		// Inside the street name: not a directional field, so expanded.
		{"123 CALLE SO 5\nSAN JUAN PR 00926", "", "CALLE SUDOESTE 5", "", "123 CALLE SUDOESTE 5"},
		{"123 CALLE NORTE 5\nSAN JUAN PR 00926", "", "CALLE NORTE 5", "", "123 CALLE NORTE 5"},
		// A directional that is the whole root name is the name, not a
		// postdirectional.
		{"150 CALLE O\nSAN JUAN PR 00926", "", "CALLE O", "", "150 CALLE O"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got, line := normalizedStreet(t, tc.source)
			if got.Predirectional != tc.wantPre || got.StreetName != tc.wantName || got.Postdirectional != tc.wantPost {
				t.Errorf("pre/name/post = %q/%q/%q, want %q/%q/%q",
					got.Predirectional, got.StreetName, got.Postdirectional, tc.wantPre, tc.wantName, tc.wantPost)
			}
			if line != tc.wantLine {
				t.Errorf("street line = %q, want %q", line, tc.wantLine)
			}
		})
	}
}

// The postdirectional is a claim of its own, placed by Candidates — not a
// word the street-name normalizer happens to leave alone. This is the claim
// to field wiring Aaron asked to see on PR #175.
func TestTheDirectionalClaimLandsInTheField(t *testing.T) {
	whole := false
	for _, c := range candidates("1510 CALLE 3 NOROESTE\nSAN JUAN PR 00926") {
		if c.Address.Postdirectional != "NO" || c.Address.StreetName != "CALLE 3" {
			t.Errorf("candidate postdirectional/name = %q/%q, want NO/CALLE 3", c.Address.Postdirectional, c.Address.StreetName)
		}
		// Some last-line readings strand a token; the one that reads the
		// whole last line must leave nothing over.
		if len(c.Leftover) == 0 {
			whole = true
		}
	}
	if !whole {
		t.Error("no candidate accounts for every token")
	}
}

// Region abbreviation runs last in NormalizePuertoRicoStreetName, as in the
// generic normalizer (#165): MONTANA abbreviates to MT and must not then be
// read back as MOUNT by ExpandCityInStreetNameFn.
func TestRegionAbbreviationRunsLast(t *testing.T) {
	got, err := puertorico.NormalizePuertoRicoStreetName("MONTANA TREASURE", normalizer.AddressNormalizationOptions{})
	if err != nil && err != normalizer.Done {
		t.Fatalf("NormalizePuertoRicoStreetName: %v", err)
	}
	if got != "MT TREASURE" {
		t.Errorf("NormalizePuertoRicoStreetName(MONTANA TREASURE) = %q, want MT TREASURE", got)
	}
}
