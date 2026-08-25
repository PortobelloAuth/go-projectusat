package military

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// MilitaryAddress is the AddressType for an overseas APO, FPO or DPO address.
//
// The street line is a facility designator with its assigned number followed by
// a box and its number, which Claims reads as a street name and a primary
// address number. Formatting is putting those two back in that order; there is
// no suffix, no directional and no secondary unit in this shape, and anything
// sitting in those fields did not come from this address type.
type MilitaryAddress struct{}

// FormatStreetLine renders "PSC 3 BOX 4120".
//
// Detail is deliberately absent, where every other address type renders it. It
// holds a private mailbox number, and a private mailbox is a box rented from a
// commercial mail receiving agency. An overseas military address is delivered
// through a military post office, so the case cannot arise, and rendering a
// value that could only have arrived by mistake would hide the mistake.
func (m *MilitaryAddress) FormatStreetLine(a *address.Address) string {
	return textutil.JoinNonEmpty(" ", a.StreetName, a.PrimaryNumber)
}

// Candidates returns this package's readings of the address under the given
// last line.
//
// Overseas military addresses are the ones this type can recognize. The
// standard requires the designation, the AE/AP/AA region and the box style
// street line, and forbids a city or country name, so all three of those are
// tests this can actually apply. Domestic military addresses are deliberately
// not claimed: the standard says they take a conventional street address with
// an ordinary city and state, which makes them indistinguishable from any other
// address by inspection, and a candidate that fires on every address is not
// evidence of anything.
//
// The street line is required rather than merely valued. The standard says a
// military address MUST have it, so an APO last line above something else is
// not a weak military address; it is an address that is not this shape. Nothing
// is returned in that case, and the parser is left to prefer some other type
// rather than to weigh a military reading it should never have been offered.
//
// One last line is taken at a time rather than the whole set, so that a caller
// that has several readings of the last line gets candidates for each and can
// see which one the address type liked. Combining them is the parser's job.
//
// The tokens are passed alongside the claims because the claims alone do not
// say who made them. A military street line and a rural route street line have
// the same shape — a street name and a primary address number over one run —
// so a package that recognized its own work by shape would happily build an
// address out of the other one's claim. Re-reading the tokens with this
// package's own recognizer is what tells the two apart.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	if !isOverseasLastLine(tokens, line) {
		return nil
	}

	var candidates []*address.CandidateAddress

	for _, c := range claims {
		// The street line has to end where the last line begins. A reading
		// whose city ran back over the street line would assign those tokens
		// twice; the designation check above rules that out today, since a
		// swallowing city is not APO, FPO or DPO, but the overlap is what
		// actually makes the reading wrong.
		if !isStreetLine(tokens, c) || c.End() > line.Span.Start {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&MilitaryAddress{}, len(tokens), []claim.Claim{c}))
	}

	return candidates
}

// isOverseasLastLine reports whether a last line reading is a military one.
//
// Two things have to hold, and the recognizer settles both. NormalizeLastLine
// says whether the tokens are the pattern — the standard fixes the order and
// the adjacency as well as the vocabularies, so a reading that found AE ahead
// of APO, or anything at all between them, is not this however military both
// words look. Then the reading has to agree with what the recognizer found,
// because the same tokens support other readings: lastline will offer "APO AE"
// as a city with no region at all, which normalizes perfectly well and would
// still build an address with an empty region.
//
// Deferring to NormalizeLastLine also holds this package to one statement of
// what a military last line is, rather than to a second one here that could
// drift from it.
func isOverseasLastLine(tokens []token.Token, line lastline.LineClaim) bool {
	if line.Span.Start < 0 || line.Span.End() > len(tokens) {
		return false
	}

	city, region, postal, err := NormalizeLastLine(
		token.Join(tokens[line.Span.Start:line.Span.End()]))
	if err != nil {
		return false
	}

	want := map[claim.Part]string{
		claim.PartCity:   city,
		claim.PartRegion: region,
		claim.PartPostal: postal,
	}

	for _, p := range line.Claim.Parts {
		if want[p.Part] != p.Value {
			return false
		}

		delete(want, p.Part)
	}

	return len(want) == 0
}

// isStreetLine reports whether a claim is the street line this package makes.
//
// NormalizeStreetLine is the recognizer — it returns an error for anything that
// is not this pattern — so asking it about the claim's own tokens answers both
// halves of the question at once: that the run is a military street line, and
// that it is this package's claim rather than a look-alike from another.
func isStreetLine(tokens []token.Token, c claim.Claim) bool {
	if c.Start() < 0 || c.End() > len(tokens) {
		return false
	}

	_, err := NormalizeStreetLine(token.Join(tokens[c.Start():c.End()]))

	return err == nil
}
