package ruralroute

import (
	"slices"
	"strings"

	"github.com/poetic-systems/addresstables/puertorico"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// Claims returns every reading of tokens this vocabulary can support.
//
// These do not go into the parser's shared claim pool, and this package is not
// registered with it. A Puerto Rico address is one whose last line says so, and
// nothing above that line can tell — #71, and the reasoning on
// puertorico.Candidates, which is the only caller. RR 2 BOX 1980 is a perfectly
// good mainland address, and the Spanish vocabulary must not be offered one.
//
// A route is claimed as a whole pattern or not at all, exactly as on the
// mainland: the route is a numbered street that runs a long way and the boxes
// on it are primary addresses, so "RUTA RURAL 3 BUZON 12000" is one claim
// assigning a street name of RR 3 and a primary address number of BOX 12000. A
// bare BUZON, or an RD standing alone, says nothing a parser can use.
//
// What is deliberately absent is any reading of the sector name. p. 30 writes
// its examples as two lines —
//
//	RR 2 BOX 1980
//	SECTOR EL BRINCO
//
// — and requires that a developer "MUST eliminate this information". A claim
// never spans a line break (see token.LineEnd), so the sector line is not
// absorbed into the route claim the way a mainland rural route absorbs the
// trailing town name on its own line. It is simply left unread, and the
// candidate that carries this claim strands it. That is the standard's
// instruction taken literally: the tokens are not evidence for a competing
// reading, they are text to delete.
func Claims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	// The scan steps over a claim it has made rather than into it. This
	// vocabulary nests — RUTA RURAL contains RURAL, and RUTA ESTRELLA does not
	// contain HC only by luck — so a start one token inside a route reads the
	// same route from one word later. That is not a second reading of the
	// address, it is the same reading with a word stranded, and offering it
	// would put a candidate on the table that can only lose.
	for start := 0; start < len(tokens); {
		c, ok := routeClaim(tokens, start)
		if !ok {
			start++
			continue
		}

		claims = append(claims, c)
		start = c.End()
	}

	return claims
}

// maxSpan is the longest route the vocabulary accepts, in tokens: a two word
// designator (RUTA ESTRELLA), the route number, a box word, and the box
// number.
const maxSpan = 5

// boxWords are the spellings that open the box half of the pattern.
var boxWords = distinct(func(w puertorico.RouteWord) (string, bool) {
	return w.Spelling, w.Standard == "BOX"
})

// routeClaim reads the route pattern beginning at start, and nothing beyond
// it.
//
// Normalize accepts a span that continues past the pattern — p. 30's sector
// examples are exactly that — so a longer span normalizes just as happily as
// the pattern alone. The shortest span that normalizes is therefore the one
// that says how far the pattern itself runs.
func routeClaim(tokens []token.Token, start int) (claim.Claim, bool) {
	// The pattern is a delivery address line and cannot run past the end of
	// one. See token.LineEnd.
	limit := min(maxSpan, token.LineEnd(tokens, start)-start)

	for length := 1; length <= limit; length++ {
		span := tokens[start : start+length]

		normalized, err := Normalize(token.Join(span))
		if err != nil {
			continue
		}

		split, ok := boxIndex(span)
		if !ok {
			// The pattern matched inside a single glued token, so there is no
			// boundary in the token slice to divide the parts on.
			return claim.Claim{}, false
		}

		// Normalize emits exactly "DESIGNATOR ROUTE BOX BOXNUM".
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

// boxIndex reports where the box half of the pattern starts, as an offset into
// the span.
//
// The scan begins after the designator so that a designator which is also a
// box word could not be mistaken for one. None is today; the vocabulary is
// data and may grow.
func boxIndex(span []token.Token) (int, bool) {
	for i := designatorLength(span); i < len(span); i++ {
		if slices.Contains(boxWords, fold(span[i].Text)) {
			return i, true
		}
	}

	return 0, false
}

// designatorLength reports how many whole tokens the route designator
// occupies.
//
// routeWords is ordered longest spelling first, so the first that matches is
// the longest that could — the same property routeReplacer depends on.
//
// A number may be glued to the designator's last word rather than standing on
// its own: p. 30 writes RR03, and "RUTA RURAL03" would read the same way.
// Since a spelling and a lead are both words joined by single spaces, a lead
// that merely begins with a spelling matches it up to its last word and glues
// the rest, so the designator occupies n-1 whole tokens.
func designatorLength(span []token.Token) int {
	for _, w := range routeWords {
		n := len(strings.Fields(w.Spelling))
		if n > len(span) {
			continue
		}

		lead := fold(token.Join(span[:n]))
		if lead == w.Spelling {
			return n
		}
		if strings.HasPrefix(lead, w.Spelling) {
			return n - 1
		}
	}

	return 0
}

// fold renders a token the way the vocabulary is written: upper case, without
// the diacritics the standard's own "Buzón" carries. A token that cannot be
// folded is returned upper-cased, since it can then only fail to match.
func fold(text string) string {
	folded, err := diacritics.Substitute(text)
	if err != nil {
		return strings.ToUpper(text)
	}

	return strings.ToUpper(folded)
}
