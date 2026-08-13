package pobox

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// A post office box is claimed as a whole pattern or not at all. The
// designator names the street line and the box number is the primary address
// on it, so PO BOX 11890 is one claim assigning a street name of PO BOX and a
// primary address number of 11890. Every recognized spelling claims the same
// street name, since the standard requires developers to rewrite the synonyms
// to PO BOX.
//
// A designator with no number after it is not claimed. That is not this
// package hedging: a post office box without a box number is not an address,
// so a lone DRAWER is a fragment of a pattern that did not match rather than
// weak evidence of one.
//
// Requiring the whole pattern is also what keeps BOX out of the argument. The
// token appears here, in a rural route, and in a military street line, and in
// none of those does it mean anything without the form around it.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		if c, ok := boxClaim(tokens, start); ok {
			claims = append(claims, c)
		}
	}

	return claims
}

// maxSpan is the longest post office box the vocabulary accepts, in tokens:
// the three word designator POST OFFICE BOX and the box number.
const maxSpan = 4

// boxClaim reads a post office box beginning at start.
//
// Normalize is the recognizer, not a formatter — it returns an error for
// anything that is not this pattern — so the rule for what counts stays in one
// place. It anchors at both ends, so unlike a rural route there is no trailing
// text to guard against: the span either is the whole pattern or is rejected.
func boxClaim(tokens []token.Token, start int) (claim.Claim, bool) {
	limit := min(maxSpan, len(tokens)-start)

	// The box number is the final token, so a designator alone cannot match.
	for length := 2; length <= limit; length++ {
		normalized, err := Normalize(token.Join(tokens[start : start+length]))
		if err != nil {
			continue
		}

		// Normalize emits exactly "PO BOX BOXNUM".
		fields := strings.Fields(normalized)
		designator := length - 1

		return claim.Claim{
			Confidence: designatorConfidence(token.Join(tokens[start : start+designator])),
			Parts: []claim.ClaimPart{
				{
					Start:  start,
					Length: designator,
					Part:   claim.PartStreetName,
					Value:  fields[0] + " " + fields[1],
				},
				{
					Start:  start + designator,
					Length: 1,
					Part:   claim.PartPrimaryNumber,
					Value:  fields[2],
				},
			},
		}, true
	}

	return claim.Claim{}, false
}

// designatorConfidence rates the spelling that opened the pattern.
//
// The forms that name a post office box outright are reserved for it, and with
// a box number after them nothing else reads that way. The synonyms the
// standard tells developers to rewrite are ordinary English words: DRAWER 214
// is a good match for this pattern and still not the only thing those tokens
// could be.
func designatorConfidence(designator string) claim.Confidence {
	if reservedDesignators[strings.ToUpper(designator)] {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}

// reservedDesignators are the spellings that name a post office box and mean
// nothing else. Listed rather than derived: LOCKBOX is built from the same
// words and is not one of them, so the distinction is which forms the standard
// reserved, not how they are spelled.
var reservedDesignators = map[string]bool{
	"POST OFFICE BOX": true,
	"PO BOX":          true,
	"POB":             true,
}
