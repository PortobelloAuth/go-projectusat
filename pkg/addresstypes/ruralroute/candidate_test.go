package ruralroute_test

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
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

// candidates runs the vocabularies an address is assembled from and returns
// every rural route candidate, over every reading of the last line. military is
// among them so the tests see the look-alike street line this package has to
// tell its own work apart from, and streetsuffixes so they see the competing
// reading of the text a rural route line should not carry.
func candidates(source string) []*address.CandidateAddress {
	tokens := token.Tokenize(source)

	var claims []claim.Claim
	claims = append(claims, region.Claims(tokens)...)
	claims = append(claims, postalcode.Claims(tokens)...)
	claims = append(claims, country.Claims(tokens)...)
	claims = append(claims, streetsuffixes.Claims(tokens)...)
	claims = append(claims, military.Claims(tokens)...)
	claims = append(claims, ruralroute.Claims(tokens)...)

	var found []*address.CandidateAddress
	for _, line := range lastline.LineClaims(tokens, claims) {
		found = append(found, ruralroute.Candidates(tokens, claims, line)...)
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

func TestARuralRouteIsRecognizedWhole(t *testing.T) {
	top, ok := best(candidates("RR 2 BOX 18\nBRYAN OH 43506"))
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
	if a.StreetName != "RR 2" || a.PrimaryNumber != "BOX 18" {
		t.Errorf("street line = %q %q, want %q %q",
			a.StreetName, a.PrimaryNumber, "RR 2", "BOX 18")
	}

	if a.City != "BRYAN" || a.Region != "OH" || a.Postal != "43506" {
		t.Errorf("last line = %q %q %q, want BRYAN OH 43506", a.City, a.Region, a.Postal)
	}
}

func TestAStreetLineOnlyTypeStillProducesAWholeAddress(t *testing.T) {
	// This package reads nothing but the street line, and still has to say what
	// the whole address looks like — otherwise its evidence never meets the last
	// line and the parser has nothing to weigh against a competing street
	// reading.
	top, ok := best(candidates("RR 2 BOX 18\nBRYAN OH 43506"))
	if !ok {
		t.Fatal("no candidate")
	}

	if top.Address.City == "" || top.Address.Region == "" || top.Address.Postal == "" {
		t.Errorf("candidate carries no last line: %+v", *top.Address)
	}
}

func TestTrailingTextIsReportedAsLeftoverRatherThanSwallowed(t *testing.T) {
	// The standard says a rural route line should not carry a town or street
	// name. Claims offers both the bare pattern and a reading that absorbs the
	// extra tokens; the absorbing one is deliberately the weaker, so the winner
	// keeps RR 2 BOX 18 and reports BRYAN DAIRY RD as unexplained.
	top, ok := best(candidates("RR 2 BOX 18 BRYAN DAIRY RD\nLARGO FL 33777"))
	if !ok {
		t.Fatal("no candidate")
	}

	if top.Address.PrimaryNumber != "BOX 18" {
		t.Errorf("PrimaryNumber = %q, want %q", top.Address.PrimaryNumber, "BOX 18")
	}

	if len(top.Leftover) == 0 {
		t.Fatal("Leftover = none, want the trailing tokens")
	}

	if top.Confidence >= claim.ConfidenceExact {
		t.Errorf("Confidence = %d, want below %d with tokens unexplained",
			top.Confidence, claim.ConfidenceExact)
	}
}

func TestAMilitaryStreetLineIsNotMistakenForARuralRoute(t *testing.T) {
	// The two street lines have the same claim shape, so a package that
	// recognized its own work by shape would build a rural route here.
	for _, c := range candidates("PSC 3 BOX 4120\nDENVER CO 80201") {
		t.Errorf("got candidate %+v, want none: PSC 3 BOX 4120 is not a rural route", *c.Address)
	}
}

func TestNoRuralRouteIsNoCandidates(t *testing.T) {
	if found := candidates("123 MAIN ST\nDENVER CO 80201"); len(found) != 0 {
		t.Errorf("got %d candidates, want none", len(found))
	}
}

func TestNoCandidateAssignsATokenTwice(t *testing.T) {
	// Written on one line, the address gives lastline readings whose city runs
	// back over the street line — "RR 2 BOX 18 BRYAN" — and they are rated as
	// highly as the correct one. Nothing about a rural route rules such a city
	// out the way a military designation does, so the reading has to be
	// declined for overlapping instead: it would put the same tokens in the
	// city and in the street name at once.
	found := candidates("RR 2 BOX 18 BRYAN OH 43506")
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
	}
}

func TestEveryCandidateIsWellFormed(t *testing.T) {
	for _, c := range candidates("RR 2 BOX 18 BRYAN DAIRY RD\nLARGO FL 33777") {
		if c.Address == nil {
			t.Fatal("candidate with no address")
		}

		if _, ok := c.Address.Type.(*ruralroute.RuralRouteAddress); !ok {
			t.Errorf("Address.Type = %T, want *ruralroute.RuralRouteAddress", c.Address.Type)
		}

		if len(c.Claims) < 2 {
			t.Errorf("Claims = %d, want the last line and at least one street claim", len(c.Claims))
		}

		for _, s := range c.Leftover {
			if s.Length <= 0 {
				t.Errorf("Leftover run %v is empty", s)
			}
		}
	}
}

// A highway contract route builds the same candidate a rural route does. It
// shares this package's address type, so nothing downstream has to know which
// designator it was.
func TestAHighwayContractRouteIsRecognizedWhole(t *testing.T) {
	top, ok := best(candidates("HC 4 BOX 125\nBRYAN OH 43506"))
	if !ok {
		t.Fatal("no candidate for a highway contract route")
	}

	if top.Address.StreetName != "HC 4" {
		t.Errorf("StreetName = %q, want %q", top.Address.StreetName, "HC 4")
	}
	if top.Address.PrimaryNumber != "BOX 125" {
		t.Errorf("PrimaryNumber = %q, want %q", top.Address.PrimaryNumber, "BOX 125")
	}
	if top.Address.City != "BRYAN" {
		t.Errorf("City = %q, want %q", top.Address.City, "BRYAN")
	}
	if _, isRoute := top.Address.Type.(*ruralroute.RuralRouteAddress); !isRoute {
		t.Errorf("Type = %T, want *ruralroute.RuralRouteAddress", top.Address.Type)
	}
}
