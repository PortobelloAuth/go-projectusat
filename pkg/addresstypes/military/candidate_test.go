package military_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
)

// candidates runs the vocabularies an address is assembled from and returns
// every military candidate, over every reading of the last line. ruralroute is
// among them so the tests see the look-alike street line this package has to
// tell its own work apart from.
func candidates(source string) []*address.CandidateAddress {
	tokens := token.Tokenize(source)

	var claims []claim.Claim
	claims = append(claims, region.Claims(tokens)...)
	claims = append(claims, postalcode.Claims(tokens)...)
	claims = append(claims, country.Claims(tokens)...)
	claims = append(claims, military.Claims(tokens)...)
	claims = append(claims, ruralroute.Claims(tokens)...)

	var found []*address.CandidateAddress
	for _, line := range lastline.LineClaims(tokens, claims) {
		found = append(found, military.Candidates(tokens, claims, line)...)
	}

	return found
}

// best returns the highest confidence candidate, and whether there was one.
func best(found []*address.CandidateAddress) (*address.CandidateAddress, bool) {
	var top *address.CandidateAddress
	for _, c := range found {
		if top == nil || c.Confidence > top.Confidence {
			top = c
		}
	}

	return top, top != nil
}

func TestAnOverseasAddressIsRecognizedWhole(t *testing.T) {
	cases := []struct {
		name                     string
		source                   string
		street, number           string
		city, regionCode, postal string
	}{
		{
			name:       "APO",
			source:     "PSC 3 BOX 4120\nAPO AE 09021-0002",
			street:     "PSC 3",
			number:     "BOX 4120",
			city:       "APO",
			regionCode: "AE",
			postal:     "09021-0002",
		},
		{
			name:       "FPO",
			source:     "UNIT 100100 BOX 4120\nFPO AP 96691-0104",
			street:     "UNIT 100100",
			number:     "BOX 4120",
			city:       "FPO",
			regionCode: "AP",
			postal:     "96691-0104",
		},
		{
			name:       "DPO",
			source:     "UNIT 9900 BOX 0500\nDPO AE 09701-0500",
			street:     "UNIT 9900",
			number:     "BOX 0500",
			city:       "DPO",
			regionCode: "AE",
			postal:     "09701-0500",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			top, ok := best(candidates(tc.source))
			if !ok {
				t.Fatalf("no candidate for %q", tc.source)
			}

			if top.Confidence != claim.ConfidenceExact {
				t.Errorf("Confidence = %d, want %d", top.Confidence, claim.ConfidenceExact)
			}

			if len(top.Leftover) != 0 {
				t.Errorf("Leftover = %v, want none", top.Leftover)
			}

			a := top.Address
			if a.StreetName != tc.street || a.PrimaryNumber != tc.number {
				t.Errorf("street line = %q %q, want %q %q",
					a.StreetName, a.PrimaryNumber, tc.street, tc.number)
			}

			if a.City != tc.city || a.Region != tc.regionCode || a.Postal != tc.postal {
				t.Errorf("last line = %q %q %q, want %q %q %q",
					a.City, a.Region, a.Postal, tc.city, tc.regionCode, tc.postal)
			}
		})
	}
}

func TestALastLineWithoutAStreetLineIsNotAWeakMilitaryAddress(t *testing.T) {
	if found := candidates("APO AE 09021-0002"); len(found) != 0 {
		t.Errorf("got %d candidates, want none: the standard requires the street line", len(found))
	}
}

func TestAMilitaryStreetLineUnderAnOrdinaryLastLineIsNotClaimed(t *testing.T) {
	if found := candidates("PSC 3 BOX 4120\nDENVER CO 80201"); len(found) != 0 {
		t.Errorf("got %d candidates, want none: a domestic last line is not this type", len(found))
	}
}

func TestARuralRouteStreetLineIsNotMistakenForAMilitaryOne(t *testing.T) {
	// The two street lines have the same claim shape, so a package that
	// recognized its own work by shape would build a military address here.
	if found := candidates("RR 2 BOX 18\nAPO AE 09021-0002"); len(found) != 0 {
		t.Errorf("got %d candidates, want none: RR 2 BOX 18 is not a military street line", len(found))
	}
}

func TestHalfAMilitaryLastLineIsNotAMilitaryLastLine(t *testing.T) {
	cases := []struct{ name, source string }{
		{
			name:   "a military region under an ordinary city",
			source: "PSC 3 BOX 4120\nDENVER AE 80201",
		},
		{
			name:   "a designation over an ordinary region",
			source: "PSC 3 BOX 4120\nAPO CO 80201",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Each of these has one half of the pattern and a valid military
			// street line above it. The standard pairs the designation with
			// AE, AP or AA, so half of it is not weak evidence of the whole.
			if found := candidates(tc.source); len(found) != 0 {
				t.Errorf("got %d candidates, want none", len(found))
			}
		})
	}
}

func TestTheDesignationAndRegionMustBeInOrderAndAdjacent(t *testing.T) {
	cases := []struct{ name, source string }{
		{
			name:   "region ahead of the designation",
			source: "PSC 3 BOX 4120\nAE APO 09021-0002",
		},
		{
			name:   "something between them",
			source: "PSC 3 BOX 4120\nAPO BERLIN AE 09021-0002",
		},
		{
			name:   "a country name the standard forbids",
			source: "PSC 3 BOX 4120\nAPO AE 09021-0002 GERMANY",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Both words are military vocabulary in every one of these, so a
			// check that only asked which parts were assigned would accept
			// them. The standard fixes the order and the adjacency too.
			if found := candidates(tc.source); len(found) != 0 {
				t.Errorf("got %d candidates, want none", len(found))
			}
		})
	}
}

func TestNoCandidateAssignsATokenTwice(t *testing.T) {
	// Written on one line, the address gives lastline readings whose city runs
	// back over the street line — "PSC 3 BOX 4120 APO" — and they are rated as
	// highly as the correct one. Building an address from those would put the
	// same tokens in the city and in the street name at once.
	found := candidates("PSC 3 BOX 4120 APO AE 09021-0002")
	if len(found) == 0 {
		t.Fatal("no candidate for the single line form")
	}

	for _, c := range found {
		assigned := map[int]bool{}
		for _, cl := range c.Claims {
			for _, part := range cl.Parts {
				for i := part.Start; i < part.End(); i++ {
					if assigned[i] {
						t.Fatalf("token %d assigned twice by %+v", i, *c.Address)
					}

					assigned[i] = true
				}
			}
		}

		if c.Address.City != "APO" {
			t.Errorf("City = %q, want APO", c.Address.City)
		}
	}
}

func TestEveryCandidateNamesTheMilitaryType(t *testing.T) {
	for _, c := range candidates("PSC 3 BOX 4120\nAPO AE 09021-0002") {
		if _, ok := c.Address.Type.(*military.MilitaryAddress); !ok {
			t.Errorf("Address.Type = %T, want *military.MilitaryAddress", c.Address.Type)
		}
	}
}

func TestFormatStreetLine(t *testing.T) {
	top, ok := best(candidates("PSC 3 BOX 4120\nAPO AE 09021-0002"))
	if !ok {
		t.Fatal("no candidate")
	}

	if got := top.Address.Type.FormatStreetLine(top.Address); got != "PSC 3 BOX 4120" {
		t.Errorf("FormatStreetLine() = %q, want %q", got, "PSC 3 BOX 4120")
	}
}

// Detail is the one field an overseas military street line does not render.
// A private mailbox is rented from a commercial mail receiving agency, and an
// overseas military address is delivered through a military post office, so a
// value here could only have arrived by mistake. See FormatStreetLine.
func TestFormatStreetLineOmitsTheDetail(t *testing.T) {
	a := &address.Address{
		StreetName:    "PSC 3",
		PrimaryNumber: "BOX 4120",
		Detail:        "PMB 234",
	}

	want := "PSC 3 BOX 4120"
	if got := (&military.MilitaryAddress{}).FormatStreetLine(a); got != want {
		t.Errorf("FormatStreetLine() = %q, want %q", got, want)
	}
}
