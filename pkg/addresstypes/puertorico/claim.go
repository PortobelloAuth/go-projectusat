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
// The designator must open the line, with one exception: a primary address
// number immediately before it, and nothing before that. p.28's own example
// puts one there — "A17 URB JARDINES FAGOTA" -> "A17 JARD FAGOTA" — a
// standalone urbanization name acting as the street name, with the primary
// number that always leads a Puerto Rico street line still in front of it.
// Anything else preceding the designator is the shape this rule otherwise
// guards against: claiming from there to the line end would swallow
// "ACME CORP" or similar into an address component that cannot contain it,
// and a number is not exempt from that risk unless it is the only thing
// ahead — see precededOnlyByPrimaryNumber.
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
	if start > 0 && tokens[start-1].Line == tokens[start].Line && !precededOnlyByPrimaryNumber(tokens, start) {
		return claim.Claim{}, false
	}

	end := token.LineEnd(tokens, start)
	if end-start < 2 || end == len(tokens) {
		return claim.Claim{}, false
	}

	designator, desErr := NormalizeUrbanization(tokens[start].Text)

	// p.28-29's Exceptions table: a standalone urbanization name "stand[s]
	// alone and MUST NOT require the use of the abbreviation URB" — the rule
	// is a plain word substitution on whichever word opens the name, whether
	// or not URB precedes it. So the name's first word is checked against
	// that table before falling back to the ordinary URB-plus-free-text
	// reading below; when it matches, any URB ahead of it is dropped and the
	// word is replaced by its abbreviation, per the standard's own examples:
	// "URB EXT VISTA BELLA" -> "EXT VISTA BELLA", "URB ALTS DE CANA" ->
	// "ALTS DE CANA".
	nameStart := start
	if desErr == nil {
		nameStart = start + 1
	}
	if short, err := NormalizeStandaloneUrbanization(tokens[nameStart].Text); err == nil {
		rest := strings.ToUpper(token.Join(tokens[nameStart+1 : end]))
		value := short
		if rest != "" {
			value += " " + rest
		}

		return claim.Claim{
			Confidence: claim.ConfidenceExact,
			Parts: []claim.ClaimPart{
				{
					Start:  start,
					Length: end - start,
					Part:   claim.PartArea,
					Value:  value,
				},
			},
		}, true
	}

	if desErr != nil {
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

// precededOnlyByPrimaryNumber reports whether the token immediately before
// start is a primary address number (see normalizePrimaryNumber) that itself
// opens the line — the one thing #132 case 1 needs to admit ahead of the
// designator on the same line.
//
// Requiring the number to open the line, rather than merely to sit
// immediately before the designator, is what keeps "ACME CORP 17 URB
// HIGHLAND GDNS" from being read the same way "A17 URB JARDINES FAGOTA" is:
// a number is only unambiguous evidence of p.28's pattern when there is
// nothing ahead of it that the claim would otherwise have to explain away.
func precededOnlyByPrimaryNumber(tokens []token.Token, start int) bool {
	prev := start - 1
	if prev < 0 || tokens[prev].Line != tokens[start].Line {
		return false
	}
	if prev > 0 && tokens[prev-1].Line == tokens[prev].Line {
		return false
	}

	_, ok := normalizePrimaryNumber(strings.ToUpper(tokens[prev].Text))
	return ok
}
