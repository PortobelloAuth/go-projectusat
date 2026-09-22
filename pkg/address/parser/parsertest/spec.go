package parsertest

import (
	"fmt"
	"regexp"
	"strings"
)

// specExample is one Incorrect/Correct row from the Project US@ standard's
// tables (v1.0 PDF, pp. 16-35). Transcribed from
// poetic-systems/addressparsers's internal/probe/specexamples.go
// (addressparsers#16), which measured go-projectusat#107 and #109 before
// this package existed; every page number here traces back to that probe.
type specExample struct {
	page      int
	incorrect string
	correct   string
}

// mainland is every example from the standard's mainland tables (p.16-22,
// p.35). A fragment among them (a street line with no last line) is
// completed with mainlandLastLine, the standard's own general-delivery city
// (p.22), rather than a Puerto Rico last line: San Juan would engage the
// Puerto Rico address type these examples are not illustrating.
var mainland = []specExample{
	// Predirectional
	{16, "NORTH BAY STREET", "N BAY STREET"},
	{16, "EAST END AVE", "E END AVE"},
	// Postdirectional
	{16, "BAY DRIVE WEST", "BAY DRIVE W"},
	// Two directionals
	{16, "NORTH E MAIN STREET", "NE MAIN ST"},
	{16, "SOUTHEAST FREEWAY NORTH", "SOUTHEAST FWY N"},
	{17, "COUNTY ROAD N EAST", "COUNTY ROAD NE"},
	// Directional as part of street name
	{17, "BAY W DRIVE", "BAY WEST DRIVE"},
	{17, "NORTH AVENUE", "NORTH AVE"},
	// Street suffix as part of the name
	{19, "789 MAIN AVENUE DRIVE", "789 MAIN AVENUE DR"},
	{19, "4513 3RD STREET CIRCLE WEST", "4513 3RD STREET CIR W"},
	{19, "1000 AVE E", "1000 AVENUE E"},
	// Rural route
	{21, "RURAL ROUTE 91 BOX A7", "RR 91 BOX A7"},
	{21, "RFD 82 BOX 12", "RR 82 BOX 12"},
	{21, "RD 51 # 25", "RR 51 BOX 25"},
	{21, "RFD Route 4 #87a", "RR 4 BOX 87A"},
	{21, "RR 2 BOX 18 Bryan Dairy Rd", "RR 2 BOX 18"},
	{21, "RR03 BOX 98D", "RR 3 BOX 98D"},
	// General delivery
	{22, "GEN DELIVERY\nTAMPA, FL 33602", "GENERAL DELIVERY\nTAMPA FL 33602-9999"},
	// Post office box
	{22, "POST OFFICE BOX 11890", "PO BOX 11890"},
	{22, "POST OFFICE BOX G", "PO BOX G"},
	// Business addresses
	{35, "BIG BUSINESS INCORPORATED\n12 EAST BUSINESS LANE, SUITE-209\nKRYTON,TN\n38188-0002", "BIG BUSINESS INC\n12 E BUSINESS LN STE 209\nKRYTON, TN 38188-0022"},
	{35, "PIZZA DELIVERY COMPANY\n61-20 EAST RIVER DRIVE\nNEW YORK, NY 10021-0905", "PIZZA DELIVERY COMPANY\n61-20 E RIVER DR\nNEW YORK NY 10021-0905"},
}

// puertoRico is every example from the standard's Puerto Rico tables (p.25-
// 31): apartment buildings, urbanizations, and highway contract routes all
// engage the Puerto Rico address type, so a fragment among them is completed
// with puertoRicoLastLine instead of a mainland city.
var puertoRico = []specExample{
	// Puerto Rico: apartment buildings and condominiums
	{25, "COND VERDE APT 1120", "1 COND VERDE APT 1120"},
	{25, "VISTA SUITES III APT 104", "3 VISTA SUITES APT 104"},
	// Puerto Rico: house number before the street name
	{26, "CALLE 1 A17", "A17 CALLE 1"},
	{26, "CALLE 191 B113", "13 CALLE 191"},
	{27, "CALLE 125 C-19", "C19 CALLE 125"},
	{27, "A-17 CALLE AMAPOLA", "A17 CALLE AMAPOLA"},
	{27, "B-17A CALLE 1", "B17A CALLE 1"},
	// Puerto Rico: block and house
	{27, "CALLE 19 BLQ 199 Casa 31", "199-31 CALLE 19"},
	{27, "CALLE 117 Bloque 23 Núm.18", "23-18 CALLE 117"},
	// Urbanizations
	{28, "URBANIZATION GOLDEN GATE", "URB GOLDEN GATE"},
	{28, "A17 URB JARDINES FAGOTA\nPONCE PR 00731", "A17 JARD FAGOTA\nPONCE PR 00731"},
	{29, "URB EXT VISTA BELLA", "EXT VISTA BELLA"},
	{29, "URB ALTS DE CANÁ", "ALTS DE CANA"},
	// Puerto Rico: post office box
	{29, "XYZ COMPANY\nAPARTADO 2018", "XYZ COMPANY\nPO BOX 2018"},
	{29, "ABC COMPANY\nGPO BOX 1118", "ABC COMPANY\nPO BOX 1118"},
	// Puerto Rico: postal station above the delivery line
	{30, "PO BOX 1190\nOLD SAN JUAN STA\nSAN JUAN PR 00902-1190", "OLD SAN JUAN STA\nPO BOX 1190\nSAN JUAN PR 00902-1190"},
	// Puerto Rico: rural route
	{30, "RR03 BOX 9800", "RR 3 BOX 9800"},
	{30, "RFD ROUTE 4 BZN 1725", "RR 4 BOX 1725"},
	{30, "RUTA RURAL 3 BUZON 12000", "RR 3 BOX 12000"},
	{30, "RFD 1 Bzn 17-A", "RR 1 BOX 17A"},
	{30, "RR 2 BOX 1980\nSECTOR EL BRINCO", "RR 2 BOX 1980"},
	{30, "RR 3 BOX 3415\nBARRIO VISTA ALEGRE", "RR 3 BOX 3415"},
	// Highway contract routes
	{31, "Ruta Estrella 1 Buzón 18", "HC 1 BOX 18"},
	{31, "HC 03 Bzn 1050", "HC 1 BOX 1050"},
}

const (
	mainlandLastLine   = "\nTAMPA FL 33602"
	puertoRicoLastLine = "\nSAN JUAN PR 00907"
)

// hasLastLine tells a complete example from a fragment (a street line with
// no city/region/postal of its own), the same way specexamples.go does.
var hasLastLine = regexp.MustCompile(`\b[A-Z]{2},? [0-9]{5}(-[0-9]{4})?$`).MatchString

// SpecCases is the standard's 46 Incorrect/Correct pairs (pp. 16-35, mainland
// and Puerto Rico), each turned into two cases: whether the Correct form is a
// fixed point under Normalize, and whether the Incorrect form reaches it.
// That is the same pair of questions go-projectusat#109 measured this corpus
// against.
var SpecCases = buildSpecCases()

func buildSpecCases() []Case {
	groups := []struct {
		examples []specExample
		lastLine string
	}{
		{mainland, mainlandLastLine},
		{puertoRico, puertoRicoLastLine},
	}

	var cases []Case
	for _, g := range groups {
		for _, e := range g.examples {
			correct, incorrect := e.correct, e.incorrect
			if !hasLastLine(correct) {
				correct += g.lastLine
				incorrect += g.lastLine
			}
			source := fmt.Sprintf("p.%d", e.page)
			cases = append(cases,
				Case{
					Source: source,
					Note:   fmt.Sprintf("Correct form %q is a fixed point", strings.TrimSuffix(e.correct, g.lastLine)),
					Input:  correct,
					Want:   correct,
				},
				Case{
					Source: source,
					Note:   fmt.Sprintf("Incorrect form %q reaches %q", strings.TrimSuffix(e.incorrect, g.lastLine), strings.TrimSuffix(e.correct, g.lastLine)),
					Input:  incorrect,
					Want:   correct,
				},
			)
		}
	}
	return cases
}
