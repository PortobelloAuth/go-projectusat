package ordinarystreet

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// Candidates returns this package's readings of the address under the given
// last line.
//
// Unlike the pattern types, this one does not recognize anything. It assembles:
// the street line is read from the right by the vocabulary claims that end it —
// a secondary unit, a postdirectional, a suffix — and from the left by the
// primary number and a predirectional, and whatever is left in between is the
// street name. Every combination of those that leaves a non-empty name is a
// reading, and all of them are returned. Choosing between them is the parser's
// job, exactly as it is one level down.
//
// The street line is the last line of tokens ahead of the last line proper,
// except where that line is nothing but a secondary unit. A unit standing alone
// under the street is a delivery address written across two lines, not a street
// line of its own, and taking it for one discards the street: "123 MAIN ST /
// APT 4 / DENVER CO 80201" read a street named APT 4 and dropped 123 MAIN ST
// entirely. The span reaches back over the line above so the unit is read where
// it always is, at the end of the delivery address.
//
// A unit or a private mailbox on the line above the street is the other half
// of that shape, and is read the same way in the other direction: where the
// whole of that line is exactly one claim this package already admits at the
// end of its own line, a second reading accepts it too. Pub 28 §213.3 puts a
// secondary unit there when it does not fit on the street line, and §285's
// four-line CMRA form puts PMB 234 or #234 there. See aboveLineClaim.
//
// Anything else above the street line — a business name, an urbanization — is
// not this package's business and falls out as leftover, which costs the
// candidate a step of confidence. That is the honest rating: this reading
// really has not accounted for those tokens.
//
// Only the pool is consulted, never a recognizer, because there is nothing to
// recognize. That is also why this package cannot be confused with another the
// way #52 confused a rural route with a military address: it makes no claim
// that a look-alike could imitate.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	end := line.Span.Start
	if end <= 0 || end > len(tokens) {
		return nil
	}

	start := lineStart(tokens, end-1)
	if start > 0 && isSecondaryUnitLine(claims, start, end) {
		start = lineStart(tokens, start-1)
	}

	// The street begins at the house number, and anything ahead of it is not
	// street. aboveLineClaim keeps the line start: it looks at the line above.
	unit := businessBoundary(tokens, claims, start, end)

	var business []claim.Claim
	if unit > start {
		business = append(business, businessClaim(tokens, start, unit))
	}

	// A unit written ahead of the house number is read where it always is, at
	// the end of the delivery address. See prefixUnitEnd.
	street := prefixUnitEnd(tokens, claims, unit, end)

	placed := placements(tokens, claims, street, end)

	var candidates []*address.CandidateAddress
	for _, t := range tails(claims, street, end) {
		for _, h := range heads(tokens, claims, street, t.nameEnd) {
			for _, name := range nameReadings(tokens, claims, h.nameStart, t.nameEnd) {
				accepted := make([]claim.Claim, 0, len(business)+len(h.claims)+len(t.claims)+2)
				accepted = append(accepted, business...)
				accepted = append(accepted, h.claims...)
				accepted = append(accepted, t.claims...)
				accepted = append(accepted, streetClaim(claims, placed, h, t, name))

				if street > unit {
					accepted = withPrefixUnit(claims, accepted, unit, street)
				}

				candidates = append(candidates,
					line.Candidate(&OrdinaryStreetAddress{}, len(tokens), accepted))

				if above, ok := aboveLineClaim(tokens, claims, start,
					hasPart(accepted, claim.PartSecondaryDesignator), hasPart(accepted, claim.PartDetail)); ok {
					candidates = append(candidates,
						line.Candidate(&OrdinaryStreetAddress{}, len(tokens), append(append([]claim.Claim{}, accepted...), above)))
				}
			}
		}
	}

	return candidates
}

// aboveLineClaim returns the claim this package accepts from the line
// immediately above the street line, if the pool offers one.
//
// It is isSecondaryUnitLine's rule applied in the other direction: only a
// line covered by exactly one such claim qualifies, and the same two
// elements this package already admits at the end of its own line — a
// secondary unit or a private mailbox — are the only ones offered here.
//
// unitPlaced and detailPlaced describe what this reading of the street line
// itself has already accepted. A second secondary unit is never offered
// beside one already placed — the standard has one designator per address —
// and admitMailbox's ruling on # carries over unchanged: #78 is about what a
// bare # means, not about which line it sits on, so a placed unit still
// turns a # above the street into the patient's mailbox at Strong, and PMB is
// still taken wherever it stands. A second Detail is never offered beside one
// the street line's own tail already supplied, for the same reason as the
// second secondary unit.
func aboveLineClaim(tokens []token.Token, claims []claim.Claim, start int, unitPlaced, detailPlaced bool) (claim.Claim, bool) {
	if start <= 0 {
		return claim.Claim{}, false
	}

	return spanClaim(claims, lineStart(tokens, start-1), start, unitPlaced, detailPlaced)
}

// spanClaim returns the secondary unit or private mailbox claim this package
// accepts for exactly the tokens from start to end, if the pool offers one. It
// is the rule aboveLineClaim describes, and it is the same rule for a unit
// written ahead of the house number on the street line itself: which line, or
// which end of the line, the unit is written on does not change what it is.
func spanClaim(claims []claim.Claim, start, end int, unitPlaced, detailPlaced bool) (claim.Claim, bool) {
	for _, c := range claims {
		if c.Start() != start || c.End() != end {
			continue
		}

		if !detailPlaced && assigns(c, claim.PartDetail) {
			if admitted, ok := admitMailbox(c, unitPlaced); ok {
				return admitted, true
			}
			continue
		}

		if !unitPlaced && assigns(c, claim.PartSecondaryDesignator) {
			return c, true
		}
	}

	return claim.Claim{}, false
}

// hasPart reports whether any accepted claim assigns the named part.
func hasPart(claims []claim.Claim, part claim.Part) bool {
	for _, c := range claims {
		if assigns(c, part) {
			return true
		}
	}

	return false
}

// isSecondaryUnitLine reports whether a claim covers the span exactly and reads
// it as a secondary unit. Exactly is the point: a line holding a unit and
// something else is a street line that happens to carry a unit, which the tail
// readings already handle.
func isSecondaryUnitLine(claims []claim.Claim, start, end int) bool {
	for _, c := range claims {
		if c.Start() == start && c.End() == end && assigns(c, claim.PartSecondaryDesignator) {
			return true
		}
	}

	return false
}

// houseNumber is a bare run of digits, the shape of a house number written
// alone. It is deliberately narrower than primaryNumbers' digit test: "42ND" is
// an ordinal in a street name, as is "12TH.". "100", "33-55", "A17", "A17B", and
// "N6W23001" are all house numbers documented in the specification. We are not
// supporting primary address numbers synthisized from building numbers here; that
// logic is specific to Puerto Rico in the standard.
var houseNumber = regexp.MustCompile(`^([0-9]+|[A-Z]?\d+[A-Z]?|[NS]\d+[EW]\d+|\d+-\d+)$`)

// businessBoundary returns where the street begins when a business name opens
// the line, and from otherwise.
//
// A single line address has no line break to say where the business name ends.
// The street begins at the last house number that opens a complete street line
// and is not itself part of the street line that an earlier number opens:
//
//   - "CENTER OF HOPE 110 EAST 7TH STREET" reads its business as CENTER OF HOPE.
//   - "1ST STREET PIZZA COMPANY 511 MAIN ST" reads it as 1ST STREET PIZZA
//     COMPANY, because 1ST is an ordinal and not a house number.
//   - "3M CORPORATION 100 MAIN ST" reads it as 3M CORPORATION. 3M is house number
//     shaped, but a line can open on a business name that starts with one, so
//     the first token is only a candidate like any other.
//   - "ACME 3M CORPORATION 100 MAIN ST" keeps its name whole, because the last
//     qualifying number is taken.
//
// A later number is not a new street when it belongs to the street line already
// found: the 500 of a grid address 100 N 500 E (opensHead), the 12 of COUNTY
// ROAD 12 (insideClaim), the unit identifier of 100 MAIN ST 4 B
// (closesStreetLine), and the 11 of 10 MAIN ST STE 11 (unitAfterStreetLine).
//
// It stays put where the line opens with a unit or route designator, and a
// unit written directly ahead of the house number is not business either
// (unitBeforeStreetLine): "Apartment 3200 152 South Tech Dr" and "#3200 152
// South Tech Dr" have no business name. Candidates reads that unit; see
// prefixUnitEnd.
func businessBoundary(tokens []token.Token, claims []claim.Claim, from, end int) int {
	if from >= end || isUnitOrRoute(tokens[from].Text) {
		return from
	}

	street := from
	for i := from + 1; i < end-1; i++ {
		if !houseNumber.MatchString(tokens[i].Text) || routeNumber(tokens, i, end) || insideClaim(claims, i) {
			continue
		}
		if opensHead(tokens, claims, street, i) || closesStreetLine(tokens, claims, street, i) {
			continue
		}
		if unitAfterStreetLine(tokens, claims, street, i) {
			continue
		}
		if !readsStreetLine(tokens, claims, i, end) {
			continue
		}

		street = i
	}

	return unitBeforeStreetLine(claims, from, street)
}

// insideClaim reports whether the token at i lies inside a street name some
// vocabulary claims from ahead of it: the 12 of COUNTY ROAD 12, the 30 of
// HIGHWAY 30. A number already read as part of a street's name is not where a
// street begins.
//
// Only a street name counts. A secondary unit claim is not enough, because
// BUILDING is both a business word and a unit designator: the 847 of UCENT
// BUILDING 847 NORTH 49TH STREET is claimed as BLDG 847 and is still the house
// number.
func insideClaim(claims []claim.Claim, i int) bool {
	for _, c := range claims {
		if c.Start() < i && c.End() > i && assigns(c, claim.PartStreetName) {
			return true
		}
	}

	return false
}

// unitAfterStreetLine reports whether the number at i is the identifier of a
// unit claimed from ahead of it, and the street line from start already closes
// where that unit begins: the 420 of 450 JANE STANFORD WAY BUILDING 420, the 11
// of 10 MAIN ST STE 11. A unit word that does not follow a closed street line,
// the BUILDING of UCENT BUILDING 847 NORTH 49TH STREET, is read as part of a
// business name instead.
func unitAfterStreetLine(tokens []token.Token, claims []claim.Claim, start, i int) bool {
	for _, c := range claims {
		if c.Start() < i && c.End() > i && c.Start() > start && closesStreetLine(tokens, claims, start, c.Start()) {
			return true
		}
	}

	return false
}

// unitBeforeStreetLine returns where a secondary unit or private mailbox
// written directly ahead of the house number at street begins, or street where
// there is none. The unit is not part of the business name ahead of it: ACME
// SUITE 3200 152 S TECH DR names ACME, and SUITE 3200 is its unit.
//
// Only a claim that ends exactly at the house number counts. A unit word
// further back, with business words between it and the street, is read as part
// of the business name, as BUILDING is in UCENT BUILDING 847 N 49TH ST.
func unitBeforeStreetLine(claims []claim.Claim, from, street int) int {
	first := street
	for _, c := range claims {
		if c.End() == street && c.Start() >= from && c.Start() < first && isUnitClaim(c) {
			first = c.Start()
		}
	}

	return first
}

// prefixUnitEnd returns where the street line begins when it opens at from
// with a secondary unit or private mailbox written ahead of the house number,
// and from otherwise.
//
// "#3200 152 S TECH DR" and "SUITE 3200 152 S TECH DR" put the unit first. It
// is still the unit, rendered where the standard puts it, after the street:
// 152 S TECH DR STE 3200. Taking the line from the unit read SUITE 3200 152
// SOUTH TECH as a street name, and #3200 as a house number.
//
// The unit only opens the line where a house number follows it and the rest
// reads as a street line, the same test businessBoundary puts to a number.
// Where more than one claim qualifies, the longest is taken: STE 3200 and not
// STE with 3200 as the house number.
func prefixUnitEnd(tokens []token.Token, claims []claim.Claim, from, end int) int {
	street := from
	for _, c := range claims {
		at := c.End()
		if c.Start() != from || at <= street || at >= end || !isUnitClaim(c) {
			continue
		}
		if !houseNumber.MatchString(tokens[at].Text) || !readsStreetLine(tokens, claims, at, end) {
			continue
		}

		street = at
	}

	return street
}

// withPrefixUnit returns a reading's accepted claims with the unit written
// ahead of its house number, from start to end, accounted for.
//
// Where the reading's own tail placed no unit, the prefix is taken by
// spanClaim's rule, so #3200 is the unit # 3200 and PMB 456 the mailbox.
//
// Where the tail did place one, the standard still has one designator slot,
// and the prefix leads the chain into it, exactly as the leftmost designator
// leads a chain at the end of the line (joinSecondary): UNIT 3200 152 TECH DR
// ROOM 12 is 152 TECH DR UNIT 3200 RM 12, and UNIT 3200 152 TECH DR UPPER is
// 152 TECH DR UNIT 3200 UPPR. A prefix that cannot chain — # beside a placed
// unit is the mailbox (#78) — falls back to spanClaim, and one spanClaim will
// not take either is left over and charged for.
func withPrefixUnit(claims, accepted []claim.Claim, start, end int) []claim.Claim {
	for i, placed := range accepted {
		if !assigns(placed, claim.PartSecondaryDesignator) {
			continue
		}

		for _, prefix := range claims {
			if prefix.Start() != start || prefix.End() != end {
				continue
			}
			if merged, ok := joinPrefixUnit(prefix, placed); ok {
				out := append([]claim.Claim{}, accepted...)
				out[i] = merged

				return out
			}
		}
	}

	if prefix, ok := spanClaim(claims, start, end,
		hasPart(accepted, claim.PartSecondaryDesignator), hasPart(accepted, claim.PartDetail)); ok {
		return append(accepted, prefix)
	}

	return accepted
}

// joinPrefixUnit chains a numbered unit written ahead of the house number into
// the unit the tail placed after the street, the prefix leading.
//
// The two are not adjacent, and a claim part is one run of tokens, so the
// chained number is carried on both runs with the same value: the prefix's
// number and the whole of the tail's unit each read as SecondaryNumber 3200 RM
// 12. Every token is accounted for, and whichever part is assigned last, the
// field reads the same.
//
// Unlike joinSecondary the tail's unit may be unnumbered, UPPER as much as RM
// 12, because the prefix already supplies the number the chain hangs off. #
// never chains, for the reason joinSecondary gives.
func joinPrefixUnit(prefix, placed claim.Claim) (claim.Claim, bool) {
	ld, ln, ok := secondaryParts(prefix)
	if !ok || len(prefix.Parts) != 2 || ld.Value == "#" {
		return claim.Claim{}, false
	}

	var rd, rn claim.ClaimPart
	for _, p := range placed.Parts {
		switch p.Part {
		case claim.PartSecondaryDesignator:
			rd = p
		case claim.PartSecondaryNumber:
			rn = p
		default:
			return claim.Claim{}, false
		}
	}
	if rd.Value == "#" {
		return claim.Claim{}, false
	}

	value := strings.TrimSpace(ln.Value + " " + rd.Value + " " + rn.Value)
	number := ln
	number.Value = value
	tail := claim.ClaimPart{
		Start:  placed.Start(),
		Length: placed.End() - placed.Start(),
		Part:   claim.PartSecondaryNumber,
		Value:  value,
	}

	confidence := prefix.Confidence
	if placed.Confidence < confidence {
		confidence = placed.Confidence
	}

	return claim.Claim{Confidence: confidence, Parts: []claim.ClaimPart{ld, number, tail}}, true
}

// isUnitClaim reports whether a claim reads a secondary unit or a private
// mailbox.
func isUnitClaim(c claim.Claim) bool {
	return assigns(c, claim.PartSecondaryDesignator) || assigns(c, claim.PartDetail)
}

// opensHead reports whether the tokens from start to at are exactly the head
// of a street line — a house number, a fraction after it, and a
// predirectional — so that the number at at is the street name that follows
// them and not a second house number. 100 N 500 E is a grid address whose
// street is named 500, and 2300 W 8 MILE RD a street named 8 MILE.
func opensHead(tokens []token.Token, claims []claim.Claim, start, at int) bool {
	if !houseNumber.MatchString(tokens[start].Text) {
		return false
	}

	afterNumber := start + 1
	if afterNumber < at && fraction.MatchString(tokens[afterNumber].Text) {
		afterNumber++
	}
	if afterNumber == at {
		return true
	}

	for _, pre := range startingAt(claims, claim.PartPredirectional, afterNumber)[1:] {
		if pre.End() == at {
			return true
		}
	}

	return false
}

// closesStreetLine reports whether the tokens from start to at, opening on a
// house number, are already a whole street line closed by a suffix or a
// postdirectional. What follows them is the remainder of that line, an
// undesignated unit as in 100 MAIN ST 4 B, and not a business name ahead of a
// second street.
func closesStreetLine(tokens []token.Token, claims []claim.Claim, start, at int) bool {
	if !houseNumber.MatchString(tokens[start].Text) {
		return false
	}

	for _, t := range tails(claims, start, at) {
		if !hasPart(t.claims, claim.PartStreetSuffix) && !hasPart(t.claims, claim.PartPostdirectional) {
			continue
		}
		if len(heads(tokens, claims, start, t.nameEnd)) > 0 {
			return true
		}
	}

	return false
}

// readsStreetLine reports whether the tokens from start to end read as a
// street line: some tail and some head fit between them. It is the test a
// boundary candidate has to pass before the words ahead of it are a business.
func readsStreetLine(tokens []token.Token, claims []claim.Claim, start, end int) bool {
	for _, t := range tails(claims, start, end) {
		if len(heads(tokens, claims, start, t.nameEnd)) > 0 {
			return true
		}
	}

	return false
}

// routeNumber reports whether the number at i is a rural route's number, the
// 2 of "RR 2 BOX 18": a number followed by BOX is never where a street begins.
// RR alone is not a route, so isUnitOrRoute does not see it.
func routeNumber(tokens []token.Token, i, end int) bool {
	return i+1 < end && normalizeWord(tokens[i+1].Text) == "BOX"
}

// isUnitOrRoute reports whether the word opens a secondary unit or rural route.
func isUnitOrRoute(text string) bool {
	word := normalizeWord(text)
	if _, err := secondaryunit.Info(word); err == nil {
		return true
	}
	_, err := ruralroute.Normalize(word)

	return err == nil
}

// normalizeWord upper-cases a token and strips its punctuation for vocabulary lookup.
func normalizeWord(text string) string {
	return strings.ToUpper(textutil.StripPunctuation(text, textutil.StripOptions{KeepHyphen: false, KeepSlash: false}))
}

// businessClaim reads the words ahead of the street as the business name, so
// they are kept rather than dropped as leftover: an address that names its
// business should still say so.
//
// It is rated Strong, the same as the street it stands ahead of, because the
// house number is what confirms where the words end. A lower rating would be a
// confidence minimum that caps every reading at the business claim's level and
// leaves the best one tied with its own variants. Nothing confirms the name
// itself, which is the same ceiling streetConfidence gives a street name.
func businessClaim(tokens []token.Token, from, to int) claim.Claim {
	clean := textutil.StripPunctuation(token.Join(tokens[from:to]), textutil.StripOptions{
		KeepHyphen: false,
		KeepSlash:  false,
	})

	return claim.Claim{
		Confidence: claim.ConfidenceStrong,
		Parts: []claim.ClaimPart{{
			Start:  from,
			Length: to - from,
			Part:   claim.PartBusinessName,
			Value:  strings.ToUpper(clean),
		}},
	}
}

// lineStart returns the index of the first token on the same line as at.
func lineStart(tokens []token.Token, at int) int {
	start := at
	for start > 0 && tokens[start-1].Line == tokens[at].Line {
		start--
	}

	return start
}

// tail is one reading of the right hand end of the street line: the elements
// that follow the street name, and where the name therefore stops.
type tail struct {
	claims  []claim.Claim
	nameEnd int
}

// tails returns every reading of the elements that close the street line.
//
// They are taken in the order the standard puts them — suffix, then
// postdirectional, then secondary unit, then the private mailbox — and each is
// optional, so the readings range from a bare name to all four present. A
// missing element is not an error and not a lower rating here; whether its
// absence matters is settled by what the name then has to absorb. See
// streetConfidence.
func tails(claims []claim.Claim, from, to int) []tail {
	var out []tail

	for _, detail := range endingAt(claims, claim.PartDetail, to, from) {
		afterUnit := to
		if detail != nil {
			afterUnit = detail.Start()
		}

		for _, secondary := range secondaryReadings(claims, afterUnit, from) {
			afterName := afterUnit
			var accepted []claim.Claim
			if detail != nil {
				mailbox, ok := admitMailbox(*detail, secondary != nil)
				if !ok {
					continue
				}
				accepted = append(accepted, mailbox)
			}
			if secondary != nil {
				afterName = secondary.Start()
				accepted = append(accepted, *secondary)
			}

			for _, post := range endingAt(claims, claim.PartPostdirectional, afterName, from) {
				afterSuffix := afterName
				withPost := accepted
				if post != nil {
					afterSuffix = post.Start()
					withPost = append(append([]claim.Claim{}, accepted...), *post)
				}

				for _, suffix := range endingAt(claims, claim.PartStreetSuffix, afterSuffix, from) {
					nameEnd := afterSuffix
					withSuffix := withPost
					if suffix != nil {
						nameEnd = suffix.Start()
						withSuffix = append(append([]claim.Claim{}, withPost...), *suffix)
					}

					if nameEnd <= from {
						continue
					}

					out = append(out, tail{claims: withSuffix, nameEnd: nameEnd})
				}
			}
		}
	}

	return out
}

// secondaryReadings returns every run of numbered secondary designators that
// ends at end and begins no earlier than floor, preceded by a nil standing for
// the reading with no secondary unit. A run of more than one designator, as in
// BUILDING 420 ROOM 120, is merged into a single claim so that it is accepted
// or rejected as one reading.
func secondaryReadings(claims []claim.Claim, end, floor int) []*claim.Claim {
	found := []*claim.Claim{nil}

	for _, right := range endingAt(claims, claim.PartSecondaryDesignator, end, floor)[1:] {
		found = append(found, right)

		for _, left := range secondaryReadings(claims, right.Start(), floor)[1:] {
			if merged, ok := joinSecondary(*left, *right); ok {
				found = append(found, &merged)
			}
		}
	}

	return found
}

// joinSecondary merges the secondary designator left, which ends where right
// begins, into one claim. The highest level designator is the leftmost one, so
// it stays the SecondaryDesignator, and everything after it is accumulated, in
// order, into the SecondaryNumber: BLDG 420 then RM 120 becomes BLDG with the
// number 420 RM 120. An unnumbered designator such as BSMT may lead the chain,
// so BSMT STE 480 becomes BSMT with the number STE 480 rather than dropping
// BSMT (#188); the claim it leads into must be numbered.
//
// # is never joined. Beside a placed unit it is the mailbox (#78, admitMailbox),
// and a chain through it would offer STE 11 # 234 as a rival to STE 11 PMB 234.
func joinSecondary(left, right claim.Claim) (claim.Claim, bool) {
	ld, ln, lok := secondaryParts(left)
	rd, rn, rok := secondaryParts(right)
	if !rok || ld.Value == "#" || rd.Value == "#" {
		return claim.Claim{}, false
	}

	var number claim.ClaimPart
	switch {
	case lok:
		number = ln
		number.Length = rn.End() - ln.Start
		number.Value = ln.Value + " " + rd.Value + " " + rn.Value
	case len(left.Parts) == 1 && ld.Part == claim.PartSecondaryDesignator:
		number = rn
		number.Start = rd.Start
		number.Length = rn.End() - rd.Start
		number.Value = rd.Value + " " + rn.Value
	default:
		return claim.Claim{}, false
	}

	confidence := left.Confidence
	if right.Confidence < confidence {
		confidence = right.Confidence
	}

	return claim.Claim{
		Confidence: confidence,
		Parts:      []claim.ClaimPart{ld, number},
	}, true
}

// secondaryParts returns the designator and number parts of a numbered
// secondary claim, and whether the claim has both.
func secondaryParts(c claim.Claim) (claim.ClaimPart, claim.ClaimPart, bool) {
	var designator, number claim.ClaimPart
	var hasDesignator, hasNumber bool

	for _, p := range c.Parts {
		switch p.Part {
		case claim.PartSecondaryDesignator:
			designator, hasDesignator = p, true
		case claim.PartSecondaryNumber:
			number, hasNumber = p, true
		}
	}

	return designator, number, hasDesignator && hasNumber
}

// admitMailbox returns the private mailbox reading this package accepts from a
// Detail claim, if it accepts one.
//
// privatemailbox holds PMB 234 at Exact and # 234 below it, because # is also
// the secondary unit designator of unspecified type, which secondaryunit
// claims at Exact (Publication 28 §213.2). With no unit placed elsewhere on
// the line a # is that unit, and no mailbox reading is offered: the unit
// reading wins and the address renders as # 234, which is deliverable either
// way (#78). Beside a placed unit the # is the reading left that explains the
// tokens — the standard forbids combining the CMRA's secondary element with
// the patient's mailbox, so a second unit is not a reading at all — and the
// mailbox is taken at Strong. The vocabulary rates what the tokens could be;
// this package rates what they are on this line, which is the same split as
// streetConfidence.
func admitMailbox(detail claim.Claim, unitPlaced bool) (claim.Claim, bool) {
	if detail.Confidence == claim.ConfidenceExact {
		return detail, true
	}

	if !unitPlaced {
		return claim.Claim{}, false
	}

	return claim.Claim{Confidence: claim.ConfidenceStrong, Parts: detail.Parts}, true
}

// head is one reading of the left hand end of the street line: the primary
// number and predirectional, and where the street name therefore begins.
type head struct {
	claims    []claim.Claim
	number    *claim.ClaimPart
	nameStart int
}

// heads returns every reading of the elements that open the street line.
func heads(tokens []token.Token, claims []claim.Claim, from, nameEnd int) []head {
	var out []head

	for _, number := range primaryNumbers(tokens, from, nameEnd) {
		afterNumber := from
		if number != nil {
			afterNumber = number.End()
		}

		for _, pre := range startingAt(claims, claim.PartPredirectional, afterNumber) {
			nameStart := afterNumber
			var accepted []claim.Claim
			if pre != nil {
				nameStart = pre.End()
				accepted = []claim.Claim{*pre}
			}

			if nameStart >= nameEnd {
				continue
			}

			out = append(out, head{claims: accepted, number: number, nameStart: nameStart})
		}
	}

	return out
}

// placements returns every claim some reading of the line places in a slot,
// which is the only sense in which a reading can have declined to place one.
//
// heads offers a predirectional only where it opens the name, and tails
// offers the closing elements only in the order the standard puts them, each
// ending where the next begins. A claim buried anywhere else was never
// offered a slot, so no reading left it unfilled: "1250 AVENUE OF THE
// AMERICAS" has a suffix claim on AVENUE, and no reading in which it is the
// suffix. See streetConfidence.
func placements(tokens []token.Token, claims []claim.Claim, from, to int) []claim.Claim {
	var out []claim.Claim

	for _, t := range tails(claims, from, to) {
		out = append(out, t.claims...)
		for _, h := range heads(tokens, claims, from, t.nameEnd) {
			out = append(out, h.claims...)
		}
	}

	return out
}

// fraction is a primary number written as a fraction of the one before it, the
// "1/2" of "123 1/2 MAIN ST".
var fraction = regexp.MustCompile(`^[0-9]+/[0-9]+$`)

// primaryNumbers returns the readings of the primary address number at the
// start of the line, or a single nil reading where there is no number.
//
// A number is any leading token carrying a digit, which is what makes the
// alphanumeric grid forms work: N6W23001 is a primary number by the same rule
// as 123, and neither needs a table. A fraction directly after one extends it
// rather than competing with it: Pub 28 puts a fractional address in the
// primary number, so "123 1/2 MAIN ST" is number "123 1/2" and not a street
// named "1/2 MAIN".
//
// Where a number-shaped token opens the line, no numberless reading is offered.
// A leading digit-bearing token with a street name after it is a house number;
// reading it as the first word of the name is a reading nothing in the library
// supports, and offering it would double the output of this package for every
// ordinary address to no purpose.
func primaryNumbers(tokens []token.Token, from, limit int) []*claim.ClaimPart {
	if from >= limit || !strings.ContainsFunc(tokens[from].Text, unicode.IsDigit) {
		return []*claim.ClaimPart{nil}
	}

	length := 1
	if from+1 < limit && fraction.MatchString(tokens[from+1].Text) {
		length = 2
	}

	return []*claim.ClaimPart{{
		Start:  from,
		Length: length,
		Part:   claim.PartPrimaryNumber,
		Value:  strings.ToUpper(token.Join(tokens[from : from+length])),
	}}
}

// nameReading is one reading of the residue as a street name, and whether a
// vocabulary claimed exactly those tokens as one.
type nameReading struct {
	part         claim.ClaimPart
	corroborated bool
}

// nameReadings returns the readings of the residue between the head and the
// tail as a street name.
//
// Ordinarily there is one, and its value is the tokens themselves: the name is
// arbitrary text and there is nothing to normalize it against.
//
// Where a vocabulary has claimed exactly those tokens as a street name, its
// spelling is taken instead. That vocabulary knows the form the standard wants
// and this package does not — highways rewrites through NormalizeStreetName —
// so taking the tokens verbatim over a claim that covers them would emit a name
// the library already knows how to normalize. This is the mechanism that makes
// the demotion of highways to a vocabulary work end to end: highways names the
// street, and this type builds the address around it. See #56.
//
// Only the spelling is taken, never the confidence. What a vocabulary is sure
// or unsure of is its own reading of those tokens standing alone, and that is a
// different question from how well this street line hangs together. region
// offers every state name as a possible street name at ConfidenceLikely,
// because on its own that is all it can say; inheriting it would rate
// "1600 PENNSYLVANIA AVE NW" below "1600 MAIN AVE NW" for no reason but the
// name. A vocabulary agreeing with this reading does not weaken it.
//
// A claim that covers only part of the residue is left alone. "OLD STATE ROUTE
// 9" contains a highway name and is not one, and nothing here can tell which
// reading the caller meant.
func nameReadings(tokens []token.Token, claims []claim.Claim, from, to int) []nameReading {
	seen := map[string]bool{}
	var corroborated []nameReading

	for _, c := range claims {
		if c.Start() != from || c.End() != to {
			continue
		}

		for _, p := range c.Parts {
			if p.Part != claim.PartStreetName || seen[p.Value] {
				continue
			}

			seen[p.Value] = true
			corroborated = append(corroborated, nameReading{
				part: claim.ClaimPart{
					Start:  from,
					Length: to - from,
					Part:   claim.PartStreetName,
					Value:  p.Value,
				},
				corroborated: true,
			})
		}
	}

	if len(corroborated) > 0 {
		return corroborated
	}

	// Make sure that we don't leave punctuation that doesn't belong in the street name
	clean := textutil.StripPunctuation(token.Join(tokens[from:to]), textutil.StripOptions{
		KeepHyphen: false,
		KeepSlash:  false,
	})

	return []nameReading{{
		part: claim.ClaimPart{
			Start:  from,
			Length: to - from,
			Part:   claim.PartStreetName,
			Value:  strings.ToUpper(clean),
		},
	}}
}

// streetClaim assembles the primary number and street name into the one claim
// this package makes.
//
// They are one claim rather than two because neither survives without the
// other. A bare 123 is a house number only because a street name follows it on
// the same line, and the name is bounded on the left only because the number
// opened the line. A parser that took one and rejected the other would hold a
// reading this package never offered, which is the same reason a rural route
// claims its route and box together.
func streetClaim(claims, placed []claim.Claim, h head, t tail, name nameReading) claim.Claim {
	parts := make([]claim.ClaimPart, 0, 2)
	if h.number != nil {
		parts = append(parts, *h.number)
	}
	parts = append(parts, name.part)

	return claim.Claim{Confidence: streetConfidence(claims, placed, h, t, name), Parts: parts}
}

// streetConfidence rates the number and name reading.
//
// A number opening the line is the shape the standard describes, and the
// reading is held strongly. Without one the reading is contested by
// construction: a run of words with nothing in front of it is a street name
// here and could be a city, a business name, or the tail of the line above.
//
// Neither is ever exact, because nothing confirms a street name. Confirming it
// takes a data source this library does not have — see #61 and the zipcity
// work — and until then the honest ceiling is a reading the parser can
// legitimately overrule.
//
// A name that swallows tokens another vocabulary has claimed drops one further
// step, where the element that claim reads is one this reading left unfilled.
// That is what separates "123 MAIN ST" read with its suffix from the same
// tokens read as a name of "MAIN ST": both are offered, and the one that
// explains the suffix is the better account of the line. The demotion is one
// step whatever the name absorbed, because this package cannot tell which
// absorbed claim was the one that mattered.
//
// A slot the reading did fill is not charged again for a claim buried in the
// name, because there was no reading in which that claim went there. PARK is a
// Pub 28 suffix, so "123 W FOX PARK DR" read as W, FOX PARK, DR has a suffix
// claim inside its name — but its suffix is DR, and the alternative that puts
// PARK in the suffix would strand DR, which this package never offers. Charging
// it left every reading of that address demoted and all four of them tied.
// Nor is a slot charged for a claim no reading could have put there, which is
// why only placed claims are consulted: "1250 AVENUE OF THE AMERICAS" absorbs
// a suffix claim, but a suffix closes the street line and AVENUE opens it, so
// no reading ever offered it the suffix slot, and the standard would not have
// it rendered there in any case. See placements.
//
// A corroborated name is exempt. The demotion is a guess that the name swallowed
// a component it should have left outside, and a vocabulary claiming exactly
// these tokens as a street name has already answered that question with
// knowledge this package does not have. "123 STATE ROUTE 9" is the case:
// ROUTE is a Pub 28 suffix, so the name absorbs one, and highways nonetheless
// knows the whole run is the name of the street.
//
// The same knowledge cuts the other way. A reading that takes its suffix off
// the end of a run a vocabulary claims whole as a street name has split a name
// the library knows, and drops the same one step. COUNTY ROAD is a highway
// used as a street name, so COUNTY with the suffix RD is the weaker reading of
// it (p.17, #155). See splitsName.
func streetConfidence(claims, placed []claim.Claim, h head, t tail, name nameReading) claim.Confidence {
	confidence := claim.ConfidenceLikely
	if h.number != nil {
		confidence = claim.ConfidenceStrong
	}

	swallows := !name.corroborated && !isDirectional(placed, name.part) &&
		absorbs(placed, unplaced(h, t), name.part.Start, name.part.End())
	if !swallows && !splitsName(claims, t, name) && !splitsCompound(claims, h, name) {
		return confidence
	}

	if confidence == claim.ConfidenceStrong {
		return claim.ConfidenceLikely
	}

	return claim.ConfidenceWeak
}

// splitsCompound reports whether the reading's predirectional is one half of a
// compound directional that some vocabulary claims as one span. N EAST MAIN ST
// reads N as the predirectional and leaves EAST to open the name, where the
// standard's compound NE is the reading the words were written for (#186). The
// split is charged one step, as splitsName charges a split street name.
//
// A one-token name is exempt. N E ST is N followed by the alphabet street E
// (p.17), and the residue there is the second half of the pair by design.
func splitsCompound(claims []claim.Claim, h head, name nameReading) bool {
	if name.part.Length == 1 {
		return false
	}

	for _, pre := range h.claims {
		for _, c := range claims {
			if c.Start() == pre.Start() && c.End() > pre.End() && assigns(c, claim.PartPredirectional) {
				return true
			}
		}
	}

	return false
}

// splitsName reports whether the reading's suffix closes a run that some
// vocabulary claims, from where this name begins, as one street name.
//
// A corroborated name is never charged: it is that vocabulary's own reading.
func splitsName(claims []claim.Claim, t tail, name nameReading) bool {
	if name.corroborated {
		return false
	}

	for _, suffix := range t.claims {
		if !assigns(suffix, claim.PartStreetSuffix) {
			continue
		}

		for _, c := range claims {
			if c.Start() == name.part.Start && c.End() == suffix.End() && assigns(c, claim.PartStreetName) {
				return true
			}
		}
	}

	return false
}

// unplaced returns the elements this package offers a place for that the
// reading has left empty. Those are the only ones a name can be said to have
// swallowed: an element the reading placed was not declined.
func unplaced(h head, t tail) []claim.Part {
	var open []claim.Part

	for _, part := range []claim.Part{
		claim.PartStreetSuffix,
		claim.PartPredirectional,
		claim.PartPostdirectional,
		claim.PartSecondaryDesignator,
		claim.PartDetail,
	} {
		filled := false
		for _, c := range h.claims {
			filled = filled || assigns(c, part)
		}
		for _, c := range t.claims {
			filled = filled || assigns(c, part)
		}

		if !filled {
			open = append(open, part)
		}
	}

	return open
}

// isDirectional reports whether the street name is nothing but a directional
// some reading of the line places as one.
//
// Such a name is not charged for it. The standard has directional street
// names — NORTH AVE and SOUTHEAST FWY N are its own examples, spelled out
// because the directional is the name — and the reading that takes the word
// as the predirectional instead is left with the suffix as its name, which
// is the reading that declined to place something. Charging both left them
// tied, and "123 NORTH AVENUE" read as a name of NORTH AVENUE, where the
// suffix is neither placed nor abbreviated.
//
// This only exempts a one-token name. N E ST is N followed by the alphabet
// street E (p.17), not the compound directional NORTH EAST — a name of two
// tokens ties a directional reading to that compound only because
// directionals also claims N E as one span, not because the name is nothing
// but a placed direction.
func isDirectional(placed []claim.Claim, name claim.ClaimPart) bool {
	if name.Length != 1 {
		return false
	}

	for _, c := range placed {
		if c.Start() == name.Start && c.End() == name.End() &&
			(assigns(c, claim.PartPredirectional) || assigns(c, claim.PartPostdirectional)) {
			return true
		}
	}

	return false
}

// absorbs reports whether the street name swallows tokens that some reading
// of the line places as one of the given elements.
//
// A street name claim is not one of those. A vocabulary naming the street is
// saying the same thing this reading says, and where it covers the residue
// exactly nameReadings has already adopted its spelling. Where it covers only
// part of the residue it is a longer or shorter name, not a component this
// reading declined to place.
func absorbs(placed []claim.Claim, placeable []claim.Part, from, to int) bool {
	for _, c := range placed {
		if c.Start() < from || c.End() > to {
			continue
		}

		for _, part := range placeable {
			if assigns(c, part) {
				return true
			}
		}
	}

	return false
}

// endingAt returns the claims assigning part whose extent ends at end and
// begins no earlier than floor, preceded by a nil standing for the reading in
// which that element is absent.
func endingAt(claims []claim.Claim, part claim.Part, end, floor int) []*claim.Claim {
	found := []*claim.Claim{nil}

	for i := range claims {
		if claims[i].End() == end && claims[i].Start() > floor && assigns(claims[i], part) {
			found = append(found, &claims[i])
		}
	}

	return found
}

// startingAt returns the claims assigning part whose extent begins at start,
// preceded by a nil standing for the reading in which that element is absent.
func startingAt(claims []claim.Claim, part claim.Part, start int) []*claim.Claim {
	found := []*claim.Claim{nil}

	for i := range claims {
		if claims[i].Start() == start && assigns(claims[i], part) {
			found = append(found, &claims[i])
		}
	}

	return found
}

// assigns reports whether a claim assigns the named part.
func assigns(c claim.Claim, part claim.Part) bool {
	for _, p := range c.Parts {
		if p.Part == part {
			return true
		}
	}

	return false
}
