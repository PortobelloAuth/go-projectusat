// Package goprojectusat normalizes patient address strings per Project US@.

package goprojectusat

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// Export the primary interfaces for the package

type USAtNormalizeOption func(*parser.AddressParsingOptions, *normalizer.AddressNormalizationOptions, *address.FormatOptions) error

// // Normalize parses, normalizes and formats source.
// //
// // With no options it uses a package-level USAt that is built once, on first
// // use. With options it builds a new USAt for this call only, so a caller
// // normalizing many addresses with the same options should build one with New
// // and call its Normalize method instead.
// func Normalize(source string, opts ...USAtNormalizeOption) (string, error) {
// 	var u *USAt
// 	var err error
// 	if len(opts) == 0 {
// 		u, err = defaultUSAt()
// 		if err != nil {
// 			return "", fmt.Errorf("Unable to instantiate parser: %w", err)
// 		}
// 	} else {
// 		u, err = New(opts...)
// 		if err != nil {
// 			return "", err
// 		}
// 	}
// 	return u.Normalize(source)
// }

// WithParsedAddressVerifier provides a `parser.AddessVerifier` function to a USAt instance's
// internal parser. This function is called after an address has been parsed by the internal
// parser (which may be a custom address parser or `parser.ParsingFunc`) and before the
// internal address normalizer.
func WithParsedAddressVerifier(v parser.AddressVerifier) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
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
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		popts.CustomParser = cp
		return nil
	}
}

// WithFuzzyNormalization enbales fuzzy matching in the normalization functions that support
// it - particularly `streetsuffixes` and `region` - have support for fuzzy matching using
// Damerau-Levenshtein edit distance.
func WithFuzzyNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
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
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
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
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
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
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.Fuzzy = false
		nopts.SecondaryAsHash = false
		nopts.DiacriticMode = diacritics.KeepDiacritics
		return nil
	}
}

// WithMatchingNormalization() sets the options recommended for "exchange form", also known as
// matching normalization. These options that are appropriate for address matching: fuzzy matching
// of certain vocabularies, diacritic substitution, and numbered secondary designators converted
// to `"#"`.
func WithMatchingNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.Fuzzy = true
		nopts.SecondaryAsHash = true
		nopts.DiacriticMode = diacritics.SubstituteDiacritics
		return nil
	}
}

// WithSingleLineFormatting specifies that addresses should be formatted without newlines.
func WithSingleLineFormatting() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.SingleLine = true
		return nil
	}
}

// WithDiacriticFormatting allows formatting an address that has been parsed keeping diacritics
// in a format with diacritics substituted or transliterated, without changing the source address
// data.
func WithDiacriticFormatting(d diacritics.DiacriticMode) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.DiacriticMode = d
		return nil
	}
}

// // Parse parses source into a structured address. Although opts is variadic,
// // only the first options object is used.
// //
// // With no options it uses the same package-level USAt as Normalize. With
// // options it builds a parser for this call only; a caller parsing many
// // addresses with the same options should build a USAt with New instead.
// func Parse(source string, opts ...parser.AddressParsingOptions) (*address.Address, error) {
// 	if len(opts) == 0 {
// 		u, err := defaultUSAt()
// 		if err != nil {
// 			return nil, err
// 		}
// 		return u.Parse(source)
// 	}
// 	p, err := parser.New(opts[0])
// 	if err != nil {
// 		return nil, err
// 	}
// 	return p.Parse(source)
// }
