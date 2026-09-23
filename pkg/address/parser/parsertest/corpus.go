// Package parsertest is a corpus of address test cases, plus a runner that
// scores any parser against them. It exists because "36 of 46 pass" is a
// useless report on its own — a parser has to say which 36 and which 10
// before a reviewer can tell a regression from a known gap. See
// go-projectusat#109 and libpostalhttp-parser#3.
//
// The corpus is data, not assertions. Aaron's call on libpostalhttp-parser#3
// was to build this here so addressparsers and libpostalhttp-parser can both
// score themselves against it without depending on each other, and without
// this package depending on either of them.
package parsertest

import "github.com/PortobelloAuth/go-projectusat/pkg/address"

// Case is one address test: an input string, what goprojectusat.Normalize
// should return for it, and where it came from.
//
// Source is a spec page ("p.16"), a go-projectusat issue ("go-projectusat#61"),
// "go-projectusat" for the parity suite the project has carried since before
// this package existed, or "ours" for cases this package originated. It is
// what lets a report say which cases failed instead of only how many.
//
// An empty Want means the right answer is still an open question. The case is
// carried so a parser reports what it does with it, and it is scored as
// neither a pass nor a failure until someone settles it. Corpus entries are
// ground truth, and inventing one is worse than admitting we do not have it.
type Case struct {
	Source string
	Note   string
	Input  string
	Want   string

	// WantFields, when set, is the decomposition a parser's Parse must
	// return for Input, checked field by field — including the fields it
	// leaves empty. A nil WantFields asserts nothing about the
	// decomposition, the same way an empty Want asserts nothing about the
	// rendering above: unsettled, not failing, and scored that way by Run.
	//
	// Want and WantFields are independent, so a case can pin the rendering,
	// the decomposition, both, or neither. They have to be independent
	// because rendering the same string is not evidence two parses agree:
	// "123 NORTH PARK" renders identically whether NORTH is read as a
	// predirectional before the street name PARK, or NORTH PARK is read as
	// one unsplit name, and Want cannot tell those apart. WantFields can,
	// which is the entire reason it exists (go-projectusat#123,
	// addressparsers#24).
	WantFields *address.Address
}

// Cases is the full corpus: the specification's own examples, the parity
// cases go-projectusat used to run against every custom parser before this
// package existed, and the hard cases the address stack has since found.
var Cases = concat(SpecCases, HistoricalCases, OursCases)

func concat(sets ...[]Case) []Case {
	n := 0
	for _, s := range sets {
		n += len(s)
	}
	all := make([]Case, 0, n)
	for _, s := range sets {
		all = append(all, s...)
	}
	return all
}
