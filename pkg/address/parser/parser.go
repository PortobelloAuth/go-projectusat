package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
)

// AddressVerifier functions take an address.Address and return it if it
// passes verification. Otherwise an error is returned indicating the
// issue.
type AddressVerifier func(*address.Address) (*address.Address, error)

// AddressParsingOptions controls how address parsing is done.
// The zero value has no Verifier function
type AddressParsingOptions struct {
	Verifier AddressVerifier
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
	// TODO: implement Parse
	/*
		- split the string on newlines, commas, and spaces
		- score each token
		- use scores to determine which Address parts each token belongs to
		- run verifier to check whether the address is verifiable (does not error)
	*/
	tokens := Tokenize(source)
	fmt.Printf("tokens: %v\n", tokens)
	return nil, fmt.Errorf("Not implemented")
}

type Token struct {
	Text         string
	Line         int
	OfLines      int
	Position     int
	FollowsComma int // -1 means no, 0+ is the ordinal of the comma it follows on the line
}

var bycommaspace = regexp.MustCompile(`([^,\s]+|[,\s]+)`)
var whitespace = regexp.MustCompile(`\s`)

func Tokenize(source string) []Token {
	tokens := make([]Token, 0)
	lines := strings.Split(source, "\n")
	fmt.Printf("lines: %s %d\n", lines, len(lines))
	for lnum, ln := range lines {
		texts := bycommaspace.FindAllString(ln, -1)
		follows := -1
		pcomma := -1
		pos := 0
		for _, txt := range texts {
			txt = whitespace.ReplaceAllString(txt, "")
			if len(txt) > 0 {
				if txt[0] == ',' {
					pcomma += len(txt) // there might be more than one comma
					follows = pcomma
				} else {
					tokens = append(tokens, Token{
						Text:         txt,
						Line:         lnum,
						OfLines:      len(lines),
						Position:     pos,
						FollowsComma: follows,
					})
					follows = -1
					pos += 1
				}
			}
		}
	}

	return tokens
}
