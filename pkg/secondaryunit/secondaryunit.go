package secondaryunit

import (
	"fmt"
	"maps"
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

var unitMap = maps.Collect(func(yield func(string, SecondaryUnit) bool) {
	for _, v := range unitTypes {
		if !yield(strings.ToUpper(v.Primary), v) {
			return
		}
	}
})

var unitShortMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range unitTypes {
		if !yield(v.Short, strings.ToUpper(v.Primary)) {
			return
		}
	}
})

// TODO: When comparing addresses for matching, consider converting
// all unit types to # as an option (even though it isn't correct by
// the standard.) Because the standard allows for it when the unit
// type isn't fully known it is essentially the common matchable
// unit type abbreviation for all numbered unit types.
// TODO: what should we to do with # (re-read the standard)

// Per the standard, Normalize always returns the uppercase abbreviation
// for the Unit Type.
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
