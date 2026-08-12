package directionals

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// A directional is the clearest case of a token whose meaning is entirely
// positional. The same W is a predirectional in 3253 W 9200 S, a
// postdirectional in 123 MAIN ST W, and neither in 123 MAIN ST W PALM BEACH
// FL, where it opens the city name. This package can see the first two and has
// no way to see the third, so it claims both and ranks them equally: the tie
// is real, and only the parser knows which side of the street name a token
// fell on.
//
// Claims are made over single tokens only. Two-word spellings such as NORTH
// EAST are not claimed, because splitting a directional across tokens is
// indistinguishable here from a directional followed by another directional,
// and the standard writes these as one word.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i, t := range tokens {
		abbreviation, err := AbbreviateDirectional(t.Text)
		if err != nil {
			continue
		}

		confidence := directionalConfidence(t.Text)
		for _, part := range []claim.Part{claim.PartPredirectional, claim.PartPostdirectional} {
			claims = append(claims, claim.Claim{
				Start:      i,
				Length:     1,
				Part:       part,
				Confidence: confidence,
				Value:      abbreviation,
			})
		}
	}

	return claims
}

// directionalConfidence rates a matched token. An abbreviation is a fixed code
// that means nothing else; a spelled-out direction is an ordinary English word
// that turns up in street and city names — NORTH SALT LAKE, SOUTH BEND — so it
// is rated lower even though the lookup is just as certain.
func directionalConfidence(text string) claim.Confidence {
	if _, isFullWord := directionMap[strings.ToUpper(text)]; isFullWord {
		return claim.ConfidenceStrong
	}

	return claim.ConfidenceExact
}
