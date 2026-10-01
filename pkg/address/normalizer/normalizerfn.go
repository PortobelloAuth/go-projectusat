package normalizer

import (
	"errors"
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

// Done exists as a sentinel error to allow for early returns from complex compositions
// of normalization functions without requiring a special status type. In that regard it
// is somewhat analogous to EOF for IO streams.
var Done = fmt.Errorf("normalization complete")

// An AddressNormalizationFn takes an Address and normalizes some aspect of it
type AddressNormalizationFn func(a *address.Address, o AddressNormalizationOptions) (*address.Address, error)

func ComposeNormalizationFn(fns ...AddressNormalizationFn) AddressNormalizationFn {
	return func(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
		out := a.Clone()
		var err error
		for _, fn := range fns {
			out, err = fn(out, o)
			if err != nil {
				if !errors.Is(err, Done) {
					return nil, err
				}
				return out, nil
			}
		}

		return out, nil
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

func NormalizeBusinessNameFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.BusinessName) > 0 {
		out, err := textutil.FreeTextField(a.BusinessName, o.DiacriticMode)

		if err != nil {
			return nil, fmt.Errorf("business name: %w", err)
		}

		a.BusinessName = out
	}
	return a, nil
}

func NormalizeAreaFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.Area) > 0 {
		out, err := textutil.FreeTextField(a.Area, o.DiacriticMode)
		if err != nil {
			return nil, fmt.Errorf("area: %w", err)
		}

		a.Area = out
	}
	return a, nil
}

func NormalizeDetailFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.Detail) > 0 {
		out, err := textutil.FreeTextField(a.Detail, o.DiacriticMode)
		if err != nil {
			return nil, fmt.Errorf("detail: %w", err)
		}
		a.Detail = out
	}
	return a, nil
}

func NormalizeStreetNameFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.StreetName) > 0 {
		// TODO: break NormalizeStreetName up in to composable
		out, err := NormalizeStreetName(a.StreetName, o)
		if err != nil {
			return nil, err
		}
		a.StreetName = out
	}
	return a, nil
}

func NormalizePrimaryNumberFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.PrimaryNumber) > 0 {
		out, err := textutil.FreeTextField(a.PrimaryNumber, o.DiacriticMode)
		if err != nil {
			return nil, fmt.Errorf("primary number: %w", err)
		}
		a.PrimaryNumber = out
	}
	return a, nil
}

func NormalizeSecondaryDesignatorFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	out := a.Clone()
	v := textutil.BaseField(a.SecondaryDesignator)
	if v == "" {
		out.SecondaryDesignator = v
		return out, nil
	}

	info, err := secondaryunit.Info(v)
	if err != nil {
		return nil, fmt.Errorf("secondary designator: %w", err)
	}

	// Only use SecondaryAsHash for Numbered secondary designators
	if o.SecondaryAsHash && info != nil && info.Numbered {
		out.SecondaryDesignator = "#"
		return out, nil
	}

	abbr, err := secondaryunit.Normalize(v)
	if err != nil {
		return nil, fmt.Errorf("secondary designator: %w", err)
	}

	out.SecondaryDesignator = abbr
	return out, nil
}

func NormalizePredirectionalFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	v := textutil.BaseField(a.Predirectional)
	if v == "" {
		a.Predirectional = v
		return a, nil
	}

	abbr, err := directionals.AbbreviateDirectional(v)
	if err != nil {
		return nil, fmt.Errorf("predirectional: %w", err)
	}

	a.Predirectional = abbr
	return a, nil
}

func NormalizePostdirectionalFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	v := textutil.BaseField(a.Postdirectional)
	if v == "" {
		a.Postdirectional = v
		return a, nil
	}

	abbr, err := directionals.AbbreviateDirectional(v)
	if err != nil {
		return nil, fmt.Errorf("postdirectional: %w", err)
	}

	a.Postdirectional = abbr
	return a, nil
}

func NormalizeStreetSuffixFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
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
		return nil, fmt.Errorf("street suffix: %w", err)
	}

	a.StreetSuffix = abbr
	return a, nil
}

func NormalizeSecondaryNumberFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.SecondaryNumber) > 0 {
		out, err := textutil.FreeTextField(a.SecondaryNumber, o.DiacriticMode)
		if err != nil {
			return nil, fmt.Errorf("secondary number: %w", err)
		}
		a.SecondaryNumber = out
	}
	return a, nil
}

func NormalizeCityFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.City) == 0 {
		return a, nil
	}

	out, err := textutil.FreeTextField(a.City, o.DiacriticMode)
	if err != nil {
		return nil, fmt.Errorf("city: %w", err)
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

func NormalizeRegionFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	v := textutil.BaseField(a.Region)
	if v != "" {
		var err error
		if o.Fuzzy {
			v, err = region.FuzzyNormalizeRegion(v)
		} else {
			v, err = region.NormalizeRegion(v)
		}
		if err != nil {
			return nil, fmt.Errorf("region: %w", err)
		}
	}

	a.Region = v
	return a, nil
}

func NormalizePostalFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.Postal) == 0 {
		return a, nil
	}
	out, err := postalcode.Normalize(a.Postal)
	if err != nil {
		return nil, fmt.Errorf("postal code: %w", err)
	}

	a.Postal = out
	return a, nil
}

func NormalizeCountryFn(a *address.Address, o AddressNormalizationOptions) (*address.Address, error) {
	if len(a.Country) == 0 {
		return a, nil
	}
	out, err := textutil.FreeTextField(a.Country, o.DiacriticMode)
	if err != nil {
		return nil, fmt.Errorf("country: %w", err)
	}
	a.Country = out
	return a, nil
}

// Steps for street name normalization
// - uppercase
// - diacritic fold
// - UNKNOWN removal
// - excess whitespace removal
// - early return for empty street name (after whitespace & unknown removal)
// - single-word region replacement (prefers full state names)
// - early return for single letter street name
// - single word directional abbreviation expansion (within street name, not directional fields)
// - street suffix expansion (within street name, not street suffix field)
// - “suffix-first” single letter preference (AVENUE D)
// - directional abbreviation expansion (within street name, not directional fields)
// - region name abbreviation (when other words are also present)
// - city name expansion (Saint, Sainte, Fort, Mount)
// - street suffix expansion (within street name, not street suffix field)
// - highway street name normalization

type StringNormalizationFn func(sn string, o AddressNormalizationOptions) (string, error)
type StreetNameNormalizationFn StringNormalizationFn

func ComposeStreetNameNormalizationFn(fns ...StreetNameNormalizationFn) StreetNameNormalizationFn {
	return func(sn string, o AddressNormalizationOptions) (string, error) {
		out := sn[:]
		var err error
		for _, fn := range fns {
			out, err = fn(out, o)
			if err != nil {
				if !errors.Is(err, Done) {
					return "", err
				}
				return out, nil
			}
		}

		return out, nil
	}
}

func NormalizeTextFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - uppercase
	// - diacritic fold
	// - UNKNOWN removal
	// - excess whitespace removal

	out, err := textutil.FreeTextField(sn, o.DiacriticMode)
	if err != nil {
		return "", err
	}

	// - early return for empty street name (after whitespace & unknown removal)
	if out == "" {
		return "", Done
	}
	return out, nil
}

func OnlyRegionStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - single-word region replacement (prefers full state names)
	regioninfo, _ := region.Info(sn, false)
	if regioninfo != nil && regioninfo.PossibleStreetName {
		// If we substitute the whole street name with the full region name, we're done
		return regioninfo.Primary, Done
	}

	return sn, nil
}

func OnlySingleLetterStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - early return for single letter street name
	if len(sn) == 1 {
		return sn, Done
	}

	return sn, nil
}

func PrefixAndSingleLetterStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - “suffix-first” single letter preference (AVENUE D)
	parts := strings.Split(sn, " ")
	if len(parts) == 2 && len(parts[1]) == 1 {
		fullprefix, err := streetsuffixes.NormalizeStreetSuffix(parts[0])
		if err == nil {
			return fmt.Sprintf("%s %s", strings.ToUpper(fullprefix), parts[1]), Done
		}
	}

	return sn, nil
}

func ExpandCityInStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - city name expansion (Saint, Sainte, Fort, Mount)
	parts := strings.Split(sn, " ")

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
	// TODO: adjust this hueristic is to allow MT ST HELENS ST to become
	// MOUNT SAINT HELENS ST
	if !OnlyDirectionsFollow(parts) {
		if full, err := cityabbreviations.Expand(parts[0]); err == nil {
			parts[0] = full
			return strings.Join(parts, " "), nil
		}
	}

	return sn, nil
}

func ExpandStreetTypeInStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - street suffix expansion (within street name, not street suffix field)
	parts := strings.Split(sn, " ")
	changed := false
	for i, snp := range parts {
		fullsuffix, err := streetsuffixes.NormalizeStreetSuffix(snp)
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

func AbbreviateRegionInStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - region name abbreviation (when other words are also present)
	regioninfo, _ := region.Info(sn, false)
	if regioninfo != nil && regioninfo.PossibleStreetName {
		// If the whole street name is substitutable with the full region name, we should not abbreviate it
		return sn, nil
	}

	parts := strings.Split(sn, " ")
	newparts := make([]string, 0)
	changed := false
	for i := 0; i < len(parts); i++ {
		snp := parts[i]
		for j := len(parts); j > i; j-- {
			set := parts[i:j]
			snphrase := strings.Join(set, " ")

			regioninfo, _ := region.Info(snphrase, false)
			if regioninfo != nil && regioninfo.PossibleStreetName {
				snp = regioninfo.Short
				changed = true

				// jump to j so we don't re-replace what we just replaced
				i = j
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

func ExpandDirectionalsInStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - single word directional abbreviation expansion (within street name, not directional fields)
	// - directional abbreviation expansion (within street name, not directional fields)
	parts := strings.Split(sn, " ")
	newparts := make([]string, 0)
	changed := false
	for i := 0; i < len(parts); i++ {
		snp := parts[i]
		for j := len(parts); j > i; j-- {
			set := parts[i:j]
			snphrase := strings.Join(set, " ")

			full, err := directionals.NormalizeDirectional(snphrase)
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

func NormalizeHighwayStreetNameFn(sn string, o AddressNormalizationOptions) (string, error) {
	// - highway street name normalization
	// An error means the name is not a highway, which is common.
	hw, err := highways.NormalizeStreetName(sn)
	if err == nil {
		// highway normalization takes the whole street name and handles it - including abbreviating
		// state and region names. No further normalization should be required.
		return hw, Done
	}

	return sn, nil
}

var NormalizeStreetName = ComposeStreetNameNormalizationFn(
	NormalizeTextFn,
	OnlySingleLetterStreetNameFn,
	PrefixAndSingleLetterStreetNameFn,
	OnlyRegionStreetNameFn,
	NormalizeHighwayStreetNameFn,

	ExpandDirectionalsInStreetNameFn,
	// Abbreviate region AFTER expanding directionals so that NEBRASKA doesn't get
	// converted to NORTHEAST
	AbbreviateRegionInStreetNameFn,
	ExpandCityInStreetNameFn,
	ExpandStreetTypeInStreetNameFn,
)

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
