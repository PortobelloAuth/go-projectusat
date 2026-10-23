// Package goprojectusat normalizes patient address strings per Project US@.

package goprojectusat

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizeroptions"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// Export the primary interfaces for the package

type USAtNormalizeOption func(*parser.AddressParsingOptions, *normalizeroptions.AddressNormalizationOptions, *address.FormatOptions) error

// WithParsedAddressVerifier provides a `parser.AddessVerifier` function to a USAt instance's
// internal parser. This function is called after an address has been parsed by the internal
// parser (which may be a custom address parser or `parser.ParsingFunc`) and before the
// internal address normalizer.
func WithParsedAddressVerifier(v parser.AddressVerifier) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		popts.Verifier = v
		return nil
	}
}

// WithCustomAddressParser allows passing in a `parser.ParsingFunc` to replace the core parsing
// function of a USAt instance. By default `goprojectusat.USAt` relies on a parser that uses
// [zipcity](https://github.com/poetic-systems/zipcity) to disambiguate between streets, cites,
// and regions. `zipcity`'s bloom filters are an efficient method for doing this without calling
// an API or running a database. However, they may not serve every use case. WithCustomAddressParser
// lets a caller use their own data store, API, or other custom parsing logic.
func WithCustomAddressParser(cp parser.ParsingFunc) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		popts.CustomParser = cp
		return nil
	}
}

// WithFuzzyNormalization enbales fuzzy matching in the normalization functions that support
// it - particularly `streetsuffixes` and `region` - have support for fuzzy matching using
// Damerau-Levenshtein edit distance.
func WithFuzzyNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.Fuzzy = true
		return nil
	}
}

// WithSecondaryAsHash sets the normalization functions to replace secondary designators
// with "#". [Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
// allows the secondary designator of an address to be `"#"` when it is not known. As a
// result, any address with a numbered secondary designator may also be recorded with a
// `"#"` (unless the address also contains a Private Mailbox.) For a matching / exchange
// use case normalization must resolve to a single value in order to match. This option
// normalizes numbered secondary designators to `"#"` to facilitate such matching.
// NOTE: for addresses with multiple secodary units (i.e. BLDG 5 STE 110) only the first
// will be normalized to `"#"`. This is not likely to be the desired behavior.
func WithSecondaryAsHash() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.SecondaryAsHash = true
		return nil
	}
}

// WithDiacriticNormalization sets the diacritics.DiacriticMode used during normalization.
// [Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
// allows diacritics to be retained or substituted according to a table it supplies.
// `WithDiacriticNormalization()` allows the caller to specify how they would like
// the parser and normalizer to handle diacritic substitution.
// `diacritics.DiacriticMode` is an enumeration with 3 possible values:
//   - `diacritics.KeepDiacritics` does not change diacritics
//   - `diacritics.SubstituteDiacritics` substitutes diacritics according to the
//     ProjectUS@ specification
//   - `diacritics.TransliterateDiacritics` transliterates diacritics, replacing
//     them with ASCII letter combinations that commonly represent them.
func WithDiacriticNormalization(d diacritics.DiacriticMode) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.DiacriticMode = d
		return nil
	}
}

// WithContentNormalization() explicitly sets the options recommended for "content form"
// normalization. These options that are appropriate for address intake and storage: an
// exact controlled vocabulary, diacritic preservation, and secondary designators as
// their standard abbreviations (`APT`, `STE`, …). These same options are used by default.
// However, using this Option clearly indicates the caller's intent to use content normalization.
func WithContentNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		c := normalizeroptions.Content()
		nopts.Fuzzy = c.Fuzzy
		nopts.SecondaryAsHash = c.SecondaryAsHash
		nopts.DiacriticMode = c.DiacriticMode
		return nil
	}
}

// WithMatchingNormalization() sets the options recommended for "exchange form", also known as
// matching normalization. These options that are appropriate for address matching: fuzzy matching
// of certain vocabularies, diacritic substitution, and numbered secondary designators converted
// to `"#"`.
func WithMatchingNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		m := normalizeroptions.Matching()
		nopts.Fuzzy = m.Fuzzy
		nopts.SecondaryAsHash = m.SecondaryAsHash
		nopts.DiacriticMode = m.DiacriticMode
		return nil
	}
}

// WithSingleLineFormatting specifies that addresses should be formatted without newlines.
func WithSingleLineFormatting() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.SingleLine = true
		return nil
	}
}

// WithDiacriticFormatting allows formatting an address that has been parsed keeping diacritics
// in a format with diacritics substituted or transliterated, without changing the source address
// data.
func WithDiacriticFormatting(d diacritics.DiacriticMode) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizeroptions.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.DiacriticMode = d
		return nil
	}
}

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
	nopts := normalizeroptions.AddressNormalizationOptions{}
	fopts := address.FormatOptions{}
	for _, fn := range opts {
		if err := fn(&popts, &nopts, &fopts); err != nil {
			return nil, fmt.Errorf("Error setting normailzation options: %w", err)
		}
	}

	p, err := parser.New(popts)
	if err != nil {
		return nil, fmt.Errorf("Unable to instantiate parser: %w", err)
	}

	return &USAt{
		parser:     p,
		normalizer: normalizer.NewNomalizer(nopts),
		format:     fopts,
	}, nil
}

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
