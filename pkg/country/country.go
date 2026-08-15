package country

import (
	"regexp"
	"strings"
)

// country is always the last line of an address in a multiline address according to US Project@ rules.
// Commas should be interpreted similarly. In a single line address without commas, we may parse the
// country as the set of tokens containing only letters after what looks like a zip code.

// NOTE: Project US@ doesn't describe any logic for normalizing or validating international country names
// beyond specifying that they should not be abbreviated and should be capitalized. This map handles that
// for common two-letter abbreviations for Canada and Mexico only, since the specification is specifically
// scoped to US and some Canadian addresses and does not provide guidance on normalizing abbreviations of
// international country names.
var countryNameMap = map[string]string{
	"UNITED STATES": "",
	"USA":           "",
	"US":            "",
	"CA":            "CANADA",
	"CANADA":        "CANADA",
	"MX":            "MEXICO",
	"MEXICO":        "MEXICO",
}

var andpunctuation = regexp.MustCompile(`\+|\&`)
var whitespace = regexp.MustCompile(`\s+`)

// punctuation matches what Project US@ treats as punctuation in a country
// name. Digits survive it. This package formats rather than validates, so it
// has no basis for judging a digit unreadable — and stripping one turns input
// it did not recognize into a country it did: M5X is a Toronto postal code and
// became MEXICO, and 2US became the empty string, which this package uses to
// mean "domestic US address".
var punctuation = regexp.MustCompile("[^a-zA-Z0-9 ]+")

// NormalizeCountry formats a country name. Domestic spellings are blanked and
// the recognized neighbours are canonicalized; anything else is returned
// cleaned and uppercased.
//
// This is one of the two exceptions to the rule that a Normalize function over
// a closed vocabulary errors on input it does not recognize. It never returns
// a non-nil error, including for empty input. The standard gives no basis for
// validating an international country name, so refusing input the library
// cannot judge would be worse than passing it through. Callers must not read
// a nil error here as recognition.
func NormalizeCountry(r string) (string, error) {
	// Replace "+" or "&" with "AND" in country names
	andfixed := andpunctuation.ReplaceAllString(r, " AND ")
	// condense consecutive whitespace characters to a single space
	spaced := whitespace.ReplaceAllString(andfixed, " ")
	// clean out any punctuation
	clean := punctuation.ReplaceAllString(spaced, "")
	// capitalize
	capitalized := strings.ToUpper(clean)

	substituted, ok := countryNameMap[capitalized]
	if ok {
		return substituted, nil
	}

	return capitalized, nil
}
