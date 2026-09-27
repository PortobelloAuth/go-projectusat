package normalizer

import (
	"fmt"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/cityabbreviations"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

type NormalizationStatus struct {
	Done  bool
	Error error
}

// An AddressNormalizationFn takes an Address and normalizes some aspect of it
type AddressNormalizationFn func(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus)

func ComposeNormalizationFn(fns ...AddressNormalizationFn) AddressNormalizationFn {
	return func(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
		out := a.Clone()
		var status *NormalizationStatus
		for _, fn := range fns {
			out, status = fn(out, o)
			if status != nil {
				if status.Error != nil {
					return nil, status
				}
				if status.Done {
					return out, status
				}
			}
		}

		return out, &NormalizationStatus{
			Done: false,
		}
	}
}

/*
Groupings for Normalization

- LastLine (City, Region, Postal and Country; everything will use this)
- OtherParts (Detail, Area, BusinessName - they don’t really go together, but rarely appear separately)
- StreetLine (most address types will customize; see steps for street name normalization)
- Secondary (Designator and Number go together; technically part of StreetLine)
- Directionals (go together; technically part of StreetLine)
- PrimaryStreetLine (just Primary and StreetName, the core of StreetLine)

Consumers will likely use LastLine, Other (optionally), and either StreetLine or the compositions that
make it up
*/

var normalizeAddressFn = ComposeNormalizationFn(
	NormalizeLastLine,
	NormalizeOtherParts,
	NormalizeStreetLine,
)

var NormalizeLastLine = ComposeNormalizationFn(
	NormalizeCityFn,
	NormalizeRegionFn,
	NormalizePostalFn,
	NormalizeCountryFn,
)

var NormalizeOtherParts = ComposeNormalizationFn(
	NormalizeAreaFn,
	NormalizeBusinessNameFn,
	NormalizeDetailFn,
)

var NormalizeDirectionals = ComposeNormalizationFn(
	NormalizePredirectionalFn,
	NormalizePostdirectionalFn,
)

var NormalizeStreetLine = ComposeNormalizationFn(
	NormalizePrimaryNumberFn,
	NormalizeStreetNameFn,
	NormalizeStreetSuffixFn,
	NormalizeSecondaryNumberFn,
	NormalizeSecondaryDesignatorFn,
	NormalizeDirectionals,
)

func NormalizeBusinessNameFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.BusinessName) > 0 {
		out, err := textutil.FreeTextField(a.BusinessName, o.DiacriticMode)

		if err != nil {
			return nil, &NormalizationStatus{
				Error: fmt.Errorf("business name: %w", err),
			}
		}

		a.BusinessName = out
	}
	return a, nil
}

func NormalizeAreaFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.Area) > 0 {
		out, err := textutil.FreeTextField(a.Area, o.DiacriticMode)
		if err != nil {
			return nil, &NormalizationStatus{
				Error: fmt.Errorf("area: %w", err),
			}
		}

		a.Area = out
	}
	return a, nil
}

func NormalizeDetailFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.Detail) > 0 {
		out, err := textutil.FreeTextField(a.Detail, o.DiacriticMode)
		if err != nil {
			return nil, &NormalizationStatus{
				Error: fmt.Errorf("detail: %w", err),
			}
		}
		a.Detail = out
	}
	return a, nil
}

func NormalizeStreetNameFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.StreetName) > 0 {
		// TODO: break NormalizeStreetName up in to composable
		out, err := NormalizeStreetName(a.StreetName, o)
		if err != nil {
			return nil, &NormalizationStatus{
				// Error: fmt.Errorf("street name: %w", err),
				Error: err,
			}
		}
		a.StreetName = out
	}
	return a, nil
}

func NormalizePrimaryNumberFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.PrimaryNumber) > 0 {
		out, err := textutil.FreeTextField(a.PrimaryNumber, o.DiacriticMode)
		if err != nil {
			return nil, &NormalizationStatus{
				Error: fmt.Errorf("primary number: %w", err),
			}
		}
		a.PrimaryNumber = out
	}
	return a, nil
}

func NormalizeSecondaryDesignatorFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	out := a.Clone()
	v := textutil.BaseField(a.SecondaryDesignator)
	if v == "" {
		out.SecondaryDesignator = v
		return out, nil
	}

	info, err := secondaryunit.Info(v)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("secondary designator: %w", err),
		}
	}

	// Only use SecondaryAsHash for Numbered secondary designators
	if o.SecondaryAsHash && info != nil && info.Numbered {
		out.SecondaryDesignator = "#"
		return out, nil
	}

	abbr, err := secondaryunit.Normalize(v)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("secondary designator: %w", err),
		}
	}

	out.SecondaryDesignator = abbr
	return out, nil
}

func NormalizePredirectionalFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	v := textutil.BaseField(a.Predirectional)
	if v == "" {
		a.Predirectional = v
		return a, nil
	}

	abbr, err := directionals.AbbreviateDirectional(v)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("predirectional: %w", err),
		}
	}

	a.Predirectional = abbr
	return a, nil
}

func NormalizePostdirectionalFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	v := textutil.BaseField(a.Postdirectional)
	if v == "" {
		a.Postdirectional = v
		return a, nil
	}

	abbr, err := directionals.AbbreviateDirectional(v)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("postdirectional: %w", err),
		}
	}

	a.Postdirectional = abbr
	return a, nil
}

func NormalizeStreetSuffixFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	v := textutil.BaseField(a.StreetSuffix)
	if v == "" {
		a.StreetSuffix = v
		return a, nil
	}

	var abbr string
	var err error
	if o.Fuzzy {
		abbr, err = streetsuffixes.FuzzyNormalizeStreetSuffixAbreviation(v)
	} else {
		abbr, err = streetsuffixes.NormalizeStreetSuffixAbreviation(v)
	}
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("street suffix: %w", err),
		}
	}

	a.StreetSuffix = abbr
	return a, nil
}

func NormalizeSecondaryNumberFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.SecondaryNumber) > 0 {
		out, err := textutil.FreeTextField(a.SecondaryNumber, o.DiacriticMode)
		if err != nil {
			return nil, &NormalizationStatus{
				Error: fmt.Errorf("secondary number: %w", err),
			}
		}
		a.SecondaryNumber = out
	}
	return a, nil
}

func NormalizeCityFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.City) == 0 {
		return a, nil
	}

	out, err := textutil.FreeTextField(a.City, o.DiacriticMode)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("city: %w", err),
		}
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
	a.City = out
	return a, nil
}

func NormalizeRegionFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	v := textutil.BaseField(a.Region)
	if v != "" {
		var err error
		if o.Fuzzy {
			v, err = region.FuzzyNormalizeRegion(v)
		} else {
			v, err = region.NormalizeRegion(v)
		}
		if err != nil {
			return nil, &NormalizationStatus{
				Error: fmt.Errorf("region: %w", err),
			}
		}
	}

	a.Region = v
	return a, nil
}

func NormalizePostalFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.Postal) == 0 {
		return a, nil
	}
	out, err := postalcode.Normalize(a.Postal)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("postal code: %w", err),
		}
	}

	a.Postal = out
	return a, nil
}

func NormalizeCountryFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, *NormalizationStatus) {
	if len(a.Country) == 0 {
		return a, nil
	}
	out, err := textutil.FreeTextField(a.Country, o.DiacriticMode)
	if err != nil {
		return nil, &NormalizationStatus{
			Error: fmt.Errorf("country: %w", err),
		}
	}
	a.Country = out
	return a, nil
}

func NormalizeStreetName(streetname string, o AddressNormalizationOptions) (string, error) {
	// TODO: refactor NormalizeStreetName in to composable functions
	sn, err := textutil.FreeTextField(streetname, o.DiacriticMode)
	if err != nil {
		return "", fmt.Errorf("street name: %w", err)
	}

	if sn != "" {
		// if street name has only 1 word, run it through the streetsuffix normalizer;
		// a directional street name is spelled out (NORTH AVE), as one inside a
		// longer name is below. A single letter is left as written instead: it
		// may be an alphabet indicator (1000 G ST, 100 E ST), and the standard says
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
			} else if ss, err := streetsuffixes.NormalizeStreetSuffix(sn); err == nil {
				sn = ss
			}
		} else {
			for i, snp := range snparts {
				// A one-letter final part that follows a street suffix word is
				// an alphabet indicator (AVENUE E), not a direction, and stays
				// as written (p.17). BAY W, where the preceding word is not a
				// suffix, is still a direction and is spelled out (p.18).
				if i == len(snparts)-1 && len(snp) == 1 && IsStreetSuffix(snparts[i-1]) {
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
					if i == 0 && !OnlyDirectionsFollow(snparts) {
						if full, err := cityabbreviations.Expand(snp); err == nil {
							snparts[i] = full
							continue
						}
					}

					if fullss, err := streetsuffixes.NormalizeStreetSuffix(snp); err == nil {
						snparts[i] = fullss
					}
				}
			}
			sn = strings.Join(snparts, " ")
		}

		// Highway forms normalize. An error means the name is not a highway, which
		// is the ordinary case, so the already uppercased and collapsed name stands.
		hw, err := highways.NormalizeStreetName(sn)
		if err == nil {
			return hw, nil
		}
	}
	return sn, nil
}

// Support functions for NormalizeStreetName()

// IsStreetSuffix reports whether s is a street suffix, abbreviated or spelled
// out, so a one-letter part right after it can be read as an alphabet
// indicator rather than a directional (p.17).
func IsStreetSuffix(s string) bool {
	_, err := streetsuffixes.NormalizeStreetSuffix(s)
	return err == nil
}

// OnlyDirectionsFollow reports whether every part after the first is a
// direction.
//
// It is what separates a saint from an absorbed suffix at the head of a street
// name. SAINT CLAIR is a street name and STREET CLAIR is not, which is why the
// city table is consulted there at all (#114); but ST NW is a suffix and a
// trailing direction that a reading put inside the name, and SAINT NORTHWEST
// is not a street anyone lives on. A direction cannot be the name a saint is
// named for, so what follows the abbreviation is enough to tell the two apart.
func OnlyDirectionsFollow(parts []string) bool {
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
