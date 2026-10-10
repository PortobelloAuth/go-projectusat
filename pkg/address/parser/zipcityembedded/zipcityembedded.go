package zipcityembedded

import (
	"sync"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	zcparser "github.com/PortobelloAuth/go-projectusat/pkg/address/parser/zipcityembedded/parse"
)

// zipcityembedded.Parser wraps zipcityembedded/parse.Parser in sync.OnceValue(...) so that
// if it is used as the default parser - as it currently is - zipcity will only be initialized
// on the first call to p.Parse(). This is likely overly conservative, given changes in
// zipcity itself which are necessary to make this meaningful.
type Parser struct {
	parser func() *zcparser.Parser
}

// New returns a Parser.
func New() *Parser {
	return &Parser{
		parser: sync.OnceValue(func() *zcparser.Parser {
			zc, err := zcparser.New()
			if err != nil {
				// OnceValue() lazy loading does not allow us to return or handle the error
				panic(err)
			}
			return zc
		}),
	}
}

// Parse reads source into a structured address, returning ErrNoReading when no
// address type recognizes it.
//
// The stages are go-projectusat's, in the order that library defines: tokenize,
// let every vocabulary claim what it recognizes, read the last line from those
// claims, then ask each address type for its reading of the whole. Only the
// final choice among readings belongs to this package.
func (p *Parser) Parse(source string) (*address.Address, error) {
	return p.parser().Parse(source)
}
