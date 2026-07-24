package address

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Address struct {
	BusinessName string // firm / business line (optional)

	// Street line elements
	PrimaryNumber       string
	Predirectional      string
	StreetName          string
	StreetSuffix        string
	Postdirectional     string
	SecondaryDesignator string // APT, STE, ...
	SecondaryNumber     string

	// Last line
	City   string
	Region string // state / province / military "state"
	Postal string // ZIP, ZIP+4, or Canadian postal code

	Country string // optional; often blank for domestic
}

func (a *Address) Format(opt AddressFormatOptions) string {
	sep := "\n"
	if opt.Oneline {
		sep = " "
	}
	return joinNonEmpty(sep, a.BusinessName, FormatStreetLine(*a), FormatLastLine(*a))
}

type AddressFormatOptions struct {
	Oneline    bool
	Comma      bool
	Diacritics diacritics.DiacriticMode
}

// FormatStreetLine joins street elements with single spaces; omits blanks.
// Order: PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM.
func FormatStreetLine(a Address) string {
	return joinNonEmpty(" ",
		a.PrimaryNumber,
		a.Predirectional,
		a.StreetName,
		a.StreetSuffix,
		a.Postdirectional,
		a.SecondaryDesignator,
		a.SecondaryNumber,
	)
}

// FormatLastLine joins CITY REGION POSTAL with single spaces (1+ spaces allowed
// by standard; we use one). Omits blanks.
func FormatLastLine(a Address) string {
	return joinNonEmpty(" ", a.City, a.Region, a.Postal)
}

// Format returns business line (if any), street line, and last line separated
// by newlines. Empty lines are omitted.
func Format(a Address) string {
	return joinNonEmpty("\n", a.BusinessName, FormatStreetLine(a), FormatLastLine(a))
}

// joinNonEmpty joins non-empty parts with sep.
func joinNonEmpty(sep string, parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}
