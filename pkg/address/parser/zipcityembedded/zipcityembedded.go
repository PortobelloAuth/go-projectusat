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
	parser *zcparser.Parser
}

type onceParser struct {
	parser *zcparser.Parser
	err    error
}

var singleparser = sync.OnceValue(func() onceParser {
	zc, err := zcparser.New()
	if err != nil {
		return onceParser{
			parser: nil,
			err:    err,
		}
	}

	return onceParser{
		parser: zc,
		err:    nil,
	}
})

// New returns a Parser.
func New() (*Parser, error) {
	once := singleparser()
	if once.err != nil {
		return nil, once.err
	}

	return &Parser{
		parser: once.parser,
	}, nil
}

// Parse reads source into a structured address, returning ErrNoReading when no
// address type recognizes it.
//
// The stages are go-projectusat's, in the order that library defines: tokenize,
// let every vocabulary claim what it recognizes, read the last line from those
// claims, then ask each address type for its reading of the whole. Only the
// final choice among readings belongs to this package.
func (p *Parser) Parse(source string) (*address.Address, error) {
	return p.parser.Parse(source)
}
