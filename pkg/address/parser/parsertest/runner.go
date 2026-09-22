package parsertest

import (
	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
)

// Result is one Case run through a parser.
type Result struct {
	Case Case
	Got  string
	Err  error
}

// Pass reports whether the case's expectation was met.
func (r Result) Pass() bool {
	return r.Err == nil && r.Got == r.Case.Want
}

// Run scores p against cases by running each Input through
// goprojectusat.Normalize, with content normalization, and comparing the
// result to Want. It returns one Result per case, in the order given, so a
// caller can report by Source rather than only a total.
func Run(p parser.ParsingFunc, cases []Case) []Result {
	opts := []goprojectusat.USAtNormalizeOption{
		goprojectusat.WithCustomAddressParser(p),
		goprojectusat.WithContentNormalization(),
	}

	results := make([]Result, len(cases))
	for i, c := range cases {
		got, err := goprojectusat.Normalize(c.Input, opts...)
		results[i] = Result{Case: c, Got: got, Err: err}
	}
	return results
}

// CountPass returns how many of results passed, out of the total.
func CountPass(results []Result) (pass, total int) {
	for _, r := range results {
		if r.Pass() {
			pass++
		}
	}
	return pass, len(results)
}
