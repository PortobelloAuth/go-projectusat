package puertorico

import (
	"slices"
	"strings"

	"github.com/poetic-systems/addresstables/puertorico"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
)

// poBoxWords are the Spanish spellings of the post office box designator, held
// in addresstables beside the route words.
var poBoxWords = slices.Collect(puertorico.POBoxWords())

// poBoxCandidates offers one reading for each Spanish post office box the
// vocabulary finds above the last line.
//
// The reading is the one pobox makes of its own designators: the street line
// is PO BOX and the box number, rendered by pobox's type. Only the designator
// is Spanish, and p. 29 requires the Spanish form to be rewritten, so nothing
// about the output is new. A reading is offered only when pobox's own
// recognizer accepts the rewritten line, so the rule for what counts as a box
// stays in one place.
//
// Like the route candidates, it sits behind the dialect gate in Candidates. A
// mainland address is never offered APARTADO as a box.
func poBoxCandidates(tokens []token.Token, line lastline.LineClaim) []*address.CandidateAddress {
	var candidates []*address.CandidateAddress

	// covered is the first token no accepted box has claimed. A start inside
	// an accepted box is the same box read again: BOX in PO BOX 2018, or PO BOX
	// in GPO BOX 1118. The leftmost designator claims its tokens first, so the
	// longer designator wins.
	covered := 0
	for start := range tokens {
		if start < covered {
			continue
		}

		c, ok := poBoxClaim(tokens, start)
		if !ok || c.End() > line.Span.Start {
			continue
		}

		covered = c.End()
		candidates = append(candidates,
			line.Candidate(&pobox.POBoxAddress{}, len(tokens), []claim.Claim{c}))
	}

	return candidates
}

// poBoxClaim reads a Spanish post office box beginning at start: the designator
// and the box number that follows it on the same line. A designator with no
// number after it is not a box, as on the mainland.
func poBoxClaim(tokens []token.Token, start int) (claim.Claim, bool) {
	end := token.LineEnd(tokens, start)

	for _, designator := range poBoxWords {
		n := len(strings.Fields(designator))
		if start+n >= end {
			continue
		}
		if foldUpper(token.Join(tokens[start:start+n])) != designator {
			continue
		}

		number := foldUpper(tokens[start+n].Text)
		// Pub 28 lists "PO BOX S-1190" among the designators that should be transformed
		// to "PO BOX" in Puerto Rico. We interpret this as meaning "S-" should be removed
		// from the beginning of all Puerto Rico PO BOX numbers.
		number, _ = strings.CutPrefix(number, "S-")

		if !pobox.CheckStreetLine("PO BOX " + number) {
			continue
		}

		return claim.Claim{
			Confidence: claim.ConfidenceExact,
			Parts: []claim.ClaimPart{
				{Start: start, Length: n, Part: claim.PartStreetName, Value: "PO BOX"},
				{Start: start + n, Length: 1, Part: claim.PartPrimaryNumber, Value: number},
			},
		}, true
	}

	return claim.Claim{}, false
}
