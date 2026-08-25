package lastline_test

import (
	"slices"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

// read runs the vocabularies the last line is built from and returns the
// readings, best first. The extra vocabularies are there so the tests see the
// same competition the parser will: region and postalcode supply the pattern,
// and highways and streetsuffixes supply the claims a last line has to rule
// out.
func read(source string) ([]token.Token, []lastline.LineClaim) {
	tokens := token.Tokenize(source)

	var claims []claim.Claim
	claims = append(claims, region.Claims(tokens)...)
	claims = append(claims, postalcode.Claims(tokens)...)
	claims = append(claims, country.Claims(tokens)...)
	claims = append(claims, highways.Claims(tokens)...)
	claims = append(claims, streetsuffixes.Claims(tokens)...)

	lines := lastline.LineClaims(tokens, claims)
	slices.SortFunc(lines, func(a, b lastline.LineClaim) int {
		if a.Claim.Confidence != b.Claim.Confidence {
			return int(b.Claim.Confidence) - int(a.Claim.Confidence)
		}

		return b.Claim.Length() - a.Claim.Length()
	})

	return tokens, lines
}

// valueOf returns what a reading assigns to a part, and whether it assigns one
// at all.
func valueOf(line lastline.LineClaim, part claim.Part) (string, bool) {
	for _, p := range line.Claim.Parts {
		if p.Part == part {
			return p.Value, true
		}
	}

	return "", false
}

func TestBestReadingOfACompleteLastLine(t *testing.T) {
	cases := []struct {
		name                     string
		source                   string
		city, regionCode, postal string
	}{
		{
			name:   "line break marks the city",
			source: "8011 SOUTH CAROLINA AVE\nWEST JORDAN UT 84088",
			city:   "WEST JORDAN", regionCode: "UT", postal: "84088",
		},
		{
			name:   "commas mark the city on one line",
			source: "8011 SOUTH CAROLINA AVE, WEST JORDAN, UT 84088",
			city:   "WEST JORDAN", regionCode: "UT", postal: "84088",
		},
		{
			name:   "a multi token city",
			source: "123 MAIN ST\nSALT LAKE CITY, UT 84101",
			city:   "SALT LAKE CITY", regionCode: "UT", postal: "84101",
		},
		{
			name:   "a military last line",
			source: "UNIT 2050 BOX 4190\nAPO AP 96278-2050",
			city:   "APO", regionCode: "AP", postal: "96278-2050",
		},
		{
			name:   "a ZIP+4",
			source: "123 MAIN ST\nWASHINGTON, DC 20024-2101",
			city:   "WASHINGTON", regionCode: "DC", postal: "20024-2101",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, lines := read(tc.source)
			if len(lines) == 0 {
				t.Fatalf("no readings for %q", tc.source)
			}

			best := lines[0]
			if best.Claim.Confidence != claim.ConfidenceExact {
				t.Errorf("best reading is %d, want ConfidenceExact", best.Claim.Confidence)
			}
			if len(best.Leftover) != 0 {
				t.Errorf("best reading leaves %v over, want nothing unexplained", best.Leftover)
			}

			for part, want := range map[claim.Part]string{
				claim.PartCity:   tc.city,
				claim.PartRegion: tc.regionCode,
				claim.PartPostal: tc.postal,
			} {
				got, ok := valueOf(best, part)
				if !ok {
					t.Errorf("best reading assigns no %s", part)
					continue
				}
				if got != want {
					t.Errorf("%s = %q, want %q", part, got, want)
				}
			}
		})
	}
}

// The case from #37. PR is a region, and it is also the Publication 28
// abbreviation for PRAIRIE, and highways reads PR 00926 as a state highway.
// Deciding the last line is what settles all three, and recording that it did
// is why LineClaim is not just a Claim.
func TestTheChosenLineRejectsTheClaimsItRulesOut(t *testing.T) {
	_, lines := read("URB LAS GLADIOLAS\n150 CALLE A APT 4\nSAN JUAN PR 00926")
	if len(lines) == 0 {
		t.Fatal("no readings")
	}

	best := lines[0]

	if city, _ := valueOf(best, claim.PartCity); city != "SAN JUAN" {
		t.Errorf("city = %q, want %q", city, "SAN JUAN")
	}

	var sawHighway, sawSuffix bool
	for _, rejected := range best.Rejected {
		for _, p := range rejected.Parts {
			switch {
			case p.Part == claim.PartStreetName && p.Length > 1:
				sawHighway = true
			case p.Part == claim.PartStreetSuffix:
				sawSuffix = true
			}
		}
	}

	if !sawHighway {
		t.Error("the highway reading of PR 00926 was not rejected")
	}
	if !sawSuffix {
		t.Error("the street suffix reading of PR was not rejected")
	}
}

// Claims outside the line are not this package's business, and saying nothing
// about them is different from rejecting them.
func TestClaimsOutsideTheLineAreNotRejected(t *testing.T) {
	_, lines := read("100 SOUTH CAROLINA AVE\nWEST JORDAN, UT 84088")
	if len(lines) == 0 {
		t.Fatal("no readings")
	}

	for _, rejected := range lines[0].Rejected {
		if rejected.End() <= lines[0].Span.Start {
			t.Errorf("claim at %d..%d is ahead of the line at %d and should have been left alone",
				rejected.Start(), rejected.End(), lines[0].Span.Start)
		}
	}
}

// A reading that strands tokens is a worse account of the line than one that
// does not, and that is the whole of what leftovers do.
func TestLeftoversLowerConfidence(t *testing.T) {
	_, clean := read("123 MAIN ST\n80201")
	if len(clean) != 1 {
		t.Fatalf("got %d readings for a bare postal code, want 1", len(clean))
	}
	if len(clean[0].Leftover) != 0 {
		t.Errorf("leftover = %v, want none: the line is the postal code", clean[0].Leftover)
	}

	_, stranded := read("123 MAIN ST\nFOOTOWN 80201")

	var postalOnly *lastline.LineClaim
	for i, line := range stranded {
		if _, ok := valueOf(line, claim.PartCity); !ok {
			postalOnly = &stranded[i]
			break
		}
	}
	if postalOnly == nil {
		t.Fatal("no postal only reading")
	}

	if len(postalOnly.Leftover) != 1 || postalOnly.Leftover[0].Length != 1 {
		t.Errorf("leftover = %v, want the one unexplained token", postalOnly.Leftover)
	}
	if postalOnly.Claim.Confidence >= clean[0].Claim.Confidence {
		t.Errorf("stranding a token rated %d, no worse than the clean reading at %d",
			postalOnly.Claim.Confidence, clean[0].Claim.Confidence)
	}
}

// Aaron's point on #37: a token in a reasonable position for the city is the
// city, not a leftover. The shape is not one Project US@ documents, so it is
// offered alongside the postal only reading rather than instead of it, and at
// equal confidence the tiebreak prefers the reading that explains more.
func TestACityAheadOfAPostalCodeIsClaimed(t *testing.T) {
	_, lines := read("123 MAIN ST\nDENVER 80201")

	var withCity, withoutCity bool
	for _, line := range lines {
		if city, ok := valueOf(line, claim.PartCity); ok {
			if city != "DENVER" {
				t.Errorf("city = %q, want %q", city, "DENVER")
			}
			withCity = true
			if len(line.Leftover) != 0 {
				t.Errorf("leftover = %v, want none: the reading explains the line", line.Leftover)
			}
		} else {
			withoutCity = true
		}
	}

	if !withCity {
		t.Error("DENVER was never read as the city")
	}
	if !withoutCity {
		t.Error("the documented postal only reading was not also offered")
	}

	if len(lines) == 0 {
		t.Fatal("no readings")
	}
	if _, ok := valueOf(lines[0], claim.PartCity); !ok {
		t.Error("the reading that explains DENVER did not sort first")
	}
}

// The undocumented shape must never outrank a documented one over the same
// tokens, or it would start deciding addresses the standard already decides.
func TestTheCityAndPostalShapeNeverOutranksADocumentedOne(t *testing.T) {
	_, lines := read("123 MAIN ST\nDENVER, CO 80201")
	if len(lines) == 0 {
		t.Fatal("no readings")
	}

	best := lines[0]
	if _, ok := valueOf(best, claim.PartRegion); !ok {
		t.Fatalf("best reading assigns no region, confidence %d", best.Claim.Confidence)
	}

	for _, line := range lines {
		if _, hasRegion := valueOf(line, claim.PartRegion); hasRegion {
			continue
		}
		if _, hasCity := valueOf(line, claim.PartCity); !hasCity {
			continue
		}
		if line.Claim.Confidence >= best.Claim.Confidence {
			t.Errorf("the regionless reading rated %d against the complete pattern at %d",
				line.Claim.Confidence, best.Claim.Confidence)
		}
	}
}

// A country on a line of its own makes the last line two physical lines. The
// span has to cover what the reading actually assigns.
func TestACountryOnItsOwnLineExtendsTheSpan(t *testing.T) {
	tokens, lines := read("123 MAIN ST\nWEST JORDAN, UT 84088\nUNITED STATES")
	if len(lines) == 0 {
		t.Fatal("no readings")
	}

	best := lines[0]
	if _, ok := valueOf(best, claim.PartCountry); !ok {
		t.Fatal("best reading assigns no country")
	}

	for _, p := range best.Claim.Parts {
		if p.Start < best.Span.Start || p.End() > best.Span.End() {
			t.Errorf("%s covers %d..%d, outside the span %d..%d",
				p.Part, p.Start, p.End(), best.Span.Start, best.Span.End())
		}
	}

	if best.Span.End() != len(tokens) {
		t.Errorf("span ends at %d, want %d: the last line runs to the end of the address",
			best.Span.End(), len(tokens))
	}
}

// The first token of a single line address is where the address begins, not
// where its last line does. Treating it as a boundary would rate the reading
// that swallows the street address as an exact one.
func TestTheStartOfASingleLineAddressIsNotACityBoundary(t *testing.T) {
	_, lines := read("8011 SOUTH CAROLINA AVE, WEST JORDAN, UT 84088")

	for _, line := range lines {
		city, ok := valueOf(line, claim.PartCity)
		if !ok || city != "8011 SOUTH CAROLINA AVE WEST JORDAN" {
			continue
		}
		if line.Claim.Confidence >= claim.ConfidenceExact {
			t.Errorf("the reading that swallows the street address rated %d", line.Claim.Confidence)
		}

		return
	}
}

func TestNoLastLineIsNoReadings(t *testing.T) {
	for _, source := range []string{"", "8011 SOUTH CAROLINA AVE"} {
		if _, lines := read(source); len(lines) != 0 {
			t.Errorf("read(%q) returned %d readings, want none", source, len(lines))
		}
	}
}

// Every reading has to be internally consistent, whatever else is true of it:
// the span covers the parts, the parts do not overlap each other, and the
// leftovers are exactly the tokens no part covers.
func TestEveryReadingIsWellFormed(t *testing.T) {
	sources := []string{
		"8011 SOUTH CAROLINA AVE\nWEST JORDAN UT 84088",
		"8011 SOUTH CAROLINA AVE, WEST JORDAN, UT 84088",
		"URB LAS GLADIOLAS\n150 CALLE A APT 4\nSAN JUAN PR 00926",
		"UNIT 2050 BOX 4190\nAPO AP 96278-2050",
		"123 MAIN ST\nDENVER 80201",
		"123 MAIN ST\nWEST JORDAN, UT 84088\nUNITED STATES",
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			_, lines := read(source)
			if len(lines) == 0 {
				t.Fatal("no readings")
			}

			for _, line := range lines {
				covered := map[int]bool{}

				for _, p := range line.Claim.Parts {
					if p.Start < line.Span.Start || p.End() > line.Span.End() {
						t.Errorf("%s covers %d..%d, outside the span %d..%d",
							p.Part, p.Start, p.End(), line.Span.Start, line.Span.End())
					}
					for i := p.Start; i < p.End(); i++ {
						if covered[i] {
							t.Errorf("token %d is claimed twice within one reading", i)
						}
						covered[i] = true
					}
				}

				for _, l := range line.Leftover {
					for i := l.Start; i < l.End(); i++ {
						if covered[i] {
							t.Errorf("token %d is both claimed and left over", i)
						}
						covered[i] = true
					}
				}

				for i := line.Span.Start; i < line.Span.End(); i++ {
					if !covered[i] {
						t.Errorf("token %d is inside the span and neither claimed nor left over", i)
					}
				}
			}
		})
	}
}

// A country on its own line was hiding the city: anchoring to the final line
// looked for a city after the region rather than ahead of it, and the complete
// pattern was never offered. pkg/region carries the Canadian provinces, so this
// address has a region claim on ON and there is no excuse for missing it.
func TestACountryOnItsOwnLineDoesNotHideTheCity(t *testing.T) {
	_, lines := read("123 MAIN ST\nTORONTO ON M5V 3A8\nCANADA")
	if len(lines) == 0 {
		t.Fatal("no readings")
	}

	best := lines[0]
	if best.Claim.Confidence != claim.ConfidenceExact {
		t.Errorf("best reading is %d, want ConfidenceExact", best.Claim.Confidence)
	}

	for part, want := range map[claim.Part]string{
		claim.PartCity:    "TORONTO",
		claim.PartRegion:  "ON",
		claim.PartPostal:  "M5V 3A8",
		claim.PartCountry: "CANADA",
	} {
		got, ok := valueOf(best, part)
		if !ok {
			t.Errorf("best reading assigns no %s", part)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", part, got, want)
		}
	}
}

// Candidate assigns each claim part to the field of the same name, so a part
// added to claim.Part without a case in assign fails here rather than being
// silently dropped from every address that carries it. See assign: claim.Part
// mirrors the fields of address.Address by construction, and this is the test
// that keeps the construction honest.
func TestEveryClaimPartReachesItsField(t *testing.T) {
	cases := []struct {
		part  claim.Part
		field func(*address.Address) string
	}{
		{claim.PartBusinessName, func(a *address.Address) string { return a.BusinessName }},
		{claim.PartArea, func(a *address.Address) string { return a.Area }},
		{claim.PartPrimaryNumber, func(a *address.Address) string { return a.PrimaryNumber }},
		{claim.PartPredirectional, func(a *address.Address) string { return a.Predirectional }},
		{claim.PartStreetName, func(a *address.Address) string { return a.StreetName }},
		{claim.PartStreetSuffix, func(a *address.Address) string { return a.StreetSuffix }},
		{claim.PartPostdirectional, func(a *address.Address) string { return a.Postdirectional }},
		{claim.PartSecondaryDesignator, func(a *address.Address) string { return a.SecondaryDesignator }},
		{claim.PartSecondaryNumber, func(a *address.Address) string { return a.SecondaryNumber }},
		{claim.PartDetail, func(a *address.Address) string { return a.Detail }},
		{claim.PartCity, func(a *address.Address) string { return a.City }},
		{claim.PartRegion, func(a *address.Address) string { return a.Region }},
		{claim.PartPostal, func(a *address.Address) string { return a.Postal }},
		{claim.PartCountry, func(a *address.Address) string { return a.Country }},
	}

	for _, tc := range cases {
		t.Run(string(tc.part), func(t *testing.T) {
			line := lastline.LineClaim{
				Span:  lastline.Span{Start: 1, Length: 0},
				Claim: claim.Claim{Confidence: claim.ConfidenceExact},
			}
			street := claim.Claim{
				Confidence: claim.ConfidenceExact,
				Parts: []claim.ClaimPart{
					{Start: 0, Length: 1, Part: tc.part, Value: "ASSIGNED"},
				},
			}

			got := line.Candidate(nil, 1, []claim.Claim{street})
			if value := tc.field(got.Address); value != "ASSIGNED" {
				t.Errorf("a %s part left its field %q; assign has no case for it", tc.part, value)
			}
		})
	}
}
