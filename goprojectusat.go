// Package goprojectusat normalizes structured patient addresses per Project US@.
//
// Unknown fields are blank. The token UNKNOWN (any case) is treated as empty.
// Alphabetical content is uppercased. Diacritics are not stripped by default;
// callers may pre-run diacritics.Substitute when mapping is required, or set
// Options.DiacriticMode on NormalizeWithOptions for exchange/matching.
package goprojectusat

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/parse"
	"github.com/PortobelloAuth/go-projectusat/pkg/puertorico"
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

// Options controls exchange/matching variants of normalization.
// Zero value is content form (same as Normalize).
type Options struct {
	// Fuzzy enables FuzzyNormalize* for region and street suffix.
	Fuzzy bool
	// SecondaryAsHash rewrites secondary designators to "#" for matching
	// (not correct for content storage; for exchange/matching only).
	SecondaryAsHash bool
	// DiacriticMode: "" = leave as-is, "substitute" = diacritics.Substitute then upper,
	// "transliterate" = diacritics.Transliterate (anyascii) then upper.
	DiacriticMode string
}

// usZIPCompact matches ##### or #####-#### / ######### after punctuation strip.
var usZIPCompact = regexp.MustCompile(`^(\d{5})(?:-?(\d{4}))?$`)

// Canadian postal: compact A1A1A1, or FSA (A1A) + LDU (1A1) as two tokens.
var (
	caPostalCompact = regexp.MustCompile(`^[A-Z]\d[A-Z]\d[A-Z]\d$`)
	caPostalFSA     = regexp.MustCompile(`^[A-Z]\d[A-Z]$`)
	caPostalLDU     = regexp.MustCompile(`^\d[A-Z]\d$`)
)

// Normalize returns a content-normalized copy (uppercase, standard abbreviations).
// Diacritics are preserved; callers may pre-run diacritics.Substitute if needed.
// Empty optional fields stay blank. Unrecognized non-empty controlled vocabulary
// (region, directionals, street suffix, secondary designator) returns an error.
//
// Equivalent to NormalizeWithOptions(a, Options{}).
func Normalize(a Address) (Address, error) {
	return NormalizeWithOptions(a, Options{})
}

// NormalizeWithOptions is like Normalize but applies exchange/matching options.
// Use for patient matching / comparison; prefer Normalize for content storage.
func NormalizeWithOptions(a Address, opts Options) (Address, error) {
	var out Address

	var err error
	if out.BusinessName, err = freeTextField(a.BusinessName, opts.DiacriticMode); err != nil {
		return Address{}, fmt.Errorf("business name: %w", err)
	}
	if out.City, err = freeTextField(a.City, opts.DiacriticMode); err != nil {
		return Address{}, fmt.Errorf("city: %w", err)
	}
	if out.Country, err = freeTextField(a.Country, opts.DiacriticMode); err != nil {
		return Address{}, fmt.Errorf("country: %w", err)
	}
	out.Postal = normalizePostal(a.Postal)

	// Normalize region early so Puerto Rico street/secondary vocabulary can be
	// applied when region is PR.
	if v := baseField(a.Region); v != "" {
		var abbr string
		if opts.Fuzzy {
			abbr, err = region.FuzzyNormalizeRegion(v)
		} else {
			abbr, err = region.NormalizeRegion(v)
		}
		if err != nil {
			return Address{}, fmt.Errorf("region: %w", err)
		}
		out.Region = abbr
	}
	// Puerto Rico dialect from region and/or PR ZIP ranges.
	usePR := puertorico.UsePRDialect(out.Region, out.Postal)

	// Overseas military street lines (e.g. "PSC 3 BOX 4120") live entirely in
	// StreetName. Detect on the joined candidate; on success skip civilian
	// primary/directional/suffix/secondary controlled-vocab fields.
	streetCandidate := joinNonEmpty(" ",
		baseField(a.PrimaryNumber),
		baseField(a.Predirectional),
		baseField(a.StreetName),
		baseField(a.StreetSuffix),
		baseField(a.Postdirectional),
		baseField(a.SecondaryDesignator),
		baseField(a.SecondaryNumber),
	)
	if mil, milErr := military.NormalizeStreetLine(streetCandidate); milErr == nil {
		out.StreetName = mil
	} else {
		if out.PrimaryNumber, err = freeTextField(a.PrimaryNumber, opts.DiacriticMode); err != nil {
			return Address{}, fmt.Errorf("primary number: %w", err)
		}
		if out.SecondaryNumber, err = freeTextField(a.SecondaryNumber, opts.DiacriticMode); err != nil {
			return Address{}, fmt.Errorf("secondary number: %w", err)
		}

		if sn, err := freeTextField(a.StreetName, opts.DiacriticMode); err != nil {
			return Address{}, fmt.Errorf("street name: %w", err)
		} else if sn != "" {
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
			var abbr string
			var err error
			if opts.Fuzzy {
				abbr, err = streetsuffixes.FuzzyNormalizeStreetSuffixAbreviation(v)
			} else {
				abbr, err = streetsuffixes.NormalizeStreetSuffixAbreviation(v)
			}
			// PR Spanish street types (CLL, CAM, …) are not in the USPS English
			// suffix table; fall back to puertorico when region is PR.
			if err != nil && usePR {
				if a, ok := puertorico.TryAbbreviateStreetType(v); ok {
					abbr, err = a, nil
				}
			}
			if err != nil {
				return Address{}, fmt.Errorf("street suffix: %w", err)
			}
			out.StreetSuffix = abbr
		}

		if v := baseField(a.SecondaryDesignator); v != "" {
			abbr, err := secondaryunit.Normalize(v)
			// PR Spanish secondaries (URB, EDIF, BDA, …) fall back when region is PR.
			if err != nil && usePR {
				if a, ok := puertorico.TryNormalizeSecondary(v); ok {
					abbr, err = a, nil
				}
			}
			if err != nil {
				return Address{}, fmt.Errorf("secondary designator: %w", err)
			}
			if opts.SecondaryAsHash {
				out.SecondaryDesignator = "#"
			} else {
				out.SecondaryDesignator = abbr
			}
		}
	}

	return out, nil
}

// freeTextField collapses whitespace, blanks UNKNOWN, uppercases, then optionally
// applies DiacriticMode and re-uppers (Substitute/Transliterate return lowercase).
func freeTextField(s, diacriticMode string) (string, error) {
	s = baseField(s)
	if s == "" || diacriticMode == "" {
		return s, nil
	}
	var (
		out string
		err error
	)
	switch diacriticMode {
	case "substitute":
		out, err = diacritics.Substitute(s)
	case "transliterate":
		out, err = diacritics.Transliterate(s)
	default:
		return "", fmt.Errorf("unknown DiacriticMode %q (want \"\", \"substitute\", or \"transliterate\")", diacriticMode)
	}
	if err != nil {
		return "", err
	}
	return textutil.Upper(out), nil
}

// baseField collapses whitespace, then blanks UNKNOWN and uppercases.
// Collapse must run before Upper so padded values like " UNKNOWN " become blank
// (Upper/NormalizeUnknown only match the exact token "UNKNOWN").
func baseField(s string) string {
	return textutil.Upper(textutil.CollapseSpace(s))
}

// normalizePostal formats US ZIP / ZIP+4 and Canadian A1A 1A1; other patterns
// remain uppercase alphanumerics with collapsed spacing.
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

	// Canadian: compact or spaced → "A1A 1A1"
	if caPostalCompact.MatchString(compact) {
		return compact[:3] + " " + compact[3:]
	}

	// Other international: uppercase, collapse space, drop punctuation.
	return textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{}))
}

// FormatStreetLine joins street elements with single spaces; omits blanks.
// Order: PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM.
func FormatStreetLine(a Address) string {
	return joinNonEmpty(" ",
		a.PrimaryNumber,
		a.Predirectional,
		a.StreetName,
		a.StreetSuffix,
		a.Postdirectional,
		a.SecondaryDesignator,
		a.SecondaryNumber,
	)
}

// FormatLastLine joins CITY REGION POSTAL with single spaces (1+ spaces allowed
// by standard; we use one). Omits blanks.
func FormatLastLine(a Address) string {
	return joinNonEmpty(" ", a.City, a.Region, a.Postal)
}

// Format returns business line (if any), street line, and last line separated
// by newlines. Empty lines are omitted.
func Format(a Address) string {
	return joinNonEmpty("\n", a.BusinessName, FormatStreetLine(a), FormatLastLine(a))
}

// joinNonEmpty joins non-empty parts with sep.
func joinNonEmpty(sep string, parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

// Parse converts free-text multi-line or comma-separated address text into a
// structured Address via pkg/parse (token scoring + component assignment).
// It does not call Normalize; compose Normalize(Parse(raw)) for content form.
func Parse(raw string) (Address, error) {
	a, err := parse.Parse(raw)
	if err != nil {
		return Address{}, err
	}
	return addressFromParse(a), nil
}

func addressFromParse(a parse.Address) Address {
	return Address{
		BusinessName:        a.BusinessName,
		PrimaryNumber:       a.PrimaryNumber,
		Predirectional:      a.Predirectional,
		StreetName:          a.StreetName,
		StreetSuffix:        a.StreetSuffix,
		Postdirectional:     a.Postdirectional,
		SecondaryDesignator: a.SecondaryDesignator,
		SecondaryNumber:     a.SecondaryNumber,
		City:                a.City,
		Region:              a.Region,
		Postal:              a.Postal,
		Country:             a.Country,
	}
}
