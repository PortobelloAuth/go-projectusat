package address

import (
	"reflect"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// AddressType is the special format an address follows, if it follows one.
//
// Implementations describe a shape rather than an address, so they carry no
// per-address state: RuralRouteAddress is an empty struct, and one value of it
// serves every rural route address. Equals relies on that — it compares which
// implementation is present, not which instance of it.
type AddressType interface {
	FormatStreetLine(a *Address) string
}

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
//
// Compare addresses with Equals rather than ==. See that method for why.
type Address struct {
	// Type indicates that this address is a special format, such as ruralroute,
	// military, pobox, puertorico, or streetsuffixfirst
	Type         AddressType
	BusinessName string // firm / business line (optional)

	// Area is the urbanization or similar neighborhood-like construct that
	// precedes the street line. USPS Publication 28 describes an urbanization
	// as "an area, sector, or development within a geographic area", which is
	// why the field is named for what it is rather than for where it sits.
	//
	// Project US@ requires it on Puerto Rican addresses, where it is the first
	// line of the street address block:
	//
	//	URB HIGHLAND GDNS
	//	COND LAS AMAPOLAS APT 103
	//	123 CALLE MAIN
	//
	// It is not exclusive to Puerto Rico by definition, but that is the only
	// dialect the standard requires it for today.
	Area string

	// Street line elements
	PrimaryNumber       string
	Predirectional      string
	StreetName          string
	StreetSuffix        string
	Postdirectional     string
	SecondaryDesignator string // APT, STE, ...
	SecondaryNumber     string

	// Detail is a further address number qualifying the secondary one, such as
	// the "PMB 456" of a private mailbox at a street address that already has
	// a suite. It is effectively a tertiary number, and is named for its role
	// rather than by that word because "tertiary" says its depth and not its
	// purpose.
	//
	// It is deliberately not a catchall. Anything that is not a number
	// qualifying the secondary designator belongs in the field that names it.
	Detail string

	// Last line
	City   string
	Region string // state / province / military "state"
	Postal string // ZIP, ZIP+4, or Canadian postal code

	Country string // optional; often blank for domestic
}

// Equals reports whether two addresses say the same thing. A nil address
// equals only another nil address.
//
// This exists because Address must not be compared with ==. The Type field is
// an interface, so == compares the dynamic value behind it, and that is wrong
// in two ways at once. A parsed address carries a Type while the expected
// value it is compared against usually does not, so the comparison fails on a
// field the caller never meant to assert and prints two addresses that look
// identical. And == on an interface panics at run time when the dynamic type
// is not comparable, which makes "an AddressType must never hold a slice, map,
// or func" an invariant the compiler cannot check.
//
// Type is compared by which implementation is present rather than by value.
// What separates a rural route address from a PO box is the shape it follows,
// and implementations hold no state to compare beyond that.
func (a *Address) Equals(other *Address) bool {
	if a == nil || other == nil {
		return a == other
	}

	if reflect.TypeOf(a.Type) != reflect.TypeOf(other.Type) {
		return false
	}

	return a.BusinessName == other.BusinessName &&
		a.Area == other.Area &&
		a.PrimaryNumber == other.PrimaryNumber &&
		a.Predirectional == other.Predirectional &&
		a.StreetName == other.StreetName &&
		a.StreetSuffix == other.StreetSuffix &&
		a.Postdirectional == other.Postdirectional &&
		a.SecondaryDesignator == other.SecondaryDesignator &&
		a.SecondaryNumber == other.SecondaryNumber &&
		a.Detail == other.Detail &&
		a.City == other.City &&
		a.Region == other.Region &&
		a.Postal == other.Postal &&
		a.Country == other.Country
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
	return textutil.JoinNonEmpty(sep, a.BusinessName, a.Area, a.FormatStreetLine(), a.FormatLastLine())
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
// Normal Order: PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM DETAIL.
// Rural Route Order: STREET PRIMARY SEC SECNUM.
//
// Detail qualifies the secondary number, so it follows it. Area is not a
// street line element at all — it is its own line above one — and is rendered
// by Format rather than here.
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
		a.Detail,
	)
}

// FormatLastLine joins CITY REGION POSTAL with single spaces (1+ spaces allowed
// by standard; we use one). Omits blanks.
func (a *Address) FormatLastLine() string {
	return textutil.JoinNonEmpty(" ", a.City, a.Region, a.Postal)
}
