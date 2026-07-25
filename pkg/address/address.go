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

type FormatOptions struct {
	SingleLine    bool
	DiacriticMode diacritics.DiacriticMode
}

// Format converts an Address struct to a string according to the
// supplied FormatOptions.
//
// Although options is variadic, only the first options object will
// actually be used.
func (a *Address) Format(opts ...FormatOptions) string {
	o := variadicOptions(opts)
	sep := "\n"
	if o.SingleLine {
		sep = " "
	}
	return joinNonEmpty(sep, a.BusinessName, a.FormatStreetLine(), a.FormatLastLine())
}

func (a *Address) FormatSingleLine(opts ...FormatOptions) string {
	o := variadicOptions(opts)
	o.SingleLine = true
	return a.Format(o)
}

func variadicOptions(opts []FormatOptions) FormatOptions {
	o := FormatOptions{
		SingleLine:    false,
		DiacriticMode: diacritics.KeepDiacritics,
	}
	if len(opts) > 0 {
		o = opts[0]
	}
	return o
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
