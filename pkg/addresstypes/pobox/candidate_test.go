package pobox_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
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
// every post office box candidate, over every reading of the last line.
// military and ruralroute are among them so the tests see the two look-alike
// street lines this package has to tell its own work apart from.
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

	var found []*address.CandidateAddress
	for _, line := range lastline.LineClaims(tokens, claims) {
		found = append(found, pobox.Candidates(tokens, claims, line)...)
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

func TestAPostOfficeBoxIsRecognizedWhole(t *testing.T) {
	top, ok := best(candidates("PO BOX 11890\nDENVER CO 80201"))
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
	if a.StreetName != "PO BOX" || a.PrimaryNumber != "11890" {
		t.Errorf("street line = %q %q, want %q %q",
			a.StreetName, a.PrimaryNumber, "PO BOX", "11890")
	}

	if a.City != "DENVER" || a.Region != "CO" || a.Postal != "80201" {
		t.Errorf("last line = %q %q %q, want DENVER CO 80201", a.City, a.Region, a.Postal)
	}
}

func TestAStreetLineOnlyTypeStillProducesAWholeAddress(t *testing.T) {
	// This package reads nothing but the street line, and still has to say what
	// the whole address looks like — otherwise its evidence never meets the last
	// line and the parser has nothing to weigh against a competing street
	// reading.
	top, ok := best(candidates("PO BOX 11890\nDENVER CO 80201"))
	if !ok {
		t.Fatal("no candidate")
	}

	if top.Address.City == "" || top.Address.Region == "" || top.Address.Postal == "" {
		t.Errorf("candidate carries no last line: %+v", *top.Address)
	}
}

// The standard tells developers to rewrite DRAWER to PO BOX, so the address is
// a post office box and the value says so. What the spelling costs is
// confidence, not recognition: DRAWER is an ordinary English word, and Claims
// rates it below the reserved forms. A candidate is only as good as its weakest
// accepted claim, so the whole address inherits that.
func TestASynonymDesignatorProducesAWeakerCandidate(t *testing.T) {
	reserved, ok := best(candidates("PO BOX 214\nDENVER CO 80201"))
	if !ok {
		t.Fatal("no candidate for the reserved spelling")
	}

	synonym, ok := best(candidates("DRAWER 214\nDENVER CO 80201"))
	if !ok {
		t.Fatal("no candidate for the synonym")
	}

	if synonym.Address.StreetName != "PO BOX" || synonym.Address.PrimaryNumber != "214" {
		t.Errorf("street line = %q %q, want the synonym rewritten to PO BOX 214",
			synonym.Address.StreetName, synonym.Address.PrimaryNumber)
	}

	if synonym.Confidence >= reserved.Confidence {
		t.Errorf("DRAWER candidate rated %d, want below the reserved spelling's %d",
			synonym.Confidence, reserved.Confidence)
	}
}

func TestARuralRouteStreetLineIsNotMistakenForAPostOfficeBox(t *testing.T) {
	// The two street lines have the same claim shape, so a package that
	// recognized its own work by shape would build a PO box here.
	for _, c := range candidates("RR 2 BOX 18\nBRYAN OH 43506") {
		t.Errorf("got candidate %+v, want none: RR 2 BOX 18 is not a post office box", *c.Address)
	}
}

func TestAMilitaryStreetLineIsNotMistakenForAPostOfficeBox(t *testing.T) {
	for _, c := range candidates("PSC 3 BOX 4120\nAPO AE 09021") {
		t.Errorf("got candidate %+v, want none: PSC 3 BOX 4120 is not a post office box", *c.Address)
	}
}

func TestNoPostOfficeBoxIsNoCandidates(t *testing.T) {
	if found := candidates("123 MAIN ST\nDENVER CO 80201"); len(found) != 0 {
		t.Errorf("got %d candidates, want none", len(found))
	}
}

func TestNoCandidateAssignsATokenTwice(t *testing.T) {
	// Written on one line, the address gives lastline readings whose city runs
	// back over the street line, and they are rated as highly as the correct
	// one. Nothing about a post office box rules such a city out the way a
	// military designation does, so the reading has to be declined for
	// overlapping instead: it would put the same tokens in the city and in the
	// street line at once.
	found := candidates("PO BOX 11890 DENVER CO 80201")
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
	for _, c := range candidates("PO BOX 11890 DENVER CO 80201") {
		if c.Address == nil {
			t.Fatal("candidate with no address")
		}

		if _, ok := c.Address.Type.(*pobox.POBoxAddress); !ok {
			t.Errorf("Address.Type = %T, want *pobox.POBoxAddress", c.Address.Type)
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

// Candidates takes the claims on trust for what they say, but not for what
// they point at. A claim computed from some other token slice can name tokens
// this one does not have, and reading them would take the whole parser down
// rather than lose one reading.
func TestAClaimReachingPastTheTokensIsDeclined(t *testing.T) {
	tokens := token.Tokenize("PO BOX 11890")

	beyond := claim.Claim{
		Confidence: claim.ConfidenceExact,
		Parts: []claim.ClaimPart{
			{Start: 0, Length: 2, Part: claim.PartStreetName, Value: "PO BOX"},
			{Start: 2, Length: 9, Part: claim.PartPrimaryNumber, Value: "11890"},
		},
	}

	line := lastline.LineClaim{
		Claim: claim.Claim{
			Confidence: claim.ConfidenceExact,
			Parts:      []claim.ClaimPart{{Start: 20, Length: 1, Part: claim.PartCity, Value: "DENVER"}},
		},
		Span: claim.Span{Start: 20, Length: 1},
	}

	if found := pobox.Candidates(tokens, []claim.Claim{beyond}, line); len(found) != 0 {
		t.Errorf("got %d candidates, want none for a claim past the end of the tokens", len(found))
	}
}

func TestFormatStreetLine(t *testing.T) {
	// A post office box is not a place on a street, so the fields that describe
	// one have nothing to say here and are not rendered even when set.
	a := &address.Address{
		StreetName:          "PO BOX",
		PrimaryNumber:       "11890",
		Predirectional:      "N",
		StreetSuffix:        "ST",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "3",
	}

	if got := (&pobox.POBoxAddress{}).FormatStreetLine(a); got != "PO BOX 11890" {
		t.Errorf("FormatStreetLine() = %q, want %q", got, "PO BOX 11890")
	}
}
