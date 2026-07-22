package goprojectusat

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/puertorico"
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
	street, err := parseStreetLine(streetLine, reg)
	if err != nil {
		return Address{}, err
	}
	// Multi-line business stays; same-line pre-street prefix fills BusinessName
	// when empty, otherwise is prepended with a space.
	street.BusinessName = mergeBusinessName(business, street.BusinessName)
	street.City = city
	street.Region = reg
	street.Postal = zip
	return street, nil
}

// mergeBusinessName combines a multi-line firm/business line with an optional
// same-line pre-street prefix extracted by parseStreetLine.
// multi-line takes precedence as the base; same-line only fills when empty,
// otherwise same-line is prepended: "PREFIX MULTI".
func mergeBusinessName(multiLine, sameLine string) string {
	switch {
	case sameLine == "":
		return multiLine
	case multiLine == "":
		return sameLine
	default:
		return sameLine + " " + multiLine
	}
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
			street, err := parseStreetLine(strings.Join(streetToks, " "), abbr)
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
// Special forms:
//   - Rural route and PO Box free-text lines are rewritten first (see
//     rewriteSpecialStreetLine) and stored wholly in StreetName, similar to
//     overseas military street lines.
//   - Grid-style double directionals without a street suffix (e.g. "1016 E 1700 S"
//     or "1016 East 1700 South") peel as Primary + Predir + numeric StreetName +
//     Postdir — no suffix is required.
//   - Fractional house numbers (e.g. "123 1/2 Main Street"): PrimaryNumber is
//     the integer portion ("123"); the fraction stays in StreetName ("1/2 MAIN")
//     with slash retained via StripOptions.KeepSlash. This prefers a clean
//     primary over packing "123 1/2" into PrimaryNumber (USPS-style packing is
//     also valid but harder to round-trip through field-level Normalize).
//   - Hyphenated primaries (NYC style, e.g. "112-10") keep the hyphen via
//     StripOptions.KeepHyphen.
//   - Multi-secondary units (e.g. "Building 420 Room 120") are peeled right-to-
//     left repeatedly and combined into SecondaryDesignator + SecondaryNumber
//     so Format yields "BLDG 420 RM 120".
func parseStreetLine(line string, regionCode string) (Address, error) {
	isPR := regionCode == "PR"

	// PR Spanish specials (Apartado) before general RR/PO rewrite.
	if isPR {
		if rewritten, ok := rewriteSpecialStreetLinePR(line); ok {
			return Address{StreetName: rewritten}, nil
		}
	}
	if rewritten, ok := rewriteSpecialStreetLine(line); ok {
		return Address{StreetName: rewritten}, nil
	}

	// KeepHyphen: NYC-style primary ranges (112-10). KeepSlash: fractional
	// addresses (123 1/2 Main St) so "1/2" survives into StreetName.
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
	tokens = reorderLeadingSecondary(tokens, isPR)

	// Same-line business / narrative tokens before the house number
	// (e.g. "WILLIAMSON MEDICAL CENTER 3000 EDWARD CURD LANE").
	// Skip for PR: house numbers often trail Spanish type+name ("CALLE LUNA 123"),
	// which would otherwise be misread as pre-street + primary.
	var out Address
	if !isPR {
		var preStreet string
		preStreet, tokens = splitPreStreet(tokens)
		out.BusinessName = preStreet
	}

	tokens = peelSecondary(tokens, &out, isPR)

	// Postdirectional (right)
	if len(tokens) > 1 {
		if abbr, err := directionals.AbbreviateDirectional(tokens[len(tokens)-1]); err == nil {
			out.Postdirectional = abbr
			tokens = tokens[:len(tokens)-1]
		}
	}

	// Street suffix (right) — only if a name body remains after the peel
	// (accounting for an optional leading primary number). Peel exactly one
	// suffix; a second trailing suffix stays in the name (expanded to primary
	// form below) so e.g. "Main Avenue Drive" → name MAIN AVENUE + DR.
	//
	// When the only residual after primary would be the suffix itself, leave it
	// as street-name material: "1001 Avenue E" → name AVENUE + postdir E
	// (not suffix AVE with empty name); "1000 AVE" → name AVENUE.
	if len(tokens) >= 2 {
		if abbr, err := streetsuffixes.NormalizeStreetSuffixAbreviation(tokens[len(tokens)-1]); err == nil {
			residual := tokens[:len(tokens)-1]
			nameBody := residual
			if len(nameBody) > 0 && looksLikePrimaryNumber(nameBody[0]) {
				nameBody = nameBody[1:]
			}
			if len(nameBody) > 0 {
				out.StreetSuffix = abbr
				tokens = residual
			}
		}
	}

	// PR Spanish street types as suffix (trailing) when English suffix did not match.
	if isPR && out.StreetSuffix == "" && len(tokens) >= 2 {
		if abbr, err := puertorico.AbbreviateStreetType(tokens[len(tokens)-1]); err == nil {
			residual := tokens[:len(tokens)-1]
			nameBody := residual
			if len(nameBody) > 0 && looksLikePrimaryNumber(nameBody[0]) {
				nameBody = nameBody[1:]
			}
			if len(nameBody) > 0 {
				out.StreetSuffix = abbr
				tokens = residual
			}
		}
	}

	// Primary number (left). Fractional house numbers keep the integer here
	// and leave "1/2" (etc.) in residual tokens for StreetName — see
	// parseStreetLine doc on fractional addresses.
	if len(tokens) > 0 && looksLikePrimaryNumber(tokens[0]) {
		out.PrimaryNumber = tokens[0]
		tokens = tokens[1:]
	}

	// PR trailing house number: "CALLE LUNA 123" → primary 123.
	if isPR && out.PrimaryNumber == "" && len(tokens) >= 2 && looksLikePrimaryNumber(tokens[len(tokens)-1]) {
		out.PrimaryNumber = tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
	}

	// PR leading Spanish street type: "CALLE LUNA" → suffix CLL, name LUNA.
	if isPR && out.StreetSuffix == "" && len(tokens) >= 2 {
		if abbr, err := puertorico.AbbreviateStreetType(tokens[0]); err == nil {
			out.StreetSuffix = abbr
			tokens = tokens[1:]
		}
	}

	// Predirectional (left) only when a street-name token remains after it.
	// A leading fraction token (1/2, 1/4, …) stays in the name and is skipped
	// when looking for the predirectional: "123 1/2 N MAIN" → predir N.
	// If the directional is the only remaining name token (e.g. "1005 South
	// Avenue" after peeling AVE), keep it as StreetName — do not set Predirectional.
	if len(tokens) > 1 {
		preIdx := 0
		if looksLikeFraction(tokens[0]) && len(tokens) > 2 {
			preIdx = 1
		}
		if abbr, err := directionals.AbbreviateDirectional(tokens[preIdx]); err == nil {
			// Ensure at least one name token remains after removing the predir.
			if len(tokens) > preIdx+1 {
				out.Predirectional = abbr
				tokens = append(tokens[:preIdx:preIdx], tokens[preIdx+1:]...)
			}
		}
	}

	if len(tokens) == 0 {
		return Address{}, fmt.Errorf("unrecognized street line: %q", line)
	}

	// Sole remaining name token:
	// - spell out directionals (S → SOUTH) when kept as name rather than Predirectional
	// - promote bare street-suffix tokens to primary form only when no StreetSuffix
	//   was peeled (AVE → AVENUE for "1000 AVE" / "1001 Avenue E"). When a suffix
	//   is already set, leave the token alone so state-as-name (CT Drive → CONNECTICUT)
	//   can still apply.
	if len(tokens) == 1 {
		if full, err := directionals.NormalizeDirectional(tokens[0]); err == nil {
			tokens[0] = full
		} else if out.StreetSuffix == "" {
			if primary, err := streetsuffixes.NormalizeStreetSuffix(tokens[0]); err == nil {
				tokens[0] = primary
			}
		}
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

// splitPreStreet finds non-address tokens before the primary house number on a
// street line and returns them as a business/narrative prefix. The leftmost
// token that looksLikePrimaryNumber and is followed by at least one more token
// (street body) is treated as the primary; tokens before it become the prefix.
//
// Ordinary streets that already start with a primary number are left unchanged
// (i == 0). If no primary-looking token appears after position 0 with a
// following street body, tokens are returned as-is.
//
// Call after special rewrite, hash expand, directional merge, and leading
// secondary reorder so military/RR/PO never reach here and "APT 4 123 …" is
// already reordered to start with the house number.
func splitPreStreet(tokens []string) (business string, rest []string) {
	if len(tokens) < 2 {
		return "", tokens
	}
	for i := 0; i < len(tokens)-1; i++ {
		if !looksLikePrimaryNumber(tokens[i]) {
			continue
		}
		// Primary already first: ordinary street — do not invent a pre-street.
		if i == 0 {
			return "", tokens
		}
		// At least one token after the primary remains for the street body.
		return strings.Join(tokens[:i], " "), tokens[i:]
	}
	// No primary after position 0 (or no following body): leave as today.
	return "", tokens
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
func reorderLeadingSecondary(tokens []string, isPR bool) []string {
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

	// PR Spanish secondary (e.g. APARTAMENTO 4 … / EDIF 2 …)
	if isPR {
		if short, err := puertorico.NormalizeSecondary(tokens[0]); err == nil {
			if looksLikeSecondaryNumber(tokens[1]) {
				rest := tokens[2:]
				out := make([]string, 0, len(tokens))
				out = append(out, rest...)
				out = append(out, short, tokens[1])
				return out
			}
		}
	}

	return tokens
}

// looksLikeSecondaryNumber reports whether tok is a plausible unit number
// (contains a digit). Pure alpha tokens are not treated as unit numbers so
// "APT SOUTH …" is not reordered.
func looksLikeSecondaryNumber(tok string) bool {
	return looksLikePrimaryNumber(tok)
}

// secondaryPeel is one right-to-left secondary unit match.
type secondaryPeel struct {
	designator string // short form (APT, BLDG, #, …)
	number     string // unit number, or empty for non-numbered designators
}

// peelSecondary repeatedly peels trailing secondary designators from the right.
// Overseas military "UNIT N BOX N" is handled earlier by the military fast path,
// so this peel only sees civilian street lines.
//
// The Address struct holds a single SecondaryDesignator + SecondaryNumber pair.
// Multiple trailing secondaries are peeled right-to-left and folded so Format
// yields e.g. "BLDG 420 RM 120":
//   - rightmost peel seeds designator/number
//   - each further-left numbered peel becomes the designator; prior des+num
//     is appended to the number trail
//   - non-numbered peels (UPPER/REAR) append onto SecondaryNumber
//
// So "Building 420 Room 120" → BLDG / "420 RM 120", and "Unit 3200 … Upper"
// (after leading reorder) → UNIT / "3200 UPPR".
func peelSecondary(tokens []string, out *Address, isPR bool) []string {
	var peels []secondaryPeel // rightmost first

	for len(tokens) > 0 {
		// "# 12" or "#12" (already expanded by expandHashTokens)
		if len(tokens) >= 2 && tokens[len(tokens)-2] == "#" {
			peels = append(peels, secondaryPeel{
				designator: "#",
				number:     tokens[len(tokens)-1],
			})
			tokens = tokens[:len(tokens)-2]
			continue
		}

		// Numbered designator + unit number (e.g. APT 4, BLDG 420, RM 120).
		// Unit id may be alpha-only (STE A); do not require a digit.
		if len(tokens) >= 2 {
			if info, err := secondaryunit.Info(tokens[len(tokens)-2]); err == nil && info.Numbered {
				peels = append(peels, secondaryPeel{
					designator: info.Short,
					number:     tokens[len(tokens)-1],
				})
				tokens = tokens[:len(tokens)-2]
				continue
			}
		}

		// Designator alone (non-numbered, or numbered without a number)
		if info, err := secondaryunit.Info(tokens[len(tokens)-1]); err == nil {
			peels = append(peels, secondaryPeel{designator: info.Short})
			tokens = tokens[:len(tokens)-1]
			continue
		}

		// PR Spanish secondary designators (only when region is PR).
		if isPR {
			if len(tokens) >= 2 {
				if short, err := puertorico.NormalizeSecondary(tokens[len(tokens)-2]); err == nil {
					peels = append(peels, secondaryPeel{
						designator: short,
						number:     tokens[len(tokens)-1],
					})
					tokens = tokens[:len(tokens)-2]
					continue
				}
			}
			if short, err := puertorico.NormalizeSecondary(tokens[len(tokens)-1]); err == nil {
				peels = append(peels, secondaryPeel{designator: short})
				tokens = tokens[:len(tokens)-1]
				continue
			}
		}

		break
	}

	if len(peels) == 0 {
		return tokens
	}

	// Fold rightmost-first peels into a single designator + number pair.
	// Rightmost peel is the seed; each further-left numbered secondary becomes
	// the new designator with the prior designator/number as the trail:
	//   RM+120, then BLDG+420  →  BLDG / "420 RM 120"
	// Non-numbered peels (UPPER/REAR) append to SecondaryNumber so leading
	// "Unit 3200 … Upper" (reordered to "… Upper Unit 3200") still yields
	// UNIT / "3200 UPPR" rather than promoting UPPR to designator.
	des := peels[0].designator
	num := peels[0].number
	for i := 1; i < len(peels); i++ {
		p := peels[i]
		if p.number != "" {
			trail := des
			if num != "" {
				trail = des + " " + num
			}
			des = p.designator
			num = p.number + " " + trail
			continue
		}
		// Non-numbered further left (or after a bare designator seed).
		if num != "" {
			num = num + " " + p.designator
		} else {
			num = p.designator
		}
	}
	out.SecondaryDesignator = des
	out.SecondaryNumber = num
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

// looksLikeFraction reports whether tok is a simple numeric fraction such as
// "1/2" or "3/4" (digits, slash, digits). Used so a fraction after the primary
// number stays in StreetName and does not block predirectional detection.
func looksLikeFraction(tok string) bool {
	i := strings.IndexByte(tok, '/')
	if i <= 0 || i >= len(tok)-1 {
		return false
	}
	return isAllDigits(tok[:i]) && isAllDigits(tok[i+1:])
}
