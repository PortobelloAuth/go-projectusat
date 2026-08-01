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

	// Canadian / other international: uppercase, collapse space, drop punctuation.
	return textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{})), nil
}

func MostLikelyTokens(t []token.Token) []int {
	//
	return nil
}
