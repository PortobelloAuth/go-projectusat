// Package goprojectusat normalizes structured patient addresses per Project US@.
//
// Unknown fields are blank. The token UNKNOWN (any case) is treated as empty.
// Alphabetical content is uppercased. Diacritics are not stripped by default;
// callers may pre-run diacritics.Substitute when mapping is required, or set
// Options.DiacriticMode on NormalizeWithOptions for exchange/matching.
package goprojectusat

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
// Zero value is content form (same as Normalize).
type AddressNormalizationOptions struct {
	// Fuzzy enables FuzzyNormalize* for region and street suffix.
	Fuzzy bool
	// SecondaryAsHash rewrites secondary designators to "#" for matching
	// (not correct for content storage; for exchange/matching only).
	SecondaryAsHash bool
	// DiacriticMode: "" = leave as-is, "substitute" = diacritics.Substitute then upper,
	// "transliterate" = diacritics.Transliterate (anyascii) then upper.
	DiacriticMode diacritics.DiacriticMode
}

// Normalize returns a content-normalized copy (uppercase, standard abbreviations).
// Diacritics are preserved; callers may pre-run diacritics.Substitute if needed.
// Empty optional fields stay blank. Unrecognized non-empty controlled vocabulary
// (region, directionals, street suffix, secondary designator) returns an error.
//
// Equivalent to NormalizeWithOptions(a, Options{}).
func Normalize(a address.Address) (address.Address, error) {
	return NormalizeWithOptions(a, AddressNormalizationOptions{})
}

// NormalizeWithOptions is like Normalize but applies exchange/matching options.
// Use for patient matching / comparison; prefer Normalize for content storage.
func NormalizeWithOptions(a address.Address, opts AddressNormalizationOptions) (address.Address, error) {
	var out address.Address

	var err error
	if out.BusinessName, err = textutil.FreeTextField(a.BusinessName, opts.DiacriticMode); err != nil {
		return address.Address{}, fmt.Errorf("business name: %w", err)
	}
	if out.PrimaryNumber, err = textutil.FreeTextField(a.PrimaryNumber, opts.DiacriticMode); err != nil {
		return address.Address{}, fmt.Errorf("primary number: %w", err)
	}
	if out.SecondaryNumber, err = textutil.FreeTextField(a.SecondaryNumber, opts.DiacriticMode); err != nil {
		return address.Address{}, fmt.Errorf("secondary number: %w", err)
	}
	if out.City, err = textutil.FreeTextField(a.City, opts.DiacriticMode); err != nil {
		return address.Address{}, fmt.Errorf("city: %w", err)
	}
	if out.Country, err = textutil.FreeTextField(a.Country, opts.DiacriticMode); err != nil {
		return address.Address{}, fmt.Errorf("country: %w", err)
	}
	out.Postal = postalcode.Normalize(a.Postal)

	if sn, err := textutil.FreeTextField(a.StreetName, opts.DiacriticMode); err != nil {
		return address.Address{}, fmt.Errorf("street name: %w", err)
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
			return address.Address{}, fmt.Errorf("predirectional: %w", err)
		}
		out.Predirectional = abbr
	}

	if v := textutil.BaseField(a.Postdirectional); v != "" {
		abbr, err := directionals.AbbreviateDirectional(v)
		if err != nil {
			return address.Address{}, fmt.Errorf("postdirectional: %w", err)
		}
		out.Postdirectional = abbr
	}

	if v := textutil.BaseField(a.StreetSuffix); v != "" {
		var abbr string
		var err error
		if opts.Fuzzy {
			abbr, err = streetsuffixes.FuzzyNormalizeStreetSuffixAbreviation(v)
		} else {
			abbr, err = streetsuffixes.NormalizeStreetSuffixAbreviation(v)
		}
		if err != nil {
			return address.Address{}, fmt.Errorf("street suffix: %w", err)
		}
		out.StreetSuffix = abbr
	}

	if v := textutil.BaseField(a.SecondaryDesignator); v != "" {
		abbr, err := secondaryunit.Normalize(v)
		if err != nil {
			return address.Address{}, fmt.Errorf("secondary designator: %w", err)
		}
		if opts.SecondaryAsHash {
			out.SecondaryDesignator = "#"
		} else {
			out.SecondaryDesignator = abbr
		}
	}

	if v := textutil.BaseField(a.Region); v != "" {
		var abbr string
		var err error
		if opts.Fuzzy {
			abbr, err = region.FuzzyNormalizeRegion(v)
		} else {
			abbr, err = region.NormalizeRegion(v)
		}
		if err != nil {
			return address.Address{}, fmt.Errorf("region: %w", err)
		}
		out.Region = abbr
	}

	return out, nil
}
