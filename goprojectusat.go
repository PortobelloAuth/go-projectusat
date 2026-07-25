// Package goprojectusat normalizes patient address strings per Project US@.

package goprojectusat

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// TODO: export the primary interfaces for the package

type USAtNormalizeOption func(*parser.AddressParsingOptions, *normalizer.AddressNormalizationOptions, *address.FormatOptions) error

func Normalize(source string, opts ...USAtNormalizeOption) (string, error) {
	// Normalize uses parser, normalizer, and format. It needs to support the options for all of them.
	// use function options pattern
	popts := &parser.AddressParsingOptions{}
	nopts := &normalizer.AddressNormalizationOptions{}
	fopts := &address.FormatOptions{}
	for _, fn := range opts {
		err := fn(popts, nopts, fopts)
		if err != nil {
			return "", fmt.Errorf("Error setting normailzation options: %w", err)
		}
	}

	// TODO: create a parser, then Parse(popts)
	// TODO: create a normalizer, then Normalize(nopts)
	// TODO: call *address.Format(fopts)

	return "", fmt.Errorf("Not implemented")
}

func WithParsedAddressVerifier(v parser.AddressVerifier) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		popts.Verifier = v
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

func WithDiacriticFormating(d diacritics.DiacriticMode) USAtNormalizeOption {
	return func(popts *parser.AddressParsingOptions, nopts *normalizer.AddressNormalizationOptions, fopts *address.FormatOptions) error {
		fopts.DiacriticMode = d
		return nil
	}
}

func Parse(source string, opts ...parser.AddressParsingOptions) (*address.Address, error) {
	return nil, fmt.Errorf("Not implemented")
}
