package parser

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	zipcityembedded "github.com/PortobelloAuth/go-projectusat/pkg/address/parser/zipcityembedded/parse"
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
	verifier AddressVerifier
	parser   ParsingFn
}

// New creates a new Parser using the provided AddressParsingOptions.
// Although options is variadic, only the first options object will
// actually be used.
func New(opts ...AddressParsingOptions) (*Parser, error) {
	o := AddressParsingOptions{}
	if len(opts) > 0 {
		o = opts[0]
	}
	v := IdentityVerifier
	if o.Verifier != nil {
		v = o.Verifier
	}

	var p ParsingFn
	if o.CustomParser != nil {
		p = o.CustomParser.Parse
	} else {
		zc, err := zipcityembedded.New()
		if err != nil {
			return nil, err
		}
		p = zc.Parse
	}

	return &Parser{
		verifier: v,
		parser:   p,
	}, nil
}

func (p *Parser) Parse(source string) (*address.Address, error) {
	addr, err := p.parser(source)
	if err != nil {
		return nil, err
	}
	return p.verifier(addr)
}
