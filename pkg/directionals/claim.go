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
// A compound directional written as two tokens is claimed as one span. NORTH
// EAST is a reading of NE, and it competes with reading those tokens as two
// separate directionals — both are returned, and the parser decides. Which
// pairs combine is not a list kept here: two tokens are a compound when their
// abbreviations concatenate into one this vocabulary recognizes. That is why
// EAST WEST and NORTH SOUTH are not claimed, and it is the same reason the
// standard gives for excluding them — EW and NS are not directions.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i := range tokens {
		span := min(maxSpan, len(tokens)-i)
		for length := span; length >= 1; length-- {
			abbreviation, ok := abbreviateSpan(tokens[i : i+length])
			if !ok {
				continue
			}

			confidence := spanConfidence(tokens[i : i+length])
			for _, part := range []claim.Part{claim.PartPredirectional, claim.PartPostdirectional} {
				claims = append(claims, claim.Claim{
					Confidence: confidence,
					Parts: []claim.ClaimPart{{
						Start:  i,
						Length: length,
						Part:   part,
						Value:  abbreviation,
					}},
				})
			}
		}
	}

	return claims
}

// maxSpan is the longest directional in the vocabulary, measured in tokens: a
// compound spelled as two words.
const maxSpan = 2

// abbreviateSpan reduces a run of tokens to a single directional abbreviation,
// reporting whether the vocabulary recognizes it.
//
// A run of more than one token is recognized when the abbreviations of its
// tokens concatenate into an abbreviation this vocabulary knows. NORTH EAST
// gives NE and is a compound; EAST WEST gives EW and is not a direction at
// all, so it is not claimed. Deriving the rule this way means the invalid
// pairs never have to be enumerated, and a pair in the wrong order — EAST
// NORTH — falls out of it for free.
func abbreviateSpan(tokens []token.Token) (string, bool) {
	var combined strings.Builder
	for _, t := range tokens {
		abbreviation, err := AbbreviateDirectional(t.Text)
		if err != nil {
			return "", false
		}

		combined.WriteString(abbreviation)
	}

	if len(tokens) == 1 {
		return combined.String(), true
	}

	if _, ok := directionShortMap[combined.String()]; !ok {
		return "", false
	}

	return combined.String(), true
}

// spanConfidence rates a matched run of tokens.
//
// An abbreviation is a fixed code that means nothing else; a spelled-out
// direction is an ordinary English word that turns up in street and city names
// — NORTH SALT LAKE, SOUTH BEND — so it is rated lower even though the lookup
// is just as certain.
//
// A compound written as two tokens drops one further step, because it always
// competes with reading the same tokens as two separate directionals, and the
// standard spells compounds as one word. NORTHEAST is the expected form;
// NORTH EAST is a reading of it worth offering, not the one to prefer.
func spanConfidence(tokens []token.Token) claim.Confidence {
	spelledOut := false
	for _, t := range tokens {
		if _, isFullWord := directionMap[strings.ToUpper(t.Text)]; isFullWord {
			spelledOut = true
		}
	}

	switch {
	case len(tokens) > 1 && spelledOut:
		return claim.ConfidenceLikely
	case len(tokens) > 1:
		return claim.ConfidenceStrong
	case spelledOut:
		return claim.ConfidenceStrong
	}

	return claim.ConfidenceExact
}
