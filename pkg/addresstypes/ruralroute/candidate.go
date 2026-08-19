package ruralroute

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
)

// Candidates returns this package's readings of the address under the given
// last line.
//
// A rural route address is an ordinary address whose street line happens to be
// a route and a box. Unlike a military address there is nothing about its last
// line to check: the city, region and postal code are whatever they are, and
// the standard puts no constraint on them. So every claim this package makes
// becomes a candidate, and the last line is taken as read.
//
// That the street line is the whole of this type's contribution is not a reason
// for it to sit out. A candidate is a reading of the entire address, and the
// parser chooses between readings, so a type that recognizes only the street
// line still has to say what the whole address looks like if that reading is
// taken — otherwise its evidence never meets the last line, and the choice
// between a rural route and some other street reading has nothing to weigh.
//
// Claims offers both the bare pattern and the reading that absorbs the trailing
// tokens the standard says do not belong; both come through here, at the
// confidence Claims gave them, because which one is right is a judgment about
// the address and not about this vocabulary.
//
// The tokens are passed alongside the claims because the claims alone do not
// say who made them. A rural route street line and a military one have the same
// shape — a street name and a primary address number over one run — so a
// package that recognized its own work by shape would build a rural route out
// of PSC 3 BOX 4120. Re-reading the tokens with this package's own recognizer
// is what tells the two apart.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	var candidates []*address.CandidateAddress

	for _, c := range claims {
		// The street line has to end where the last line begins. Nothing about
		// a rural route rules out a city that ran back over it the way a
		// military designation does, so declining the overlap is the only
		// thing keeping those tokens from being assigned twice.
		if !isStreetLine(tokens, c) || c.End() > line.Span.Start {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&RuralRouteAddress{}, len(tokens), []claim.Claim{c}))
	}

	return candidates
}

// isStreetLine reports whether a claim is one this package made.
//
// Normalize is the recognizer — it returns an error for anything that is not a
// rural route — so asking it about the claim's own tokens answers both halves
// of the question at once: that the run is a rural route, and that it is this
// package's claim rather than a look-alike from another. It accepts the longer
// reading too, since Normalize discards whatever trails the pattern.
func isStreetLine(tokens []token.Token, c claim.Claim) bool {
	if c.Start() < 0 || c.End() > len(tokens) {
		return false
	}

	_, err := Normalize(token.Join(tokens[c.Start():c.End()]))

	return err == nil
}
