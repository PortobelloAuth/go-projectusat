package generaldelivery

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
)

// Candidates returns this package's readings of the address under the given
// last line.
//
// Like a post office box and unlike a military address, nothing about the last
// line is checked. The standard puts one requirement on it — the ZIP Code or
// ZIP+4 MUST be correctly applied — and that is a rule about a well formed
// record rather than a test that tells this shape from another. Applying it
// here would mean that "GENERAL DELIVERY / FAIRHAVEN XA" produced no candidate
// at all, and an address that goes unread is not reported as missing its ZIP;
// it is reported as nothing. The reading the caller can act on is the one that
// says this is general delivery and the postal code is empty.
//
// That the street line is the whole of this type's contribution is not a reason
// for it to sit out. A candidate is a reading of the entire address, and the
// parser chooses between readings, so a type that recognizes only the street
// line still has to say what the whole address looks like if that reading is
// taken.
//
// The tokens are passed alongside the claims because the claims alone do not
// say who made them. This package's claim is a lone street name, which is the
// shape any street name vocabulary produces, so recognizing its own work by
// shape would build a general delivery address out of somebody else's street.
// Re-reading the tokens with Normalize is what tells them apart.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	var candidates []*address.CandidateAddress

	for _, c := range claims {
		// The street line has to end where the last line begins. A city that
		// ran back over the phrase would assign those tokens twice.
		if !isStreetLine(tokens, c) || c.End() > line.Span.Start {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&GeneralDeliveryAddress{}, len(tokens), []claim.Claim{c}))
	}

	return candidates
}

// isStreetLine reports whether a claim is one this package made.
//
// Normalize is the recognizer — it returns an error for anything that is not a
// general delivery line — so asking it about the claim's own tokens answers
// both halves of the question at once: that the run is general delivery, and
// that it is this package's claim rather than a look-alike from another.
//
// It is anchored at both ends, so there is no longer reading to allow for. A
// claim that reached past the phrase is not one this package could have made.
func isStreetLine(tokens []token.Token, c claim.Claim) bool {
	if c.Start() < 0 || c.End() > len(tokens) {
		return false
	}

	_, err := Normalize(token.Join(tokens[c.Start():c.End()]))

	return err == nil
}
