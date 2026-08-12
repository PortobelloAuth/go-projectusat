package pobox

import (
	"maps"
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// designatorSet is the recognized designators indexed for lookup, derived from
// the same slice the replacer is built from so the two cannot drift.
var designatorSet = maps.Collect(func(yield func(string, bool) bool) {
	for _, d := range recognizedDesignators {
		if !yield(d, true) {
			return
		}
	}
})

// maxSpan is the longest designator in the vocabulary, measured in tokens
// ("POST OFFICE BOX").
var maxSpan = slices.Max(slices.Collect(func(yield func(int) bool) {
	for _, d := range recognizedDesignators {
		if !yield(len(strings.Fields(d))) {
			return
		}
	}
}))

// Claims returns every reading of tokens this package can support.
//
// The designator is claimed as a street name, because that is the field it
// occupies: a post office box address is standardized as "PO BOX ___" on the
// street address line. Every recognized spelling claims the value PO BOX,
// since the standard requires developers to rewrite the synonyms to it.
//
// The box number is not claimed. It is a number that means something only
// because of the designator in front of it, which is positional knowledge this
// package does not have.
//
// Note that the two token spellings overlap a claim from another vocabulary:
// in PO BOX 11890 the BOX token is also a secondary unit designator, which
// pkg/secondaryunit claims on its own. Both readings are real — that input
// could be a PO box, or a street address with a box unit — and the parser has
// the surrounding tokens needed to choose. This package claims the longer span
// and says nothing about the shorter one.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		span := min(maxSpan, len(tokens)-start)
		for length := span; length >= 1; length-- {
			candidate := strings.ToUpper(token.Join(tokens[start : start+length]))
			if !designatorSet[candidate] {
				continue
			}

			claims = append(claims, claim.Claim{
				Start:      start,
				Length:     length,
				Part:       claim.PartStreetName,
				Confidence: designatorConfidence(candidate),
				Value:      "PO BOX",
			})
		}
	}

	return claims
}

// designatorConfidence rates a matched designator.
//
// The forms that name a post office box outright are reserved for it. The
// synonyms the standard tells developers to rewrite — CALLER, BIN, LOCKBOX,
// DRAWER — are ordinary English words that carry their own meaning elsewhere
// in an address, so a match on one of them is real but contested.
func designatorConfidence(candidate string) claim.Confidence {
	if reservedDesignators[candidate] {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}

// reservedDesignators are the spellings that name a post office box and mean
// nothing else. Listed rather than derived: LOCKBOX contains BOX and is not one
// of them, so there is no rule here to derive from.
var reservedDesignators = map[string]bool{
	"POST OFFICE BOX": true,
	"PO BOX":          true,
	"POB":             true,
}
