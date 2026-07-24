package postalcode

import (
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

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
