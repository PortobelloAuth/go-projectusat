package puertorico

import "strings"

// romanNumerals maps the canonical spelling of a building number p. 25 asks
// for — "Where there are multiple buildings (or towers) with the same name,
// the building number SHOULD become the primary number" — to the arabic form
// a primary number is written in. Capped at 20: Project US@ gives no example
// past a handful of towers, and a longer table would be inventing precision
// the standard does not ask for. An error means the input is not one of
// these spellings, per CONTRIBUTING §1.6.
var romanNumerals = map[string]string{
	"I": "1", "II": "2", "III": "3", "IV": "4", "V": "5",
	"VI": "6", "VII": "7", "VIII": "8", "IX": "9", "X": "10",
	"XI": "11", "XII": "12", "XIII": "13", "XIV": "14", "XV": "15",
	"XVI": "16", "XVII": "17", "XVIII": "18", "XIX": "19", "XX": "20",
}

// romanNumeral reports the arabic form of a token that is exactly one of the
// canonical roman numeral spellings above. Folds case so VISTA SUITES III and
// vista suites iii are read alike.
func romanNumeral(s string) (string, bool) {
	number, ok := romanNumerals[strings.ToUpper(s)]
	return number, ok
}
