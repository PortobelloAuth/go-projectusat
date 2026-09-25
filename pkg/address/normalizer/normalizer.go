package normalizer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/cityabbreviations"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

var whitespace = regexp.MustCompile(`\s+`)

// AddressNormalizationOptions controls exchange/matching variants of normalization.
// Zero value is content form (the same settings used by NewContentNormalizer).
type AddressNormalizationOptions struct {
	// Fuzzy enables FuzzyNormalize* for region and street suffix.
	Fuzzy bool
	// SecondaryAsHash rewrites secondary designators to "#" for matching
	// (not correct for content storage; for exchange/matching only).
	SecondaryAsHash bool
	DiacriticMode   diacritics.DiacriticMode
}

// NormalizingAddressType is an interface describing an AddressType that has a
// Normalize() method because it implements its own Normalization rules. It may
// or may not choose to employ normalization from shared address components, etc.
type NormalizingAddressType interface {
	address.AddressType
	Normalize(a *address.Address, o AddressNormalizationOptions) (*address.Address, error)
}

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Normalizer struct {
	Options AddressNormalizationOptions
}

func NewNomalizer(opts AddressNormalizationOptions) *Normalizer {
	return &Normalizer{
		Options: opts,
	}
}

// NewContentNomalizer returns a Normalizer with options appropriate for normalizing
// an address as it is being captured. The Project US@ standard refers to this as
// the "Content" use case. The address is returned in uppercase, with standard
// abbreviations. Diacritics are preserved; callers may pre-run diacritics.Substitute
// if needed. Empty optional fields stay blank. Unrecognized non-empty controlled
// vocabulary (region, directionals, street suffix, secondary designator) returns an
// error.
func NewContentNomalizer() *Normalizer {
	return &Normalizer{}
}

// NewMatchingNomalizer returns a Normalizer with options appropriate for normalizing
// an address for comparison. The Project US@ standard refers to this as
// the "Exchange" use case. The address is returned in uppercase, with standard
// abbreviations. Diacritics are substituted. Numbered Secondary Units use the hash symbol
// to allow for matching addresses where the secondary unit type is not known. Empty
// optional fields stay blank. Unrecognized non-empty controlled vocabulary (region,
// directionals, street suffix, secondary designator) returns an error.
func NewMatchingNomalizer() *Normalizer {
	return &Normalizer{
		Options: AddressNormalizationOptions{
			// Fuzzy:           true,  // TODO: decide whether we really want this to be true or false
			SecondaryAsHash: true,
			DiacriticMode:   diacritics.SubstituteDiacritics,
		},
	}
}

func NormalizeBusinessName(bname string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(bname, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("business name: %w", err)
	}
	return out, nil
}

func NormalizeArea(area string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(area, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("area: %w", err)
	}
	return out, nil
}

func NormalizeDetail(detail string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(detail, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("detail: %w", err)
	}
	return out, nil
}

func NormalizePrimaryNumber(number string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(number, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("primary number: %w", err)
	}
	return out, nil
}

func NormalizeSecondaryNumber(number string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(number, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("secondary number: %w", err)
	}
	return out, nil
}

func NormalizeCity(city string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(city, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("city: %w", err)
	}
	if out != "" {
		// Publication 28 §223 and Project US@ (p.20) both require a city name
		// spelled out in its entirety: ST CLOUD must read SAINT CLOUD, not
		// stay abbreviated (go-projectusat#115). The table states the
		// position rule (addresstables/cityabbreviations): a word is spelled
		// out only when another word follows it, so a lone or trailing ST
		// is left as written rather than expanded.
		cityparts := whitespace.Split(out, -1)
		for i := 0; i < len(cityparts)-1; i++ {
			if full, err := cityabbreviations.Expand(cityparts[i]); err == nil {
				cityparts[i] = full
			}
		}
		out = strings.Join(cityparts, " ")
	}
	return out, nil
}

func NormalizeCountry(country string, o AddressNormalizationOptions) (string, error) {
	out, err := textutil.FreeTextField(country, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("country: %w", err)
	}
	return out, nil
}

func NormalizePostal(postal string, o AddressNormalizationOptions) (string, error) {
	out, err := postalcode.Normalize(postal)
	if err != nil {
		return "", fmt.Errorf("postal code: %w", err)
	}
	return out, nil
}

func NormalizeStreetName(streetname string, o AddressNormalizationOptions) (string, error) {
	sn, err := textutil.FreeTextField(streetname, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("street name: %w", err)
	}

	if sn != "" {
		// TODO: move this in to pobox NormalizingAddressType.Normalize()
		// poboxsn, err := pobox.Normalize(sn)
		// if err == nil {
		// 	sn = poboxsn
		// }

		// TODO: move to puertorico NormalizingAddressType.Normalize()
		// // A Puerto Rico address uses only its own Spanish street-type
		// // vocabulary, never the English suffix table: AVE and BLVD collide
		// // between the two (go-projectusat#95), so a PR address run through
		// // the English table silently mistranslates (1234 AVE ASHFORD ->
		// // 1234 AVENUE ASHFORD instead of staying Spanish). Decided once,
		// // from the pre-normalization Region/Postal, since the loop below
		// // reads it more than once.
		// prDialect := puertorico.UsePRDialect(a.Region, a.Postal)

		// if street name has only 1 word, run it through the streetsuffix normalizer;
		// a directional street name is spelled out (NORTH AVE), as one inside a
		// longer name is below. A single letter is left as written instead: it
		// may be an alphabet indicator (1000 AVENUE E), and the standard says
		// directional letters SHOULD NOT be combined with alphabet indicators
		// (p.17). Anything longer is a spelled-out or abbreviated direction,
		// never an alphabet indicator, and is spelled out (p.18: BAY WEST DRIVE).
		snparts := whitespace.Split(sn, -1)
		if snparts[0] == sn {
			if len(sn) == 1 {
				if ss, err := streetsuffixes.NormalizeStreetSuffix(sn); err == nil {
					sn = ss
				}
			} else if full, err := directionals.NormalizeDirectional(sn); err == nil {
				sn = full
				// TODO: move to puertorico NormalizingAddressType.Normalize()
				// } else if prDialect {
				// 	if pr, err := puertorico.NormalizeStreetType(sn); err == nil {
				// 		sn = pr
				// 	}
			} else if ss, err := streetsuffixes.NormalizeStreetSuffix(sn); err == nil {
				sn = ss
			}
		} else {
			for i, snp := range snparts {
				// A one-letter final part that follows a street suffix word is
				// an alphabet indicator (AVENUE E), not a direction, and stays
				// as written (p.17). BAY W, where the preceding word is not a
				// suffix, is still a direction and is spelled out (p.18).
				if i == len(snparts)-1 && len(snp) == 1 && isStreetSuffix(snparts[i-1]) {
					continue
				}

				// directionals left in the street name should be the full text
				full, err := directionals.NormalizeDirectional(snp)
				if err == nil {
					// replace the part
					snparts[i] = full
				}

				if i < len(snparts)-1 {
					// state names in the street name should be full text if there are not
					// other, non-suffix elements in the street name.
					regioninfo, _ := region.Info(snp, false)
					if regioninfo != nil && regioninfo.PossibleStreetName {
						if i == 0 && len(snparts) <= 2 {
							// replace the part with the full state name
							snparts[i] = regioninfo.Primary
						} else {
							snparts[i] = regioninfo.Short
						}
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
					if i == 0 && !onlyDirectionsFollow(snparts) {
						if full, err := cityabbreviations.Expand(snp); err == nil {
							snparts[i] = full
							continue
						}
					}

					// TODO: move to puertorico NormalizingAddressType.Normalize()
					// // Street suffixes left inside the street name should be the full text
					// // Only replace street suffix abreviations if we have not already
					// // replaced this index with a state / region. A Puerto Rico address
					// // uses its own Spanish vocabulary instead (go-projectusat#95).
					// if prDialect {
					// 	if pr, err := puertorico.NormalizeStreetType(snp); err == nil {
					// 		snparts[i] = pr
					// 	}
					// } else
					if fullss, err := streetsuffixes.NormalizeStreetSuffix(snp); err == nil {
						snparts[i] = fullss
					}
				}
			}
			sn = strings.Join(snparts, " ")
		}

		// Highway forms normalize. An error means the name is not a highway, which
		// is the ordinary case, so the already uppercased and collapsed name stands.
		// TODO: check for an errantly parsed predirectional as well
		hw, err := highways.NormalizeStreetName(sn)
		if err == nil {
			return hw, nil
		}
	}
	return sn, nil
}

func NormalizePredirectional(predirectional string, o AddressNormalizationOptions) (string, error) {
	v := textutil.BaseField(predirectional)
	if v == "" {
		return "", nil
	}

	abbr, err := directionals.AbbreviateDirectional(v)
	if err != nil {
		return "", fmt.Errorf("predirectional: %w", err)
	}
	return abbr, nil
}

func NormalizePostdirectional(postdirectional string, o AddressNormalizationOptions) (string, error) {
	v := textutil.BaseField(postdirectional)
	if v == "" {
		return "", nil
	}

	abbr, err := directionals.AbbreviateDirectional(v)
	if err != nil {
		return "", fmt.Errorf("postdirectional: %w", err)
	}
	return abbr, nil
}

func NormalizeStreetSuffix(suffix string, o AddressNormalizationOptions) (string, error) {
	v := textutil.BaseField(suffix)
	if v == "" {
		return "", nil
	}

	var abbr string
	var err error
	if o.Fuzzy {
		abbr, err = streetsuffixes.FuzzyNormalizeStreetSuffixAbreviation(v)
	} else {
		abbr, err = streetsuffixes.NormalizeStreetSuffixAbreviation(v)
	}
	if err != nil {
		return "", fmt.Errorf("street suffix: %w", err)
	}
	return abbr, nil
}

func NormalizeSecondaryDesingator(designator string, o AddressNormalizationOptions) (string, error) {
	v := textutil.BaseField(designator)
	if v == "" {
		return "", nil
	}
	info, err := secondaryunit.Info(v)
	if err != nil {
		return "", fmt.Errorf("secondary designator: %w", err)
	}

	// Only use SecondaryAsHash for Numbered secondary designators
	if o.SecondaryAsHash && info != nil && info.Numbered {
		return "#", nil
	}
	abbr, err := secondaryunit.Normalize(v)
	if err != nil {
		return "", fmt.Errorf("secondary designator: %w", err)
	}
	return abbr, nil
}

func NormalizeRegion(r string, o AddressNormalizationOptions) (string, error) {
	v := textutil.BaseField(r)
	if v == "" {
		return "", nil
	}

	var abbr string
	var err error
	if o.Fuzzy {
		abbr, err = region.FuzzyNormalizeRegion(v)
	} else {
		abbr, err = region.NormalizeRegion(v)
	}
	if err != nil {
		return "", fmt.Errorf("region: %w", err)
	}
	return abbr, nil
}

// Normalize applies the Normalizer's AddressNormalizationOptions to the Address
func (n *Normalizer) Normalize(a *address.Address) (*address.Address, error) {
	// if the address type is an AddressNormalizingType (it implements its own Normalization
	// rules) employ that Normalization instead of the default.
	if normalizing, ok := a.Type.(NormalizingAddressType); ok {
		return normalizing.Normalize(a, n.Options)
	}

	// The type is how the address formats; normalizing the fields does not
	// change which kind of address they make.
	out := address.Address{Type: a.Type}

	var err error
	if out.BusinessName, err = NormalizeBusinessName(a.BusinessName, n.Options); err != nil {
		return nil, err
	}
	if out.Area, err = NormalizeArea(a.Area, n.Options); err != nil {
		return nil, err
	}
	if out.Detail, err = NormalizeDetail(a.Detail, n.Options); err != nil {
		return nil, err
	}
	if out.PrimaryNumber, err = NormalizePrimaryNumber(a.PrimaryNumber, n.Options); err != nil {
		return nil, err
	}
	if out.SecondaryNumber, err = NormalizeSecondaryNumber(a.SecondaryNumber, n.Options); err != nil {
		return nil, err
	}
	if out.City, err = NormalizeCity(a.City, n.Options); err != nil {
		return nil, err
	}
	if out.Country, err = NormalizeCountry(a.Country, n.Options); err != nil {
		return nil, err
	}
	if out.Postal, err = NormalizePostal(a.Postal, n.Options); err != nil {
		return nil, err
	}
	if out.StreetName, err = NormalizeStreetName(a.StreetName, n.Options); err != nil {
		return nil, err
	}
	if out.Predirectional, err = NormalizePredirectional(a.Predirectional, n.Options); err != nil {
		return nil, err
	}
	if out.Postdirectional, err = NormalizePostdirectional(a.Postdirectional, n.Options); err != nil {
		return nil, err
	}
	if out.StreetSuffix, err = NormalizeStreetSuffix(a.StreetSuffix, n.Options); err != nil {
		return nil, err
	}
	if out.SecondaryDesignator, err = NormalizeSecondaryDesingator(a.SecondaryDesignator, n.Options); err != nil {
		return nil, err
	}
	if out.Region, err = NormalizeRegion(a.Region, n.Options); err != nil {
		return nil, err
	}

	return &out, nil
}

// isStreetSuffix reports whether s is a street suffix, abbreviated or spelled
// out, so a one-letter part right after it can be read as an alphabet
// indicator rather than a directional (p.17).
func isStreetSuffix(s string) bool {
	_, err := streetsuffixes.NormalizeStreetSuffix(s)
	return err == nil
}

// onlyDirectionsFollow reports whether every part after the first is a
// direction.
//
// It is what separates a saint from an absorbed suffix at the head of a street
// name. SAINT CLAIR is a street name and STREET CLAIR is not, which is why the
// city table is consulted there at all (#114); but ST NW is a suffix and a
// trailing direction that a reading put inside the name, and SAINT NORTHWEST
// is not a street anyone lives on. A direction cannot be the name a saint is
// named for, so what follows the abbreviation is enough to tell the two apart.
func onlyDirectionsFollow(parts []string) bool {
	if len(parts) < 2 {
		return false
	}

	for _, p := range parts[1:] {
		if _, err := directionals.NormalizeDirectional(p); err != nil {
			return false
		}
	}

	return true
}
