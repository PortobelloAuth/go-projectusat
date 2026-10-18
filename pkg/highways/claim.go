package highways

import (
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
)

// maxSpan bounds how many tokens a highway name can cover: four for the
// longest form the standard gives — HIGHWAY 66 FRONTAGE ROAD, HIGHWAY 3
// BYPASS RD — plus four for the longest region name that can prefix one,
// FEDERATED STATES OF MICRONESIA.
//
// It is a constant rather than a derivation because the region half of it is
// not reachable: pkg/region exports no way to enumerate its names or measure
// its longest. Adding one is an API change and wants its own issue, so the
// number is written down here with what it is made of, and this comment is the
// only thing keeping the two in step.
const maxSpan = 8

// Claims returns every reading of tokens this package can support.
//
// Highway names are claimed as street names, because that is what they are:
// the standard's rule is that county, state, and local highways are used as
// street names and so are not abbreviated.
//
// NormalizeStreetName is the recognizer. It returns an error for a name that
// matches no highway rule, so an error is the evidence that a span is not a
// highway, and a value is the evidence that it is.
//
// Shorter spans inside a longer match are claimed too. CA COUNTY ROAD 150
// contains COUNTY ROAD 150, and both are real highway names; the parser
// decides whether the state prefix belongs to the name.
//
// A span that ends partway through a compound directional is not claimed. A
// letter is a route designator — COUNTY ROAD N is a real road — but where that
// letter and the token after it are one compound directional, the standard
// reads them as the directional: COUNTY ROAD N EAST is COUNTY ROAD with the
// postdirectional NE (p.17). Taking N as the route would leave EAST to be
// either a postdirectional on its own or a word swallowed into the highway
// name, and both are readings the standard does not give. Whether two tokens
// are a compound is directionals' question, so it is asked there.
//
// Every claim is rated the same. This vocabulary has no fixed codes that can
// mean nothing else — HIGHWAY, COUNTY, and ROUTE are ordinary words, and a
// match is a match of structure around them rather than a table lookup — so
// there is no basis here for ranking one highway form above another. Whether a
// state prefix or a longer span is the better reading is a question about the
// surrounding tokens, and belongs to the parser.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		// Every span from start begins with the first word of its first
		// token, so when that word cannot lead a highway none of them is one.
		// A token that is all space has no first word; its spans fall
		// through to NormalizeStreetName.
		if lead, ok := leadingWord(tokens[start].Text); ok && !mayLeadHighway(lead) {
			continue
		}

		// A phrase hard wrapped across two lines is not one phrase. See
		// token.LineEnd.
		span := min(maxSpan, token.LineEnd(tokens, start)-start)
		for length := span; length >= 1; length-- {
			text := token.Join(tokens[start : start+length])

			normalized, err := NormalizeStreetName(text)
			if err != nil {
				continue
			}
			normalized = inRegion(normalized, tokens[start+length:])

			// Both checks are filters with no side effects, so their order
			// does not change which spans are claimed. The recognizer goes
			// first because it rejects almost every span, and the compound
			// check asks directionals about every token pair in the span.
			if splitsCompoundDirectional(tokens, start, start+length) {
				continue
			}

			claims = append(claims, claim.Claim{
				Confidence: claim.ConfidenceStrong,
				Parts: []claim.ClaimPart{{
					Start:  start,
					Length: length,
					Part:   claim.PartStreetName,
					Value:  normalized,
				}},
			})
		}
	}

	return claims
}

// splitsCompoundDirectional reports whether the span from start to end takes
// a token that, with the token after it on the same line, directionals claims
// as one compound. That covers a pair inside the span, the EAST swallowed in
// COUNTY ROAD N EAST, and a pair the span's end cuts in half, the N of COUNTY
// ROAD N with EAST still to come.
func splitsCompoundDirectional(tokens []token.Token, start, end int) bool {
	for i := start; i < end; i++ {
		if token.LineEnd(tokens, i)-i < 2 {
			continue
		}

		for _, c := range directionals.Claims(tokens[i : i+2]) {
			if c.Length() == 2 {
				return true
			}
		}
	}

	return false
}

// inRegion reads a highway name in the region its address is in, which is the
// last region named after it. PR leads Puerto Rico's highways everywhere but
// Texas, where it is the state's Park Road: PR 462 in LEONARD TX is PARK ROAD
// 462, and a Texas address cannot be on a Puerto Rico highway (#238).
func inRegion(name string, after []token.Token) string {
	if route, ok := strings.CutPrefix(name, "PR HIGHWAY "); ok && lastRegion(after) == "TX" {
		return "PARK ROAD " + route
	}
	return name
}

// lastRegion returns the code of the last region named in tokens, or "" when
// none is.
func lastRegion(tokens []token.Token) string {
	for _, t := range slices.Backward(tokens) {
		if r, err := region.NormalizeRegion(t.Text); err == nil {
			return r
		}
	}
	return ""
}
