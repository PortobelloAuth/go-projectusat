package parse

import (
	"fmt"
	"regexp"
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

// Address is a structured patient address produced by Parse.
// Empty string means unknown / not present. The root package maps this to its Address type.
type Address struct {
	BusinessName string

	PrimaryNumber       string
	Predirectional      string
	StreetName          string
	StreetSuffix        string
	Postdirectional     string
	SecondaryDesignator string
	SecondaryNumber     string

	City   string
	Region string
	Postal string

	Country string
}

// usZIPCompact matches ##### or #####-#### / ######### after punctuation strip.
var usZIPCompact = regexp.MustCompile(`^(\d{5})(?:-?(\d{4}))?$`)

// Canadian postal: compact A1A1A1, or FSA (A1A) + LDU (1A1) as two tokens.
var (
	caPostalCompact = regexp.MustCompile(`^[A-Z]\d[A-Z]\d[A-Z]\d$`)
	caPostalFSA     = regexp.MustCompile(`^[A-Z]\d[A-Z]$`)
	caPostalLDU     = regexp.MustCompile(`^\d[A-Z]\d$`)
)

// normalizePostal formats US ZIP / ZIP+4 and Canadian A1A 1A1; other patterns
// remain uppercase alphanumerics with collapsed spacing.
func normalizePostal(s string) string {
	s = textutil.Upper(textutil.CollapseSpace(s))
	if s == "" {
		return ""
	}
	cleaned := textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{KeepHyphen: true}))
	compact := strings.ReplaceAll(cleaned, " ", "")
	if m := usZIPCompact.FindStringSubmatch(compact); m != nil {
		if m[2] != "" {
			return m[1] + "-" + m[2]
		}
		return m[1]
	}
	if caPostalCompact.MatchString(compact) {
		return compact[:3] + " " + compact[3:]
	}
	return textutil.CollapseSpace(textutil.StripPunctuation(s, textutil.StripOptions{}))
}

// Parse converts free-text multi-line or comma-separated address text into a
// structured Address.
//
// Architecture (Project US@ Go layout):
//  1. Split the input into Tokens (position, comma/newline context).
//  2. Score tokens against pkg/* vocabularies (region, directionals, suffixes, …).
//  3. Assign tokens to address components using score, order, and punctuation.
//
// Special forms (military, rural route, PO Box) are recognized when scores and
// patterns align. Business/narrative prefixes are residual tokens before the
// street components when parsing from the end.
//
// Parse does not call content Normalize; callers compose as needed. Empty input
// returns an error.
func Parse(raw string) (Address, error) {
	// Tokenize for position/punctuation context (used by scoring helpers and
	// available for multi-interpretation assignment). Line routing still uses
	// splitAddressLines for military comma bipartition and multi-line shapes.
	if len(Tokenize(raw)) == 0 {
		return Address{}, fmt.Errorf("empty address")
	}
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

		// Military street form with civilian last line (e.g. UNIT BOX + Springfield IL).
		// Street line alone normalizes as military; last line is ordinary city/region/ZIP.
		if street, err := military.NormalizeStreetLine(lines[len(lines)-2]); err == nil {
			if city, reg, zip, country, err := parseLastLine(lines[len(lines)-1]); err == nil {
				var business string
				if len(lines) > 2 {
					business = strings.ToUpper(textutil.CollapseSpace(strings.Join(lines[:len(lines)-2], " ")))
				}
				return Address{
					BusinessName: business,
					StreetName:   street,
					City:         city,
					Region:       reg,
					Postal:       zip,
					Country:      country,
				}, nil
			}
		}
	}

	// Single-line space-separated military: "PSC 3 BOX 4120 APO AE 09021-0002"
	// (comma-separated form already becomes 2 lines via splitAddressLines).
	if len(lines) == 1 {
		if addr, ok := tryParseSingleLineMilitary(lines[0]); ok {
			return addr, nil
		}
	}

	// Fall through for non-military.
	return parseCivilian(lines)
}

// tryParseSingleLineMilitary recognizes a space-only overseas military address
// on one logical line: optional business tokens, then TYPE N BOX N, then
// APO|FPO|DPO AE|AP|AA ZIP. Consumes last-line and street from the ends so
// multi-token business prefixes are preserved.
func tryParseSingleLineMilitary(line string) (Address, bool) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	// Minimum: TYPE N BOX N + CITY REGION ZIP (7 tokens).
	if len(tokens) < 7 {
		return Address{}, false
	}

	// Last 3 tokens must be a military last line.
	lastCand := strings.Join(tokens[len(tokens)-3:], " ")
	city, reg, zip, err := military.NormalizeLastLine(lastCand)
	if err != nil {
		return Address{}, false
	}

	// Preceding 4 tokens must be a military street line.
	streetCand := strings.Join(tokens[len(tokens)-7:len(tokens)-3], " ")
	street, err := military.NormalizeStreetLine(streetCand)
	if err != nil {
		return Address{}, false
	}

	var business string
	if len(tokens) > 7 {
		business = strings.Join(tokens[:len(tokens)-7], " ")
	}
	return Address{
		BusinessName: business,
		StreetName:   street,
		City:         city,
		Region:       reg,
		Postal:       zip,
	}, true
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
		line = textutil.CollapseSpace(line)
		if line == "" {
			continue
		}
		if left, right, ok := splitMilitaryCommaLine(line); ok {
			out = append(out, left, right)
			continue
		}
		// Commas are punctuation within a line, not segment boundaries.
		line = textutil.CollapseSpace(strings.ReplaceAll(line, ",", " "))
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// splitMilitaryCommaLine bipartitions a single physical line on a comma when
// the left side is a military street line and the right is a military last line.
// Remaining commas (if any) on either side are treated as spaces.
func splitMilitaryCommaLine(line string) (left, right string, ok bool) {
	if !strings.Contains(line, ",") {
		return "", "", false
	}
	// Try each comma as the street/last boundary (usually exactly one).
	for i := 0; i < len(line); i++ {
		if line[i] != ',' {
			continue
		}
		left = textutil.CollapseSpace(strings.ReplaceAll(line[:i], ",", " "))
		right = textutil.CollapseSpace(strings.ReplaceAll(line[i+1:], ",", " "))
		if left == "" || right == "" {
			continue
		}
		if _, err := military.NormalizeStreetLine(left); err != nil {
			continue
		}
		if _, _, _, err := military.NormalizeLastLine(right); err != nil {
			continue
		}
		return left, right, true
	}
	return "", "", false
}

func parseCivilian(lines []string) (Address, error) {
	if len(lines) == 1 {
		// Single segment: peel trailing last-line tokens; remainder is street.
		return parseSingleLineCivilian(lines[0])
	}
	city, reg, zip, country, err := parseLastLine(lines[len(lines)-1])
	if err != nil {
		return Address{}, err
	}
	streetLine := lines[len(lines)-2]
	business := ""
	if len(lines) > 2 {
		// Everything before the street+last-line address is business/narrative residual.
		business = strings.ToUpper(textutil.CollapseSpace(strings.Join(lines[:len(lines)-2], " ")))
	}
	// Region and PR ZIP ranges both engage PR dialect (score-collocated in puertorico).
	streetRegion := reg
	if usePRDialect(reg, zip) {
		streetRegion = "PR"
	}
	street, err := parseStreetLine(streetLine, streetRegion)
	if err != nil {
		return Address{}, err
	}
	// Multi-line business stays; same-line pre-street prefix fills BusinessName
	// when empty, otherwise is prepended with a space.
	street.BusinessName = mergeBusinessName(business, street.BusinessName)
	street.City = city
	street.Region = reg
	street.Postal = zip
	street.Country = country
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
// Trailing country tokens (USA/US/UNITED STATES) are stripped into country.
//
// Assignment is score-driven from the right: postal score, then region.Score /
// ScorePhrase on longest-right candidates, residual tokens form the city.
func parseLastLine(line string) (city, reg, postal, country string, err error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	if len(tokens) < 2 {
		return "", "", "", "", fmt.Errorf("invalid last line: %q", line)
	}
	tokens, country = stripTrailingCountry(tokens)
	if len(tokens) < 2 {
		return "", "", "", "", fmt.Errorf("invalid last line: %q", line)
	}
	postal, rest, ok := extractPostal(tokens)
	if !ok {
		return "", "", "", "", fmt.Errorf("invalid last line postal: %q", line)
	}
	if len(rest) < 1 {
		return "", "", "", "", fmt.Errorf("invalid last line: missing region in %q", line)
	}
	// Prefer the longest right-hand region phrase with the best region.Score.
	bestN, bestScore := 0, 0
	var bestAbbr string
	for n := len(rest); n >= 1; n-- {
		cand := strings.Join(rest[len(rest)-n:], " ")
		sc, _ := region.ScorePhrase(cand)
		if sc == 0 {
			continue
		}
		abbr, e := region.NormalizeRegion(cand)
		if e != nil {
			continue
		}
		// Prefer longer matches when scores tie (DISTRICT OF COLUMBIA over COLUMBIA).
		if sc > bestScore || (sc == bestScore && n > bestN) {
			bestScore, bestN, bestAbbr = sc, n, abbr
		}
	}
	if bestN == 0 {
		return "", "", "", "", fmt.Errorf("unrecognized region in last line: %q", line)
	}
	city = strings.Join(rest[:len(rest)-bestN], " ")
	if city == "" {
		return "", "", "", "", fmt.Errorf("invalid last line: missing city in %q", line)
	}
	return city, bestAbbr, postal, country, nil
}

// stripTrailingCountry removes a trailing country designator (USA, US, UNITED
// STATES) from tokens. Returns the remaining tokens and the country string
// (empty if none). Does not strip CA/CANADA — CA collides with California.
func stripTrailingCountry(tokens []string) (rest []string, country string) {
	if len(tokens) == 0 {
		return tokens, ""
	}
	last := tokens[len(tokens)-1]
	switch last {
	case "USA", "US":
		return tokens[:len(tokens)-1], last
	case "STATES":
		// UNITED STATES
		if len(tokens) >= 2 && tokens[len(tokens)-2] == "UNITED" {
			return tokens[:len(tokens)-2], "UNITED STATES"
		}
	}
	return tokens, ""
}

// extractPostal removes US ZIP / ZIP+4 or Canadian postal from the right of tokens.
func extractPostal(tokens []string) (postal string, rest []string, ok bool) {
	if len(tokens) == 0 {
		return "", nil, false
	}
	last := tokens[len(tokens)-1]

	// ##### or #####-#### (and compact #########)
	if usZIPCompact.MatchString(last) {
		return normalizePostal(last), tokens[:len(tokens)-1], true
	}

	// Compact Canadian single token: K1A0B1 → K1A 0B1
	if caPostalCompact.MatchString(last) {
		return normalizePostal(last), tokens[:len(tokens)-1], true
	}

	if len(tokens) >= 2 {
		prev := tokens[len(tokens)-2]
		// Two-token US ZIP+4: 62701 1234
		if usZIPCompact.MatchString(prev+last) || usZIPCompact.MatchString(prev+"-"+last) {
			return normalizePostal(prev + "-" + last), tokens[:len(tokens)-2], true
		}
		// Canadian two-token FSA + LDU only: K1A 0B1 (not region+garbage).
		if caPostalFSA.MatchString(prev) && caPostalLDU.MatchString(last) {
			return normalizePostal(prev + last), tokens[:len(tokens)-2], true
		}
	}

	// Single alphanumeric postal fallback (unknown international forms).
	return normalizePostal(last), tokens[:len(tokens)-1], true
}

// parseSingleLineCivilian peels postal + region from the right, then splits the
// remaining tokens into street + multi-word city by trying city lengths and
// preferring a parseStreetLine result with PrimaryNumber and StreetSuffix.
func parseSingleLineCivilian(line string) (Address, error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	tokens, country := stripTrailingCountry(tokens)
	postal, rest, ok := extractPostal(tokens)
	if !ok || len(rest) < 2 {
		return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
	}
	// Region up to 3 tokens e.g. DISTRICT OF COLUMBIA
	for n := min(3, len(rest)); n >= 1; n-- {
		cand := strings.Join(rest[len(rest)-n:], " ")
		if abbr, e := region.NormalizeRegion(cand); e == nil {
			before := rest[:len(rest)-n]
			// Need ≥1 street token and ≥1 city token.
			if len(before) < 2 {
				return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
			}
			streetRegion := abbr
			if usePRDialect(abbr, postal) {
				streetRegion = "PR"
			}
			street, city, err := splitStreetAndCity(before, streetRegion)
			if err != nil {
				return Address{}, err
			}
			street.City = city
			street.Region = abbr
			street.Postal = postal
			street.Country = country
			return street, nil
		}
	}
	return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
}

// splitStreetAndCity partitions tokens (street… city…) after region/postal are
// peeled. Tries city as the last 1..len-1 tokens; prefers the longest city where
// parseStreetLine succeeds with a PrimaryNumber or StreetName. When multiple
// candidates succeed, prefer those with StreetSuffix (or SecondaryDesignator)
// so "123 MAIN STREET NEW YORK" yields city NEW YORK rather than STREET NEW YORK.
func splitStreetAndCity(tokens []string, regionCode string) (Address, string, error) {
	type cand struct {
		street Address
		city   string
		cityN  int
		score  int
	}
	var best *cand
	// cityLen from longest to shortest so equal scores keep the longer city.
	for cityLen := len(tokens) - 1; cityLen >= 1; cityLen-- {
		streetToks := tokens[:len(tokens)-cityLen]
		if len(streetToks) == 0 {
			continue
		}
		street, err := parseStreetLine(strings.Join(streetToks, " "), regionCode)
		if err != nil {
			continue
		}
		if street.PrimaryNumber == "" && street.StreetName == "" {
			continue
		}
		score := 0
		if street.PrimaryNumber != "" {
			score += 2
		}
		if street.StreetName != "" {
			score++
		}
		if street.StreetSuffix != "" {
			score += 3
		}
		if street.SecondaryDesignator != "" {
			score++
		}
		// Longer city breaks ties (prefer multi-word cities).
		score = score*100 + cityLen
		if best == nil || score > best.score {
			c := cand{street: street, city: strings.Join(tokens[len(tokens)-cityLen:], " "), cityN: cityLen, score: score}
			best = &c
		}
	}
	if best == nil {
		return Address{}, "", fmt.Errorf("cannot split street and city from %q", strings.Join(tokens, " "))
	}
	return best.street, best.city, nil
}

// usePRDialect reports whether Spanish PR vocabulary applies for this address.
func usePRDialect(regionCode, postal string) bool {
	return puertorico.UsePRDialect(regionCode, postal)
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
//   - Multi-secondary units (e.g. "Building 420 Room 120") are peeled and
//     combined into SecondaryDesignator + SecondaryNumber so Format yields
//     "BLDG 420 RM 120". Leading secondaries (e.g. "Unit 3200 … Room 12") are
//     extracted first so fold order preserves original left-to-right appearance
//     → UNIT / "3200 RM 12".
//   - State as portion of street name is abbreviated (MONTANA TREASURE → MT
//     TREASURE); multi-word state prefixes are not stolen as predirectionals
//     (SOUTH CAROLINA COUNTY ROAD 22 → SC COUNTY ROAD 22).
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
	// Protect digit.digit periods (grid designators like 39.4) so StripPunctuation
	// does not glue them into 394; restore after strip.
	protected := protectDecimalPeriods(line)
	cleaned := strings.ToUpper(textutil.CollapseSpace(
		textutil.StripPunctuation(protected, textutil.StripOptions{KeepHyphen: true, KeepSlash: true}),
	))
	cleaned = restoreDecimalPeriods(cleaned)
	tokens := expandHashTokens(strings.Fields(cleaned))
	if len(tokens) == 0 {
		return Address{}, fmt.Errorf("empty street line")
	}
	// Hyphenated compounds: NORTH-EAST → NE; Main-Street → MAIN STREET.
	tokens = expandHyphenatedDirectionals(tokens)
	tokens = expandHyphenatedStreetTokens(tokens)
	// Merge multi-token directionals (SOUTH WEST → SW) before peels so pre-
	// and postdirectionals resolve as single compound abbreviations.
	tokens = mergeDirectionTokens(tokens)
	// Mid-line "# 12" secondary → trailing for reverse peels.
	tokens = reorderMidLineHashSecondary(tokens)

	// Extract leading secondary designator/# + number so it stays first in
	// multi-secondary fold order (e.g. "UNIT 3200 152 TECH DR ROOM 12" →
	// UNIT 3200 then RM 12). Also clears the leading designator so
	// splitPreStreet sees the house number first.
	var leadingSecondary *secondaryPeel
	// Extract leading UNIT/APT/# pairs first so multi-secondary fold keeps
	// original left-to-right order (Unit 3200 … Room 12 → UNIT 3200 RM 12).
	// Then reorder remaining mid-line pairs (Suite 480 411 …) to the end.
	tokens, leadingSecondary = takeLeadingSecondary(tokens, isPR)
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

	tokens = extractSecondary(tokens, &out, isPR, leadingSecondary)

	// Postdirectional (right) only when a street-name token remains after the peel
	// (accounting for an optional leading primary). Same empty-name guard as
	// predirectional: "123 South" keeps SOUTH as StreetName, not Postdirectional.
	if len(tokens) > 1 {
		if abbr, err := directionals.AbbreviateDirectional(tokens[len(tokens)-1]); err == nil {
			residual := tokens[:len(tokens)-1]
			nameBody := residual
			if len(nameBody) > 0 && looksLikePrimaryNumber(nameBody[0]) {
				nameBody = nameBody[1:]
			}
			if len(nameBody) > 0 {
				out.Postdirectional = abbr
				tokens = residual
			}
		}
	}

	// Street suffix (right) via peelStreetSuffix (handles Annex after Boulevard).
	if suffix, rest, ok := extractStreetSuffix(tokens); ok {
		out.StreetSuffix = suffix
		tokens = rest
	}

	// PR Spanish street type (trailing token), kept as Spanish primary word
	// (CALLE not CLL). Combined into StreetName below so Format yields
	// "123 CALLE LUNA" rather than English-style "123 LUNA CLL".
	var prStreetType string
	if isPR && out.StreetSuffix == "" && len(tokens) >= 2 {
		if primary, err := puertorico.NormalizeStreetType(tokens[len(tokens)-1]); err == nil {
			residual := tokens[:len(tokens)-1]
			nameBody := residual
			if len(nameBody) > 0 && looksLikePrimaryNumber(nameBody[0]) {
				nameBody = nameBody[1:]
			}
			if len(nameBody) > 0 {
				prStreetType = primary
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

	// PR leading Spanish street type: "CALLE LUNA" → type CALLE, name LUNA.
	if isPR && prStreetType == "" && out.StreetSuffix == "" && len(tokens) >= 2 {
		if primary, err := puertorico.NormalizeStreetType(tokens[0]); err == nil {
			prStreetType = primary
			tokens = tokens[1:]
		}
	}

	// Predirectional (left) only when a street-name token remains after it.
	// A leading fraction token (1/2, 1/4, …) stays in the name and is skipped
	// when looking for the predirectional: "123 1/2 N MAIN" → predir N.
	// If the directional is the only remaining name token (e.g. "1005 South
	// Avenue" after peeling AVE), keep it as StreetName — do not set Predirectional.
	//
	// Multi-word state names that begin with a directional word (SOUTH CAROLINA,
	// NORTH DAKOTA, WEST VIRGINIA, …) must not lose their first token as predir:
	// "SOUTH CAROLINA COUNTY ROAD 22" stays intact for highway normalize → SC …
	// and "SOUTH CAROLINA AVENUE" keeps the full state as StreetName.
	if len(tokens) > 1 && !leadingTokensAreMultiWordState(tokens) {
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

	// Mid-name single-letter directionals (not peeled as pre/post dir) expand to
	// full words: "BAY W DRIVE" → name BAY WEST + DR. Multi-letter compounds
	// (NE/SW/…) are left as-is so mid-name "N E" merges stay context-sensitive.
	for i, tok := range tokens {
		if len(tok) != 1 {
			continue
		}
		if full, err := directionals.NormalizeDirectional(tok); err == nil {
			tokens[i] = full
		}
	}

	// When a US state name is only a *portion* of the street name (more name
	// tokens remain), abbreviate it to the two-letter code per Project US@:
	// "MONTANA TREASURE" → "MT TREASURE". Entire-name state spelling is handled
	// below via fullySpelledUSState when a street suffix is present.
	tokens = abbreviateLeadingStatePortion(tokens)

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

	// Project US@ keeps Spanish street types as Spanish words (not English-style
	// trailing abbreviations). Prefix type onto StreetName: "CALLE LUNA".
	if prStreetType != "" {
		if out.StreetName != "" {
			out.StreetName = prStreetType + " " + out.StreetName
		} else {
			out.StreetName = prStreetType
		}
		out.StreetSuffix = ""
	}
	return out, nil
}

// fullySpelledUSState returns the fully spelled US state/possession name when
// name (possibly multi-word) normalizes to a known US region code. Military
// and Canadian codes are excluded.
func fullySpelledUSState(name string) (string, bool) {
	abbr, err := region.NormalizeRegion(name)
	if err != nil {
		return "", false
	}
	return region.FullName(abbr)
}

// leadingTokensAreMultiWordState reports whether tokens begin with a multi-word
// US state/possession name (SOUTH CAROLINA, NEW YORK, WEST VIRGINIA, …). Used to
// avoid peeling the first word as a predirectional when it belongs to the state.
func leadingTokensAreMultiWordState(tokens []string) bool {
	if len(tokens) < 2 {
		return false
	}
	code, n, ok := region.LeadingStateMatch(tokens, false)
	if !ok || n < 2 {
		return false
	}
	full, ok := region.FullName(code)
	return ok && len(strings.Fields(full)) >= 2
}

// abbreviateLeadingStatePortion replaces a leading full US state name with its
// two-letter postal code when the state is only a portion of the street name
// (at least one name token remains after the state). Entire-name state spelling
// is left alone for fullySpelledUSState when a street suffix is present.
//
//	MONTANA TREASURE → MT TREASURE
//	SOUTH CAROLINA COUNTY ROAD 22 → SC COUNTY ROAD 22
//	OKLAHOMA (alone) → unchanged
func abbreviateLeadingStatePortion(tokens []string) []string {
	if len(tokens) < 2 {
		return tokens
	}
	abbr, n, ok := region.LeadingStateMatch(tokens, true)
	if !ok {
		return tokens
	}
	out := make([]string, 0, len(tokens)-n+1)
	out = append(out, abbr)
	out = append(out, tokens[n:]...)
	return out
}

// expandHyphenatedStreetTokens splits Main-Street style compounds when the
// right side is a street suffix.
func expandHyphenatedStreetTokens(tokens []string) []string {
	if len(tokens) == 0 {
		return tokens
	}
	out := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		i := strings.LastIndexByte(t, '-')
		if i <= 0 || i >= len(t)-1 {
			out = append(out, t)
			continue
		}
		left, right := t[:i], t[i+1:]
		if left == "" || right == "" {
			out = append(out, t)
			continue
		}
		// Only split when the right side is a street suffix (STREET→ST, etc.).
		if _, err := streetsuffixes.NormalizeStreetSuffixAbreviation(right); err == nil {
			// Preserve multi-hyphen left side as a single token (rare).
			out = append(out, left, right)
			continue
		}
		out = append(out, t)
	}
	return out
}

// extractStreetSuffix extracts a street suffix from the right of tokens.
// Returns the abbreviated suffix, residual tokens, and whether a peel occurred.
//
// When the rightmost token is a trailing-junk suffix (ANX/ANNEX) and the
// previous token is also a street suffix with a non-empty name body, peel the
// previous (primary) suffix and leave the junk token in the residual name.

func isTrailingJunkSuffix(abbr string) bool {
	return abbr == "ANX"
}

// reorderMidLineHashSecondary moves a mid-line "# NUMBER" pair to the end of
// the token slice so reverse peels capture it as secondary. Patterns:
//
//	… HOUSE # NUM rest…  →  … HOUSE rest… # NUM
//	… # NUM rest…        →  … rest… # NUM   (when not already leading)
//
// Leading "# NUM rest" is already handled by reorderLeadingSecondary; this
// covers house-number-first forms like "100 #12 MAIN STREET".

func extractStreetSuffix(tokens []string) (suffix string, rest []string, ok bool) {
	if len(tokens) < 2 {
		return "", tokens, false
	}
	last := tokens[len(tokens)-1]
	lastAbbr, lastErr := streetsuffixes.NormalizeStreetSuffixAbreviation(last)
	if lastErr != nil {
		return "", tokens, false
	}

	// Prefer primary street suffixes over trailing obscure ones (ANNEX/ANX)
	// when both apply: "Oak Boulevard Annex" → BLVD, residual OAK ANNEX.
	if isTrailingJunkSuffix(lastAbbr) && len(tokens) >= 3 {
		prev := tokens[len(tokens)-2]
		if prevAbbr, err := streetsuffixes.NormalizeStreetSuffixAbreviation(prev); err == nil {
			core := tokens[:len(tokens)-2]
			nameBody := core
			if len(nameBody) > 0 && looksLikePrimaryNumber(nameBody[0]) {
				nameBody = nameBody[1:]
			}
			if len(nameBody) > 0 {
				// Keep junk token in residual name; peel previous as StreetSuffix.
				rest = append(append([]string{}, core...), last)
				return prevAbbr, rest, true
			}
		}
	}

	// Standard rightmost peel when a name body remains.
	residual := tokens[:len(tokens)-1]
	nameBody := residual
	if len(nameBody) > 0 && looksLikePrimaryNumber(nameBody[0]) {
		nameBody = nameBody[1:]
	}
	if len(nameBody) > 0 {
		return lastAbbr, residual, true
	}
	return "", tokens, false
}

// isTrailingJunkSuffix reports suffixes that should yield to a preceding
// primary street suffix when both appear at the end of a street line.

func reorderMidLineHashSecondary(tokens []string) []string {
	if len(tokens) < 4 {
		// Need at least HOUSE # NUM rest (4 tokens) for a useful mid-line move.
		// (Leading form with 3 tokens is handled by reorderLeadingSecondary.)
		return tokens
	}
	// Find first mid-line "# NUMBER" where "#" is not at index 0.
	for i := 1; i+1 < len(tokens); i++ {
		if tokens[i] != "#" {
			continue
		}
		if !looksLikeSecondaryNumber(tokens[i+1]) {
			continue
		}
		// Require at least one token after the number (street body).
		if i+2 >= len(tokens) {
			// Already trailing — peelSecondary will handle it.
			return tokens
		}
		hash, num := tokens[i], tokens[i+1]
		out := make([]string, 0, len(tokens))
		out = append(out, tokens[:i]...)
		out = append(out, tokens[i+2:]...)
		out = append(out, hash, num)
		return out
	}
	return tokens
}

// splitPreStreet finds non-address tokens before the primary house number on a
// street line and returns them as a business/narrative prefix. The leftmost
// token that looksLikeHouseNumber and is followed by at least one more token
// (street body) is treated as the primary; tokens before it become the prefix.
//
// Uses looksLikeHouseNumber (not looksLikePrimaryNumber) so digit-leading firm
// names like "3M" are not mistaken for house numbers when a real primary
// follows ("3M Corporation 100 Main Street" → business "3M CORPORATION",
// rest starting at "100").
//
// Ordinary streets that already start with a house number are left unchanged
// (i == 0). If no house-number token appears after position 0 with a
// following street body, tokens are returned as-is.
//
// Call after special rewrite, hash expand, directional merge, and leading
// secondary reorder so military/RR/PO never reach here and "APT 4 123 …" is
// already reordered to start with the house number.

func expandHashTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		if strings.HasPrefix(t, "#") && len(t) > 1 {
			out = append(out, "#", t[1:])
			continue
		}
		if i := strings.IndexByte(t, '#'); i > 0 {
			left := t[:i]
			right := t[i+1:]
			out = append(out, left, "#")
			if right != "" {
				out = append(out, right)
			}
			continue
		}
		out = append(out, t)
	}
	return out
}

// splitPreStreet finds non-address tokens before the primary house number on a
// street line and returns them as a business/narrative prefix. The leftmost
// token that looksLikeHouseNumber and is followed by at least one more token
// (street body) is treated as the primary; tokens before it become the prefix.
//
// Uses looksLikeHouseNumber (not looksLikePrimaryNumber) so digit-leading firm
// names like "3M" are not mistaken for house numbers when a real primary
// follows ("3M Corporation 100 Main Street" → business "3M CORPORATION",
// rest starting at "100").
//
// Ordinary streets that already start with a house number are left unchanged
// (i == 0). If no house-number token appears after position 0 with a
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
		if !looksLikeHouseNumber(tokens[i]) {
			continue
		}
		// Primary already first: ordinary street — do not invent a pre-street.
		if i == 0 {
			return "", tokens
		}
		// At least one token after the primary remains for the street body.
		return strings.Join(tokens[:i], " "), tokens[i:]
	}
	// No house number after position 0 (or no following body): leave as today.
	return "", tokens
}

// decimalPeriodSentinel replaces '.' between digits during StripPunctuation so
// grid designators like "39.4" survive cleaning. Private-use code point; not
// expected in address input.
const decimalPeriodSentinel = '\uE000'

// protectDecimalPeriods replaces periods that sit between two digits with a
// sentinel so StripPunctuation does not remove them (grid / milepost style).
func protectDecimalPeriods(s string) string {
	runes := []rune(s)
	if len(runes) < 3 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(runes); i++ {
		if runes[i] == '.' && i > 0 && i < len(runes)-1 &&
			unicode.IsDigit(runes[i-1]) && unicode.IsDigit(runes[i+1]) {
			b.WriteRune(decimalPeriodSentinel)
			continue
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

// restoreDecimalPeriods turns protectDecimalPeriods sentinels back into '.'.
func restoreDecimalPeriods(s string) string {
	return strings.ReplaceAll(s, string(decimalPeriodSentinel), ".")
}

func reorderLeadingSecondary(tokens []string, isPR bool) []string {
	if len(tokens) < 3 {
		// Need designator/number plus at least one rest token to move.
		return tokens
	}

	// Trailing secondaries on ordinary streets stay put for reverse peels.
	if looksLikeStreetPrimaryNumber(tokens[0]) {
		return tokens
	}

	// Locate the primary house number: first street-primary candidate that is
	// not the unit id of a secondary designator preceding a later primary.
	primaryIdx := -1
	for i := 0; i < len(tokens)-1; i++ {
		if !looksLikeStreetPrimaryNumber(tokens[i]) {
			continue
		}
		if i > 0 && isSecondaryUnitPairAt(tokens, i-1, isPR) {
			// designator + this number — unit id when another primary follows
			// (Suite 480 411 … / # 3200 152 …).
			if hasLaterStreetPrimary(tokens, i+1, isPR) {
				continue
			}
			// No later primary. Pure leading secondary at start of line
			// ("# 3200 South Tech Drive") has no house number — reorder to end.
			// Mid-prefix designator without later primary is business text and
			// this token is the street primary ("UCENT Building 847 N …").
			if i-1 == 0 {
				return reorderLeadingSecondaryAtStart(tokens, isPR)
			}
		}
		primaryIdx = i
		break
	}
	if primaryIdx <= 0 {
		// No mid-line primary after a prefix; fall back to pure leading forms
		// at index 0 ("APT 4 123 …" / "# 4 123 …").
		return reorderLeadingSecondaryAtStart(tokens, isPR)
	}

	// Pull numbered secondary pairs out of the pre-primary prefix; leave other
	// tokens (business/narrative text, bare "Building" without a unit id) in place.
	prefix := tokens[:primaryIdx]
	var pairs []string
	var kept []string
	for i := 0; i < len(prefix); {
		if isSecondaryUnitPairAt(prefix, i, isPR) {
			desig, _ := secondaryPairTokens(prefix, i, isPR)
			pairs = append(pairs, desig, prefix[i+1])
			i += 2
			continue
		}
		kept = append(kept, prefix[i])
		i++
	}
	if len(pairs) == 0 {
		return tokens
	}

	out := make([]string, 0, len(tokens))
	out = append(out, kept...)
	out = append(out, tokens[primaryIdx:]...)
	out = append(out, pairs...)
	return out
}

// isSecondaryUnitPairAt reports designator/# + unit number at tokens[i].

func reorderLeadingSecondaryAtStart(tokens []string, isPR bool) []string {
	if len(tokens) < 3 {
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
	if desig, ok := numberedSecondaryDesignatorToken(tokens[0], isPR); ok {
		if looksLikeSecondaryNumber(tokens[1]) {
			rest := tokens[2:]
			out := make([]string, 0, len(tokens))
			out = append(out, rest...)
			out = append(out, desig, tokens[1])
			return out
		}
	}

	return tokens
}

// isNumberedSecondaryDesignator reports whether tok is a numbered secondary
// unit designator (English USPS or, when isPR, Puerto Rico Spanish).

func isSecondaryUnitPairAt(tokens []string, i int, isPR bool) bool {
	if i < 0 || i+1 >= len(tokens) {
		return false
	}
	if tokens[i] == "#" && looksLikeSecondaryNumber(tokens[i+1]) {
		return true
	}
	if _, ok := numberedSecondaryDesignatorToken(tokens[i], isPR); ok && looksLikeSecondaryNumber(tokens[i+1]) {
		return true
	}
	return false
}

func hasLaterStreetPrimary(tokens []string, start int, isPR bool) bool {
	for j := start; j < len(tokens)-1; j++ {
		if isSecondaryUnitPairAt(tokens, j, isPR) {
			j++ // skip unit number (loop also advances)
			continue
		}
		if looksLikeStreetPrimaryNumber(tokens[j]) {
			return true
		}
	}
	return false
}

// looksLikeStreetPrimaryNumber is looksLikeHouseNumber excluding ordinal street
// name tokens (12TH, 3RD) so "49TH" is not treated as a house number when
// deciding whether "Building 847" has a later primary.

func looksLikeStreetPrimaryNumber(tok string) bool {
	if isOrdinalHouseToken(tok) {
		return false
	}
	return looksLikeHouseNumber(tok)
}

// isOrdinalHouseToken reports digits + ST/ND/RD/TH (12TH, 3RD, 1ST, 2ND).

func secondaryPairTokens(tokens []string, i int, isPR bool) (desig, num string) {
	num = tokens[i+1]
	if tokens[i] == "#" {
		return "#", num
	}
	desig, _ = numberedSecondaryDesignatorToken(tokens[i], isPR)
	return desig, num
}

// hasLaterStreetPrimary reports whether tokens[start:] contain a street primary
// house number (skipping secondary designator+unit pairs).

func isOrdinalHouseToken(tok string) bool {
	upper := strings.ToUpper(tok)
	if len(upper) < 3 {
		return false
	}
	suf := upper[len(upper)-2:]
	body := upper[:len(upper)-2]
	if !isAllDigits(body) {
		return false
	}
	switch suf {
	case "ST", "ND", "RD", "TH":
		return true
	default:
		return false
	}
}

// reorderLeadingSecondaryAtStart handles pure leading "APT N rest" / "# N rest"
// when a primary house number was not detected mid-line (legacy path).

func isNumberedSecondaryDesignator(tok string, isPR bool) bool {
	_, ok := numberedSecondaryDesignatorToken(tok, isPR)
	return ok
}

// numberedSecondaryDesignatorToken returns the designator token to emit when
// reordering. English keeps the input spelling (peelSecondary normalizes);
// PR Spanish uses the short form from NormalizeSecondary.

func numberedSecondaryDesignatorToken(tok string, isPR bool) (string, bool) {
	if info, err := secondaryunit.Info(tok); err == nil && info.Numbered {
		return tok, true
	}
	if isPR {
		if short, err := puertorico.NormalizeSecondary(tok); err == nil {
			// PR secondaries that accept unit numbers are always treated as numbered.
			return short, true
		}
	}
	return "", false
}

// looksLikeSecondaryNumber reports whether tok is a plausible unit number
// (contains a digit). Pure alpha tokens are not treated as unit numbers so
// "APT SOUTH …" is not reordered.

// takeLeadingSecondary extracts a leading secondary designator (or #) plus its
// unit number so it can be folded first in multi-secondary order (preserving
// original left-to-right appearance). Patterns:
//
//	DESIGNATOR + NUMBER + rest  →  rest, peel{DES, NUMBER}
//	# + NUMBER + rest           →  rest, peel{"#", NUMBER}
//
// Glued "#NUMBER" is already split by expandHashTokens. Numbered designators
// without a following number are left unchanged (no invented unit number).
// Non-numbered designators at the start are not extracted.
func takeLeadingSecondary(tokens []string, isPR bool) ([]string, *secondaryPeel) {
	if len(tokens) < 3 {
		// Need designator/number plus at least one rest token.
		return tokens, nil
	}

	// "# NUMBER rest…"
	if tokens[0] == "#" && looksLikeSecondaryNumber(tokens[1]) {
		return tokens[2:], &secondaryPeel{designator: "#", number: tokens[1]}
	}

	// "APT NUMBER rest…" (numbered secondary designator only)
	if info, err := secondaryunit.Info(tokens[0]); err == nil && info.Numbered {
		if looksLikeSecondaryNumber(tokens[1]) {
			return tokens[2:], &secondaryPeel{designator: info.Short, number: tokens[1]}
		}
		// Numbered designator with no unit number — do not invent.
	}

	// PR Spanish secondary (e.g. APARTAMENTO 4 … / EDIF 2 …)
	if isPR {
		if short, err := puertorico.NormalizeSecondary(tokens[0]); err == nil {
			if looksLikeSecondaryNumber(tokens[1]) {
				return tokens[2:], &secondaryPeel{designator: short, number: tokens[1]}
			}
		}
	}

	return tokens, nil
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

// extractSecondary repeatedly extracts trailing secondary designators from the right,
// then folds them with an optional leading secondary (from takeLeadingSecondary)
// so Format preserves left-to-right designator order of original appearance.
//
// Overseas military "UNIT N BOX N" is handled earlier by the military fast path,
// so this peel only sees civilian street lines.
//
// The Address struct holds a single SecondaryDesignator + SecondaryNumber pair.
// Multiple secondaries fold as:
//   - ordered left-to-right: leading peel first, then trailing peels LTR
//   - first numbered peel becomes SecondaryDesignator (or first peel if none numbered)
//   - subsequent peels append onto SecondaryNumber ("420 RM 120", "3200 RM 12")
//   - non-numbered peels (UPPER/REAR) before the designator append at the end
//
// So "Building 420 Room 120" → BLDG / "420 RM 120", and
// "Unit 3200 152 Tech Drive Room 12" → UNIT / "3200 RM 12".
func extractSecondary(tokens []string, out *Address, isPR bool, leading *secondaryPeel) []string {
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

		// Designator alone (non-numbered, or numbered without a number).
		// When the bare token is also a street suffix (e.g. KEY → KY, TRAILER →
		// TRLR), leave it for the suffix peel rather than treating it as a unit
		// type: "8007 EAST KENTUCKY KEY" → E KENTUCKY KY, not secondary KEY.
		if info, err := secondaryunit.Info(tokens[len(tokens)-1]); err == nil {
			// Prefer street suffix when the same token scores as a suffix (e.g. KEY).
			if sc, _ := streetsuffixes.Score(tokens[len(tokens)-1]); sc > 0 {
				// Also a street suffix — prefer suffix over bare secondary.
				break
			}
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

	// Build left-to-right order: leading (original first) then reverse of
	// rightmost-first trailing peels.
	var ordered []secondaryPeel
	if leading != nil {
		ordered = append(ordered, *leading)
	}
	for i := len(peels) - 1; i >= 0; i-- {
		ordered = append(ordered, peels[i])
	}
	if len(ordered) == 0 {
		return tokens
	}

	des, num := foldSecondaryPeels(ordered)
	out.SecondaryDesignator = des
	out.SecondaryNumber = num
	return tokens
}

// foldSecondaryPeels combines ordered (left-to-right) secondary peels into a
// single designator + number pair for Address / Format.
func foldSecondaryPeels(ordered []secondaryPeel) (des, num string) {
	if len(ordered) == 0 {
		return "", ""
	}
	// Prefer the first numbered peel as the primary designator so non-numbered
	// markers (UPPER/REAR) do not displace UNIT/BLDG when they appear first.
	firstNum := -1
	for i, p := range ordered {
		if p.number != "" {
			firstNum = i
			break
		}
	}
	if firstNum < 0 {
		des = ordered[0].designator
		for i := 1; i < len(ordered); i++ {
			if num != "" {
				num += " "
			}
			num += ordered[i].designator
		}
		return des, num
	}
	des = ordered[firstNum].designator
	num = ordered[firstNum].number
	for i := firstNum + 1; i < len(ordered); i++ {
		p := ordered[i]
		if p.number != "" {
			num += " " + p.designator + " " + p.number
		} else {
			num += " " + p.designator
		}
	}
	// Non-numbered peels that appeared before the designator append at the end
	// (Unit 3200 … Upper with Upper left of Unit after extraction is rare;
	// trailing "… Upper Unit 3200" still yields UNIT / "3200 UPPR").
	for i := 0; i < firstNum; i++ {
		num += " " + ordered[i].designator
	}
	return des, num
}

// looksLikePrimaryNumber reports whether tok is a plausible primary address number
// (contains a digit; hyphenated ranges like 112-10 qualify). Broader than
// looksLikeHouseNumber — used when consuming an already-positioned primary token.
func looksLikePrimaryNumber(tok string) bool {
	for _, r := range tok {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// looksLikeHouseNumber reports whether tok is a primarily numeric house number
// suitable for pre-street splitting. Stricter than looksLikePrimaryNumber so
// digit-leading product/firm tokens like "3M" are rejected while common US
// forms still match:
//
//	100, 5          pure digits
//	100A            digits (≥2) + single trailing letter (apartment style)
//	112-10, 112-10A hyphenated ranges
//	12TH, 3RD       ordinals
//
// Rejected: "3M" (single digit + letter product code), "A1" (letter-leading),
// bare fractions ("1/2"), grid decimals ("39.4").
func looksLikeHouseNumber(tok string) bool {
	if tok == "" {
		return false
	}
	upper := strings.ToUpper(tok)

	// Must start with a digit.
	if upper[0] < '0' || upper[0] > '9' {
		return false
	}

	// Ordinal: digits + ST/ND/RD/TH (12TH, 3RD, 1ST, 2ND).
	if len(upper) >= 3 {
		suf := upper[len(upper)-2:]
		body := upper[:len(upper)-2]
		if isAllDigits(body) {
			switch suf {
			case "ST", "ND", "RD", "TH":
				return true
			}
		}
	}

	// Hyphenated primary range: 112-10, 112-10A, 19-01.
	if strings.Contains(upper, "-") {
		parts := strings.Split(upper, "-")
		if len(parts) < 2 || !isAllDigits(parts[0]) {
			return false
		}
		for _, p := range parts[1:] {
			if p == "" || !isAlphanumeric(p) {
				return false
			}
		}
		return true
	}

	// Pure digits.
	if isAllDigits(upper) {
		return true
	}

	// Digits + single trailing letter (100A). Require ≥2 digits so single-digit
	// product codes like "3M" are not treated as house numbers during pre-street
	// split. (Bare "1A Main St" still parses via looksLikePrimaryNumber after
	// splitPreStreet leaves tokens unchanged.)
	if len(upper) >= 3 {
		last := upper[len(upper)-1]
		digits := upper[:len(upper)-1]
		if last >= 'A' && last <= 'Z' && isAllDigits(digits) && len(digits) >= 2 {
			return true
		}
	}

	return false
}

// isAlphanumeric reports whether s is non-empty and only letters/digits.
func isAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
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
