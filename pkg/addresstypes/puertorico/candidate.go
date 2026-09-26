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
	"github.com/PortobelloAuth/go-projectusat/pkg/cityabbreviations"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
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

func normalizePRStreetName(streetname string, o normalizer.AddressNormalizationOptions) (string, error) {
	sn, err := textutil.FreeTextField(streetname, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("puertorico street name: %w", err)
	}

	if sn == "" {
		return "", nil
	}

	regioninfo, _ := region.Info(sn, false)
	if regioninfo != nil && regioninfo.PossibleStreetName {
		// if a region is the whole street name, replace it with the full state name
		return regioninfo.Primary, nil
	}

	// if street name has only 1 word, run it through the streetsuffix normalizer;
	// a directional street name is spelled out (NORTH AVE), as one inside a
	// longer name is below. A single letter is left as written instead: it
	// may be an alphabet indicator (1000 G ST, 100 E ST), and the standard says
	// directional letters SHOULD NOT be combined with alphabet indicators
	// (p.17). Anything longer is a spelled-out or abbreviated direction,
	// never an alphabet indicator, and is spelled out (p.18: BAY WEST DRIVE).
	snparts := whitespace.Split(sn, -1)
	if snparts[0] == sn {
		if len(sn) > 1 {
			// Single letter street names are probably alphabetical(?)
			// TODO: figure out a better way to differentiate E ST from EAST ST
			// (or at least to make sure we check the zip/city street data.)
			if full, err := directionals.NormalizeDirectional(sn); err == nil {
				return full, nil
			}
		}

		if ss, err := streetsuffixes.NormalizeStreetSuffix(sn); err == nil {
			return ss, nil
		}
		// TODO: figure out if we actually need to also substitute via highways.NormalizeStreetName()
		return sn, nil
	}

	for i, snp := range snparts {
		// A one-letter final part that follows a street suffix word is
		// an alphabet indicator (AVENUE E), not a direction, and stays
		// as written (p.17). BAY W, where the preceding word is not a
		// suffix, is still a direction and is spelled out (p.18).
		if i == len(snparts)-1 && len(snp) == 1 && normalizer.IsStreetSuffix(snparts[i-1]) {
			continue
		}

		// directionals left in the street name should be the full text
		full, err := directionals.NormalizeDirectional(snp)
		if err == nil {
			// replace the part
			snparts[i] = full
		}

		if i < len(snparts)-1 {
			// state names in the street name should only be full text if there are not
			// other, non-suffix elements in the street name. We took care of that
			// earlier, so we substitute the abbreviation instead.
			// TODO: fix this to support multi-word region names or abbreviations
			regioninfo, _ := region.Info(snp, false)
			if regioninfo != nil && regioninfo.PossibleStreetName {
				snparts[i] = regioninfo.Short
				continue
			}

			// ST/STE/MT/FT heading the name is read as
			// SAINT/SAINTE/MOUNT/FORT from the city table
			// (addresstables/cityabbreviations), not as the STREET
			// suffix word: no street is named STREET CLAIR, and
			// SAINT CLAIR is common (#114). This is checked ahead of
			// the suffix table so it wins the collision.
			//
			// Only at the head. The table's position rule — spelled
			// out when another word follows — is a rule about city
			// names, where ST can only be SAINT. Inside a street name
			// a word follows it routinely without that being true:
			// MAIN THING ST NORTH EAST is a suffix and a trailing
			// direction, not a saint.
			// And only where a word that is not a direction follows
			// it. SAINT CLAIR is a name; SAINT NORTHWEST is not
			// anything, and a name of nothing but ST and a direction
			// is a suffix that was absorbed into the name rather than
			// a saint — E ST NW in Washington is read that way by a
			// parser that puts the direction in the name.
			if i == 0 && !normalizer.OnlyDirectionsFollow(snparts) {
				if full, err := cityabbreviations.Expand(snp); err == nil {
					snparts[i] = full
					continue
				}
			}

			// Street suffixes left inside the street name should be the full text
			// Only replace street suffix abreviations if we have not already
			// replaced this index with a state / region. A Puerto Rico address
			// uses its own Spanish vocabulary instead of the more general one (see
			// go-projectusat#95).
			if pr, err := NormalizeStreetType(snp); err == nil {
				snparts[i] = pr
			}
		}
	}
	sn = strings.Join(snparts, " ")

	// Highway forms normalize. An error means the name is not a highway, which
	// is the ordinary case, so the already uppercased and collapsed name stands.
	// TODO: check for an errantly parsed predirectional as well
	hw, err := highways.NormalizeStreetName(sn)
	if err == nil {
		return hw, nil
	}

	return sn, nil
}

func (p *PuertoRicoAddress) Normalize(a *address.Address, o normalizer.AddressNormalizationOptions) (*address.Address, error) {
	if _, ok := a.Type.(*PuertoRicoAddress); !ok {
		return nil, fmt.Errorf("address is not a *PuertoRicoAddress")
	}

	// The type is how the address formats; normalizing the fields does not
	// change which kind of address they make.
	out := address.Address{Type: a.Type}

	var err error
	if out.Postal, err = normalizer.NormalizePostal(a.Postal, o); err != nil {
		return nil, err
	}
	if out.Region, err = normalizer.NormalizeRegion(a.Region, o); err != nil {
		return nil, err
	}
	if !UsePRDialect(out.Region, out.Postal) {
		return nil, fmt.Errorf("Not a Puerto Rico address")
	}

	// A Puerto Rico address uses only its own Spanish street-type
	// vocabulary, never the English suffix table: AVE and BLVD collide
	// between the two (go-projectusat#95), so a PR address run through
	// the English table silently mistranslates (1234 AVE ASHFORD ->
	// 1234 AVENUE ASHFORD instead of staying Spanish).
	if out.StreetName, err = normalizePRStreetName(a.StreetName, o); err != nil {
		return nil, err
	}

	if out.PrimaryNumber, err = normalizer.NormalizePrimaryNumber(a.PrimaryNumber, o); err != nil {
		return nil, err
	}

	// TODO: make sure that we should do this for puertorico addresses (most won't have a StreetSuffix)
	if out.StreetSuffix, err = normalizer.NormalizeStreetSuffix(a.StreetSuffix, o); err != nil {
		return nil, err
	}

	if len(a.SecondaryNumber) > 0 {
		if out.SecondaryNumber, err = normalizer.NormalizeSecondaryNumber(a.SecondaryNumber, o); err != nil {
			return nil, err
		}
	}
	if len(a.SecondaryDesignator) > 0 {
		if out.SecondaryDesignator, err = NormalizeSecondary(a.SecondaryDesignator); err != nil {
			return nil, err
		}
	}

	if out.Predirectional, err = normalizer.NormalizePredirectional(a.Predirectional, o); err != nil {
		return nil, err
	}
	if out.Postdirectional, err = normalizer.NormalizePostdirectional(a.Postdirectional, o); err != nil {
		return nil, err
	}

	// TODO: Make sure that the result of normalizing the street name and the primary number is a
	// valid puertorico street line.
	// if !CheckStreetLine(...) {
	// 	return nil, fmt.Errorf("Failed to normalize puertorico street line")
	// }

	if out.BusinessName, err = normalizer.NormalizeBusinessName(a.BusinessName, o); err != nil {
		return nil, err
	}
	if out.Area, err = normalizer.NormalizeArea(a.Area, o); err != nil {
		return nil, err
	}
	if out.Detail, err = normalizer.NormalizeDetail(a.Detail, o); err != nil {
		return nil, err
	}
	if out.City, err = normalizer.NormalizeCity(a.City, o); err != nil {
		return nil, err
	}
	if out.Country, err = normalizer.NormalizeCountry(a.Country, o); err != nil {
		return nil, err
	}

	return &out, nil
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
