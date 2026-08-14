package streetsuffixes

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// Street suffixes are single tokens, so every claim has Length 1. The value is
// the standard abbreviation, which is the form Project US@ requires.
//
// Lookups are exact, not fuzzy. Info accepts a fuzzy flag and the Normalize
// entry points expose it, but a fuzzy match is a guess that some other word was
// meant, and this package has no basis for that guess — CRESENT is not evidence
// of CRESCENT unless something knows the input is misspelled. Spelling
// correction is a separate kind of evidence and belongs to whatever performs
// it, not to the vocabulary.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i, t := range tokens {
		info, err := Info(t.Text, false)
		if err != nil {
			continue
		}

		claims = append(claims, claim.Claim{
			Confidence: suffixConfidence(t.Text, info),
			Parts: []claim.ClaimPart{{
				Start:  i,
				Length: 1,
				Part:   claim.PartStreetSuffix,
				Value:  info.Short,
			}},
		})
	}

	return claims
}

// suffixConfidence rates a matched token. The standard abbreviation is a code
// that means nothing else in this vocabulary. Anything else — the spelled out
// name, or one of the alternate spellings — is a word that frequently appears
// in street names in its own right: PARK, COMMON, KEY, and PLAZA are all real
// suffixes and all real street names.
func suffixConfidence(text string, info *StreetSuffix) claim.Confidence {
	if strings.ToUpper(punctuation.ReplaceAllString(text, "")) == info.Short {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}
