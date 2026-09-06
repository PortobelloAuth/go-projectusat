// Package directionals reads a directional out of an address.
//
// The rows are the Publication 28 directionals, held once in
// github.com/poetic-systems/addresstables and read from there. Only the
// English rows are read: the Spanish spellings abbreviate to O, NO and SO,
// which the ZIP+4 file does not use and which this package must not emit, so
// accepting them is a separate decision about Puerto Rico addresses (#71)
// rather than a property of this table.
package directionals

import (
	"fmt"
	"maps"
	"strings"

	"github.com/poetic-systems/addresstables/directionals"
)

var directionMap = maps.Collect(func(yield func(string, string) bool) {
	for d := range directionals.English() {
		if !yield(d.Full, d.Short) {
			return
		}
	}
})

var directionShortMap = maps.Collect(func(yield func(string, string) bool) {
	for k, v := range directionMap {
		if !yield(v, k) {
			return
		}
	}
})

func AbbreviateDirectional(d string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(d)

	// look up the abbreviation
	abrev, ok := directionMap[capitalized]
	if ok {
		return abrev, nil
	}

	// See if it is an abbreviation
	_, ok = directionShortMap[capitalized]
	if ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized directional")
}

func NormalizeDirectional(d string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(d)

	// look up the primary (full) direction word
	full, ok := directionShortMap[capitalized]
	if ok {
		return full, nil
	}

	// See if it is the primary (full) direction word
	_, ok = directionMap[capitalized]
	if ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized directional")
}
