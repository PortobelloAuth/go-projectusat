package lastline

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
)

// Candidate assembles a whole address from this last line and the street line
// claims an address type has decided to accept.
//
// Every address type ends with a last line, and the standard gives them all the
// same one. This is the shared implementation of that half: an address type
// supplies what only it knows — which shape its street line takes and which
// AddressType value names it — and the last line contributes the rest, along
// with the arithmetic that is identical everywhere. What is deliberately not
// shared is the decision to produce a candidate at all. That is where the
// address types actually differ, and folding it in here would make every type
// answer for every address.
//
// street is the claims the type accepted for everything ahead of the last line.
// The type may accept more than one — a business name and a street line are two
// claims — and it must not accept overlapping ones, for the same reason a
// single Claim may not assign a token twice.
//
// Confidence is the weakest accepted claim, then a step down for each run of
// leftover tokens. See address.CandidateAddress.Confidence for why a candidate
// inherits its weakest part when a last line does not.
//
// tokenCount bounds the address, so that tokens ahead of everything accepted
// are counted as leftover rather than quietly disappearing.
func (l LineClaim) Candidate(kind address.AddressType, tokenCount int, street []claim.Claim) *address.CandidateAddress {
	accepted := append([]claim.Claim{l.Claim}, street...)

	a := &address.Address{Type: kind}
	var parts []claim.ClaimPart
	confidence := claim.Confidence(0)

	for i, c := range accepted {
		if i == 0 || c.Confidence < confidence {
			confidence = c.Confidence
		}
		for _, p := range c.Parts {
			assign(a, p)
			parts = append(parts, p)
		}
	}

	leftover := claim.Span{Start: 0, Length: tokenCount}.Gaps(parts)
	for range leftover {
		confidence = demote(confidence)
	}

	return &address.CandidateAddress{
		Address:    a,
		Confidence: confidence,
		Claims:     accepted,
		Leftover:   leftover,
	}
}

// assign writes one claim part into the address field it names.
//
// A part whose Part this does not recognize is dropped rather than reported.
// claim.Part mirrors the fields of Address by construction, so an unrecognized
// one means the two have drifted apart, and that is a bug to fix in the type
// rather than an error for every caller to handle on every address.
func assign(a *address.Address, p claim.ClaimPart) {
	switch p.Part {
	case claim.PartBusinessName:
		a.BusinessName = p.Value
	case claim.PartPrimaryNumber:
		a.PrimaryNumber = p.Value
	case claim.PartPredirectional:
		a.Predirectional = p.Value
	case claim.PartStreetName:
		a.StreetName = p.Value
	case claim.PartStreetSuffix:
		a.StreetSuffix = p.Value
	case claim.PartPostdirectional:
		a.Postdirectional = p.Value
	case claim.PartSecondaryDesignator:
		a.SecondaryDesignator = p.Value
	case claim.PartSecondaryNumber:
		a.SecondaryNumber = p.Value
	case claim.PartDetail:
		a.Detail = p.Value
	case claim.PartArea:
		a.Area = p.Value
	case claim.PartCity:
		a.City = p.Value
	case claim.PartRegion:
		a.Region = p.Value
	case claim.PartPostal:
		a.Postal = p.Value
	case claim.PartCountry:
		a.Country = p.Value
	}
}
