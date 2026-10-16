package goprojectusat

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
)

// USAt is a configured Project US@ address pipeline: parse, normalize, then
// format. Its options are applied once, by New, and every call to Normalize
// or Parse reuses the parser, normalizer and format options built then,
// rather than rebuilding them per call the way the package-level Normalize
// and Parse functions do when they are given options.
//
// A USAt is safe for concurrent use as long as any CustomParser or Verifier
// it was configured with is. It holds no per-call state of its own.
type USAt struct {
	parser     *parser.Parser
	normalizer *normalizer.Normalizer
	format     address.FormatOptions
}

// Option configures a USAt. It is the same type as USAtNormalizeOption, so
// every existing With* option works with both New and the package-level
// Normalize.
type Option = USAtNormalizeOption

// New builds a USAt from opts. Options are applied in order, so a later
// option overrides an earlier one that sets the same field.
//
// Unless WithCustomAddressParser is given, the USAt uses the default
// zipcity-backed parser. That parser's ~58MB of state is built once per
// process (see zipcityembedded) no matter how many USAts are built, so
// building several differently configured USAts is cheap. With a custom
// parser, the default parser's state is never loaded at all.
func New(opts ...Option) (*USAt, error) {
	popts := parser.AddressParsingOptions{}
	nopts := normalizer.AddressNormalizationOptions{}
	fopts := address.FormatOptions{}
	for _, fn := range opts {
		if err := fn(&popts, &nopts, &fopts); err != nil {
			return nil, fmt.Errorf("Error setting normailzation options: %w", err)
		}
	}

	// u, err := build(popts, nopts, fopts)
	p, err := parser.New(popts)
	if err != nil {
		return nil, fmt.Errorf("Unable to instantiate parser: %w", err)
	}
	// return u, nil
	return &USAt{
		parser:     p,
		normalizer: normalizer.NewNomalizer(nopts),
		format:     fopts,
	}, nil
}

// // build returns parser.New's error unwrapped, so the package-level Parse can
// // keep returning exactly what it always has.
// func build(popts parser.AddressParsingOptions, nopts normalizer.AddressNormalizationOptions, fopts address.FormatOptions) (*USAt, error) {
// 	p, err := parser.New(popts)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &USAt{
// 		parser:     p,
// 		normalizer: normalizer.NewNomalizer(nopts),
// 		format:     fopts,
// 	}, nil
// }

// Normalize parses source, normalizes the result, and formats it, using the
// options u was built with.
func (u *USAt) Normalize(source string) (string, error) {
	addr, err := u.parser.Parse(source)
	if err != nil {
		return "", fmt.Errorf("Unable to parse address: %w", err)
	}
	addr, err = u.normalizer.Normalize(addr)
	if err != nil {
		return "", fmt.Errorf("Unable to normalize address: %w", err)
	}
	return addr.Format(u.format)
}

// Parse parses source into a structured address, using the parser and
// verifier u was built with. source may be written on one line or on several,
// and its parts may be separated by commas or not.
//
// The parser reads each element as the standard names it, so the result is
// already upper case, with the USPS Publication 28 abbreviations: South Tech
// Drive Suite 3200 is parsed as Predirectional S, StreetName TECH,
// StreetSuffix DR, SecondaryDesignator STE. What Parse does not do is apply
// u's normalization options, such as WithMatchingNormalization, or format the
// result. Pass the result to NormalizeAddress for the first, and call its
// Format method for the second. Normalize does all three.
//
// The returned address's Type records which kind of address it was read as,
// for example a PO box or an ordinary street address. Type is what the
// address formats by, so keep it on the result rather than clearing it.
//
// It returns an error, and no address, when source cannot be read as any kind
// of address, or when the verifier given to WithParsedAddressVerifier rejects
// the reading.
func (u *USAt) Parse(source string) (*address.Address, error) {
	return u.parser.Parse(source)
}

// NormalizeAddress normalizes a structured address, using the normalization
// options u was built with, and returns the result as a new address. a itself
// is not modified.
//
// a is usually the result of Parse, but it does not have to be. A caller that
// has an address in fields already, or only some of them, can normalize just
// those: an Address holding only City "salt lake city" and Region "utah"
// normalizes to SALT LAKE CITY and UT, and its other fields stay empty. Such
// an address has no Type, and is normalized by the rules for an ordinary
// street address.
//
// It does not format. Call Format on the result for the string Normalize
// would have returned.
func (u *USAt) NormalizeAddress(a *address.Address) (*address.Address, error) {
	return u.normalizer.Normalize(a)
}

// // defaultUSAt is the USAt the package-level Normalize and Parse use when they
// // are given no options. It is built on first use, not at import, so a
// // program that only ever uses a custom parser never loads the default one.
// // Like zipcityembedded's own singleton, a construction error is kept and
// // returned to every caller rather than retried.
// var defaultUSAt = sync.OnceValues(func() (*USAt, error) {
// 	return build(parser.AddressParsingOptions{}, normalizer.AddressNormalizationOptions{}, address.FormatOptions{})
// })
