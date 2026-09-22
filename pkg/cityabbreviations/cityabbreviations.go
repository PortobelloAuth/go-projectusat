// Package cityabbreviations reads a city-name abbreviation out of an address.
//
// The rows are held once in github.com/poetic-systems/addresstables and read
// from there: the four words a city name is commonly abbreviated with —
// SAINT, SAINTE, MOUNT, FORT — and the spelling Publication 28 §223 and
// Project US@ (p. 20) require in their place: city names SHALL be spelled out
// in their entirety.
//
// Where a word stands in a name is not this package's business. The shared
// table's doc comment states the rule: ST is SAINT at the head of ST LOUIS
// and nothing at the end of it. Only the caller knows whether another word
// follows, so Expand answers just "what does this word spell out to" and
// leaves the position rule to the caller. That same table is what a street
// name's head word is read against too (go-projectusat#114): no street is
// named STREET CLAIR, and SAINT CLAIR is common.
package cityabbreviations

import (
	"fmt"
	"maps"
	"strings"

	"github.com/poetic-systems/addresstables/cityabbreviations"
)

var expansionMap = maps.Collect(func(yield func(string, string) bool) {
	for a := range cityabbreviations.All() {
		if !yield(a.Short, a.Full) {
			return
		}
	}
})

// Expand returns the spelled-out form of a city-name abbreviation, e.g. ST ->
// SAINT. It reports an error if word is not one of the abbreviations this
// table holds; the caller decides, from what follows word, whether to call
// Expand at all.
func Expand(word string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(word)

	full, ok := expansionMap[capitalized]
	if ok {
		return full, nil
	}

	return "", fmt.Errorf("Unrecognized city abbreviation")
}
