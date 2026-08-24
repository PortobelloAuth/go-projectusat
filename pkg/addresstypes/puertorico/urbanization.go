package puertorico

import (
	"fmt"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// urbanizationDesignators are the spellings that open an urbanization, mapped
// to the abbreviation the standard requires.
//
// The Spanish spelling carries an accent — URBANIZACIÓN — and this table holds
// only the unaccented form. Lookups fold the input rather than the table
// carrying both spellings, so a form that adds another accented designator
// needs one entry here and no second thought about how it is typed.
var urbanizationDesignators = map[string]string{
	"URB":          "URB",
	"URBANIZACION": "URB",
	"URBANIZATION": "URB",
}

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
