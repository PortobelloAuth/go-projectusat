package country

import (
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// maxSpan is the longest country name in the vocabulary, measured in tokens
// ("UNITED STATES"). It bounds how far Claims looks ahead.
var maxSpan = slices.Max(slices.Collect(func(yield func(int) bool) {
	for name := range countryNameMap {
		if !yield(len(strings.Fields(name))) {
			return
		}
	}
}))

// Claims returns every reading of tokens this package can support.
//
// Only names in countryNameMap are claimed. NormalizeCountry deliberately
// accepts any input and returns it capitalized, because Project US@ gives no
// guidance for validating international country names — but "this package will
// format anything you hand it" is not evidence that a token is a country, so
// it is not a basis for a claim.
//
// The domestic entries normalize to the empty string, since Project US@ omits
// the country on a US address. Those are still claims: the tokens are the
// country component, and the normalized form of that component is nothing.
//
// Note that a two letter abbreviation is exact only within this vocabulary. CA
// is unambiguously Canada here and unambiguously California in pkg/region, and
// both packages are right to claim it. Resolving that is the parser's job.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		span := min(maxSpan, len(tokens)-start)
		for length := span; length >= 1; length-- {
			candidate := token.Join(tokens[start : start+length])

			normalized, ok := countryNameMap[candidate]
			if !ok {
				continue
			}

			claims = append(claims, claim.Claim{
				Start:      start,
				Length:     length,
				Part:       claim.PartCountry,
				Confidence: countryConfidence(candidate, length),
				Value:      normalized,
			})
		}
	}

	return claims
}

// countryConfidence rates a matched candidate. An abbreviation is a fixed code
// that carries no other meaning in this vocabulary; a spelled-out name is a
// word that could be doing another job in the address, so it is rated lower
// even though the lookup is just as certain.
func countryConfidence(candidate string, length int) claim.Confidence {
	if length == 1 && len(candidate) <= 3 {
		return claim.ConfidenceExact
	}

	return claim.ConfidenceStrong
}
