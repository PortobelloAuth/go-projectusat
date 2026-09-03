package generaldelivery

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// The whole of a general delivery street line is one closed phrase, so there is
// one claim per occurrence of it and nothing is claimed in pieces. GENERAL
// alone is a word that appears in real street names — a street named for a
// general — and DELIVERY alone says nothing at all; it is the pair that carries
// the meaning, which is exactly the reason the standard spells it out.
//
// The phrase is claimed wherever it appears, not only at the start of a line.
// Its extent is fixed by the vocabulary rather than by the line, so there is no
// ambiguity for a line boundary to resolve, and a claim that covers less than
// the line it sits on leaves the rest as leftover tokens, which is how a
// candidate says it did not account for everything. That is the reading the
// parser should see and weigh, not one this package should suppress.
//
// Nothing here decides that an address is a general delivery address. That is a
// judgment about the whole address and belongs with AddressType selection.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		// The phrase is a delivery address line and cannot run past the end of
		// one. See token.LineEnd.
		limit := min(maxSpan, token.LineEnd(tokens, start)-start)

		for length := limit; length >= 1; length-- {
			text := token.Join(tokens[start : start+length])
			if _, err := Normalize(text); err != nil {
				continue
			}

			claims = append(claims, claim.Claim{
				Confidence: spellingConfidence(text),
				Parts: []claim.ClaimPart{{
					Start:  start,
					Length: length,
					Part:   claim.PartStreetName,
					Value:  standardForm,
				}},
			})
		}
	}

	return claims
}

// spellingConfidence rates a matched span by which spelling it used.
//
// The spelled-out form is the one the standard requires, so a line that carries
// it is not merely recognizable as general delivery, it is conformant, and
// there is no other thing those two words together could be.
//
// The abbreviations are rated a step below, and the reason is not that the
// lookup is less certain — it is exact either way — but that the words are less
// distinctive. DEL is an ordinary word inside Spanish street names, so
// GEN DEL has a shape that a longer name could produce by accident in a way
// GENERAL DELIVERY does not. Rating the two the same would leave the parser
// nothing to prefer with.
func spellingConfidence(text string) claim.Confidence {
	if text == standardForm {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}
