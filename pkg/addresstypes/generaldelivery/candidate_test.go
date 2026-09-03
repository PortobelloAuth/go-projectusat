package generaldelivery_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/generaldelivery"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

// candidates runs the vocabularies an address is assembled from and returns
// every general delivery candidate, over every reading of the last line. The
// other street line types are among them so the tests see the claims this
// package has to tell its own work apart from.
func candidates(source string) []*address.CandidateAddress {
	tokens := token.Tokenize(source)

	var claims []claim.Claim
	claims = append(claims, region.Claims(tokens)...)
	claims = append(claims, postalcode.Claims(tokens)...)
	claims = append(claims, country.Claims(tokens)...)
	claims = append(claims, streetsuffixes.Claims(tokens)...)
	claims = append(claims, military.Claims(tokens)...)
	claims = append(claims, ruralroute.Claims(tokens)...)
	claims = append(claims, pobox.Claims(tokens)...)
	claims = append(claims, generaldelivery.Claims(tokens)...)

	var found []*address.CandidateAddress
	for _, line := range lastline.LineClaims(tokens, claims) {
		found = append(found, generaldelivery.Candidates(tokens, claims, line)...)
	}

	return found
}

func best(found []*address.CandidateAddress) (*address.CandidateAddress, bool) {
	var top *address.CandidateAddress
	for _, c := range found {
		if top == nil || c.Confidence > top.Confidence {
			top = c
		}
	}

	return top, top != nil
}

func TestAGeneralDeliveryAddressIsReadWhole(t *testing.T) {
	top, ok := best(candidates("GENERAL DELIVERY\nSPRINGFIELD IL 62701-9999"))
	if !ok {
		t.Fatal("no candidate")
	}

	if top.Confidence != claim.ConfidenceExact {
		t.Errorf("Confidence = %d, want %d", top.Confidence, claim.ConfidenceExact)
	}

	if len(top.Leftover) != 0 {
		t.Errorf("Leftover = %v, want none", top.Leftover)
	}

	a := top.Address
	if a.StreetName != "GENERAL DELIVERY" {
		t.Errorf("street name = %q, want %q", a.StreetName, "GENERAL DELIVERY")
	}

	if a.City != "SPRINGFIELD" || a.Region != "IL" || a.Postal != "62701-9999" {
		t.Errorf("last line = %q %q %q, want SPRINGFIELD IL 62701-9999",
			a.City, a.Region, a.Postal)
	}

	if _, isType := top.Address.Type.(*generaldelivery.GeneralDeliveryAddress); !isType {
		t.Errorf("Type = %T, want *generaldelivery.GeneralDeliveryAddress", top.Address.Type)
	}
}

func TestAMissingPostalCodeStillProducesAReading(t *testing.T) {
	// The standard's MUST about the ZIP is a rule about a well formed record,
	// not a test that tells this shape from another. An address that goes
	// unread is not reported as missing its ZIP; it is reported as nothing.
	top, ok := best(candidates("GENERAL DELIVERY\nSPRINGFIELD IL"))
	if !ok {
		t.Fatal("no candidate")
	}

	a := top.Address
	if a.StreetName != "GENERAL DELIVERY" {
		t.Errorf("street name = %q, want %q", a.StreetName, "GENERAL DELIVERY")
	}

	if a.Postal != "" {
		t.Errorf("Postal = %q, want the empty string", a.Postal)
	}
}

func TestALookAlikeStreetLineIsNotClaimedAsGeneralDelivery(t *testing.T) {
	// A general delivery claim is a lone street name, which is the shape any
	// street name vocabulary produces. Recognizing its own work by shape would
	// build a general delivery address out of somebody else's street.
	for _, source := range []string{
		"PO BOX 11890\nSPRINGFIELD IL 62701",
		"RR 4 BOX 125\nSPRINGFIELD IL 62701",
		"PSC 3 BOX 4120\nAPO AE 09021",
	} {
		t.Run(source, func(t *testing.T) {
			if found := candidates(source); len(found) != 0 {
				t.Errorf("candidates = %d, want none", len(found))
			}
		})
	}
}

func TestTokensThePhraseDidNotAccountForCostConfidence(t *testing.T) {
	// A street named for a general is the competing reading here, and the
	// leftover LN is what says this candidate did not read the whole line.
	top, ok := best(candidates("GENERAL DELIVERY LN\nSPRINGFIELD IL 62701"))
	if !ok {
		t.Fatal("no candidate")
	}

	if len(top.Leftover) == 0 {
		t.Error("Leftover = none, want the unread suffix token")
	}

	if top.Confidence >= claim.ConfidenceExact {
		t.Errorf("Confidence = %d, want less than %d", top.Confidence, claim.ConfidenceExact)
	}
}

func TestTheStreetLineMustEndWhereTheLastLineBegins(t *testing.T) {
	// A last line reading whose city ran back over the phrase would assign
	// those tokens twice.
	for _, found := range candidates("GENERAL DELIVERY\nSPRINGFIELD IL 62701") {
		street := found.Claims[1]
		line := found.Claims[0]
		if street.End() > line.Start() {
			t.Errorf("street line [%d,%d) overlaps last line [%d,%d)",
				street.Start(), street.End(), line.Start(), line.End())
		}
	}
}
