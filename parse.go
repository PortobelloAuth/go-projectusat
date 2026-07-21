package goprojectusat

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// Parse converts free-text multi-line or comma-separated address text into a
// structured Address. It does not call Normalize; compose Normalize(Parse(raw))
// for content form. Empty input returns an error.
func Parse(raw string) (Address, error) {
	lines := splitAddressLines(raw)
	if len(lines) == 0 {
		return Address{}, fmt.Errorf("empty address")
	}

	// Military fast path: last line military last-line AND some earlier line military street.
	if len(lines) >= 2 {
		if city, reg, zip, err := military.NormalizeLastLine(lines[len(lines)-1]); err == nil {
			// Prefer the line immediately before last as street; else scan.
			for i := len(lines) - 2; i >= 0; i-- {
				if street, err := military.NormalizeStreetLine(lines[i]); err == nil {
					var business string
					if i > 0 {
						business = strings.Join(lines[:i], " ")
					}
					return Address{
						BusinessName: business,
						StreetName:   street,
						City:         city,
						Region:       reg,
						Postal:       zip,
					}, nil
				}
			}
		}
	}

	// Single-line military "STREET, LAST" already split by splitAddressLines into 2 lines via comma.
	// Fall through to Task 3/4 for non-military.
	return parseCivilian(lines)
}

// splitAddressLines splits on newlines, then on commas within lines, collapses
// space per segment, drops empties. Uppercases lightly via CollapseSpace only
// (leave casing to Normalize) — actually for military helpers we upper inside
// military package; keep segments trimmed collapsed.
func splitAddressLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		// Comma-separated segments within a line become separate logical lines.
		for _, part := range strings.Split(line, ",") {
			part = textutil.CollapseSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func parseCivilian(lines []string) (Address, error) {
	if len(lines) == 1 {
		// Single segment: peel trailing last-line tokens; remainder is street.
		return parseSingleLineCivilian(lines[0])
	}
	city, reg, zip, err := parseLastLine(lines[len(lines)-1])
	if err != nil {
		return Address{}, err
	}
	streetLine := lines[len(lines)-2]
	business := ""
	if len(lines) > 2 {
		business = strings.ToUpper(textutil.CollapseSpace(strings.Join(lines[:len(lines)-2], " ")))
	}
	street, err := parseStreetLine(streetLine)
	if err != nil {
		return Address{}, err
	}
	street.BusinessName = business
	street.City = city
	street.Region = reg
	street.Postal = zip
	return street, nil
}

// parseLastLine extracts city, region (abbreviated), and postal from a last line
// like "SPRINGFIELD IL 62701" or "OTTAWA ON K1A 0B1". City is multi-word capable.
func parseLastLine(line string) (city, reg, postal string, err error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	if len(tokens) < 2 {
		return "", "", "", fmt.Errorf("invalid last line: %q", line)
	}
	postal, rest, ok := peelPostal(tokens)
	if !ok {
		return "", "", "", fmt.Errorf("invalid last line postal: %q", line)
	}
	if len(rest) < 1 {
		return "", "", "", fmt.Errorf("invalid last line: missing region in %q", line)
	}
	// Longest region match from the right of rest (e.g. DISTRICT OF COLUMBIA).
	for n := len(rest); n >= 1; n-- {
		cand := strings.Join(rest[len(rest)-n:], " ")
		if abbr, e := region.NormalizeRegion(cand); e == nil {
			city = strings.Join(rest[:len(rest)-n], " ")
			if city == "" {
				return "", "", "", fmt.Errorf("invalid last line: missing city in %q", line)
			}
			return city, abbr, postal, nil
		}
	}
	return "", "", "", fmt.Errorf("unrecognized region in last line: %q", line)
}

// peelPostal removes US ZIP / ZIP+4 or Canadian postal from the right of tokens.
func peelPostal(tokens []string) (postal string, rest []string, ok bool) {
	if len(tokens) == 0 {
		return "", nil, false
	}
	last := tokens[len(tokens)-1]

	// ##### or #####-#### (and compact #########)
	if usZIPCompact.MatchString(last) {
		return normalizePostal(last), tokens[:len(tokens)-1], true
	}

	if len(tokens) >= 2 {
		prev := tokens[len(tokens)-2]
		// Two-token US ZIP+4: 62701 1234
		if usZIPCompact.MatchString(prev+last) || usZIPCompact.MatchString(prev+"-"+last) {
			return normalizePostal(prev + "-" + last), tokens[:len(tokens)-2], true
		}
		// Canadian two-token: K1A 0B1
		two := prev + " " + last
		np := normalizePostal(two)
		if len(np) >= 6 && containsLetter(np) {
			return np, tokens[:len(tokens)-2], true
		}
	}

	// Single alphanumeric postal fallback
	return normalizePostal(last), tokens[:len(tokens)-1], true
}

func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// parseSingleLineCivilian peels postal + region + single-token city from the
// right; remainder is componentized as a street line. Multi-word cities require
// multi-line form.
func parseSingleLineCivilian(line string) (Address, error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	postal, rest, ok := peelPostal(tokens)
	if !ok || len(rest) < 2 {
		return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
	}
	// Region up to 3 tokens e.g. DISTRICT OF COLUMBIA
	for n := min(3, len(rest)); n >= 1; n-- {
		cand := strings.Join(rest[len(rest)-n:], " ")
		if abbr, e := region.NormalizeRegion(cand); e == nil {
			before := rest[:len(rest)-n]
			if len(before) < 2 {
				return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
			}
			// Conservative: last token of before = city; rest = street.
			city := before[len(before)-1]
			streetToks := before[:len(before)-1]
			street, err := parseStreetLine(strings.Join(streetToks, " "))
			if err != nil {
				return Address{}, err
			}
			street.City = city
			street.Region = abbr
			street.Postal = postal
			return street, nil
		}
	}
	return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
}

// parseStreetLine reverse-token peels secondary, postdirectional, and suffix,
// then peels primary number and predirectional from the left. Residual tokens
// form StreetName (with highway rewrite when applicable).
//
// Rural route and PO Box free-text lines are rewritten first (see
// rewriteSpecialStreetLine) and stored wholly in StreetName, similar to
// overseas military street lines.
func parseStreetLine(line string) (Address, error) {
	if rewritten, ok := rewriteSpecialStreetLine(line); ok {
		return Address{StreetName: rewritten}, nil
	}

	cleaned := strings.ToUpper(textutil.CollapseSpace(
		textutil.StripPunctuation(line, textutil.StripOptions{KeepHyphen: true, KeepSlash: true}),
	))
	tokens := expandHashTokens(strings.Fields(cleaned))
	if len(tokens) == 0 {
		return Address{}, fmt.Errorf("empty street line")
	}
	// Merge multi-token directionals (SOUTH WEST → SW) before peels so pre-
	// and postdirectionals resolve as single compound abbreviations.
	tokens = mergeDirectionTokens(tokens)

	// Move leading secondary designator/# + number to the end so reverse peels
	// see them as trailing (e.g. "APT 4 123 MAIN ST" → "123 MAIN ST APT 4").
	tokens = reorderLeadingSecondary(tokens)

	var out Address
	tokens = peelSecondary(tokens, &out)

	// Postdirectional (right)
	if len(tokens) > 1 {
		if abbr, err := directionals.AbbreviateDirectional(tokens[len(tokens)-1]); err == nil {
			out.Postdirectional = abbr
			tokens = tokens[:len(tokens)-1]
		}
	}

	// Street suffix (right) — only if something remains for the street body.
	// Peel exactly one suffix; a second trailing suffix stays in the name
	// (expanded to primary form below) so e.g. "Main Avenue Drive" → name MAIN AVENUE + DR.
	if len(tokens) >= 2 {
		if abbr, err := streetsuffixes.NormalizeStreetSuffixAbreviation(tokens[len(tokens)-1]); err == nil {
			out.StreetSuffix = abbr
			tokens = tokens[:len(tokens)-1]
		}
	}

	// Primary number (left)
	if len(tokens) > 0 && looksLikePrimaryNumber(tokens[0]) {
		out.PrimaryNumber = tokens[0]
		tokens = tokens[1:]
	}

	// Predirectional (left) only when a street-name token remains after it.
	if len(tokens) > 1 {
		if abbr, err := directionals.AbbreviateDirectional(tokens[0]); err == nil {
			out.Predirectional = abbr
			tokens = tokens[1:]
		}
	}

	if len(tokens) == 0 {
		return Address{}, fmt.Errorf("unrecognized street line: %q", line)
	}

	// Double-suffix: after peeling one StreetSuffix, if the last remaining name
	// token is also a street suffix and a name body remains before it, keep it
	// in StreetName as the spelled-out primary form (do not peel a second suffix).
	if out.StreetSuffix != "" && len(tokens) >= 2 {
		last := tokens[len(tokens)-1]
		if primary, err := streetsuffixes.NormalizeStreetSuffix(last); err == nil {
			tokens[len(tokens)-1] = primary
		}
	}

	name := strings.Join(tokens, " ")
	// highways.NormalizeStreetName always succeeds for non-empty input: it
	// rewrites highway forms and otherwise returns the uppercased pass-through.
	// Prefer a highway rewrite over state-as-street-name (e.g. "TN 431" → "TN HIGHWAY 431").
	hw, hwErr := highways.NormalizeStreetName(name)
	if hwErr == nil && hw != name {
		out.StreetName = hw
	} else if out.StreetSuffix != "" {
		// When a US state name/abbrev is the entire street name and a suffix is set,
		// spell the state out fully (OK AVE → OKLAHOMA + AVE; CT CT → CONNECTICUT + CT).
		if full, ok := fullySpelledUSState(name); ok {
			out.StreetName = full
		} else if hwErr == nil {
			out.StreetName = hw
		} else {
			out.StreetName = name
		}
	} else if hwErr == nil {
		out.StreetName = hw
	} else {
		out.StreetName = name
	}
	return out, nil
}

// usStateFullNames maps US state/possession two-letter codes to their fully
// spelled primary names. Excludes military AE/AP/AA and Canadian provinces.
var usStateFullNames = map[string]string{
	"AL": "ALABAMA",
	"AK": "ALASKA",
	"AS": "AMERICAN SAMOA",
	"AZ": "ARIZONA",
	"AR": "ARKANSAS",
	"CA": "CALIFORNIA",
	"CO": "COLORADO",
	"CT": "CONNECTICUT",
	"DE": "DELAWARE",
	"DC": "DISTRICT OF COLUMBIA",
	"FM": "FEDERATED STATES OF MICRONESIA",
	"FL": "FLORIDA",
	"GA": "GEORGIA",
	"GU": "GUAM",
	"HI": "HAWAII",
	"ID": "IDAHO",
	"IL": "ILLINOIS",
	"IN": "INDIANA",
	"IA": "IOWA",
	"KS": "KANSAS",
	"KY": "KENTUCKY",
	"LA": "LOUISIANA",
	"ME": "MAINE",
	"MH": "MARSHALL ISLANDS",
	"MD": "MARYLAND",
	"MA": "MASSACHUSETTS",
	"MI": "MICHIGAN",
	"MN": "MINNESOTA",
	"MS": "MISSISSIPPI",
	"MO": "MISSOURI",
	"MT": "MONTANA",
	"NE": "NEBRASKA",
	"NV": "NEVADA",
	"NH": "NEW HAMPSHIRE",
	"NJ": "NEW JERSEY",
	"NM": "NEW MEXICO",
	"NY": "NEW YORK",
	"NC": "NORTH CAROLINA",
	"ND": "NORTH DAKOTA",
	"MP": "NORTHERN MARIANA ISLANDS",
	"OH": "OHIO",
	"OK": "OKLAHOMA",
	"OR": "OREGON",
	"PW": "PALAU",
	"PA": "PENNSYLVANIA",
	"PR": "PUERTO RICO",
	"RI": "RHODE ISLAND",
	"SC": "SOUTH CAROLINA",
	"SD": "SOUTH DAKOTA",
	"TN": "TENNESSEE",
	"TX": "TEXAS",
	"UT": "UTAH",
	"VT": "VERMONT",
	"VI": "VIRGIN ISLANDS",
	"VA": "VIRGINIA",
	"WA": "WASHINGTON",
	"WV": "WEST VIRGINIA",
	"WI": "WISCONSIN",
	"WY": "WYOMING",
}

// fullySpelledUSState returns the fully spelled US state/possession name when
// name (possibly multi-word) normalizes to a known US region code. Military
// and Canadian codes are excluded.
func fullySpelledUSState(name string) (string, bool) {
	abbr, err := region.NormalizeRegion(name)
	if err != nil {
		return "", false
	}
	full, ok := usStateFullNames[abbr]
	return full, ok
}

// expandHashTokens splits glued forms like "#12" into "#", "12".
func expandHashTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		if strings.HasPrefix(t, "#") && len(t) > 1 {
			out = append(out, "#", t[1:])
			continue
		}
		out = append(out, t)
	}
	return out
}

// reorderLeadingSecondary moves a leading secondary designator (or #) plus its
// unit number to the end of the token slice so reverse-token peels can capture
// them. Patterns:
//
//	DESIGNATOR + NUMBER + rest  →  rest + DESIGNATOR + NUMBER
//	# + NUMBER + rest           →  rest + # + NUMBER
//
// Glued "#NUMBER" is already split by expandHashTokens. Numbered designators
// without a following number are left unchanged (no invented unit number).
// Non-numbered designators at the start are not reordered.
func reorderLeadingSecondary(tokens []string) []string {
	if len(tokens) < 3 {
		// Need designator/number plus at least one rest token to move.
		return tokens
	}

	// "# NUMBER rest…"
	if tokens[0] == "#" && looksLikeSecondaryNumber(tokens[1]) {
		rest := tokens[2:]
		out := make([]string, 0, len(tokens))
		out = append(out, rest...)
		out = append(out, "#", tokens[1])
		return out
	}

	// "APT NUMBER rest…" (numbered secondary designator only)
	if info, err := secondaryunit.Info(tokens[0]); err == nil && info.Numbered {
		if looksLikeSecondaryNumber(tokens[1]) {
			rest := tokens[2:]
			out := make([]string, 0, len(tokens))
			out = append(out, rest...)
			out = append(out, tokens[0], tokens[1])
			return out
		}
		// Numbered designator with no unit number — do not invent.
	}

	return tokens
}

// looksLikeSecondaryNumber reports whether tok is a plausible unit number
// (contains a digit). Pure alpha tokens are not treated as unit numbers so
// "APT SOUTH …" is not reordered.
func looksLikeSecondaryNumber(tok string) bool {
	return looksLikePrimaryNumber(tok)
}

// peelSecondary removes a trailing secondary designator and optional number.
func peelSecondary(tokens []string, out *Address) []string {
	if len(tokens) == 0 {
		return tokens
	}

	// "# 12" or "#12" (already expanded)
	if len(tokens) >= 2 && tokens[len(tokens)-2] == "#" {
		out.SecondaryDesignator = "#"
		out.SecondaryNumber = tokens[len(tokens)-1]
		return peelTrailingNonNumberedSecondary(tokens[:len(tokens)-2], out)
	}

	// Numbered designator + unit number (e.g. APT 4). Overseas military
	// "UNIT N BOX N" is handled earlier by the military fast path, so this
	// peel only sees civilian street lines.
	if len(tokens) >= 2 {
		if info, err := secondaryunit.Info(tokens[len(tokens)-2]); err == nil && info.Numbered {
			out.SecondaryDesignator = info.Short
			out.SecondaryNumber = tokens[len(tokens)-1]
			return peelTrailingNonNumberedSecondary(tokens[:len(tokens)-2], out)
		}
	}

	// Designator alone (non-numbered, or numbered without a number)
	if info, err := secondaryunit.Info(tokens[len(tokens)-1]); err == nil {
		out.SecondaryDesignator = info.Short
		return tokens[:len(tokens)-1]
	}

	return tokens
}

// peelTrailingNonNumberedSecondary peels one more trailing non-numbered
// secondary designator (e.g. UPPER/REAR) after a numbered secondary was taken.
// Appended to SecondaryNumber so Format yields "UNIT 3200 UPPR".
func peelTrailingNonNumberedSecondary(tokens []string, out *Address) []string {
	if len(tokens) == 0 {
		return tokens
	}
	if info, err := secondaryunit.Info(tokens[len(tokens)-1]); err == nil && !info.Numbered {
		if out.SecondaryNumber != "" {
			out.SecondaryNumber = out.SecondaryNumber + " " + info.Short
		} else if out.SecondaryDesignator != "" {
			out.SecondaryDesignator = out.SecondaryDesignator + " " + info.Short
		} else {
			out.SecondaryDesignator = info.Short
		}
		return tokens[:len(tokens)-1]
	}
	return tokens
}

// looksLikePrimaryNumber reports whether tok is a plausible primary address number
// (contains a digit; hyphenated ranges like 112-10 qualify).
func looksLikePrimaryNumber(tok string) bool {
	for _, r := range tok {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
