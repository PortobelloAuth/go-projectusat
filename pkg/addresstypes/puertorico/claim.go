package puertorico

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// Today that is the urbanization, and only the urbanization. The standard
// gives the Puerto Rico street address line three components stacked on three
// lines:
//
//	Urbanization Name
//	Secondary Address Identifier and Number
//	Primary Address Number and Street Name
//
//	URB HIGHLAND GDNS
//	COND LAS AMAPOLAS APT 103
//	123 CALLE MAIN
//
// The lower two are read by the vocabularies that already own them. The
// urbanization is the component nothing else can read, because it is the one
// the standard gives no closed vocabulary for: URB opens it, and what follows
// is a development name of whatever length its developer chose.
//
// Nothing here decides that an address is a Puerto Rico address. That is a
// judgment about the whole address — the region or the postal code settles it,
// see UsePRDialect — and it belongs with AddressType selection.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		if c, ok := urbanizationClaim(tokens, start); ok {
			claims = append(claims, c)
		}
	}

	return claims
}

// urbanizationClaim reads an urbanization line beginning at start.
//
// The name has no closed vocabulary and no length the standard states, so the
// only thing that says where it ends is the line it sits on. That is not a
// weakness of this reading, it is the whole of it: the standard puts the
// urbanization on a line by itself, so the line boundary is the evidence.
//
// Three things have to hold, and each rules out a reading that would otherwise
// absorb tokens belonging to something else.
//
// The designator must open the line. A URB appearing partway along one is not
// the shape the standard describes, and claiming from there to the line end
// would swallow whatever preceded it into an address component that cannot
// contain it.
//
// A name must follow. A designator alone is a fragment of a pattern that did
// not match, the same way a lone DRAWER is not a post office box.
//
// The line must not be the last one. An urbanization sits above the street
// line, so at least the last line has to follow it. Without this a
// single-line "URB LAS GLADIOLAS 150 CALLE A SAN JUAN PR 00926" reads as one
// enormous urbanization — the reading is offered no confidence rather than a
// weak one, because there is nothing in the tokens to tell where such a name
// would stop.
func urbanizationClaim(tokens []token.Token, start int) (claim.Claim, bool) {
	if start > 0 && tokens[start-1].Line == tokens[start].Line {
		return claim.Claim{}, false
	}

	end := token.LineEnd(tokens, start)
	if end-start < 2 || end == len(tokens) {
		return claim.Claim{}, false
	}

	designator, err := NormalizeUrbanization(tokens[start].Text)
	if err != nil {
		return claim.Claim{}, false
	}

	name := strings.ToUpper(token.Join(tokens[start+1 : end]))

	return claim.Claim{
		Confidence: claim.ConfidenceExact,
		Parts: []claim.ClaimPart{
			{
				Start:  start,
				Length: end - start,
				Part:   claim.PartArea,
				Value:  designator + " " + name,
			},
		},
	}, true
}
