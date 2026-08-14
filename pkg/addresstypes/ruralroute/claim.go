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
// Where the pattern does not reach the end of its line, two readings are
// offered. See streetLineClaim for why, and for what the second one does with
// the tokens the standard says should not be there.
//
// Nothing here decides that an address is a rural route address. That is a
// judgment about the whole address and belongs with AddressType selection.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		c, ok := routeClaim(tokens, start)
		if !ok {
			continue
		}

		claims = append(claims, c)
		if extended, ok := streetLineClaim(tokens, c); ok {
			claims = append(claims, extended)
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

// routeClaim reads the rural route pattern beginning at start, and nothing
// beyond it.
//
// Normalize is the recognizer, not a formatter: it returns an error for
// anything that is not this pattern, so the rule for what counts stays in one
// place. It also drops whatever follows the pattern — the standard's own
// example turns "RR 2 BOX 18 Bryan Dairy Rd" into "RR 2 BOX 18" — so a longer
// span normalizes just as happily as the pattern alone. The shortest span that
// normalizes is therefore the one that says how far the pattern itself runs.
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

// streetLineClaim extends a route claim over the rest of its line, and reports
// false when the pattern already reached the end of it.
//
// The standard says a rural route line "SHOULD NOT allow additional
// designations, such as town or street names", so text after the pattern is
// not a street name the parser should keep — it is text that does not belong
// on the line. Normalize agrees: it discards the remainder. Claiming only the
// pattern would leave those tokens free for another vocabulary to read as a
// street, which is the reading the standard rules out.
//
// The claim absorbs them into the primary number part, whose value stays what
// Normalize produced. See claim.ClaimPart.Value: a part's tokens are what it
// covers, not what it says. That is already the ordinary case here, where
// "RFD ROUTE 4" is three tokens valued "RR 4".
//
// It is the weaker reading of the two, because where the street line ends is
// the uncertain part. A token slice carries no notion of a line ending, only
// of which line each token is on, so the extent is bounded by that and cannot
// run into a city or region that follows on its own line.
func streetLineClaim(tokens []token.Token, pattern claim.Claim) (claim.Claim, bool) {
	end := lineEnd(tokens, pattern.Start())
	if end <= pattern.End() {
		return claim.Claim{}, false
	}

	parts := make([]claim.ClaimPart, len(pattern.Parts))
	copy(parts, pattern.Parts)

	last := &parts[len(parts)-1]
	last.Length = end - last.Start

	return claim.Claim{Confidence: claim.ConfidenceLikely, Parts: parts}, true
}

// lineEnd reports the index one past the last token on start's line.
func lineEnd(tokens []token.Token, start int) int {
	end := start
	for end < len(tokens) && tokens[end].Line == tokens[start].Line {
		end++
	}

	return end
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
