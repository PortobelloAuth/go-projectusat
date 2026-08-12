package highways

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// maxSpan bounds how many tokens a highway name can cover. The longest forms
// the standard gives are four words — CALIFORNIA COUNTY ROAD 555, HIGHWAY 66
// FRONTAGE ROAD, HIGHWAY 3 BYPASS RD — and one more is allowed for a state
// prefix in front of them.
const maxSpan = 5

// Claims returns every reading of tokens this package can support.
//
// Highway names are claimed as street names, because that is what they are:
// the standard's rule is that county, state, and local highways are used as
// street names and so are not abbreviated.
//
// Note that NormalizeStreetName is not the recognizer here. It returns an
// error only for empty input and passes anything else through uppercased, so
// it cannot distinguish a highway from an ordinary street name — asking it
// about MAIN would yield MAIN. The evidence is whether a highway rule actually
// matched, which is what normalizeTokens reports.
//
// Shorter spans inside a longer match are claimed too. CA COUNTY ROAD 150
// contains COUNTY ROAD 150, and both are real highway names; the parser
// decides whether the state prefix belongs to the name.
//
// Every claim is rated the same. This vocabulary has no fixed codes that can
// mean nothing else — HIGHWAY, COUNTY, and ROUTE are ordinary words, and a
// match is a match of structure around them rather than a table lookup — so
// there is no basis here for ranking one highway form above another. Whether a
// state prefix or a longer span is the better reading is a question about the
// surrounding tokens, and belongs to the parser.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		span := min(maxSpan, len(tokens)-start)
		for length := span; length >= 1; length-- {
			text := token.Join(tokens[start : start+length])

			normalized, ok := normalizeTokens(strings.Fields(expandGluedInterstate(strings.ToUpper(text))))
			if !ok {
				continue
			}

			claims = append(claims, claim.Claim{
				Start:      start,
				Length:     length,
				Part:       claim.PartStreetName,
				Confidence: claim.ConfidenceStrong,
				Value:      normalized,
			})
		}
	}

	return claims
}
