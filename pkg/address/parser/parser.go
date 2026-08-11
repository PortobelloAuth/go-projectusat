package parser

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
)

// AddressVerifier functions take an address.Address and return it if it
// passes verification. Otherwise an error is returned indicating the
// issue.
type AddressVerifier func(*address.Address) (*address.Address, error)

// AddressParsingOptions controls how address parsing is done.
// The zero value has no Verifier function
type AddressParsingOptions struct {
	Verifier     AddressVerifier
	CustomParser ParsingFunc
}

type ParsingFunc interface {
	Parse(source string) (*address.Address, error)
}
type ParsingFn func(source string) (*address.Address, error)

func (pf ParsingFn) Parse(source string) (*address.Address, error) {
	return pf(source)
}

func IdentityVerifier(a *address.Address) (*address.Address, error) {
	return a, nil
}

// Parser is used to parse a string in to a Project US@ structured
// patient address.
type Parser struct {
	Options AddressParsingOptions
}

// New creates a new Parser using the provided AddressParsingOptions.
// Although options is variadic, only the first options object will
// actually be used.
func New(opts ...AddressParsingOptions) *Parser {
	o := AddressParsingOptions{}
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Verifier == nil {
		o.Verifier = IdentityVerifier
	}

	return &Parser{
		Options: o,
	}
}

func (p *Parser) Parse(source string) (*address.Address, error) {
	if p.Options.CustomParser != nil {
		return p.Options.CustomParser.Parse(source)
	}
	// TODO: implement Parse
	/*
		- split the string on newlines, commas, and spaces
		- score each token
		- use scores to determine which Address parts each token belongs to
		- run verifier to check whether the address is verifiable (does not error)
	*/
	tokens := token.Tokenize(source)

	postalcode.MostLikelyTokens(tokens)
	// use the most likely zip code and/or region to select Puerto Rico or military
	// specific parsers.
	return nil, fmt.Errorf("Not implemented")
}
