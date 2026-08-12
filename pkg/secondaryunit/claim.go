package secondaryunit

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// Only the designator is claimed, never the number that follows it. In APT 4B
// the 4B is a secondary number solely because APT precedes it — on its own it
// is evidence of nothing, and could as easily be a primary number. That
// relationship is positional, and position is the parser's to read.
//
// The consequence is worth stating plainly: Claims is not sufficient to fill
// Address.SecondaryNumber. It identifies the designator; a parser that wants to
// know whether a number is expected after one reads SecondaryUnit.Numbered via
// Info, and pairs them itself.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i, t := range tokens {
		info, err := Info(t.Text)
		if err != nil {
			continue
		}

		claims = append(claims, claim.Claim{
			Start:      i,
			Length:     1,
			Part:       claim.PartSecondaryDesignator,
			Confidence: designatorConfidence(t.Text, info),
			Value:      info.Short,
		})
	}

	return claims
}

// designatorConfidence rates a matched token. The standard abbreviation is a
// code; the spelled out word is ordinary English that appears in street and
// business names, where FRONT, REAR, SIDE, KEY, and LOWER all turn up.
func designatorConfidence(text string, info *SecondaryUnit) claim.Confidence {
	if strings.ToUpper(text) == info.Short {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}
