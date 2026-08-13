package ruralroute

import (
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// Claims returns every reading of tokens this package can support.
//
// A rural route is claimed as a whole pattern or not at all. The standard
// standardizes it as "RR ___ BOX ___", and both halves carry meaning: the
// route is a numbered street that runs a long way, and the boxes on it are
// primary addresses. So RR 4 BOX 125 is one claim assigning a street name of
// RR 4 and a primary address number of BOX 125.
//
// No part of the pattern is claimed alone. A bare RR says nothing a parser can
// use, RD would otherwise collide with the far more common ROAD, and BOX
// belongs to whichever pattern surrounds it — here, a PO box, or a military
// street line. Requiring the whole form is what keeps those apart without any
// package needing to know about the others.
//
// Nothing here decides that an address is a rural route address. That is a
// judgment about the whole address and belongs with AddressType selection.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		if c, ok := routeClaim(tokens, start); ok {
			claims = append(claims, c)
		}
	}

	return claims
}

// maxSpan is the longest rural route the vocabulary accepts, in tokens: a two
// word designator, an optional number marker, the route number, BOX, another
// optional marker, and the box number.
const maxSpan = 7

// boxMarker matches the token that opens the box half of the pattern. Normalize
// accepts a number sign or a spelled-out number word in place of BOX, so the
// same alternatives have to be recognized here to know where the split falls.
var boxMarker = regexp.MustCompile(`^(BOX|#|NUMBER|NUM|NO)`)

// routeClaim reads a rural route beginning at start.
//
// Normalize is the recognizer, not a formatter: it returns an error for
// anything that is not this pattern, so the rule for what counts stays in one
// place. It also drops whatever follows the pattern — the standard's own
// example turns "RR 2 BOX 18 Bryan Dairy Rd" into "RR 2 BOX 18" — which is why
// the shortest span that normalizes is the one taken. A longer span would
// normalize just as happily and claim tokens the pattern never covered.
func routeClaim(tokens []token.Token, start int) (claim.Claim, bool) {
	limit := min(maxSpan, len(tokens)-start)

	for length := 1; length <= limit; length++ {
		normalized, err := Normalize(token.Join(tokens[start : start+length]))
		if err != nil {
			continue
		}

		split, ok := boxTokenIndex(tokens[start : start+length])
		if !ok {
			// The pattern matched inside a single glued token, so there is no
			// boundary in the token slice to divide the parts on.
			return claim.Claim{}, false
		}

		// Normalize emits exactly "RR ROUTE BOX BOXNUM".
		fields := strings.Fields(normalized)

		return claim.Claim{
			Confidence: claim.ConfidenceExact,
			Parts: []claim.ClaimPart{
				{
					Start:  start,
					Length: split,
					Part:   claim.PartStreetName,
					Value:  fields[0] + " " + fields[1],
				},
				{
					Start:  start + split,
					Length: length - split,
					Part:   claim.PartPrimaryNumber,
					Value:  fields[2] + " " + fields[3],
				},
			},
		}, true
	}

	return claim.Claim{}, false
}

// boxTokenIndex reports where the box half of the pattern starts, as an offset
// into the span. The route designator occupies at least one token, so a marker
// at offset zero is not a boundary.
func boxTokenIndex(span []token.Token) (int, bool) {
	for i, t := range span {
		if i > 0 && boxMarker.MatchString(strings.ToUpper(t.Text)) {
			return i, true
		}
	}

	return 0, false
}
