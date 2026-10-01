package puertorico

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico/ruralroute"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
)

var whitespace = regexp.MustCompile(`\s+`)

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
// restating it. There is no suffix and no trailing directional to place: a
// Puerto Rico street type leads the name and stays inside it, per p. 26. See
// CONTRIBUTING §1.2, and ordinarystreet, which delegates the same way.
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

var NormalizePuertoRicoStreetName = normalizer.ComposeStreetNameNormalizationFn(
	normalizer.NormalizeTextFn,
	normalizer.OnlySingleLetterStreetNameFn,
	PrefixAndSingleLetterStreetNameFn,
	normalizer.OnlyRegionStreetNameFn,
	normalizer.NormalizeHighwayStreetNameFn,

	normalizer.ExpandDirectionalsInStreetNameFn,
	// Abbreviate region AFTER expanding directionals so that NEBRASKA doesn't get
	// converted to NORTHEAST
	normalizer.AbbreviateRegionInStreetNameFn,
	normalizer.ExpandCityInStreetNameFn,
	ExpandPuertoRicoStreetTypeInStreetNameFn,
)

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
	normalizer.NormalizeDirectionals,

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

	// TODO: Make sure that the result of normalizing the street name and the primary number is a
	// valid puertorico street line.
	// if !CheckStreetLine(...) {
	// 	return nil, fmt.Errorf("Failed to normalize puertorico street line")
	// }

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

	street, ok := streetLine(tokens, line)
	if !ok {
		return candidates
	}

	candidates = append(candidates,
		line.Candidate(&PuertoRicoAddress{}, len(tokens), []claim.Claim{street}))

	for _, c := range claims {
		if !isUrbanization(c) || c.End() > street.Start() {
			continue
		}

		candidates = append(candidates,
			line.Candidate(&PuertoRicoAddress{}, len(tokens), []claim.Claim{c, street}))
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
func streetLine(tokens []token.Token, line lastline.LineClaim) (claim.Claim, bool) {
	end := line.Span.Start
	if end <= 0 || end > len(tokens) {
		return claim.Claim{}, false
	}

	start := end - 1
	for start > 0 && tokens[start-1].Line == tokens[end-1].Line {
		start--
	}

	text := token.Join(tokens[start:end])

	if number, name, err := NormalizeStreetLine(text); err == nil {
		return streetClaim(
			claim.ClaimPart{Start: start, Length: 1, Part: claim.PartPrimaryNumber, Value: number},
			claim.ClaimPart{Start: start + 1, Length: end - start - 1, Part: claim.PartStreetName, Value: name},
		), true
	}

	// The number covers everything after the street name, identifiers
	// included: p. 27 says those words MUST NOT be included in the address, so
	// the tokens are spoken for by the part whose value drops them rather than
	// left for another vocabulary to read. See claim.ClaimPart.Value.
	if number, name, err := NormalizeNumberedStreetLine(text); err == nil {
		nameEnd := start + numberedStreetNameFields

		return streetClaim(
			claim.ClaimPart{Start: start, Length: numberedStreetNameFields, Part: claim.PartStreetName, Value: name},
			claim.ClaimPart{Start: nameEnd, Length: end - nameEnd, Part: claim.PartPrimaryNumber, Value: number},
		), true
	}

	return claim.Claim{}, false
}

// streetClaim holds the confidence a street line reading is offered at, so the
// two shapes streetLine reads cannot drift apart on it. Both are exact: the
// shapes are disjoint, and within this vocabulary neither line can be read any
// other way.
func streetClaim(parts ...claim.ClaimPart) claim.Claim {
	return claim.Claim{Confidence: claim.ConfidenceExact, Parts: parts}
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
