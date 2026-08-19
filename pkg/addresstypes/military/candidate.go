package military

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
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
func (m *MilitaryAddress) FormatStreetLine(a *address.Address) string {
	return joinNonEmpty(a.StreetName, a.PrimaryNumber)
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
	if !isOverseasLastLine(line) {
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

// isOverseasLastLine reports whether a last line reading names an APO, FPO or
// DPO and one of the three military regions. Both are required: AE is a region
// this package claims and APO is a city it claims, but either alone is a
// fragment, and the standard pairs them.
func isOverseasLastLine(line lastline.LineClaim) bool {
	var city, region bool

	for _, p := range line.Claim.Parts {
		switch p.Part {
		case claim.PartCity:
			city = validCities[p.Value]
		case claim.PartRegion:
			region = validRegions[p.Value]
		}
	}

	return city && region
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

func joinNonEmpty(parts ...string) string {
	var out string
	for _, p := range parts {
		if p == "" {
			continue
		}
		if out != "" {
			out += " "
		}
		out += p
	}

	return out
}
