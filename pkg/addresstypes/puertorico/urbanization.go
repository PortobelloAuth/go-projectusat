package puertorico

import (
	"fmt"
	"maps"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/poetic-systems/addresstables/puertorico"
)

// urbanizationDesignators are the spellings that open an urbanization, mapped
// to the abbreviation the standard requires.
//
// The Spanish spelling carries an accent — URBANIZACIÓN — and the shared table
// holds only the unaccented form. Lookups fold the input rather than the table
// carrying both spellings, so a form that adds another accented designator
// needs one row upstream and no second thought about how it is typed.
var urbanizationDesignators = maps.Collect(func(yield func(string, string) bool) {
	for u := range puertorico.Urbanizations() {
		if !yield(u.Full, u.Short) {
			return
		}
	}
})

// NormalizeUrbanization maps an urbanization designator to its abbreviation.
// Example: "Urbanización", "URBANIZACION" or "urb" -> "URB".
//
// The designator is a closed vocabulary, so an error here means the input is
// not one — see CONTRIBUTING §1.6. The development name that follows a
// designator is free text and is not this function's business; there is
// nothing in the standard to validate it against.
func NormalizeUrbanization(s string) (string, error) {
	folded, err := diacritics.Substitute(s)
	if err != nil {
		return "", fmt.Errorf("normalizing urbanization designator: %w", err)
	}

	if short, ok := urbanizationDesignators[strings.ToUpper(strings.TrimSpace(folded))]; ok {
		return short, nil
	}

	return "", fmt.Errorf("Unrecognized urbanization designator")
}

// standaloneUrbanizationMap maps a standalone urbanization name — the
// unabbreviated primary form or the abbreviation itself — to its
// abbreviation. Built from addresstables' StandaloneUrbanizations rather than
// duplicated here, per CONTRIBUTING §1.2.
var standaloneUrbanizationMap = maps.Collect(func(yield func(string, string) bool) {
	for u := range puertorico.StandaloneUrbanizations() {
		if !yield(u.Full, u.Short) {
			return
		}
	}
})

// standaloneUrbanizationShortMap is standaloneUrbanizationMap inverted, so an
// input that already arrives abbreviated is recognized too — the same
// primary-or-short symmetry NormalizeSecondary and NormalizeStreetType give
// their own vocabularies.
var standaloneUrbanizationShortMap = maps.Collect(func(yield func(string, string) bool) {
	for primary, short := range standaloneUrbanizationMap {
		if !yield(short, primary) {
			return
		}
	}
})

// NormalizeStandaloneUrbanization maps a Puerto Rico standalone urbanization
// name — the exceptions table from Technical Specification v1.0 pp. 28-29 —
// to its uppercase abbreviation. Example: "Extensión" (folds to "EXTENSION")
// or "EXT" -> "EXT"; "Alturas" or "ALTS" -> "ALTS".
//
// The standard requires these names to stand alone: "these urbanizations...
// MUST NOT require the use of the abbreviation URB." That is a claim.go
// concern (whether URB precedes the name at all); this function only answers
// whether a single word is one of the 38 names, the same closed-vocabulary
// question NormalizeUrbanization answers for the ordinary designators. An
// error here means the input is not one of them — see CONTRIBUTING §1.6.
func NormalizeStandaloneUrbanization(s string) (string, error) {
	folded, err := diacritics.Substitute(s)
	if err != nil {
		return "", fmt.Errorf("normalizing standalone urbanization name: %w", err)
	}

	capitalized := strings.ToUpper(strings.TrimSpace(folded))

	if short, ok := standaloneUrbanizationMap[capitalized]; ok {
		return short, nil
	}
	if _, ok := standaloneUrbanizationShortMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized standalone urbanization name")
}
