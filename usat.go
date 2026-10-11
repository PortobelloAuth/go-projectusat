package goprojectusat

import (
	"fmt"
	"sync"

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

	u, err := build(popts, nopts, fopts)
	if err != nil {
		return nil, fmt.Errorf("Unable to instantiate parser: %w", err)
	}
	return u, nil
}

// build returns parser.New's error unwrapped, so the package-level Parse can
// keep returning exactly what it always has.
func build(popts parser.AddressParsingOptions, nopts normalizer.AddressNormalizationOptions, fopts address.FormatOptions) (*USAt, error) {
	p, err := parser.New(popts)
	if err != nil {
		return nil, err
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
	return addr.Format(u.format), nil
}

// Parse parses source into a structured address, using the parser and
// verifier u was built with. It does not normalize or format.
func (u *USAt) Parse(source string) (*address.Address, error) {
	return u.parser.Parse(source)
}

// defaultUSAt is the USAt the package-level Normalize and Parse use when they
// are given no options. It is built on first use, not at import, so a
// program that only ever uses a custom parser never loads the default one.
// Like zipcityembedded's own singleton, a construction error is kept and
// returned to every caller rather than retried.
var defaultUSAt = sync.OnceValues(func() (*USAt, error) {
	return build(parser.AddressParsingOptions{}, normalizer.AddressNormalizationOptions{}, address.FormatOptions{})
})
