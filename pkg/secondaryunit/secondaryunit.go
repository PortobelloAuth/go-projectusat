package secondaryunit

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

/*
Apartment APT
Basement BSMT**
Building BLDG
Department DEPT
Floor FL
Front FRNT**
Hanger HNGR
Key KEY
Lobby LBBY**
Lot LOT
Lower LOWR**
Office OFC**
Penthouse PH**
Pier PIER
Rear REAR**
Room RM
Side SIDE**
Slip SLIP
Space SPC
Stop STOP
Suite STE
Trailer TRLR
Unit UNIT
Upper UPPR**
*/

type SecondaryUnit struct {
	Primary  string
	Short    string
	Numbered bool
}

var unitTypes = []SecondaryUnit{
	{Primary: "Apartment", Short: "APT", Numbered: true},
	{Primary: "Basement", Short: "BSMT", Numbered: false},
	{Primary: "Building", Short: "BLDG", Numbered: true},
	{Primary: "Department", Short: "DEPT", Numbered: true},
	{Primary: "Floor", Short: "FL", Numbered: true},
	{Primary: "Front", Short: "FRNT", Numbered: false},
	{Primary: "Hanger", Short: "HNGR", Numbered: true},
	{Primary: "Key", Short: "KEY", Numbered: true},
	{Primary: "Lobby", Short: "LBBY", Numbered: false},
	{Primary: "Lot", Short: "LOT", Numbered: true},
	{Primary: "Lower", Short: "LOWR", Numbered: false},
	{Primary: "Office", Short: "OFC", Numbered: false},
	{Primary: "Penthouse", Short: "PH", Numbered: false},
	{Primary: "Pier", Short: "PIER", Numbered: true},
	{Primary: "Rear", Short: "REAR", Numbered: false},
	{Primary: "Room", Short: "RM", Numbered: true},
	{Primary: "Side", Short: "SIDE", Numbered: false},
	{Primary: "Slip", Short: "SLIP", Numbered: true},
	{Primary: "Space", Short: "SPC", Numbered: true},
	{Primary: "Stop", Short: "STOP", Numbered: true},
	{Primary: "Suite", Short: "STE", Numbered: true},
	{Primary: "Trailer", Short: "TRLR", Numbered: true},
	{Primary: "Unit", Short: "UNIT", Numbered: true},
	{Primary: "Upper", Short: "UPPR", Numbered: false},
}

// DEVIATION FROM THE SPECIFICATION.
//
// Project US@ does not list BOX among the secondary unit designators above. It
// appears only in the rural route, military, and post office box sections,
// each of which describes the same thing: a designator followed by a number,
// standardized as "BOX ___".
//
//	RR 4 BOX 125
//	PSC 3 BOX 4120
//
// Recognizing it here rather than in each of those packages keeps one source of
// truth for the word. The alternative — pkg/addresstypes/ruralroute and
// pkg/addresstypes/military each knowing BOX independently — means two copies
// of the same knowledge, which is the failure this package's table exists to
// avoid.
//
// It is kept in a separate slice so that unitTypes remains a faithful
// transcription of the list the standard gives, which is quoted above it.
// Anything added here is this library's judgment, not the standard's.
var nonStandardUnitTypes = []SecondaryUnit{
	{Primary: "Box", Short: "BOX", Numbered: true},
}

// recognizedUnitTypes is every designator this library accepts, from the
// standard and otherwise. The lookup maps derive from it.
var recognizedUnitTypes = slices.Concat(unitTypes, nonStandardUnitTypes)

var unitMap = maps.Collect(func(yield func(string, SecondaryUnit) bool) {
	for _, v := range recognizedUnitTypes {
		if !yield(strings.ToUpper(v.Primary), v) {
			return
		}
	}
})

var unitShortMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range recognizedUnitTypes {
		if !yield(v.Short, strings.ToUpper(v.Primary)) {
			return
		}
	}
})

// hashUnit is the Project US@ exchange/matching designator for numbered
// secondary units when the unit type is unknown or intentionally collapsed.
// It is accepted by Info/Normalize so SecondaryAsHash results re-normalize.
var hashUnit = SecondaryUnit{
	Primary:  "#",
	Short:    "#",
	Numbered: true,
}

// Per the standard, Normalize always returns the uppercase abbreviation
// for the Unit Type. "#" is accepted as the matching-form designator.
func Normalize(u string) (string, error) {
	info, err := Info(u)
	if err != nil {
		return "", err
	}

	return info.Short, nil
}

func Info(u string) (*SecondaryUnit, error) {
	// capitalize
	capitalized := strings.ToUpper(u)

	// "#" is the standard matching/unknown unit designator (numbered).
	if capitalized == "#" {
		h := hashUnit
		return &h, nil
	}

	// look up the primary (full) unit type word
	full, ok := unitShortMap[capitalized]
	if !ok {
		full = capitalized
	}

	// See if it is the primary (full) unit type word
	info, ok := unitMap[full]
	if ok {
		return &SecondaryUnit{
			Primary:  info.Primary,
			Short:    info.Short,
			Numbered: info.Numbered,
		}, nil
	}

	return nil, fmt.Errorf("Unrecognized unit type")
}
