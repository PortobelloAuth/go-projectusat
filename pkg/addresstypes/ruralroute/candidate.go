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
//
// Where a private mailbox claim follows the box number on the same line, a
// second candidate offers it too. See trailingDetail for which readings that
// is (#77). The candidate without it is offered as well, so the leftover step
// on the one that strands the mailbox is what separates the two.
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

		if detail, ok := trailingDetail(tokens, claims, c, line); ok {
			candidates = append(candidates,
				line.Candidate(&RuralRouteAddress{}, len(tokens), []claim.Claim{c, detail}))
		}
	}

	return candidates
}

// trailingDetail returns the private mailbox claim that follows the route's
// box number on its own line, if the pool offers one, re-rated to Exact
// because this package knows something the vocabulary that made the claim
// cannot.
//
// Pub 28 §285's own three-line CMRA example is "RR 1 BOX 12" with "PMB 234"
// above it; the trailing form this library emits is the same tokens read the
// other way around. §241 gives the rural route delivery line as RR ## BOX ##
// and says "Do not use the words RURAL, NUMBER, NO., or the pound sign (#)",
// and §245 adds that the line carries no additional designations, so unlike
// an ordinary street line a rural route line never carries a secondary unit
// (Aaron on #102). That is what privatemailbox's demotion of # depends on —
// # is also the secondary unit designator of unspecified type, and
// secondaryunit claims it there at Exact under §213.2 — and a rural route
// line has no secondary unit position for # to be read into instead. So on
// this line both identifiers §285 permits, PMB and #, can only be the
// patient's mailbox, and both are admitted: the value privatemailbox already
// normalized to PMB n is what renders, so "# 234" renders "PMB 234" the same
// as "PMB 234" does.
//
// The vocabulary rates what the tokens could be; this package rates what
// they are on this line — the same split ordinarystreet.admitMailbox draws
// for the street line. Exact rather than Strong is the rating that matters:
// the route-and-box claim is Exact, and the candidate without the mailbox is
// the same claim alone, demoted one step by lastline.Candidate for the
// leftover run the trailing tokens leave — Exact down to Strong. A mailbox
// admitted at Strong would only tie that candidate; at Exact, the reading
// that accounts for every token wins outright, which is what the leftover
// step exists to do.
func trailingDetail(tokens []token.Token, claims []claim.Claim, c claim.Claim, line lastline.LineClaim) (claim.Claim, bool) {
	for _, d := range claims {
		if d.Start() < 0 || d.End() > len(tokens) {
			continue
		}
		if d.Start() != c.End() || d.End() > line.Span.Start {
			continue
		}
		if len(d.Parts) != 1 || d.Parts[0].Part != claim.PartDetail {
			continue
		}
		if tokens[d.Start()].Line != tokens[c.Start()].Line {
			continue
		}

		return claim.Claim{Confidence: claim.ConfidenceExact, Parts: d.Parts}, true
	}

	return claim.Claim{}, false
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
