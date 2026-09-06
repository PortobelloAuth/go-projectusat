// Package streetsuffixes reads a street suffix out of an address.
//
// The rows are Publication 28 Appendix C1, held once in
// github.com/poetic-systems/addresstables and read from there. What is here is
// the part that is about parsing: which spelling a token is, the fuzzy match,
// and Claims with its confidence rules.
//
// Alt is the only key the maps below are built from, so an entry is findable
// by its own Primary or Short only when that spelling is repeated into Alt.
// The shared table states that invariant and enforces it, so the maps may be
// built from Alt alone.
package streetsuffixes

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hbollon/go-edlib"
	"github.com/poetic-systems/addresstables/streetsuffixes"
)

type StreetSuffix struct {
	Primary string
	Short   string
	Alt     []string
}

// streetSuffixes is Publication 28 Appendix C1. Alt is the only lookup key:
// the maps below are built from it alone, so an entry is findable by its own
// Primary or Short only when that spelling is repeated into Alt. That is an
// invariant of this table rather than a property of the type, and it is
// enforced by TestTablePrimaryAndShortAreLookupKeys instead of by the lookup,
// because indexing Primary directly would silently resolve a name two entries
// both claim rather than reporting it.

var streetSuffixPrimaryMap = maps.Collect(func(yield func(string, StreetSuffix) bool) {
	for s := range streetsuffixes.All() {
		v := StreetSuffix{Primary: s.Primary, Short: s.Short, Alt: s.Alt}
		for _, a := range v.Alt {
			if !yield(a, v) {
				return
			}
		}
	}
})
var streetSuffixShortMap = maps.Collect(func(yield func(string, string) bool) {
	for s := range streetsuffixes.All() {
		for _, a := range s.Alt {
			if !yield(a, s.Short) {
				return
			}
		}
	}
})
var streetSuffixKeys = slices.Collect(maps.Keys(streetSuffixPrimaryMap))

// punctuation matches everything a street suffix is not made of. Suffix keys in
// this table are letters and spaces, so a digit surviving the strip is what
// makes a lookup of 1ST fail instead of finding STREET.
var punctuation = regexp.MustCompile("[^a-zA-Z0-9 ]+")

func normalizeStreetSuffix(src string, primary bool, fuzzy bool) (string, error) {
	info, err := Info(src, fuzzy)
	if err != nil {
		return "", err
	}

	if primary {
		return info.Primary, nil
	}

	return info.Short, nil
}

func Info(src string, fuzzy bool) (*StreetSuffix, error) {
	// clean out any punctuation
	clean := punctuation.ReplaceAllString(src, "")
	// capitalize
	capitalized := strings.ToUpper(clean)
	// if requested, fuzzy match keys
	rkey := capitalized
	if fuzzy && len(capitalized) > 3 {
		matched, err := edlib.FuzzySearchThreshold(capitalized, streetSuffixKeys, 0.7, edlib.DamerauLevenshtein)
		if err != nil {
			// TODO: figure out how to let the user control logging in this library
			// log warn "Unable to fuzzy match supplied region string"
			matched = capitalized
		} else {
			rkey = matched
		}
	}

	// look up the primary
	short, ok := streetSuffixShortMap[rkey]
	if !ok {
		short = rkey
	}

	info, ok := streetSuffixPrimaryMap[short]
	if ok {
		return &StreetSuffix{
			Primary: info.Primary,
			Short:   info.Short,
			Alt: slices.Collect(func(yield func(string) bool) {
				for _, a := range info.Alt {
					if !yield(a) {
						return
					}
				}
			}),
		}, nil
	}

	return nil, fmt.Errorf("Unrecognized street suffix")
}

func NormalizeStreetSuffix(r string) (string, error) {
	return normalizeStreetSuffix(r, true, false)
}

func FuzzyNormalizeStreetSuffix(r string) (string, error) {
	return normalizeStreetSuffix(r, true, true)
}

func NormalizeStreetSuffixAbreviation(r string) (string, error) {
	return normalizeStreetSuffix(r, false, false)
}

func FuzzyNormalizeStreetSuffixAbreviation(r string) (string, error) {
	return normalizeStreetSuffix(r, false, true)
}
