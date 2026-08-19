// Package lastline recognizes the Project US@ Last Line.
//
// The shapes it accepts on input, each optionally followed by a country:
//
//	{City}, {Region} {Postal Code}
//	{City}, {Region}
//	{Postal Code}
//	{City} {Postal Code}
//
// Where those come from matters, because it is not one citation. The first
// three are amadsen's statement of the pattern on #37; neither Project US@ nor
// Publication 28 writes them out as a list of accepted input forms. What both
// standards do state is the preferred form for *output*, which is
// {City} {Region} {Postal Code} with no comma, and Project US@ strips
// punctuation. So the standards constrain what this library should emit, and
// the input grammar here is a reading of the last line rather than a quotation
// of a rule.
//
// The fourth is neither: see confidenceFor. It is offered because a city in a
// reasonable position ahead of a postal code is a city, and it is rated below
// every other shape precisely because nothing but that argument supports it.
//
// The practical consequence of the punctuation point is that a comma is a
// bonus and not a premise. On normalized input there will not be one, and the
// complete pattern is then rated ConfidenceStrong rather than exact — it still
// matches, on a boundary this package inferred rather than one the address
// stated. Distinguishing an unmarked last line by other means, such as the end
// of a street suffix or secondary unit claim, is open work.
//
// This package is not a vocabulary and does not compete with one. It consumes
// the claims that region, postalcode and country already make and reports where
// a last line could be, which is a different kind of statement: a vocabulary
// says what tokens mean, and this says which reading of a line makes the most
// sense of them.
//
// It exists as a package rather than as parser code because it has to run
// before the rest of the sweep. streetsuffixes reads AVE as AVENUE and
// puertorico reads it as AVENIDA, and UsePRDialect chooses between them from
// the region and the postal code — the two things the last line produces. So
// the last line is a first pass over region and postal claims alone, and the
// vocabulary sweep is a second pass that runs knowing its answer.
//
// One thing it does do that a vocabulary cannot: claim the city. There is no
// city table in this library and Project US@ does not supply one. A city is
// identified by where it sits — the tokens ahead of the region on the last line
// — so the only thing in a position to claim it is whatever recognizes the
// line. Tokens that are not the city and have no other place to live are not
// claimed at all; they are reported as leftovers, and they lower the confidence
// of the reading that leaves them stranded.
package lastline

import (
	"slices"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// maxCitySpan bounds how far back an unmarked city may reach. Where a comma or
// a line break marks the start of the city there is no need to guess, and where
// nothing marks it this is how much of the line ahead of the region is worth
// offering as one. Four tokens covers the longest city names in ordinary use.
const maxCitySpan = 4

// Span is a run of tokens. It is claim.Span, aliased so that a LineClaim and a
// CandidateAddress name their leftovers with the same type rather than with two
// identical ones.
type Span = claim.Span

// LineClaim is one reading of the last line.
//
// It is deliberately more than a Claim. A Claim says these tokens mean this; a
// LineClaim says that under this reading of the line, these claims are the ones
// that survive and those other ones do not. The rejection is half of what makes
// the reading useful — deciding the last line of SAN JUAN PR 00926 is what
// rules out the state highway reading of PR 00926 and the street suffix reading
// of PR, and a bare []Claim has nowhere to record that it did.
type LineClaim struct {
	// Span is the extent of the line under this reading, including any tokens
	// the reading does not explain.
	Span Span

	// Claim is what the line says: the city, region, postal and country parts,
	// grouped because a last line stands or falls as one. Its confidence is the
	// confidence of the whole reading.
	Claim claim.Claim

	// Rejected are the input claims overlapping Span that this reading rules
	// out. Claims lying entirely outside Span are neither accepted nor rejected
	// — they belong to the street address line, which is not this package's
	// business.
	Rejected []claim.Claim

	// Leftover are runs of tokens inside Span that the reading does not
	// explain. Every leftover run lowers Claim.Confidence by one step: a
	// reading that strands tokens is a worse account of the line than one that
	// does not, and that is the whole of what leftovers are for. They are not a
	// bucket for a parser to assign from.
	Leftover []Span
}

// LineClaims returns every reading of the last line that the given claims
// support.
//
// The name is deliberately not Claims. A vocabulary's Claims says what tokens
// mean; this says which reading of a line makes the most sense of them, and the
// two are not interchangeable even where the signatures rhyme.
//
// Readings are not ranked against each other here beyond the confidence each
// carries; sorting them is claim.Claim.Compare's job, and choosing among them
// is the parser's. Nothing is returned when no reading is possible, which is
// the honest answer for an address with no recognizable last line.
//
// tokens is needed for structure rather than for vocabulary. Line, Position and
// FollowsComma decide where a line begins and whether a city boundary is marked
// or guessed at, and none of that is recoverable from the claims.
func LineClaims(tokens []token.Token, claims []claim.Claim) []LineClaim {
	if len(tokens) == 0 {
		return nil
	}

	end := len(tokens)

	var lines []LineClaim

	// Every pattern is offered with a country and without one, so the absent
	// country is a candidate in its own right.
	countries := append([]indexedClaim{{}}, byPartEndingAt(claims, claim.PartCountry, end)...)

	for _, country := range countries {
		innerEnd := end
		if country.ok {
			innerEnd = claims[country.index].Start()
		}
		if innerEnd == 0 {
			continue
		}

		// The line the pattern sits on is the one the country does not. A
		// country written on a line of its own makes the last line two
		// physical lines, and anchoring to the final line would look for the
		// city after the region rather than ahead of it.
		lineStart := startOfLineContaining(tokens, innerEnd-1)

		// {City}, {Region} {Postal Code}
		for _, postal := range byPartEndingAt(claims, claim.PartPostal, innerEnd) {
			postalStart := claims[postal.index].Start()

			for _, region := range byPartEndingAt(claims, claim.PartRegion, postalStart) {
				lines = append(lines, cityReadings(tokens, claims, patternCityRegionPostal,
					lineStart, claims[region.index].Start(), []indexedClaim{region, postal, country})...)
			}

			// {Postal Code}, with whatever precedes it on the line left over.
			lines = append(lines, assemble(claims, lineStart, end,
				confidenceFor(patternPostal, false), nil, []indexedClaim{postal, country}))

			// And the same tokens read as a city, which is not one of the
			// documented shapes. See confidenceFor.
			lines = append(lines, cityReadings(tokens, claims, patternCityPostal,
				lineStart, postalStart, []indexedClaim{postal, country})...)
		}

		// {City}, {Region}
		for _, region := range byPartEndingAt(claims, claim.PartRegion, innerEnd) {
			lines = append(lines, cityReadings(tokens, claims, patternCityRegion,
				lineStart, claims[region.index].Start(), []indexedClaim{region, country})...)
		}
	}

	return lines
}

// indexedClaim refers to a claim in the input slice. Readings are built from
// the caller's claims rather than from copies so that a rejected claim can be
// identified by which claim it is, not by whether it happens to compare equal
// to an accepted one.
type indexedClaim struct {
	index int
	ok    bool
}

// byPartEndingAt finds the single-part claims of the given part that finish
// exactly at end. Several may, and each one is a different reading of the line
// rather than a set to choose from, so all of them are returned.
func byPartEndingAt(claims []claim.Claim, part claim.Part, end int) []indexedClaim {
	var found []indexedClaim

	for i, c := range claims {
		if len(c.Parts) == 1 && c.Parts[0].Part == part && c.End() == end {
			found = append(found, indexedClaim{index: i, ok: true})
		}
	}

	return found
}

// cityReadings offers the city for a pattern whose region has been placed. The
// city is the run of tokens ending at cityEnd, and where it starts is the only
// open question, so one reading is returned per defensible start.
func cityReadings(tokens []token.Token, claims []claim.Claim, shape pattern, lineStart, cityEnd int, members []indexedClaim) []LineClaim {
	var lines []LineClaim

	for _, start := range cityStarts(tokens, lineStart, cityEnd) {
		parts := []claim.ClaimPart{{
			Start:  start,
			Length: cityEnd - start,
			Part:   claim.PartCity,
			Value:  textutil.Upper(token.Join(tokens[start:cityEnd])),
		}}

		lines = append(lines, assemble(claims, start, len(tokens),
			confidenceFor(shape, marked(tokens, lineStart, start)), parts, members))
	}

	return lines
}

// assemble builds the reading: the accepted parts in token order, the leftover
// runs, the confidence after the leftovers are charged for, and the input
// claims this reading rules out.
func assemble(claims []claim.Claim, start, end int, confidence claim.Confidence, parts []claim.ClaimPart, members []indexedClaim) LineClaim {
	accepted := map[int]bool{}

	for _, m := range members {
		if !m.ok {
			continue
		}
		accepted[m.index] = true
		parts = append(parts, claims[m.index].Parts...)
	}

	slices.SortFunc(parts, func(a, b claim.ClaimPart) int { return a.Start - b.Start })

	// The span runs to the end of the address. It begins where the reading
	// begins, which is normally the start of the final line but is earlier when
	// a country sits on a line of its own: the last line is then two physical
	// lines, and a span that excluded the region and postal ahead of the
	// country would be reporting an extent the reading does not have.
	if len(parts) > 0 {
		start = min(start, parts[0].Start)
	}
	span := Span{Start: start, Length: end - start}

	leftover := span.Gaps(parts)
	for range leftover {
		confidence = demote(confidence)
	}

	var rejected []claim.Claim
	for i, c := range claims {
		if accepted[i] || len(c.Parts) == 0 {
			continue
		}
		if c.Start() < span.End() && span.Start < c.End() {
			rejected = append(rejected, c)
		}
	}

	return LineClaim{
		Span:     span,
		Claim:    claim.Claim{Confidence: confidence, Parts: parts},
		Rejected: rejected,
		Leftover: leftover,
	}
}

// cityStarts offers the places the city could begin, between lineStart and the
// region. A line break or a comma marks a boundary outright. Where neither is
// present the boundary is a guess, and the guesses are bounded by maxCitySpan
// rather than allowed to swallow the street address.
func cityStarts(tokens []token.Token, lineStart, cityEnd int) []int {
	var starts []int

	for i := lineStart; i < cityEnd; i++ {
		if i == lineStart || tokens[i].FollowsComma >= 0 || cityEnd-i <= maxCitySpan {
			starts = append(starts, i)
		}
	}

	return starts
}

// marked reports whether the start of the city is stated by the address rather
// than inferred by this package.
//
// The start of the final line only counts when there is a line ahead of it. On
// a single line address the first token is where the address begins, not where
// the last line does, and treating it as a boundary would rate the reading that
// swallows the whole street address as an exact one.
func marked(tokens []token.Token, lineStart, start int) bool {
	if tokens[start].FollowsComma >= 0 {
		return true
	}

	return start == lineStart && tokens[start].Line > 0
}

// pattern names the shape a reading matched. Project US@ documents the first
// two and the fourth; patternCityPostal is not one of them and is offered
// anyway, for the reason given on confidenceFor.
type pattern int

const (
	patternCityRegionPostal pattern = iota
	patternCityRegion
	patternCityPostal
	patternPostal
)

// confidenceFor rates a shape before leftovers are charged for.
//
// The complete pattern with a marked city is the strongest statement an address
// can make about its own last line, and it is rated exact even though a bare
// five digit ZIP is claimed weakly on its own. That is not double counting; it
// is the point of matching a pattern. postalcode rates 84101 weakly because the
// shape alone cannot tell a ZIP from a primary number, and a five digit number
// sitting after a region abbreviation at the end of the final line is no longer
// the shape alone. The line corroborates its members, so a reading is not
// bounded by the weakest claim in it.
//
// patternCityPostal — a city with a postal code and no region — is the one
// shape here the standard does not list, and it is rated below every documented
// one so that it can never outrank them on the same tokens. It is offered
// because a city in a reasonable position ahead of the postal code is a city,
// and the alternative is to report DENVER on DENVER 80201 as a leftover, which
// says the line does not account for it when plainly it does. Both readings are
// returned; this one only wins where nothing better applies.
//
// marked is whether the address states where the city begins — a comma, or a
// line break with a line ahead of it — rather than this package inferring it.
func confidenceFor(shape pattern, marked bool) claim.Confidence {
	switch shape {
	case patternCityRegionPostal:
		if marked {
			return claim.ConfidenceExact
		}

		return claim.ConfidenceStrong

	case patternCityRegion:
		if marked {
			return claim.ConfidenceStrong
		}

		return claim.ConfidenceLikely

	case patternCityPostal:
		if marked {
			return claim.ConfidenceLikely
		}

		return claim.ConfidenceWeak

	default:
		return claim.ConfidenceStrong
	}
}

// demote lowers a confidence by one step on the shared scale. It is how a
// leftover is charged for: not a numeric penalty, which would invite arithmetic
// the scale does not support, but a step down a list of named readings.
func demote(confidence claim.Confidence) claim.Confidence {
	switch {
	case confidence > claim.ConfidenceStrong:
		return claim.ConfidenceStrong
	case confidence > claim.ConfidenceLikely:
		return claim.ConfidenceLikely
	default:
		return claim.ConfidenceWeak
	}
}

// startOfLineContaining finds the first token of the line the given token is
// on. A single line address has one line, and the whole of it is a candidate.
func startOfLineContaining(tokens []token.Token, at int) int {
	for i := at; i > 0; i-- {
		if tokens[i-1].Line != tokens[at].Line {
			return i
		}
	}

	return 0
}
