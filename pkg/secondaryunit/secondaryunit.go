// Package secondaryunit reads a secondary unit designator out of an address.
//
// The rows are Publication 28 Appendix C2, held once in
// github.com/poetic-systems/addresstables and read from there. What is here is
// the part that is about parsing: which spelling a token is, whether the
// designator takes a number, and how strongly a match should be rated.
package secondaryunit

import (
	"fmt"
	"maps"
	"strings"

	"github.com/poetic-systems/addresstables/secondaryunit"
)

type SecondaryUnit struct {
	Primary  string
	Short    string
	Numbered bool
}

var unitMap = maps.Collect(func(yield func(string, SecondaryUnit) bool) {
	for u := range secondaryunit.All() {
		if !yield(u.Full, SecondaryUnit{Primary: u.Full, Short: u.Short, Numbered: u.Numbered}) {
			return
		}
	}
})

var unitShortMap = maps.Collect(func(yield func(string, string) bool) {
	for u := range secondaryunit.All() {
		if !yield(u.Short, u.Full) {
			return
		}
	}
})

// hashUnit is the Project US@ exchange/matching designator for numbered
// secondary units when the unit type is unknown or intentionally collapsed.
// It is not an Appendix C2 row, so it is not in the shared table; it is
// accepted by Info/Normalize so SecondaryAsHash results re-normalize.
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
		return &info, nil
	}

	return nil, fmt.Errorf("Unrecognized unit type")
}
