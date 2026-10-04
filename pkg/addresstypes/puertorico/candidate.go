package puertorico

import (
	"fmt"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico/ruralroute"
	endirectionals "github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// PuertoRicoAddress is the AddressType for a Puerto Rico street address.
//
// The street line is a primary address number and a street name that carries
// its own type at the front, so the fields it fills are the ordinary ones and
// the order they are rendered in is the ordinary order. That is only true
// because the Spanish type stays inside StreetName: a.Suffix is never used to
// hold it. p. 26 is explicit that the type is part of the name rather than a
// suffix that moved, and this package does not play that game for English
// prefixes either — nothing here reaches for a.Suffix to carry a leading
// word. What makes the address Puerto Rican is the vocabulary that reads it
// and the urbanization line above it, not a different arrangement of the
// street line.
type PuertoRicoAddress struct{}

func (o *PuertoRicoAddress) IsCivicAddressType() bool { return true }

// FormatStreetLine renders "A17 CALLE AMAPOLA", or "1 COND MIRAFLOR APT 104"
// where a secondary designator follows the name.
//
// The field order is the one address.Address.FormatStreetLine already owns for
// an address with no special format, so this reaches that branch rather than
// restating it. There is no suffix to place: a Puerto Rico street type leads
// the name and stays inside it, per p. 26. A pre- or postdirectional, where
// there is one, sits in its own field and renders in the ordinary position,
// "1510 CALLE 3 NO" (p. 26, go-projectusat#154). See CONTRIBUTING §1.2, and
// ordinarystreet, which delegates the same way.
func (p *PuertoRicoAddress) FormatStreetLine(a *address.Address) string {
	ordinary := *a
	ordinary.Type = nil

	return ordinary.FormatStreetLine()
}

func PrefixAndSingleLetterStreetNameFn(sn string, o normalizer.AddressNormalizationOptions) (string, error) {
	// - “suffix-first” single letter preference (AVENIDA D, CALLE C)
	parts := strings.Split(sn, " ")
	if len(parts) == 2 && len(parts[1]) == 1 {
		fullprefix, err := NormalizeStreetType(parts[0])
		if err == nil {
			return fmt.Sprintf("%s %s", strings.ToUpper(fullprefix), parts[1]), normalizer.Done
		}
	}

	return sn, nil
}

func ExpandPuertoRicoStreetTypeInStreetNameFn(sn string, o normalizer.AddressNormalizationOptions) (string, error) {
	// - street suffix expansion (within street name, not street suffix field)
	parts := strings.Split(sn, " ")
	changed := false
	for i, snp := range parts {
		fullsuffix, err := NormalizeStreetType(snp)
		if err == nil {
			parts[i] = fullsuffix
			changed = true
		}
	}

	if changed {
		return strings.Join(parts, " "), nil
	}

	return sn, nil
}

// ExpandPuertoRicoDirectionalsInStreetNameFn expands a Spanish directional
// abbreviation or word found anywhere in the free-text street name to its
// full Spanish word. It replaces normalizer.ExpandDirectionalsInStreetNameFn
// in NormalizePuertoRicoStreetName the same way PrefixAndSingleLetterStreetNameFn
// and ExpandPuertoRicoStreetTypeInStreetNameFn already replace their generic
// counterparts (PR #175 review): the generic step reads English only, so it
// would leave NOROESTE unrecognized, and p.25 forbids expanding a Spanish
// directional to its English row anyway.
func ExpandPuertoRicoDirectionalsInStreetNameFn(sn string, o normalizer.AddressNormalizationOptions) (string, error) {
	parts := strings.Split(sn, " ")
	newparts := make([]string, 0)
	changed := false
	for i := 0; i < len(parts); i++ {
		snp := parts[i]
		for j := len(parts); j > i; j-- {
			set := parts[i:j]
			snphrase := strings.Join(set, " ")

			full, err := normalizeSpanishDirectional(snphrase)
			if err == nil && len(full) > 0 {
				snp = full
				changed = true

				// jump to j - 1 so we don't re-replace what we just replaced
				i = j - 1
				break
			}
		}
		newparts = append(newparts, snp)
	}

	if changed {
		return strings.Join(newparts, " "), nil
	}

	return sn, nil
}

var NormalizePuertoRicoStreetName = normalizer.ComposeStreetNameNormalizationFn(
	normalizer.NormalizeTextFn,
	normalizer.OnlySingleLetterStreetNameFn,
	PrefixAndSingleLetterStreetNameFn,
	normalizer.OnlyRegionStreetNameFn,
	normalizer.NormalizeHighwayStreetNameFn,

	// A directional INSIDE the street name is expanded (Pub 28 / Project US@);
	// one that is a pre- or postdirectional is not in StreetName at all —
	// Candidates places it in Predirectional/Postdirectional, and
	// normalizePRDirectionals abbreviates it there (go-projectusat#154).
	ExpandPuertoRicoDirectionalsInStreetNameFn,
	normalizer.ExpandCityInStreetNameFn,
	ExpandPuertoRicoStreetTypeInStreetNameFn,
	// Abbreviate region LAST, after every step that can expand an
	// abbreviation: NEBRASKA must not become NORTHEAST, and MONTANA's MT must
	// not be read back as MOUNT by ExpandCityInStreetNameFn. Same ordering,
	// for the same reason, as normalizer.NormalizeStreetName (#165, fab3c94).
	normalizer.AbbreviateRegionInStreetNameFn,
)

// abbreviatePRDirectional abbreviates a pre- or postdirectional on a Puerto
// Rico address. A Spanish directional keeps its own Spanish abbreviation
// (NOROESTE -> NO, never NW: p.25 "developers MUST NOT translate
// directionals"); anything else falls back to the shared English vocabulary,
// so EAST -> E still works on a PR address.
func abbreviatePRDirectional(field, v string) (string, error) {
	v = textutil.BaseField(v)
	if v == "" {
		return "", nil
	}

	if abbr, err := abbreviateSpanishDirectional(v); err == nil {
		return abbr, nil
	}

	abbr, err := endirectionals.AbbreviateDirectional(v)
	if err != nil {
		return "", fmt.Errorf("%s: %w", field, err)
	}

	return abbr, nil
}

// normalizePRDirectionals replaces normalizer.NormalizeDirectionals for a
// Puerto Rico address, the same way PrefixAndSingleLetterStreetNameFn and
// ExpandPuertoRicoStreetTypeInStreetNameFn replace their generic
// counterparts: the generic step reads English only, so it rejects NO and SO
// outright. Pre- and postdirectionals are ABBREVIATED (go-projectusat#154).
func normalizePRDirectionals(a *address.Address, o normalizer.AddressNormalizationOptions) (*address.Address, error) {
	pre, err := abbreviatePRDirectional("predirectional", a.Predirectional)
	if err != nil {
		return nil, err
	}

	post, err := abbreviatePRDirectional("postdirectional", a.Postdirectional)
	if err != nil {
		return nil, err
	}

	a.Predirectional = pre
	a.Postdirectional = post

	return a, nil
}

func normalizePRStreetNameFn(a *address.Address, o normalizer.AddressNormalizationOptions) (*address.Address, error) {
	if len(a.StreetName) > 0 {
		out, err := NormalizePuertoRicoStreetName(a.StreetName, o)
		if err != nil {
			return nil, err
		}
		a.StreetName = out
	}
	return a, nil
}

func normalizePRSecondaryDesignatorFn(a *address.Address, o normalizer.AddressNormalizationOptions) (*address.Address, error) {
	if len(a.SecondaryDesignator) > 0 {
		out, err := NormalizeSecondary(a.SecondaryDesignator)
		if err != nil {
			return nil, err
		}
		a.SecondaryDesignator = out
	}
	return a, nil
}

var normalizePRStreetLine = normalizer.ComposeNormalizationFn(
	normalizer.NormalizePrimaryNumberFn,
	// NOTE: most puerto rico addresses won't have a suffix. It is not how Spanish streets are
	// named.
	normalizer.NormalizeStreetSuffixFn,
	normalizer.NormalizeSecondaryNumberFn,
	normalizePRSecondaryDesignatorFn,
	normalizePRDirectionals,

	// A Puerto Rico address uses only its own Spanish street-type
	// vocabulary, never the English suffix table: AVE and BLVD collide
	// between the two (go-projectusat#95), so a PR address run through
	// the English table silently mistranslates (1234 AVE ASHFORD ->
	// 1234 AVENUE ASHFORD instead of staying Spanish).
	normalizePRStreetNameFn,
)

var normalizePRAddressFn = normalizer.ComposeNormalizationFn(
	normalizer.NormalizeLastLine,
	normalizer.NormalizeOtherParts,
	normalizePRStreetLine,
)

func (p *PuertoRicoAddress) Normalize(a *address.Address, o normalizer.AddressNormalizationOptions) (*address.Address, error) {
	if _, ok := a.Type.(*PuertoRicoAddress); !ok {
		return nil, fmt.Errorf("address is not a *PuertoRicoAddress")
	}

	if !UsePRDialect(a.Region, a.Postal) {
		return nil, fmt.Errorf("Not a Puerto Rico address")
	}

	// The type is how the address formats; normalizing the fields does not
	// change which kind of address they make.
	out := a.Clone()

	var err error
	if out, err = normalizePRAddressFn(out, o); err != nil {
		return nil, err
	}

	// Make sure that the result of normalizing the street name and the primary number is a
	// valid puertorico street line.
	if _, _, err := NormalizeStreetLine(out.FormatStreetLine()); err != nil {
		return nil, err
	}

	return out, nil
}

// Candidates returns this package's reading of the address under the given
// last line.
//
// The dialect gate comes first: a Puerto Rico address is one whose last line
// says PR or carries a Puerto Rico ZIP, and nothing below that line can tell.
// That is the answer this package has to the question raised on #71 — which
// vocabularies may claim on a Puerto Rico address. Claims(tokens) cannot
// answer it, because the dialect is a judgment about the whole address and a
// Claims function sees only tokens. So the Spanish street vocabulary makes no
// claim into the shared pool at all, and reads the street line here instead,
// where the last line is known. A mainland address never meets it, which is
// the same protection in the other direction: 1000 AVE E in Florida is not
// offered a Spanish reading.
//
// Reading the tokens rather than the pool is also what #70 asks of an address
// type, and it is what lets a single-line address be read: the street line
// ends where the last line begins, whether or not a line break says so.
//
// Where an urbanization sits above the street line, two readings are offered —
// with it and without — so that a reading which strands the urbanization is
// ranked against one that does not, rather than being the only one on offer.
//
// A rural route or highway contract route is read from the same gate, by the
// ruralroute sub-package, and is not conditional on there being a street line
// at all: p. 30's route is the whole of the delivery address, and the line the
// standard prints beneath it is a sector name to be eliminated rather than a
// street.
func Candidates(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) []*address.CandidateAddress {
	if !isPuertoRicoLastLine(line) {
		return nil
	}

	candidates := routeCandidates(tokens, line)

	street, ok := streetLine(tokens, claims, line)
	if !ok {
		if start, end, boundsOK := streetLineBounds(tokens, line); boundsOK {
			// The two urbanization-only shapes below and the condominium
			// fallback can both match the same bounds when a word that
			// happens to sit in the pp.28-29 standalone-exceptions table
			// (e.g. "VISTA") opens a building name: condominiumStreetLine's
			// own precondition — a secondary-unit claim flush at the end of
			// the line, see its comment — is the stronger, unambiguous
			// signal in that case, so the urbanization shapes defer to it
			// rather than compete with it. A true standalone urbanization
			// line, per the issue, never carries a secondary designator of
			// its own; it is the line above one, not the line with one.
			if _, hasTrailingSecondary := trailingSecondaryUnit(claims, start, end); !hasTrailingSecondary {
				if urb, urbOK := standaloneUrbanizationStreetLine(claims, start, end); urbOK {
					candidates = append(candidates,
						line.Candidate(&PuertoRicoAddress{}, len(tokens), []claim.Claim{urb}))
				}

				if combined, combinedOK := primaryNumberedUrbanizationStreetLine(tokens, claims, start, end); combinedOK {
					candidates = append(candidates,
						line.Candidate(&PuertoRicoAddress{}, len(tokens), []claim.Claim{combined}))
				}
			}

			if name, secondary, hasSecondary, primaryNumber, condoOK := condominiumStreetLine(tokens, claims, start, end); condoOK {
				parts := []claim.Claim{name}
				if hasSecondary {
					parts = append(parts, secondary)
				}

				candidate := line.Candidate(&PuertoRicoAddress{}, len(tokens), parts)
				candidate.Address.PrimaryNumber = primaryNumber
				candidates = append(candidates, candidate)
			}
		}

		return candidates
	}

	candidates = append(candidates,
		line.Candidate(&PuertoRicoAddress{}, len(tokens), street))

	for _, c := range claims {
		if !isUrbanization(c) || c.End() > street[0].Start() {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&PuertoRicoAddress{}, len(tokens), append([]claim.Claim{c}, street...)))
	}

	return candidates
}

// routeCandidates offers one reading for each route the Spanish vocabulary
// finds above the last line.
//
// A route claim reaching into the last line is discarded rather than trimmed:
// the two would assign the same tokens, and a candidate may not claim one
// twice. That is also the guard against reading a Puerto Rico ZIP range as a
// box number.
func routeCandidates(tokens []token.Token, line lastline.LineClaim) []*address.CandidateAddress {
	var candidates []*address.CandidateAddress

	for _, c := range ruralroute.Claims(tokens) {
		if c.End() > line.Span.Start {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&ruralroute.PuertoRicoRouteAddress{}, len(tokens), []claim.Claim{c}))
	}

	return candidates
}

// isPuertoRicoLastLine reports whether the last line puts the address in
// Puerto Rico.
//
// Either the region or the postal code settles it on its own — see
// UsePRDialect — so an address that arrives with one and not the other is
// still read in Spanish.
func isPuertoRicoLastLine(line lastline.LineClaim) bool {
	var region, postal string
	for _, p := range line.Claim.Parts {
		switch p.Part {
		case claim.PartRegion:
			region = p.Value
		case claim.PartPostal:
			postal = p.Value
		}
	}

	return UsePRDialect(region, postal)
}

// streetLine reads the delivery line immediately above the last line.
//
// The street line is the bottom line of the street address block, so the one
// that ends where the last line begins is the one to read. Everything above it
// is the urbanization and the secondary identifier, which have their own
// readings.
//
// The standard writes that line two ways round and both are read. The one it
// requires puts the primary address number first, "A17 CALLE 1", and
// NormalizeStreetLine reads it. The one pp. 26-27 print in their Incorrect
// Form column puts the street first and the house number after it, "CALLE 1
// A17", and NormalizeNumberedStreetLine reads that. The order is which
// recognizer is asked, not a preference between readings: the two shapes do
// not overlap, so at most one of them answers.
//
// A directional flush against the end of the line is a postdirectional, and
// one immediately after the primary number is a predirectional: either is
// read as its own claim, so Candidates places it in Postdirectional or
// Predirectional rather than leaving it inside StreetName
// (go-projectusat#154). That is what keeps p. 26's "1510 CALLE 3 NO" NO: a
// directional left inside the street name is expanded by
// NormalizePuertoRicoStreetName, which is right for a directional that is part
// of the name and wrong for one that is not. A directional is only split off
// where the rest of the line is still a street line on its own, so "150 CALLE
// O" keeps O as its root name. The first slice element is always the street
// claim itself; a directional claim, where there is one, follows it.
func streetLine(tokens []token.Token, claims []claim.Claim, line lastline.LineClaim) ([]claim.Claim, bool) {
	start, end, ok := streetLineBounds(tokens, line)
	if !ok {
		return nil, false
	}

	if post, ok := directionalClaim(claims, claim.PartPostdirectional, func(p claim.ClaimPart) bool {
		return p.End() == end && p.Start > start+2
	}); ok {
		nameEnd := post.Start()
		if number, name, err := NormalizeStreetLine(token.Join(tokens[start:nameEnd])); err == nil {
			return []claim.Claim{streetClaim(
				claim.ClaimPart{Start: start, Length: 1, Part: claim.PartPrimaryNumber, Value: number},
				claim.ClaimPart{Start: start + 1, Length: nameEnd - start - 1, Part: claim.PartStreetName, Value: name},
			), post}, true
		}
	}

	if pre, ok := directionalClaim(claims, claim.PartPredirectional, func(p claim.ClaimPart) bool {
		return p.Start == start+1 && p.End() < end-1
	}); ok {
		nameStart := pre.End()
		text := token.Join(append([]token.Token{tokens[start]}, tokens[nameStart:end]...))
		if number, name, err := NormalizeStreetLine(text); err == nil {
			return []claim.Claim{streetClaim(
				claim.ClaimPart{Start: start, Length: 1, Part: claim.PartPrimaryNumber, Value: number},
				claim.ClaimPart{Start: nameStart, Length: end - nameStart, Part: claim.PartStreetName, Value: name},
			), pre}, true
		}
	}

	text := token.Join(tokens[start:end])

	if number, name, err := NormalizeStreetLine(text); err == nil {
		return []claim.Claim{streetClaim(
			claim.ClaimPart{Start: start, Length: 1, Part: claim.PartPrimaryNumber, Value: number},
			claim.ClaimPart{Start: start + 1, Length: end - start - 1, Part: claim.PartStreetName, Value: name},
		)}, true
	}

	// The number covers everything after the street name, identifiers
	// included: p. 27 says those words MUST NOT be included in the address, so
	// the tokens are spoken for by the part whose value drops them rather than
	// left for another vocabulary to read. See claim.ClaimPart.Value.
	if number, name, err := NormalizeNumberedStreetLine(text); err == nil {
		nameEnd := start + numberedStreetNameFields

		return []claim.Claim{streetClaim(
			claim.ClaimPart{Start: start, Length: numberedStreetNameFields, Part: claim.PartStreetName, Value: name},
			claim.ClaimPart{Start: nameEnd, Length: end - nameEnd, Part: claim.PartPrimaryNumber, Value: number},
		)}, true
	}

	return nil, false
}

// directionalClaim finds the directional claim of the given part that match
// accepts, preferring the longer span (a compound such as NORTE ESTE over the
// ESTE it ends in, the preference pkg/directionals.Claims documents) and then
// the stronger claim. Both this package's Spanish vocabulary and the shared
// English one (pkg/directionals) contribute candidates; a single-part claim
// is all either offers.
func directionalClaim(claims []claim.Claim, part claim.Part, match func(claim.ClaimPart) bool) (claim.Claim, bool) {
	var best claim.Claim
	found := false

	for _, c := range claims {
		if len(c.Parts) != 1 || c.Parts[0].Part != part || !match(c.Parts[0]) {
			continue
		}

		if found {
			p, b := c.Parts[0], best.Parts[0]
			if p.Length < b.Length || (p.Length == b.Length && c.Confidence <= best.Confidence) {
				continue
			}
		}

		best, found = c, true
	}

	return best, found
}

// streetClaim holds the confidence a street line reading is offered at, so the
// two shapes streetLine reads cannot drift apart on it. Both are exact: the
// shapes are disjoint, and within this vocabulary neither line can be read any
// other way.
func streetClaim(parts ...claim.ClaimPart) claim.Claim {
	return claim.Claim{Confidence: claim.ConfidenceExact, Parts: parts}
}

// streetLineBounds finds the extent of the delivery line immediately above the
// last line, the span streetLine and condominiumStreetLine both read from.
//
// end is where the last line begins; start walks back from there to the
// beginning of that physical line, so a line break — where one exists — is
// what bounds it, and its absence (a single-line address) still leaves end
// itself as the boundary.
func streetLineBounds(tokens []token.Token, line lastline.LineClaim) (start, end int, ok bool) {
	end = line.Span.Start
	if end <= 0 || end > len(tokens) {
		return 0, 0, false
	}

	start = end - 1
	for start > 0 && tokens[start-1].Line == tokens[end-1].Line {
		start--
	}

	return start, end, true
}

// isSecondaryUnit reports whether a claim is a secondaryunit reading: a
// designator, or a designator and its number, and nothing else. Mirrors
// isUrbanization below, which does the same test for a different vocabulary's
// part.
func isSecondaryUnit(c claim.Claim) bool {
	for _, p := range c.Parts {
		if p.Part != claim.PartSecondaryDesignator && p.Part != claim.PartSecondaryNumber {
			return false
		}
	}

	return len(c.Parts) > 0
}

// trailingSecondaryUnit finds the secondary-unit claim sitting flush against
// the end of a condominium street line — "APT 1120" in "COND VERDE APT 1120" —
// so condominiumStreetLine knows where the building name ends.
//
// Flush against end, not merely inside [start, end), because a secondary unit
// anywhere else in the line is not this pattern: p. 25's fallback is for a
// line that ends in a secondary designator with nothing after it, not for one
// that happens to contain one.
func trailingSecondaryUnit(claims []claim.Claim, start, end int) (claim.Claim, bool) {
	for _, c := range claims {
		if isSecondaryUnit(c) && c.Start() >= start && c.End() == end {
			return c, true
		}
	}

	return claim.Claim{}, false
}

// condominiumStreetLine reads the one shape streetLine's two recognizers both
// miss: a building name with a secondary unit but no primary address number at
// all, such as "COND VERDE APT 1120" or "VISTA SUITES III APT 104".
//
// p. 25 gives two rules for this case. Where no building number exists, the
// primary number defaults to "1" (go-projectusat#158). Where the building name
// ends in a number — "Where there are multiple buildings (or towers) with the
// same name, the building number SHOULD become the primary number" — that
// number becomes the primary number instead (go-projectusat#159), and this
// package only recognizes that building number when it is spelled as a
// closed-vocabulary roman numeral (see romannumeral.go); an arabic building
// number is already read by NormalizeStreetLine's ordinary shape.
//
// Either way the primary number is never a ClaimPart: claim.ClaimPart.Length
// must cover at least one real token (see its doc comment), and in the "1"
// case there is no token to attribute "1" to at all — it is pure default, not
// a reading of anything on the line. So the caller writes it directly onto
// the assembled *address.CandidateAddress instead of into a claim part, the
// same way streetLine's numbered-street branch already absorbs tokens into a
// Value without covering them with their own part (see the comment above its
// NormalizeNumberedStreetLine branch for that precedent). That absorption
// happens here too: when the trailing token is a roman numeral, the
// street-name claim's Length still covers it, even though the numeral
// contributes nothing to the Value.
//
// This fallback only fires when a secondary-unit claim sits flush at the end
// of the line. A bare unrecognized line with no secondary unit at all is not
// this pattern — that would be a much broader change than either #158 or #159
// asked for — so ok is false whenever trailingSecondaryUnit finds nothing.
func condominiumStreetLine(tokens []token.Token, claims []claim.Claim, start, end int) (name claim.Claim, secondary claim.Claim, hasSecondary bool, primaryNumber string, ok bool) {
	secondary, hasSecondary = trailingSecondaryUnit(claims, start, end)
	if !hasSecondary {
		return claim.Claim{}, claim.Claim{}, false, "", false
	}

	leadingEnd := secondary.Start()
	leading := tokens[start:leadingEnd]
	if len(leading) == 0 {
		return claim.Claim{}, claim.Claim{}, false, "", false
	}

	var nameTokens []token.Token
	if number, numeralOK := romanNumeral(leading[len(leading)-1].Text); numeralOK {
		nameTokens = leading[:len(leading)-1]
		if len(nameTokens) == 0 {
			return claim.Claim{}, claim.Claim{}, false, "", false
		}
		primaryNumber = number
	} else {
		nameTokens = leading
		primaryNumber = "1"
	}

	name = claim.Claim{
		Confidence: claim.ConfidenceLikely,
		Parts: []claim.ClaimPart{
			{Start: start, Length: leadingEnd - start, Part: claim.PartStreetName, Value: token.Join(nameTokens)},
		},
	}

	return name, secondary, true, primaryNumber, true
}

// isUrbanization reports whether a claim is the urbanization line this package
// reads. Area is the part it writes, and no other vocabulary writes it.
func isUrbanization(c claim.Claim) bool {
	for _, p := range c.Parts {
		if p.Part != claim.PartArea {
			return false
		}
	}

	return len(c.Parts) > 0
}

// areaValue returns the Area value an urbanization claim carries.
// isUrbanization guarantees a claim with that shape has exactly one kind of
// part, so this just unwraps it under a name that says what it holds rather
// than reindexing Parts[0] at each call site.
func areaValue(c claim.Claim) string {
	for _, p := range c.Parts {
		if p.Part == claim.PartArea {
			return p.Value
		}
	}

	return ""
}

// standaloneUrbanizationStreetLine finds an urbanization claim that is, by
// itself, the entire delivery line streetLine failed to read: no primary
// number, and no separate street name below it.
//
// pp.28-29's own worked examples take this shape — "URB GOLDEN GATE" on a
// line by itself, or one of the standalone exceptions such as "EXT VISTA
// BELLA" — where the urbanization is the whole of the street-equivalent
// content rather than a line sitting above an ordinary one. streetLine has
// nothing to read there (there is no primary number or street type on the
// line at all), which is why this is checked only once streetLine has
// already failed, not as an alternative to it.
//
// The claim this returns carries only a PartArea part, so the candidate it
// builds leaves PrimaryNumber and StreetName empty. address.Address.Format
// renders Area on its own line and FormatStreetLine on an empty one is
// omitted entirely (see textutil.JoinNonEmpty), which is exactly these
// examples' expected output: the urbanization line and the last line, with
// nothing in between.
func standaloneUrbanizationStreetLine(claims []claim.Claim, start, end int) (claim.Claim, bool) {
	for _, c := range claims {
		if isUrbanization(c) && c.Start() == start && c.End() == end {
			return c, true
		}
	}

	return claim.Claim{}, false
}

// primaryNumberedUrbanizationStreetLine finds the one shape
// standaloneUrbanizationStreetLine does not cover: a primary address number
// opening the line, immediately followed by an urbanization name that runs to
// the end of it.
//
// p.28's own example is exactly this: "A17 URB JARDINES FAGOTA" ->
// "A17 JARD FAGOTA", the standalone urbanization name standing in as the
// street name, with the primary number that always leads a Puerto Rico
// street line still in front of it. urbanizationClaim already restricts which
// claims admit a number ahead of the designator at all (see
// precededOnlyByPrimaryNumber on that type) — this is the candidate.go side
// of the same rule, reassembling that claim's Area value as a street name
// rather than leaving it a reading nothing ever reads back out, which is what
// the ordinary PrimaryNumber-and-StreetName claim shape asks for.
func primaryNumberedUrbanizationStreetLine(tokens []token.Token, claims []claim.Claim, start, end int) (claim.Claim, bool) {
	if start >= end {
		return claim.Claim{}, false
	}

	number, ok := normalizePrimaryNumber(strings.ToUpper(tokens[start].Text))
	if !ok {
		return claim.Claim{}, false
	}

	for _, c := range claims {
		if !isUrbanization(c) || c.Start() != start+1 || c.End() != end {
			continue
		}

		return claim.Claim{
			Confidence: claim.ConfidenceExact,
			Parts: []claim.ClaimPart{
				{Start: start, Length: 1, Part: claim.PartPrimaryNumber, Value: number},
				{Start: c.Start(), Length: c.Length(), Part: claim.PartStreetName, Value: areaValue(c)},
			},
		}, true
	}

	return claim.Claim{}, false
}
