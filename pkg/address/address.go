package address

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

type AddressType interface {
	FormatStreetLine(a *Address) string
}

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Address struct {
	// Type indicates that this address is a special format, such as ruralroute,
	// military, pobox, puertorico, or streetsuffixfirst
	Type         AddressType
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
	return textutil.JoinNonEmpty(sep, a.BusinessName, a.FormatStreetLine(), a.FormatLastLine())
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
// Normal Order: PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM.
// Rural Route Order: STREET PRIMARY SEC SECNUM.
func (a *Address) FormatStreetLine() string {
	if a.Type != nil {
		return a.Type.FormatStreetLine(a)
	}
	return textutil.JoinNonEmpty(" ",
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
	return textutil.JoinNonEmpty(" ", a.City, a.Region, a.Postal)
}
