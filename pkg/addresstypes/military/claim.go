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
//   - CMR, OMC, PSC, UMR, and UNIT open the street line, in the position an
//     ordinary address gives the street name.
//
// The assigned number and the box number that follow a facility designator are
// not claimed. Like a secondary number after APT, they are what they are
// because of the token in front of them, and that is positional knowledge this
// package does not have. See pkg/secondaryunit for the same reasoning.
//
// BOX is likewise not claimed, though it appears in every military street
// line. Rural route addresses use the same word for the same purpose, and two
// packages emitting identical claims for one word would be two sources of
// truth for it.
//
// Nothing here decides that an address is a military address. That is a
// judgment about the whole address rather than about a run of tokens, and it
// belongs with the AddressType selection that consumes claims rather than with
// the vocabulary that produces them.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i, t := range tokens {
		text := strings.ToUpper(t.Text)

		part, ok := partFor(text)
		if !ok {
			continue
		}

		claims = append(claims, claim.Claim{
			Start:      i,
			Length:     1,
			Part:       part,
			Confidence: claim.ConfidenceExact,
			Value:      text,
		})
	}

	return claims
}

// partFor reports which part of the address a military token occupies.
//
// Every one of these is rated ConfidenceExact by Claims. They are fixed codes
// that mean nothing else in this vocabulary, and unlike a spelled-out region
// or suffix they are not ordinary words that turn up in place names. Where one
// of them collides with another vocabulary — UNIT is also a secondary unit
// designator — both packages are right, and the parser resolves it.
func partFor(text string) (claim.Part, bool) {
	if validCities[text] {
		return claim.PartCity, true
	}

	if validRegions[text] {
		return claim.PartRegion, true
	}

	if _, ok := validAddressTypes[text]; ok {
		return claim.PartStreetName, true
	}

	return "", false
}
