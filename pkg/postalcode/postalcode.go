package postalcode

import (
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// TODO: create or generate an efficient data structure to look up known
// possible city and state combinations by zip or zip+4 from the 3 sheets
// exported from https://postalpro.usps.com/ZIP_Locale_Detail as of July 2026.
// TODO: if nothing that looks like a reasonable city and state exists, try
// using the city and state of the physical delivery address (the post office?)
// TODO: identify a strategy for checking whether the ZIP Locale spreadsheet
// has been updated - likely via an environment variable

// usZIPCompact matches ##### or #####-#### / ######### after punctuation strip.
var usZIPCompact = regexp.MustCompile(`^(\d{5})(?:-?(\d{4}))?$`)

// caPostalCompact matches a Canadian postal code with the separating space
// removed (A1A1A1). Canada Post writes these as "A1A 1A1".
var caPostalCompact = regexp.MustCompile(`^[A-Z]\d[A-Z]\d[A-Z]\d$`)

// Normalize formats US ZIP / ZIP+4 and leaves Canadian (and other) patterns
// as uppercase alphanumerics with collapsed spacing.
func Normalize(s string) (string, error) {
	s = textutil.Upper(textutil.CollapseSpace(s))
	if s == "" {
		return "", nil
	}

	// Keep hyphen for ZIP+4; strip other Project US@ punctuation.
	cleaned := textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{KeepHyphen: true}))
	compact := strings.ReplaceAll(cleaned, " ", "")
	if m := usZIPCompact.FindStringSubmatch(compact); m != nil {
		if m[2] != "" {
			return m[1] + "-" + m[2], nil
		}
		return m[1], nil
	}

	// Canadian: the forward sortation area and local delivery unit are separated
	// by a single space, whether or not the input separated them at all. Unlike
	// ZIP+4 the hyphen carries no meaning here, so treat it as a separator.
	if caCompact := strings.ReplaceAll(compact, "-", ""); caPostalCompact.MatchString(caCompact) {
		return caCompact[:3] + " " + caCompact[3:], nil
	}

	// Other international: uppercase, collapse space, drop punctuation.
	return textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{})), nil
}

func MostLikelyTokens(t []token.Token) []int {
	//
	return nil
}
