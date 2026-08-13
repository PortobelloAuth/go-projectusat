package region

import (
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// maxSpan is the longest region name in the vocabulary, measured in tokens
// ("FEDERATED STATES OF MICRONESIA"). It bounds how far Claims looks ahead.
var maxSpan = slices.Max(slices.Collect(func(yield func(int) bool) {
	for name := range regionMap {
		if !yield(len(strings.Fields(name))) {
			return
		}
	}
}))

// Claims returns every reading of tokens this package can support.
//
// A two-letter postal abbreviation is exact: nothing else in the region
// vocabulary spells it that way. A full name or documented alias is strong —
// the region is certain, but the word may still be doing another job in the
// address.
//
// Where a region can also be a street name, Claims says so by returning a
// second, competing claim over the same tokens. That is the case Project US@
// calls out directly:
//
//	when the state name is the complete Primary Street Name, such as
//	OKLAHOMA AVE, then the state name SHOULD be spelled out completely
//
// so the street name reading carries the spelled-out Primary form while the
// region reading carries the Short abbreviation. Both are returned; choosing
// between them needs the neighbouring tokens, which is the parser's job, not
// this package's.
//
// The narrower case of a region appearing as one word inside a longer street
// name ("MONTANA TREASURE AVE" -> "MT TREASURE AVE") is deliberately not
// claimed here. That reading depends on what surrounds the token, and the
// abbreviation it produces differs from the one above.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		span := min(maxSpan, len(tokens)-start)
		for length := span; length >= 1; length-- {
			candidate := token.Join(tokens[start : start+length])

			info, err := Info(candidate, false)
			if err != nil {
				continue
			}

			claims = append(claims, claim.Claim{
				Confidence: regionConfidence(candidate, length),
				Parts: []claim.ClaimPart{{
					Start:  start,
					Length: length,
					Part:   claim.PartRegion,
					Value:  info.Short,
				}},
			})

			if info.PossibleStreetName {
				claims = append(claims, claim.Claim{
					Confidence: claim.ConfidenceLikely,
					Parts: []claim.ClaimPart{{
						Start:  start,
						Length: length,
						Part:   claim.PartStreetName,
						Value:  info.Primary,
					}},
				})
			}
		}
	}

	return claims
}

// regionConfidence rates a matched candidate. Only a single two-letter token is
// unambiguous within this vocabulary; everything else is a name that could be
// carrying ordinary meaning somewhere else in the address.
func regionConfidence(candidate string, length int) claim.Confidence {
	if length == 1 && len(candidate) == 2 {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}
