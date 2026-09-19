package pobox

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// POBoxAddress is the AddressType for a post office box address.
//
// The street line is the designator and the box number, which Claims reads as
// a street name and a primary address number. Formatting is putting those two
// back in that order.
//
// There is no suffix, no directional and no secondary unit in this shape. A
// post office box is not a place on a street, so there is nothing for those
// fields to describe, and anything sitting in them did not come from this
// address type.
type POBoxAddress struct{}

// FormatStreetLine renders "PO BOX 11890", or "PO BOX 159753 PMB 3571" where
// the box carries a private mailbox number.
//
// The standard's CMRA section says:
//
//	The words POST OFFICE BOX or PO BOX and the private mailbox number MUST
//	NOT be used on the Street Address Line. The Street Address Line is the
//	standardized address of the private company.
//
// Read literally that forbids the second form, and one of the section's own
// examples is "PO BOX 159753 PMB 3571". The reading consistent with the
// examples is that PO BOX must not be used *instead of* PMB — the two name
// different things and one cannot stand in for the other. Documented here per
// CONTRIBUTING §2, since resolving an ambiguity in the standard is a decision
// and not an implementation detail.
//
// Detail follows the box number, the same place it follows the secondary
// number in the ordinary street line. The section also shows PMB on a line of
// its own above the street line; this library emits only the trailing form,
// because one output form per address is what makes two addresses comparable.
func (p *POBoxAddress) FormatStreetLine(a *address.Address) string {
	return textutil.JoinNonEmpty(" ", a.StreetName, a.PrimaryNumber, a.Detail)
}

// Candidates returns this package's readings of the address under the given
// last line.
//
// A post office box address is an ordinary address whose street line happens to
// be a box. Like a rural route and unlike a military address there is nothing
// about its last line to check: the standard puts no constraint on the city,
// region or postal code above a PO box, so the last line is taken as read and
// every claim this package makes becomes a candidate.
//
// That the street line is the whole of this type's contribution is not a reason
// for it to sit out. A candidate is a reading of the entire address, and the
// parser chooses between readings, so a type that recognizes only the street
// line still has to say what the whole address looks like if that reading is
// taken.
//
// The tokens are passed alongside the claims because the claims alone do not
// say who made them. A post office box street line has the same shape as a
// rural route and a military one — a street name and a primary address number
// over one run — so a package that recognized its own work by shape would build
// a PO box out of RR 4 BOX 125. Re-reading the tokens with this package's own
// recognizer is what tells the three apart.
//
// Where a private mailbox claim follows the box number on the same line, a
// second candidate offers it too. See trailingDetail for which readings that
// is (#77). A private mailbox on the line immediately above the box is read
// the same way; see aboveLineDetail (#98). The candidate without either is
// offered as well, so the leftover step on the one that strands the mailbox
// is what separates the two.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	var candidates []*address.CandidateAddress

	for _, c := range claims {
		// The street line has to end where the last line begins. Nothing about
		// a post office box rules out a city that ran back over it the way a
		// military designation does, so declining the overlap is the only thing
		// keeping those tokens from being assigned twice.
		if !isStreetLine(tokens, c) || c.End() > line.Span.Start {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&POBoxAddress{}, len(tokens), []claim.Claim{c}))

		if detail, ok := trailingDetail(tokens, claims, c, line); ok {
			candidates = append(candidates,
				line.Candidate(&POBoxAddress{}, len(tokens), []claim.Claim{c, detail}))
		}

		if detail, ok := aboveLineDetail(tokens, claims, c); ok {
			candidates = append(candidates,
				line.Candidate(&POBoxAddress{}, len(tokens), []claim.Claim{c, detail}))
		}
	}

	return candidates
}

// trailingDetail returns the private mailbox claim that follows the box
// number on its own line, if the pool offers one, re-rated to Exact because
// this package knows something the vocabulary that made the claim cannot.
//
// Pub 28 §285 gives "PO BOX 159753 PMB 3571" as a correct form: PMB names the
// patient's private box, immediately after the CMRA's own box number. §281
// standardizes the box delivery line as PO BOX and the box number and nothing
// more, so unlike an ordinary street line a box line never carries a
// secondary unit (Aaron on #102). That is what privatemailbox's demotion of #
// depends on — # is also the secondary unit designator of unspecified type,
// and secondaryunit claims it there at Exact under §213.2 — and a box line
// has no secondary unit position for # to be read into instead. So on this
// line both identifiers §285 permits, PMB and #, can only be the patient's
// mailbox, and both are admitted: the value privatemailbox already normalized
// to PMB n is what renders, so "# 5" renders "PMB 5" the same as "PMB 5"
// does.
//
// The vocabulary rates what the tokens could be; this package rates what
// they are on this line — the same split ordinarystreet.admitMailbox draws
// for the street line. Exact rather than Strong is the rating that matters:
// the box claim is Exact, and the candidate without the mailbox is the same
// claim alone, demoted one step by lastline.Candidate for the leftover run
// the trailing tokens leave — Exact down to Strong. A mailbox admitted at
// Strong would only tie that candidate; at Exact, the reading that accounts
// for every token wins outright, which is what the leftover step exists to
// do.
//
// A box number can never end in a glued #1234 in the first place: this
// package requires its whole pattern, POBoxAddress number included, to claim
// a box at all (Aaron on #76).
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

// aboveLineDetail returns the private mailbox claim that covers the line
// immediately above the box's own line, if the pool offers one, re-rated to
// Exact for the same reason trailingDetail is: this package knows something
// the vocabulary that made the claim cannot.
//
// Pub 28 §285's four-line CMRA form puts PMB 234 or #234 above the street
// line instead of trailing it: "Either a three line or four line address
// format can be used with a CMRA address and the PMB or # identifier." The
// reason trailingDetail rejected the # reading no longer holds — a box line
// has no secondary unit position, per trailingDetail — and that is equally
// true of the line above the box: there is nothing there for # to mean but
// the mailbox, so both identifiers are admitted the same way.
func aboveLineDetail(tokens []token.Token, claims []claim.Claim, c claim.Claim) (claim.Claim, bool) {
	start := lineStart(tokens, c.Start())
	if start <= 0 {
		return claim.Claim{}, false
	}

	above := lineStart(tokens, start-1)

	for _, d := range claims {
		if d.Start() < 0 || d.End() > len(tokens) {
			continue
		}
		if d.Start() != above || d.End() != start {
			continue
		}
		if len(d.Parts) != 1 || d.Parts[0].Part != claim.PartDetail {
			continue
		}

		return claim.Claim{Confidence: claim.ConfidenceExact, Parts: d.Parts}, true
	}

	return claim.Claim{}, false
}

// lineStart returns the index of the first token on the same line as at.
func lineStart(tokens []token.Token, at int) int {
	start := at
	for start > 0 && tokens[start-1].Line == tokens[at].Line {
		start--
	}

	return start
}

// isStreetLine reports whether a claim is one this package made.
//
// Normalize is the recognizer — it returns an error for anything that is not a
// post office box — so asking it about the claim's own tokens answers both
// halves of the question at once: that the run is a PO box, and that it is this
// package's claim rather than a look-alike from another.
//
// It is anchored at both ends, so unlike a rural route there is no longer
// reading to allow for. A claim that reached past the box number is not one
// this package could have made.
func isStreetLine(tokens []token.Token, c claim.Claim) bool {
	if c.Start() < 0 || c.End() > len(tokens) {
		return false
	}

	_, err := Normalize(token.Join(tokens[c.Start():c.End()]))

	return err == nil
}
