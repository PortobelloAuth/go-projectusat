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
var alphaspace = regexp.MustCompile("[^a-zA-Z ]+")

func NormalizeCountry(r string) (string, error) {
	// Replace "+" or "&" with "AND" in country names
	andfixed := andpunctuation.ReplaceAllString(r, " AND ")
	// condense consecutive whitespace characters to a single space
	spaced := whitespace.ReplaceAllString(andfixed, " ")
	// clean out any punctuation
	clean := alphaspace.ReplaceAllString(spaced, "")
	// capitalize
	capitalized := strings.ToUpper(clean)

	substituted, ok := countryNameMap[capitalized]
	if ok {
		return substituted, nil
	}

	return capitalized, nil
}
