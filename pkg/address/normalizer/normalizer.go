package normalizer

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

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

// Normalize applies the Normalizer's AddressNormalizationOptions to the Address
func (n *Normalizer) Normalize(a *address.Address) (*address.Address, error) {
	var out address.Address

	var err error
	if out.BusinessName, err = textutil.FreeTextField(a.BusinessName, n.Options.DiacriticMode); err != nil {
		return nil, fmt.Errorf("business name: %w", err)
	}
	if out.PrimaryNumber, err = textutil.FreeTextField(a.PrimaryNumber, n.Options.DiacriticMode); err != nil {
		return nil, fmt.Errorf("primary number: %w", err)
	}
	if out.SecondaryNumber, err = textutil.FreeTextField(a.SecondaryNumber, n.Options.DiacriticMode); err != nil {
		return nil, fmt.Errorf("secondary number: %w", err)
	}
	if out.City, err = textutil.FreeTextField(a.City, n.Options.DiacriticMode); err != nil {
		return nil, fmt.Errorf("city: %w", err)
	}
	if out.Country, err = textutil.FreeTextField(a.Country, n.Options.DiacriticMode); err != nil {
		return nil, fmt.Errorf("country: %w", err)
	}
	if out.Postal, err = postalcode.Normalize(a.Postal); err != nil {
		return nil, fmt.Errorf("postal code: %w", err)
	}

	if sn, err := textutil.FreeTextField(a.StreetName, n.Options.DiacriticMode); err != nil {
		return nil, fmt.Errorf("street name: %w", err)
	} else if sn != "" {
		// Highway forms normalize; ordinary free-text street names pass through uppercased.
		// On error (e.g. empty after internal trim), keep collapsed uppercase name.
		if hw, err := highways.NormalizeStreetName(sn); err == nil {
			out.StreetName = hw
		} else {
			out.StreetName = sn
		}
	}

	if v := textutil.BaseField(a.Predirectional); v != "" {
		abbr, err := directionals.AbbreviateDirectional(v)
		if err != nil {
			return nil, fmt.Errorf("predirectional: %w", err)
		}
		out.Predirectional = abbr
	}

	if v := textutil.BaseField(a.Postdirectional); v != "" {
		abbr, err := directionals.AbbreviateDirectional(v)
		if err != nil {
			return nil, fmt.Errorf("postdirectional: %w", err)
		}
		out.Postdirectional = abbr
	}

	if v := textutil.BaseField(a.StreetSuffix); v != "" {
		var abbr string
		var err error
		if n.Options.Fuzzy {
			abbr, err = streetsuffixes.FuzzyNormalizeStreetSuffixAbreviation(v)
		} else {
			abbr, err = streetsuffixes.NormalizeStreetSuffixAbreviation(v)
		}
		if err != nil {
			return nil, fmt.Errorf("street suffix: %w", err)
		}
		out.StreetSuffix = abbr
	}

	if v := textutil.BaseField(a.SecondaryDesignator); v != "" {
		abbr, err := secondaryunit.Normalize(v)
		if err != nil {
			return nil, fmt.Errorf("secondary designator: %w", err)
		}
		if n.Options.SecondaryAsHash {
			out.SecondaryDesignator = "#"
		} else {
			out.SecondaryDesignator = abbr
		}
	}

	if v := textutil.BaseField(a.Region); v != "" {
		var abbr string
		var err error
		if n.Options.Fuzzy {
			abbr, err = region.FuzzyNormalizeRegion(v)
		} else {
			abbr, err = region.NormalizeRegion(v)
		}
		if err != nil {
			return nil, fmt.Errorf("region: %w", err)
		}
		out.Region = abbr
	}

	return &out, nil
}
