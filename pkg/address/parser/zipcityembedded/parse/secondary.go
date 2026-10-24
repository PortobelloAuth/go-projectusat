package parse

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ordinarystreet"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// agrees reports whether zipcity recognizes a reading's street.
func (p *Parser) agrees(r reference, a *address.Address) bool {
	ans, ok := p.streetAgreement(r, a)
	return ok && ans == agrees
}

// recognizable reports whether a reading's street names a place zipcity could
// have seen. It is streetAgreement's own type guard, asked ahead of the
// question so a window is never adopted on a closed form's pseudo-name — see
// streetAgreement for why "PO BOX" is not a street.
func recognizable(a *address.Address) bool {
	civic, ok := a.Type.(address.DataDependentAddressType)
	return ok && civic.IsCivicAddressType() && a.StreetName != ""
}

// leadingSecondary moves a secondary designator written ahead of the street
// to where the standard puts it, rewriting best in place.
//
// Step 6 of #17: the secondary designator and its details sit between the
// street and the city, and finding them is what fixes the right edge of the
// street window step 7 slides. APARTMENT 3200 152 SOUTH TECH DR writes the
// designator on the wrong side of the number, which Publication 28 §213 does
// not allow on a standardized record but real intake data does constantly.
//
// The grammar cannot recover from it. ordinarystreet reads a street line left
// to right, so a leading designator is not a designator to it — the whole run
// becomes one street name, the primary number is never found, and there is no
// candidate in the pile holding the right answer for choose to pick. Measured:
// APARTMENT 3200 152 SOUTH TECH DR offers exactly two readings and both keep
// APARTMENT in the name. So this is a re-read rather than a tie-break, which
// is why it runs here beside firmLine rather than inside choose.
//
// The guards are what keep it from firing on an address that means what it
// says. Only the catchall type is touched: a post office box, a rural route
// and the rest have each read their own line whole and there is nothing about
// them for a designator to be ahead of. Only a reading with no primary number
// of its own, which is the symptom the leading designator causes. Only a
// numbered designator — APT 3200, # 3200, PMB 456 — since a bare designator
// word at the start of a line is far more often a street name (KEY WEST is
// the vocabulary's own example). And only when the shorter re-read finds a
// street and claims no secondary of its own, so a designator is never stacked
// on a reading that already has one.
func (p *Parser) leadingSecondary(tokens []token.Token, claims []claim.Claim, best *address.CandidateAddress, lastlinepatch bool) {
	a := best.Address
	if _, ok := a.Type.(*ordinarystreet.OrdinaryStreetAddress); !ok {
		return
	}
	// Detail is excluded here, not folded into the rest.Address.SecondaryDesignator
	// check below: a leading designator stacked on a reading whose own tail
	// already claimed Detail (a trailing PMB) is a private-mailbox conflict
	// this function has no standing to resolve. SecondaryDesignator is not
	// excluded here on purpose — UNIT 3200 152 TECH DR ROOM 12 reads with no
	// primary number of its own (the documented symptom) and ROOM 12 already
	// in SecondaryDesignator, because the whole leading run was swallowed into
	// StreetName. primaryWithin below is what actually tests for "no primary
	// number of its own"; this field is the wrong proxy for that and bailing
	// on it here hid the stacked-designator case entirely.
	if a.Detail != "" {
		return
	}

	unit, width := leadingUnit(claims)
	if width == 0 || width >= len(tokens) || tokens[width-1].Line != tokens[0].Line {
		return
	}
	if !primaryWithin(best, width) {
		return
	}

	rest, _ := p.read(tokens[width:], lastlinepatch)
	if rest == nil || rest.Address == nil || rest.Address.StreetName == "" {
		return
	}
	if rest.Address.Detail != "" {
		return
	}

	// The re-read's own tail can already hold a secondary designator of its
	// own — UNIT 3200 152 TECH DR ROOM 12 re-reads as 152 TECH DR with ROOM
	// 12 already in SecondaryDesignator/SecondaryNumber. The standard still
	// has one designator slot, so the leading designator does not displace
	// it; it folds in ahead of it as a compound SecondaryNumber, the same
	// shape FormatStreetLine already joins with single spaces. The leading
	// designator's own word becomes the slot's designator, since it is the
	// one that sits closest to the street per Publication 28 §213.
	if rest.Address.SecondaryDesignator != "" {
		rest.Address.SecondaryNumber = textutil.JoinNonEmpty(" ",
			unit.number, rest.Address.SecondaryDesignator, rest.Address.SecondaryNumber)
		rest.Address.SecondaryDesignator = unit.designator
	} else {
		rest.Address.SecondaryDesignator = unit.designator
		rest.Address.SecondaryNumber = unit.number
		rest.Address.Detail = unit.detail
	}
	rest.Address.BusinessName = a.BusinessName
	*best = *rest
}

// primaryWithin reports whether a reading spent nothing outside the leading
// run on its primary address number — either it found none, or the tokens it
// found one in are the very tokens the designator wants back.
//
// #3200 152 SOUTH TECH DR is why this is a span test and not "has no primary
// number". ordinarystreet reads #3200 as the primary number and 152 SOUTH
// TECH DR as the street, which is a complete account of the line and looks
// nothing like the symptom the guard above is watching for. But Publication
// 28 has no primary number written with a pound sign — §213.2 makes # the
// secondary designator of unspecified type, and secondaryunit claims that
// token Exact as exactly that — so the number the reading found is the unit,
// in the one place the standard does not put it. Peeling it moves it; nothing
// is lost, because the run it came out of is the run being reassigned.
func primaryWithin(best *address.CandidateAddress, width int) bool {
	if best.Address.PrimaryNumber == "" {
		return true
	}
	for _, held := range best.Claims {
		for _, part := range held.Parts {
			if part.Part == claim.PartPrimaryNumber && part.End() > width {
				return false
			}
		}
	}
	return true
}

// unitParts is a numbered secondary designator read off one claim, in the
// fields Address keeps it in.
type unitParts struct {
	designator string
	number     string
	detail     string
}

// leadingUnit is the widest numbered secondary designator or private mailbox
// starting at the first token, and how many tokens it covers, or a zero width
// when there is none.
//
// Widest first because the vocabularies offer every extent they can support
// and the longer reading is the one that accounts for more of the input;
// secondaryunit's own numbered claim is two tokens where privatemailbox's can
// be three. A designator claimed without a number is not offered here at all
// — see leadingSecondary for why.
func leadingUnit(claims []claim.Claim) (unitParts, int) {
	var best unitParts
	width := 0

	for _, held := range claims {
		if held.Start() != 0 || held.End() <= width {
			continue
		}
		var unit unitParts
		ok := true
		for _, part := range held.Parts {
			switch part.Part {
			case claim.PartSecondaryDesignator:
				unit.designator = part.Value
			case claim.PartSecondaryNumber:
				unit.number = part.Value
			case claim.PartDetail:
				unit.detail = part.Value
			default:
				ok = false
			}
		}
		if !ok || (unit.number == "" && unit.detail == "") {
			continue
		}
		best, width = unit, held.End()
	}

	return best, width
}
