package address

import (
	"slices"
	"strings"
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

type FormatOptions int

const (
	SingleLine              FormatOptions = iota // 0 (Default/Zero-value)
	SubstituteDiacritics                         // 1
	TransliterateDiacritics                      // 2
)

func (a *Address) Format(opts ...FormatOptions) string {
	sep := "\n"
	if slices.Contains(opts, SingleLine) {
		sep = " "
	}
	return joinNonEmpty(sep, a.BusinessName, a.FormatStreetLine(), a.FormatLastLine())
}

func (a *Address) FormatSingleLine() string {
	return a.Format(SingleLine)
}

// FormatStreetLine joins street elements with single spaces; omits blanks.
// Order: PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM.
func (a *Address) FormatStreetLine() string {
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
func (a *Address) FormatLastLine() string {
	return joinNonEmpty(" ", a.City, a.Region, a.Postal)
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
