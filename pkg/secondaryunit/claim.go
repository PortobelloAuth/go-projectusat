package secondaryunit

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// A designator the standard marks as numbered is claimed together with its
// number, as one indivisible reading: APT 4B assigns both the designator and
// the secondary number, and a parser that took one without the other would
// have a reading this package never offered. Such a designator standing alone
// is not claimed at all. A lone APT is not weak evidence of a secondary unit,
// it is a fragment of a pattern that did not match.
//
// Designators the standard marks as unnumbered — BSMT, LBBY, REAR — are
// claimed on their own, because standing alone is the whole of what they look
// like.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i, t := range tokens {
		info, err := Info(t.Text)
		if err != nil {
			continue
		}

		designator := claim.ClaimPart{
			Start:  i,
			Length: 1,
			Part:   claim.PartSecondaryDesignator,
			Value:  info.Short,
		}

		if !info.Numbered {
			claims = append(claims, claim.Claim{
				Confidence: designatorConfidence(t.Text, info),
				Parts:      []claim.ClaimPart{designator},
			})

			continue
		}

		if i+1 >= len(tokens) || !isUnitNumber(tokens[i+1].Text) {
			continue
		}

		claims = append(claims, claim.Claim{
			Confidence: designatorConfidence(t.Text, info),
			Parts: []claim.ClaimPart{
				designator,
				{
					Start:  i + 1,
					Length: 1,
					Part:   claim.PartSecondaryNumber,
					Value:  strings.ToUpper(tokens[i+1].Text),
				},
			},
		})
	}

	return claims
}

// isUnitNumber reports whether a token can be the number of a secondary unit.
//
// Unit numbers are not reliably numeric: UNIT A and APT 4B are both ordinary,
// so the rule is a digit anywhere in the token, or a single alphanumeric
// character. Without it any word following a designator would qualify, and
// KEY WEST would read as unit WEST of a key.
func isUnitNumber(text string) bool {
	if strings.ContainsFunc(text, unicode.IsDigit) {
		return true
	}

	if utf8.RuneCountInString(text) != 1 {
		return false
	}

	r, _ := utf8.DecodeRuneInString(text)

	return unicode.IsLetter(r)
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
