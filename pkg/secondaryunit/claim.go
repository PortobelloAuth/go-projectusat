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
//
// What can be a unit number is not restricted. Unit numbers are not reliably
// numeric, and a landlord is free to name a unit anything at all, so refusing
// a reading because the number does not look like one would drop real
// addresses. The shape of the number sets the confidence instead: APT 4B is
// exact, KEY WEST is a reading worth offering and losing.
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

		if i+1 >= len(tokens) {
			continue
		}

		claims = append(claims, claim.Claim{
			Confidence: numberedConfidence(t.Text, info, tokens[i+1].Text),
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

// numberedConfidence rates a designator claimed together with its number.
//
// A number that looks like one leaves the rating to the designator. A number
// that does not drops the whole reading to contested: KEY WEST is a city far
// more often than it is unit WEST of a key, and the same is true of any
// designator followed by an ordinary word. The reading is still offered —
// someone will have named a unit KEY WEST — it just loses to anything else
// competing for those tokens.
func numberedConfidence(text string, info *SecondaryUnit, number string) claim.Confidence {
	if !looksLikeUnitNumber(number) {
		return claim.ConfidenceLikely
	}

	return designatorConfidence(text, info)
}

// looksLikeUnitNumber reports whether a token reads as a unit number: a digit
// anywhere, or a single letter, which covers 4B, 200, and UNIT A alike.
//
// This is not a test of whether the token can be a unit number. Anything can.
// It only separates the ordinary case from the one worth doubting.
func looksLikeUnitNumber(text string) bool {
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
