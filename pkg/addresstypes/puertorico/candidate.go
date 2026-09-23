package puertorico

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico/ruralroute"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
)

// PuertoRicoAddress is the AddressType for a Puerto Rico street address.
//
// The street line is a primary address number and a street name that carries
// its own type at the front, so the fields it fills are the ordinary ones and
// the order they are rendered in is the ordinary order. That is only true
// because the Spanish type stays inside StreetName: a.Suffix is never used to
// hold it. p. 26 is explicit that the type is part of the name rather than a
// suffix that moved, and this package does not play that game for English
// prefixes either — nothing here reaches for a.Suffix to carry a leading
// word. What makes the address Puerto Rican is the vocabulary that reads it
// and the urbanization line above it, not a different arrangement of the
// street line.
type PuertoRicoAddress struct{}

// FormatStreetLine renders "A17 CALLE AMAPOLA", or "1 COND MIRAFLOR APT 104"
// where a secondary designator follows the name.
//
// The field order is the one address.Address.FormatStreetLine already owns for
// an address with no special format, so this reaches that branch rather than
// restating it. There is no suffix and no trailing directional to place: a
// Puerto Rico street type leads the name and stays inside it, per p. 26. See
// CONTRIBUTING §1.2, and ordinarystreet, which delegates the same way.
func (p *PuertoRicoAddress) FormatStreetLine(a *address.Address) string {
	ordinary := *a
	ordinary.Type = nil

	return ordinary.FormatStreetLine()
}

// Candidates returns this package's reading of the address under the given
// last line.
//
// The dialect gate comes first: a Puerto Rico address is one whose last line
// says PR or carries a Puerto Rico ZIP, and nothing below that line can tell.
// That is the answer this package has to the question raised on #71 — which
// vocabularies may claim on a Puerto Rico address. Claims(tokens) cannot
// answer it, because the dialect is a judgment about the whole address and a
// Claims function sees only tokens. So the Spanish street vocabulary makes no
// claim into the shared pool at all, and reads the street line here instead,
// where the last line is known. A mainland address never meets it, which is
// the same protection in the other direction: 1000 AVE E in Florida is not
// offered a Spanish reading.
//
// Reading the tokens rather than the pool is also what #70 asks of an address
// type, and it is what lets a single-line address be read: the street line
// ends where the last line begins, whether or not a line break says so.
//
// Where an urbanization sits above the street line, two readings are offered —
// with it and without — so that a reading which strands the urbanization is
// ranked against one that does not, rather than being the only one on offer.
//
// A rural route or highway contract route is read from the same gate, by the
// ruralroute sub-package, and is not conditional on there being a street line
// at all: p. 30's route is the whole of the delivery address, and the line the
// standard prints beneath it is a sector name to be eliminated rather than a
// street.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	if !isPuertoRicoLastLine(line) {
		return nil
	}

	candidates := routeCandidates(tokens, line)

	street, ok := streetLine(tokens, line)
	if !ok {
		return candidates
	}

	candidates = append(candidates,
		line.Candidate(&PuertoRicoAddress{}, len(tokens), []claim.Claim{street}))

	for _, c := range claims {
		if !isUrbanization(c) || c.End() > street.Start() {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&PuertoRicoAddress{}, len(tokens), []claim.Claim{c, street}))
	}

	return candidates
}

// routeCandidates offers one reading for each route the Spanish vocabulary
// finds above the last line.
//
// A route claim reaching into the last line is discarded rather than trimmed:
// the two would assign the same tokens, and a candidate may not claim one
// twice. That is also the guard against reading a Puerto Rico ZIP range as a
// box number.
func routeCandidates(tokens []token.Token, line lastline.LineClaim) []*address.CandidateAddress {
	var candidates []*address.CandidateAddress

	for _, c := range ruralroute.Claims(tokens) {
		if c.End() > line.Span.Start {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&ruralroute.PuertoRicoRouteAddress{}, len(tokens), []claim.Claim{c}))
	}

	return candidates
}

// isPuertoRicoLastLine reports whether the last line puts the address in
// Puerto Rico.
//
// Either the region or the postal code settles it on its own — see
// UsePRDialect — so an address that arrives with one and not the other is
// still read in Spanish.
func isPuertoRicoLastLine(line lastline.LineClaim) bool {
	var region, postal string
	for _, p := range line.Claim.Parts {
		switch p.Part {
		case claim.PartRegion:
			region = p.Value
		case claim.PartPostal:
			postal = p.Value
		}
	}

	return UsePRDialect(region, postal)
}

// streetLine reads the delivery line immediately above the last line.
//
// The street line is the bottom line of the street address block, so the one
// that ends where the last line begins is the one to read. Everything above it
// is the urbanization and the secondary identifier, which have their own
// readings.
func streetLine(tokens []token.Token, line lastline.LineClaim) (claim.Claim, bool) {
	end := line.Span.Start
	if end <= 0 || end > len(tokens) {
		return claim.Claim{}, false
	}

	start := end - 1
	for start > 0 && tokens[start-1].Line == tokens[end-1].Line {
		start--
	}

	number, name, err := NormalizeStreetLine(token.Join(tokens[start:end]))
	if err != nil {
		return claim.Claim{}, false
	}

	return claim.Claim{
		Confidence: claim.ConfidenceExact,
		Parts: []claim.ClaimPart{
			{
				Start:  start,
				Length: 1,
				Part:   claim.PartPrimaryNumber,
				Value:  number,
			},
			{
				Start:  start + 1,
				Length: end - start - 1,
				Part:   claim.PartStreetName,
				Value:  name,
			},
		},
	}, true
}

// isUrbanization reports whether a claim is the urbanization line this package
// reads. Area is the part it writes, and no other vocabulary writes it.
func isUrbanization(c claim.Claim) bool {
	for _, p := range c.Parts {
		if p.Part != claim.PartArea {
			return false
		}
	}

	return len(c.Parts) > 0
}
