package privatemailbox

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// The identifier and the number are one indivisible claim, for the reason
// secondaryunit gives: a lone PMB is not weak evidence of a private mailbox, it
// is a fragment of a pattern that did not match. The claim is a single Detail
// part over both tokens rather than two parts, because Detail is one field and
// the standard wants the identifier in it.
//
// The pair is a street line element and cannot run past the end of one. See
// token.LineEnd.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		limit := min(maxSpan, token.LineEnd(tokens, start)-start)

		for length := limit; length >= 1; length-- {
			value, err := Normalize(token.Join(tokens[start : start+length]))
			if err != nil {
				continue
			}

			claims = append(claims, claim.Claim{
				Confidence: identifierConfidence(tokens[start].Text),
				Parts: []claim.ClaimPart{{
					Start:  start,
					Length: length,
					Part:   claim.PartDetail,
					Value:  value,
				}},
			})
		}
	}

	return claims
}

// identifierConfidence rates a matched span by which identifier it used.
//
// PMB means one thing. # is also the secondary unit designator of unspecified
// type, which secondaryunit claims at Exact, so this reading is held below it:
// offered, and losing unless an address type has reason to take it. See the
// deviations note in privatemailbox.go.
func identifierConfidence(text string) claim.Confidence {
	if strings.HasPrefix(strings.ToUpper(text), identifier) {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceLikely
}
