package military

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// Three vocabularies live here, and each maps onto a different part of the
// address:
//
//   - APO, FPO, and DPO stand where a city stands. The standard is explicit
//     that city and country names MUST NOT appear in these addresses, so the
//     designation is not merely near the city field, it is what occupies it.
//   - AE, AP, and AA stand where a region stands. They are not states, but the
//     standard calls them the two-character state abbreviation and they are
//     normalized as one.
//   - CMR, PSC, UMR, and UNIT open the street line, and are claimed only as
//     part of it.
//
// The street line is one claim over the whole pattern. PSC 3 BOX 4120 is a
// facility and a box on it, so the facility and its number are the street name
// and the box and its number are the primary address number — the same shape a
// rural route has, and the reason both need a formatter of their own.
//
// A facility designator is not claimed on its own. PSC standing by itself is
// not weak evidence of a military address, it is a fragment of a pattern that
// did not match, and UNIT alone is far more likely to be the secondary unit
// designator that pkg/secondaryunit claims. BOX is likewise never claimed
// alone: here it opens the primary number, in a rural route it does the same,
// and in a PO box it belongs to the designator. The word means nothing without
// the pattern around it.
//
// Nothing here decides that an address is a military address. That is a
// judgment about the whole address rather than about a run of tokens, and it
// belongs with the AddressType selection that consumes claims rather than with
// the vocabulary that produces them.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i, t := range tokens {
		if part, ok := lastLinePart(strings.ToUpper(t.Text)); ok {
			claims = append(claims, claim.Claim{
				Confidence: claim.ConfidenceExact,
				Parts: []claim.ClaimPart{{
					Start:  i,
					Length: 1,
					Part:   part,
					Value:  strings.ToUpper(t.Text),
				}},
			})
		}

		if c, ok := streetLineClaim(tokens, i); ok {
			claims = append(claims, c)
		}
	}

	return claims
}

// streetLineSpan is the length of a military street line in tokens, which the
// standard fixes exactly: a facility designator, its assigned number, BOX, and
// the box number.
const streetLineSpan = 4

// streetLineClaim reads a military street line starting at start.
//
// NormalizeStreetLine is the recognizer rather than a formatter — unusually
// for a Normalize function in this library, it returns an error for anything
// that is not this pattern — so the rule for what counts lives in one place
// and this function only has to say which tokens got which part.
func streetLineClaim(tokens []token.Token, start int) (claim.Claim, bool) {
	if start+streetLineSpan > len(tokens) {
		return claim.Claim{}, false
	}

	normalized, err := NormalizeStreetLine(token.Join(tokens[start : start+streetLineSpan]))
	if err != nil {
		return claim.Claim{}, false
	}

	// NormalizeStreetLine emits exactly "TYPE ASSIGNED BOX BOXNUM".
	fields := strings.Fields(normalized)

	return claim.Claim{
		Confidence: claim.ConfidenceExact,
		Parts: []claim.ClaimPart{
			{
				Start:  start,
				Length: 2,
				Part:   claim.PartStreetName,
				Value:  fields[0] + " " + fields[1],
			},
			{
				Start:  start + 2,
				Length: 2,
				Part:   claim.PartPrimaryNumber,
				Value:  fields[2] + " " + fields[3],
			},
		},
	}, true
}

// lastLinePart reports which part of the last line a military token occupies.
//
// Both vocabularies are rated ConfidenceExact by Claims. They are fixed codes
// that mean nothing else in this vocabulary, and unlike a spelled-out region
// or suffix they are not ordinary words that turn up in place names.
func lastLinePart(text string) (claim.Part, bool) {
	if validCities[text] {
		return claim.PartCity, true
	}

	if validRegions[text] {
		return claim.PartRegion, true
	}

	return "", false
}
