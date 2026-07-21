// Package goprojectusat normalizes structured patient addresses per Project US@.
//
// Unknown fields are blank. The token UNKNOWN (any case) is treated as empty.
// Alphabetical content is uppercased. Diacritics are not stripped by default;
// callers may pre-run diacritics.Substitute when mapping is required.
package goprojectusat

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Address struct {
	BusinessName string // firm / business line (optional)

	// Street line elements
	PrimaryNumber       string
	Predirectional      string
	StreetName          string
	StreetSuffix        string
	Postdirectional     string
	SecondaryDesignator string // APT, STE, ...
	SecondaryNumber     string

	// Last line
	City   string
	Region string // state / province / military "state"
	Postal string // ZIP, ZIP+4, or Canadian postal code

	Country string // optional; often blank for domestic
}

// usZIPCompact matches ##### or #####-#### / ######### after punctuation strip.
var usZIPCompact = regexp.MustCompile(`^(\d{5})(?:-?(\d{4}))?$`)

// Normalize returns a content-normalized copy (uppercase, standard abbreviations).
// Diacritics are preserved; callers may pre-run diacritics.Substitute if needed.
// Empty optional fields stay blank. Unrecognized non-empty controlled vocabulary
// (region, directionals, street suffix, secondary designator) returns an error.
func Normalize(a Address) (Address, error) {
	var out Address

	out.BusinessName = baseField(a.BusinessName)
	out.PrimaryNumber = baseField(a.PrimaryNumber)
	out.SecondaryNumber = baseField(a.SecondaryNumber)
	out.City = baseField(a.City)
	out.Country = baseField(a.Country)
	out.Postal = normalizePostal(a.Postal)

	if sn := baseField(a.StreetName); sn != "" {
		// Highway forms normalize; ordinary free-text street names pass through uppercased.
		// On error (e.g. empty after internal trim), keep collapsed uppercase name.
		if hw, err := highways.NormalizeStreetName(sn); err == nil {
			out.StreetName = hw
		} else {
			out.StreetName = sn
		}
	}

	if v := baseField(a.Predirectional); v != "" {
		abbr, err := directionals.AbbreviateDirectional(v)
		if err != nil {
			return Address{}, fmt.Errorf("predirectional: %w", err)
		}
		out.Predirectional = abbr
	}

	if v := baseField(a.Postdirectional); v != "" {
		abbr, err := directionals.AbbreviateDirectional(v)
		if err != nil {
			return Address{}, fmt.Errorf("postdirectional: %w", err)
		}
		out.Postdirectional = abbr
	}

	if v := baseField(a.StreetSuffix); v != "" {
		abbr, err := streetsuffixes.NormalizeStreetSuffixAbreviation(v)
		if err != nil {
			return Address{}, fmt.Errorf("street suffix: %w", err)
		}
		out.StreetSuffix = abbr
	}

	if v := baseField(a.SecondaryDesignator); v != "" {
		abbr, err := secondaryunit.Normalize(v)
		if err != nil {
			return Address{}, fmt.Errorf("secondary designator: %w", err)
		}
		out.SecondaryDesignator = abbr
	}

	if v := baseField(a.Region); v != "" {
		abbr, err := region.NormalizeRegion(v)
		if err != nil {
			return Address{}, fmt.Errorf("region: %w", err)
		}
		out.Region = abbr
	}

	return out, nil
}

// baseField blanks UNKNOWN, collapses whitespace, and uppercases.
func baseField(s string) string {
	return textutil.CollapseSpace(textutil.Upper(s))
}

// normalizePostal formats US ZIP / ZIP+4 and leaves Canadian (and other) patterns
// as uppercase alphanumerics with collapsed spacing.
func normalizePostal(s string) string {
	s = baseField(s)
	if s == "" {
		return ""
	}

	// Keep hyphen for ZIP+4; strip other Project US@ punctuation.
	cleaned := textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{KeepHyphen: true}))
	compact := strings.ReplaceAll(cleaned, " ", "")
	if m := usZIPCompact.FindStringSubmatch(compact); m != nil {
		if m[2] != "" {
			return m[1] + "-" + m[2]
		}
		return m[1]
	}

	// Canadian / other international: uppercase, collapse space, drop punctuation.
	return textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{}))
}
