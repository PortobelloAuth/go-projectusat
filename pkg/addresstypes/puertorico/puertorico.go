// Package puertorico reads a Puerto Rico address.
//
// The vocabularies are held once, in github.com/poetic-systems/addresstables,
// and read from there: the leading street types, the secondary address
// identifiers, and the urbanization designators in urbanization.go. What is
// here is the part that is about parsing — which spelling a token is, which
// form to return, and the claim rules.
package puertorico

import (
	"fmt"
	"maps"
	"strings"

	"github.com/poetic-systems/addresstables/puertorico"
)

// streetTypeMap maps Spanish primary street type -> abbreviation.
// Project US@ keeps Spanish forms (do not force English).
var streetTypeMap = maps.Collect(func(yield func(string, string) bool) {
	for t := range puertorico.StreetTypes() {
		if !yield(t.Full, t.Short) {
			return
		}
	}
})

var streetTypeShortMap = maps.Collect(func(yield func(string, string) bool) {
	for primary, short := range streetTypeMap {
		if !yield(short, primary) {
			return
		}
	}
})

// secondaryMap maps Spanish/English primary secondary designator -> abbreviation.
// Per Project US@ secondary designators, Normalize returns the uppercase short form.
//
// The urbanization is deliberately absent from this table, upstream as well as
// here. The standard puts it on a line of its own above the secondary address
// identifier, and this library carries it in Address.Area rather than as a
// secondary designator, so it has its own vocabulary in urbanization.go.
// Listing it in both places would make URB two things at once.
var secondaryMap = maps.Collect(func(yield func(string, string) bool) {
	for d := range puertorico.Secondaries() {
		if !yield(d.Full, d.Short) {
			return
		}
	}
})

var secondaryShortMap = maps.Collect(func(yield func(string, string) bool) {
	for primary, short := range secondaryMap {
		if !yield(short, primary) {
			return
		}
	}
})

// NormalizeStreetType maps a Spanish PR street type (primary or abbreviation)
// to the Spanish primary form. Example: "AVE" or "AVENIDA" -> "AVENIDA".
func NormalizeStreetType(s string) (string, error) {
	capitalized := strings.ToUpper(strings.TrimSpace(s))

	if primary, ok := streetTypeShortMap[capitalized]; ok {
		return primary, nil
	}
	if _, ok := streetTypeMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized street type")
}

// AbbreviateStreetType maps a Spanish PR street type (primary or abbreviation)
// to the Spanish abbreviation. Example: "AVE" or "AVENIDA" -> "AVE".
func AbbreviateStreetType(s string) (string, error) {
	capitalized := strings.ToUpper(strings.TrimSpace(s))

	if short, ok := streetTypeMap[capitalized]; ok {
		return short, nil
	}
	if _, ok := streetTypeShortMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized street type")
}

// NormalizeSecondary maps a Puerto Rico secondary designator (primary or short)
// to the uppercase abbreviation. Example: "Apartamento" or "APT" -> "APT".
func NormalizeSecondary(s string) (string, error) {
	capitalized := strings.ToUpper(strings.TrimSpace(s))
	// collapse internal whitespace for multi-word designators (e.g. GEN DEL)
	capitalized = strings.Join(strings.Fields(capitalized), " ")

	if short, ok := secondaryMap[capitalized]; ok {
		return short, nil
	}
	if _, ok := secondaryShortMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized secondary designator")
}

// prPostalPrefixes are the three-digit ZIP prefixes assigned to Puerto Rico.
// 008 is deliberately absent: it belongs to the US Virgin Islands, which does
// not use the Spanish vocabulary in this package.
var prPostalPrefixes = map[string]bool{
	"006": true,
	"007": true,
	"009": true,
}

// LooksLikePRPostal reports whether a postal code falls in a Puerto Rico ZIP
// range. Non-digits are ignored, so "00926" and "00926-1234" both match.
func LooksLikePRPostal(postal string) bool {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, postal)

	if len(digits) < 5 {
		return false
	}

	return prPostalPrefixes[digits[:3]]
}

// UsePRDialect reports whether Puerto Rico Spanish vocabulary applies to an
// address. A region code of PR is authoritative; a Puerto Rico ZIP range
// engages the dialect on its own, so an address that arrives without a region
// is still read in Spanish.
func UsePRDialect(regionCode, postal string) bool {
	if strings.EqualFold(strings.TrimSpace(regionCode), "PR") {
		return true
	}

	return LooksLikePRPostal(postal)
}
