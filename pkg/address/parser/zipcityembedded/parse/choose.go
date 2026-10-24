package parse

import (
	"sort"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
)

// lastLines is the readings of the last line the data leaves standing.
//
// lastline offers every reading the grammar supports, and where an address
// writes no comma that is most of the tokens ahead of the region: 123 NORTH
// PARK ST, PAUL, MN 55102 is offered with PAUL for its city, with ST PAUL,
// with PARK ST PAUL, and with no city at all. The grammar cannot narrow that
// and does not try. The data can, and this is the only place it can do it
// before the street question is asked — see #17 step 2, and go-projectusat#61
// for why asking afterwards is too late.
//
// So: when the data confirms any reading's city, the readings it does not
// confirm fall away, and when it confirms none they all stand and the grammar
// decides. That is the promotion rule from #17 — a window the data has seen
// beats one it has not, and where it has seen none nothing is discarded — and
// it is why a real city the Census missed still parses.
//
// Nothing about a surviving reading is rewritten. Confidence in particular is
// left exactly as lastline wrote it, so the comma in 123 MAIN ST WEST, PALM
// BEACH, FL still separates PALM BEACH from WEST PALM BEACH when the data has
// seen both: corroborating a reading the data likes would lift the unmarked
// one to the marked one's confidence and erase the only thing telling the pair
// apart. Measured; it is not a hypothetical.
func (p *Parser) lastLines(r reference, tokens []token.Token, claims []claim.Claim) []lastline.LineClaim {
	lines := lastline.LineClaims(tokens, claims)

	var confirmed []lastline.LineClaim
	for _, line := range lines {
		ans, ok := p.cityAgreement(r, linePart(line, claim.PartPostal), linePart(line, claim.PartCity), linePart(line, claim.PartRegion))
		if ok && ans == agrees {
			confirmed = append(confirmed, line)
		}
	}
	if len(confirmed) == 0 {
		return lines
	}
	return confirmed
}

// linePart reports what a last line reading says a part is, or "" when the
// reading does not claim that part at all — a {Region} {Postal Code} reading
// has no city, and that is a reading rather than a gap.
func linePart(line lastline.LineClaim, part claim.Part) string {
	for _, p := range line.Claim.Parts {
		if p.Part == part {
			return p.Value
		}
	}
	return ""
}

// cityAgreement asks the strongest city question the fields support: whether
// the ZIP Code and the city pair, or — when there is no ZIP Code to pair the
// city with — whether the state has that city at all. A city known for this
// ZIP Code is stronger evidence than a city known somewhere in the state, so
// the second is asked only when the first cannot be.
func (p *Parser) cityAgreement(r reference, postal, city, region string) (agreement, bool) {
	if ans, ok := p.zipCityAgreement(r, postal, city); ok {
		return ans, true
	}
	return p.cityZipsAgreement(r, postal, city, region)
}

// choose ranks candidates and returns the best, or nil when none survive.
//
// The score decides, coverage breaks ties — between two equally scored
// readings the one stranding fewer tokens is the better account of the input
// — and the street's own grammar breaks what is left, see streetConfidence.
// The score is this package's own integer — the grammar's rung (see rung)
// plus one for every question reference data agrees with and minus one for
// every question it contradicts, uncapped in either direction — see score.
// The data orders readings against each other and never rates one:
// CandidateAddress.Confidence is left exactly as the grammar wrote it, and
// Parse returns the Address alone, so no score this package computes is ever
// visible to a caller.
//
// A single agreement lifts a reading by one rung, which is enough to let it
// pass a reading one rung above it that the data contradicts — that
// inversion is the point of asking at all — but not enough to pass a reading
// two rungs above it that nothing was asked about. Two readings of one street
// line can differ only in which field holds a word — PENNSYLVANIA AVE as a
// name against PENNSYLVANIA with AVE as its suffix — and reference data
// cannot see that: both ask it the same key and get the same answer, so both
// scores shift by the same amount. The grammar's own gap between them
// survives the shift, and where the candidate's minimum had already closed
// that gap it is streetConfidence, below, that reopens it.
func (p *Parser) choose(r reference, candidates []*address.CandidateAddress) *address.CandidateAddress {
	type ranked struct {
		candidate *address.CandidateAddress
		answers   questionAnswers
		score     int
	}

	scored := make([]ranked, 0, len(candidates))
	for _, c := range candidates {
		if c == nil || c.Address == nil {
			continue
		}

		answers := p.agreement(r, c.Address)
		scored = append(scored, ranked{candidate: c, answers: answers})
	}
	if len(scored) == 0 {
		return nil
	}

	// A contradicts answer counts against a candidate only when some
	// competing candidate's answer to that same question is agrees —
	// go-projectusat#176. zipcity's gaps are real, so a question nothing in
	// the pile could confirm is exactly as uninformative as one never asked:
	// a Pub 28 placeholder street like KRYTON TN or a made-up one like
	// ZQXVBORK must not lose to a candidate that merely declined to ask,
	// which is what charging an unconfirmed contradicts always did before.
	// Gating is per question kind rather than per literal zipcity key, since
	// candidates in one pile are alternate readings of the same input and
	// rarely share a query string even when they are asking the same
	// question in spirit. Two candidates that both contradict the same
	// unconfirmed question (TestStreetContradictionDemotesButStillParses)
	// are ungated by the same amount, so the tie between them survives
	// exactly as it did when both were charged -1.
	var cityConfirmed, streetConfirmed bool
	for _, s := range scored {
		if s.answers.city == agrees {
			cityConfirmed = true
		}
		if s.answers.street == agrees {
			streetConfirmed = true
		}
	}
	for i := range scored {
		scored[i].score = score(scored[i].candidate.Confidence, scored[i].answers.gated(cityConfirmed, streetConfirmed))
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if a, b := len(scored[i].candidate.Leftover), len(scored[j].candidate.Leftover); a != b {
			return a < b
		}
		if a, b := streetConfidence(scored[i].candidate), streetConfidence(scored[j].candidate); a != b {
			return a > b
		}

		a := scored[i].candidate.Address
		b := scored[j].candidate.Address
		// A trailing directional is the postdirectional unless the street's
		// own data says it belongs to the name (go-projectusat#257): NE,
		// NORTHEAST, N EAST and N E after COUNTY ROAD are one directional,
		// not a lettered route. Where two readings are the same words and
		// differ only in whether that directional was split off, take the
		// split one; data has already had its say in the score above.
		if splitOff(a, b) {
			return true
		}
		if splitOff(b, a) {
			return false
		}
		// prefer the candidate that respects directionals
		idirectionals := unsplitDirectional(a.StreetName, false) + unsplitDirectional(a.StreetName, true)
		jdirectionals := unsplitDirectional(b.StreetName, false) + unsplitDirectional(b.StreetName, true)
		return idirectionals > jdirectionals
	})

	return scored[0].candidate
}

// splitOff reports whether split is the same street as whole with whole's
// trailing directional read as the postdirectional instead of as a word of the
// name.
func splitOff(split, whole *address.Address) bool {
	if split.Postdirectional == "" || whole.Postdirectional != "" {
		return false
	}
	tail, ok := strings.CutPrefix(whole.StreetName, split.StreetName+" ")
	if !ok {
		return false
	}
	want, err := directionals.NormalizeDirectional(split.Postdirectional)
	if err != nil {
		return false
	}
	dir, err := directionals.NormalizeDirectional(tail)
	return err == nil && dir == want
}

func unsplitDirectional(streetname string, post bool) int {
	parts := strings.Split(streetname, " ")
	lp := len(parts)
	if lp > 1 {
		i := 0
		if post {
			i = lp - 1
		}
		_, err := directionals.NormalizeDirectional(parts[i])
		if err == nil {
			if !post {
				// check for regions before penalizing the street
				_, err := region.Info(strings.Join(parts[0:2], " "), false)
				if err == nil {
					// the street starts with a region
					return 0
				}
			}
			return 1
		}
	}
	return 0
}

// streetConfidence is how strongly the street line alone is held: the minimum
// over the claims carrying a street field, or the candidate's own confidence
// where a reading has no street line to hold — a post office box is as
// strongly held as it already says it is, and this tie-break must not rate it
// again.
//
// It exists because CandidateAddress.Confidence is the minimum over every
// accepted claim, and the minimum is routinely decided by a claim the tied
// readings share. 123 OCEAN BOULEVARD WEST PALM BEACH FL is read both as
// OCEAN BOULEVARD and as OCEAN with the suffix BLVD, and ordinarystreet does
// rate the second a rung above the first — but the city on that unmarked last
// line is held one rung lower than either, so both candidates come out at the
// city's confidence and the grammar's own gap is invisible to the sort. The
// same minimum flattens 3253 W 9200 S WEST JORDAN UT into a street named
// W 9200 S.
//
// So this is not a second opinion about the address; it is the first one,
// read off the part the tied readings actually disagree about. It is the last
// key rather than the second, and both of the keys above it matter: reference
// data still decides, and coverage still decides, which is what keeps a
// reading that strands the whole street line from winning on the confidence
// it is left with by not reading one — measured, two of the standard's own
// fragments do exactly that when this key is asked first.
func streetConfidence(c *address.CandidateAddress) claim.Confidence {
	found := false
	lowest := claim.ConfidenceExact

	for _, held := range c.Claims {
		if !holdsStreet(held) {
			continue
		}
		if !found || held.Confidence < lowest {
			found, lowest = true, held.Confidence
		}
	}
	if !found {
		return c.Confidence
	}
	return lowest
}

// holdsStreet reports whether a claim carries any of the four fields a street
// name is made of — the same four streetForQuery renders, and for the same
// reason: they are the street, and the primary number and secondary unit are
// what sits on it.
func holdsStreet(held claim.Claim) bool {
	for _, part := range held.Parts {
		switch part.Part {
		case claim.PartPredirectional, claim.PartStreetName, claim.PartStreetSuffix, claim.PartPostdirectional:
			return true
		}
	}
	return false
}

// rung maps a grammar confidence onto this package's own 0..3 scale, in the
// order go-projectusat's claim package defines them: Weak, Likely, Strong,
// Exact.
func rung(c claim.Confidence) int {
	switch {
	case c >= claim.ConfidenceExact:
		return 3
	case c >= claim.ConfidenceStrong:
		return 2
	case c >= claim.ConfidenceLikely:
		return 1
	default:
		return 0
	}
}

// score is this package's ranking of one reading: the grammar's rung plus
// one for every agrees and minus one for every contradicts, uncapped in
// either direction. missingInZip, missingInCity, and unknown add nothing —
// see agreement's doc comment for why a split or a decline is not evidence.
//
// It is uncapped because there is nothing here left for a cap to protect.
// The old scale stepped claim.Confidence itself and stopped short of
// ConfidenceExact so that agreement could never manufacture certainty, but
// that cap also meant a reading rated Exact that the data contradicted
// (stepped down to Strong) tied a reading rated Strong that the data agreed
// with (capped at Strong) — and the tie-break on the grammar's own rating
// then picked the one the data had just spoken against. Scoring on this
// package's own integer instead of on Confidence removes the seam: nothing
// is capped, so the ordering the data is asked to make is the one it gets.
func score(c claim.Confidence, answers []agreement) int {
	s := rung(c)
	for _, ans := range answers {
		switch ans {
		case agrees:
			s++
		case contradicts:
			s--
		}
	}
	return s
}
