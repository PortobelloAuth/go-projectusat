// Package region reads a state, possession, Canadian province or military
// "state" out of an address.
//
// The rows are held once in github.com/poetic-systems/addresstables and read
// from there: Publication 28 Appendix B for the states, possessions and the
// three military "states", and the Project US@ specification pp. 31-32 for the
// Canadian provinces and territories, which Appendix B does not carry.
//
// What stays here is the part that is about parsing: the punctuation strip,
// the edit-distance fallback, whether a name could also be read as a street
// name, and Claims with its confidence rules.
package region

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hbollon/go-edlib"
	"github.com/poetic-systems/addresstables/regions"
)

// RegionInfo is a shared row plus the one thing a parser needs that a table of
// abbreviations has no opinion about.
type RegionInfo struct {
	Primary            string
	Short              string
	Alt                []string
	PossibleStreetName bool
}

// notStreetNames names the regions whose spelling must never be offered as a
// street name as well as a region. Everything absent from this set may be
// both: PENNSYLVANIA is a state and an avenue in Washington.
//
// Keyed by Short, the one field of a row that is unique.
//
// A row earns a place here only by naming something no street is named after.
// The military "states" are addressing routes rather than places. FM is the
// Federated States of Micronesia and also Farm to Market, so the two-letter
// spelling reads as a road number far more often than as a region.
//
// Deliberately not shared with highways, though FM is exactly the collision
// that raises the question. Reading FM as Farm to Market rather than as the
// Federated States of Micronesia is a judgment made from the tokens around it,
// which highways has and this package does not.
var notStreetNames = map[string]bool{
	"AS": true, // American Samoa
	"FM": true, // Federated States of Micronesia; also Farm to Market
	"AE": true, // Armed Forces Europe, the Middle East and Canada
	"AP": true, // Armed Forces Pacific
	"AA": true, // Armed Forces Americas
}

func info(r regions.Region) RegionInfo {
	return RegionInfo{
		Primary:            r.Primary,
		Short:              r.Short,
		Alt:                r.Alt,
		PossibleStreetName: !notStreetNames[r.Short],
	}
}

// regionMap is keyed by Alt alone. The shared table states and enforces that
// every row repeats its own Primary and Short into Alt, so that reaches every
// row by every spelling it answers to.
var regionMap = maps.Collect(func(yield func(string, RegionInfo) bool) {
	for r := range regions.All() {
		i := info(r)
		for _, a := range i.Alt {
			if !yield(a, i) {
				return
			}
		}
	}
})

var regionKeys = slices.Collect(maps.Keys(regionMap))

// punctuation matches everything a region name is not made of. Region names in
// this table are letters and spaces, so a digit surviving the strip is what
// makes a lookup of 2ND fail instead of finding North Dakota.
var punctuation = regexp.MustCompile("[^a-zA-Z0-9 ]+")

func normalizeRegion(r string, fuzzy bool) (string, error) {
	info, err := Info(r, fuzzy)
	if err != nil {
		return "", err
	}

	return info.Short, nil
}

func Info(r string, fuzzy bool) (*RegionInfo, error) {
	// clean out any punctuation
	clean := punctuation.ReplaceAllString(r, "")
	// capitalize
	capitalized := strings.ToUpper(clean)
	// if requested, fuzzy match keys
	rkey := capitalized
	if fuzzy && len(capitalized) > 3 {
		matched, err := edlib.FuzzySearchThreshold(capitalized, regionKeys, 0.7, edlib.DamerauLevenshtein)
		if err != nil {
			// TODO: figure out how to let the user control logging in this library
			// log warn "Unable to fuzzy match supplied region string"
			matched = capitalized
		} else {
			rkey = matched
		}
	}

	// look up the abbreviation
	info, ok := regionMap[rkey]
	if !ok {
		return nil, fmt.Errorf("Unrecognized state, possession, Canadian provice, or US Armed Forces region")
	}
	return &RegionInfo{
		Primary:            info.Primary,
		Short:              info.Short,
		Alt:                slices.Clone(info.Alt),
		PossibleStreetName: info.PossibleStreetName,
	}, nil
}

func NormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, false)
}

func FuzzyNormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, true)
}
