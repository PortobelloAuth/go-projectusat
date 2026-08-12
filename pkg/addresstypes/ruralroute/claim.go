package ruralroute

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
// ("RURAL ROUTE", "RFD ROUTE").
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
// occupies: the standard puts a rural route address on the street address line
// as "RR ___ BOX ___", and Address documents the ordering as
// STREET PRIMARY SEC SECNUM. Every recognized spelling claims the value RR,
// which is the only form the standard permits in a patient record.
//
// Neither number is claimed. The route number and the box number are what they
// are because of the tokens in front of them, which is positional knowledge
// this package does not have — the same reasoning as pkg/secondaryunit.
//
// BOX is not claimed here either. It is a secondary unit designator, owned by
// pkg/secondaryunit, and claiming it here as well would give one word two
// sources of truth.
//
// Nothing here decides that an address is a rural route address. That is a
// judgment about the whole address and belongs with AddressType selection.
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
				Value:      "RR",
			})
		}
	}

	return claims
}

// designatorConfidence rates a matched designator.
//
// RR, RFD, and the spelled out forms exist only to mark a rural route. RD does
// not: it is also the standard abbreviation for ROAD, which pkg/streetsuffixes
// claims at ConfidenceExact. Both readings are correct and neither package can
// resolve them, but rating RD below the unambiguous designators says something
// true about this vocabulary — that RD is the one entry here whose spelling was
// borrowed rather than reserved.
func designatorConfidence(candidate string) claim.Confidence {
	if candidate == "RD" {
		return claim.ConfidenceLikely
	}

	return claim.ConfidenceExact
}
