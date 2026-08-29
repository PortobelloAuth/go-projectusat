package ordinarystreet

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
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
// The street line is the last line of tokens ahead of the last line proper.
// Anything above it — a business name, an urbanization — is not this package's
// business and falls out as leftover, which costs the candidate a step of
// confidence. That is the honest rating: this reading really has not accounted
// for those tokens.
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

	var candidates []*address.CandidateAddress
	for _, t := range tails(claims, start, end) {
		for _, h := range heads(tokens, claims, start, t.nameEnd) {
			for _, name := range nameReadings(tokens, claims, h.nameStart, t.nameEnd) {
				accepted := make([]claim.Claim, 0, len(h.claims)+len(t.claims)+1)
				accepted = append(accepted, h.claims...)
				accepted = append(accepted, t.claims...)
				accepted = append(accepted, streetClaim(claims, h, name))

				candidates = append(candidates,
					line.Candidate(&OrdinaryStreetAddress{}, len(tokens), accepted))
			}
		}
	}

	return candidates
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
// postdirectional, then secondary unit — and each is optional, so the readings
// range from a bare name to all three present. A missing element is not an
// error and not a lower rating here; whether its absence matters is settled by
// what the name then has to absorb. See streetConfidence.
func tails(claims []claim.Claim, from, to int) []tail {
	var out []tail

	for _, secondary := range endingAt(claims, claim.PartSecondaryDesignator, to, from) {
		afterName := to
		var accepted []claim.Claim
		if secondary != nil {
			afterName = secondary.Start()
			accepted = []claim.Claim{*secondary}
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

	return out
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

	return []nameReading{{
		part: claim.ClaimPart{
			Start:  from,
			Length: to - from,
			Part:   claim.PartStreetName,
			Value:  strings.ToUpper(token.Join(tokens[from:to])),
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
func streetClaim(claims []claim.Claim, h head, name nameReading) claim.Claim {
	parts := make([]claim.ClaimPart, 0, 2)
	if h.number != nil {
		parts = append(parts, *h.number)
	}
	parts = append(parts, name.part)

	return claim.Claim{Confidence: streetConfidence(claims, h, name), Parts: parts}
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
// step. That is what separates "123 MAIN ST" read with its suffix from the same
// tokens read as a name of "MAIN ST": both are offered, and the one that
// explains the suffix is the better account of the line. The demotion is one
// step whatever the name absorbed, because this package cannot tell which
// absorbed claim was the one that mattered.
//
// A corroborated name is exempt. The demotion is a guess that the name swallowed
// a component it should have left outside, and a vocabulary claiming exactly
// these tokens as a street name has already answered that question with
// knowledge this package does not have. "123 STATE ROUTE 9" is the case:
// ROUTE is a Pub 28 suffix, so the name absorbs one, and highways nonetheless
// knows the whole run is the name of the street.
func streetConfidence(claims []claim.Claim, h head, name nameReading) claim.Confidence {
	confidence := claim.ConfidenceLikely
	if h.number != nil {
		confidence = claim.ConfidenceStrong
	}

	if name.corroborated || !absorbs(claims, name.part.Start, name.part.End()) {
		return confidence
	}

	if confidence == claim.ConfidenceStrong {
		return claim.ConfidenceLikely
	}

	return claim.ConfidenceWeak
}

// absorbs reports whether the street name swallows tokens some vocabulary has
// claimed as an element this package offers a place for.
//
// A street name claim is not one of those. A vocabulary naming the street is
// saying the same thing this reading says, and where it covers the residue
// exactly nameReadings has already adopted its spelling. Where it covers only
// part of the residue it is a longer or shorter name, not a component this
// reading declined to place.
func absorbs(claims []claim.Claim, from, to int) bool {
	placeable := []claim.Part{
		claim.PartStreetSuffix,
		claim.PartPredirectional,
		claim.PartPostdirectional,
		claim.PartSecondaryDesignator,
	}

	for _, c := range claims {
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
