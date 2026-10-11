// Package goprojectusat normalizes patient address strings per Project US@.

package goprojectusat

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// Export the primary interfaces for the package

type USAtNormalizeOption func(*parser.AddressParsingOptions, *normalizer.AddressNormalizationOptions, *address.FormatOptions) error

// Normalize parses, normalizes and formats source.
//
// With no options it uses a package-level USAt that is built once, on first
// use. With options it builds a new USAt for this call only, so a caller
// normalizing many addresses with the same options should build one with New
// and call its Normalize method instead.
func Normalize(source string, opts ...USAtNormalizeOption) (string, error) {
	var u *USAt
	var err error
	if len(opts) == 0 {
		u, err = defaultUSAt()
		if err != nil {
			return "", fmt.Errorf("Unable to instantiate parser: %w", err)
		}
	} else {
		u, err = New(opts...)
		if err != nil {
			return "", err
		}
	}
	return u.Normalize(source)
}

func WithParsedAddressVerifier(v parser.AddressVerifier) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		popts.Verifier = v
		return nil
	}
}

func WithCustomAddressParser(cp parser.ParsingFunc) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		popts.CustomParser = cp
		return nil
	}
}

func WithFuzzyNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.Fuzzy = true
		return nil
	}
}

func WithSecondaryAsHash() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.SecondaryAsHash = true
		return nil
	}
}

func WithDiacriticNormalization(d diacritics.DiacriticMode) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.DiacriticMode = d
		return nil
	}
}

func WithContentNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.Fuzzy = false
		nopts.SecondaryAsHash = false
		nopts.DiacriticMode = diacritics.KeepDiacritics
		return nil
	}
}

func WithMatchingNormalization() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		nopts.Fuzzy = true
		nopts.SecondaryAsHash = true
		nopts.DiacriticMode = diacritics.SubstituteDiacritics
		return nil
	}
}

func WithSingleLineFormatting() USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.SingleLine = true
		return nil
	}
}

func WithDiacriticFormatting(d diacritics.DiacriticMode) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.DiacriticMode = d
		return nil
	}
}

// Parse parses source into a structured address. Although opts is variadic,
// only the first options object is used.
//
// With no options it uses the same package-level USAt as Normalize. With
// options it builds a parser for this call only; a caller parsing many
// addresses with the same options should build a USAt with New instead.
func Parse(source string, opts ...parser.AddressParsingOptions) (*address.Address, error) {
	if len(opts) == 0 {
		u, err := defaultUSAt()
		if err != nil {
			return nil, err
		}
		return u.Parse(source)
	}
	p, err := parser.New(opts[0])
	if err != nil {
		return nil, err
	}
	return p.Parse(source)
}
