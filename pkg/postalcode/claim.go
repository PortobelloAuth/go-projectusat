package postalcode

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// maxSpan is the longest postal code in tokens. Both a ZIP+4 written with a
// space and a Canadian postal code written the way Canada Post writes it are
// two tokens.
const maxSpan = 2

// Claims returns every reading of tokens this package can support.
//
// This package has no table. Every claim it makes is a shape match against the
// patterns the standard describes, which is why none of them are rated above
// ConfidenceStrong, and why a bare five digit number is rated ConfidenceWeak:
// 84088 is a well formed ZIP, and it is also a perfectly ordinary primary
// number. Nothing here can tell those apart.
//
// The TODO at the top of this package — loading known city and region
// combinations by ZIP from the USPS ZIP Locale data — is what would change
// that. A ZIP found in that table would be a claim backed by a vocabulary
// rather than by a shape, and could be rated accordingly.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		span := min(maxSpan, len(tokens)-start)
		for length := span; length >= 1; length-- {
			value, confidence, ok := classify(token.Join(tokens[start : start+length]))
			if !ok {
				continue
			}

			claims = append(claims, claim.Claim{
				Confidence: confidence,
				Parts: []claim.ClaimPart{{
					Start:  start,
					Length: length,
					Part:   claim.PartPostal,
					Value:  value,
				}},
			})
		}
	}

	return claims
}

// classify matches a candidate against the postal shapes this package knows,
// returning the normalized form and how strongly the shape identifies it.
func classify(candidate string) (string, claim.Confidence, bool) {
	// Keep the hyphen: it separates a ZIP from its add-on. Everything else is
	// punctuation, and spaces are removed so that a code split across tokens
	// matches the same patterns as one written solid.
	cleaned := textutil.CollapseSpace(textutil.StripPunctuation(
		textutil.Upper(candidate), textutil.StripOptions{KeepHyphen: true}))
	compact := strings.ReplaceAll(cleaned, " ", "")

	if m := usZIPCompact.FindStringSubmatch(compact); m != nil {
		if m[2] != "" {
			// A nine digit run is a shape no other component of a US address
			// takes.
			return m[1] + "-" + m[2], claim.ConfidenceStrong, true
		}

		return m[1], claim.ConfidenceWeak, true
	}

	// The alternating letter and digit pattern is likewise unique among
	// address components. The hyphen carries no meaning here, unlike ZIP+4.
	if caCompact := strings.ReplaceAll(compact, "-", ""); caPostalCompact.MatchString(caCompact) {
		return caCompact[:3] + " " + caCompact[3:], claim.ConfidenceStrong, true
	}

	return "", 0, false
}
