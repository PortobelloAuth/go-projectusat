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

// Case is one address test: an input string, what goprojectusat.Normalize
// should return for it, and where it came from.
//
// Source is a spec page ("p.16"), a go-projectusat issue ("go-projectusat#61"),
// "go-projectusat" for the parity suite the project has carried since before
// this package existed, or "ours" for cases this package originated. It is
// what lets a report say which cases failed instead of only how many.
type Case struct {
	Source string
	Note   string
	Input  string
	Want   string
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
