package highways

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/region"
)

/*
- Because county, state, and local highways are used as street names, these are not abbreviated.
- Please note that if the highway is a state highway, then the state is abbreviated following abbreviations in
Appendix D, but the word highway is not abbreviated.
- Note: When the name of a state is used as a portion of the Primary Street Name, developers SHOULD use
the standard two–letter abbreviation. However, when the state name is the complete Primary Street Name,
such as OKLAHOMA AVE, then the state name SHOULD be spelled out completely

  Example                               Project US@
COUNTY HIGHWAY 140             <->    COUNTY HIGHWAY 140
The word county is not abbreviated if part of a street name.

COUNTY HWY 60E                 <->    COUNTY HIGHWAY 60E
CNTY HWY 20                    <->    COUNTY HIGHWAY 20
Neither the word county nor the word highway should be abbreviated because they are
part of the street name.

COUNTY RD 441                  <->    COUNTY ROAD 441
Road is not abbreviated because it is part of the street name.

CR 1185                        <->    COUNTY ROAD 1185
CNTY RD 33                     <->    COUNTY ROAD 33
CA COUNTY RD 150               <->    CA COUNTY ROAD 150
Road is not abbreviated because it is part of the street name.

CALIFORNIA COUNTY ROAD 555     <->    CA COUNTY ROAD 555
EXPRESSWAY 55                  <->    EXPRESSWAY 55
FARM to MARKET 1200            <->    FM 1200
Farm to Market is always abbreviated to FM.

FM 187                         <->    FM 187

HWY FM 1320                    <->    FM 1320
It is incorrect to place the word Highway (or HWY) before FM

HWY 64                         <->    HIGHWAY 64
The word highway is not abbreviated because it is part of the street name.

HWY 11 BYPASS                  <->    HIGHWAY 11 BYP
The word bypass is abbreviated as a suffix, and not part of the street name.

HWY 66 FRONTAGE ROAD           <->    HIGHWAY 66 FRONTAGE RD
The word frontage is never abbreviated. Road is abbreviated as a suffix.

HIGHWAY 3 BYP ROAD             <->    HIGHWAY 3 BYPASS RD
The word bypass is not abbreviated because it is part of the street name.

I10                            <->    INTERSTATE 10
The word interstate is never abbreviated.

IH280                          <->    INTERSTATE 280
INTERSTATE HWY 680             <->    INTERSTATE 680

I 55 BYPASS                    <->    INTERSTATE 55 BYP
Interstate is never abbreviated. Bypass is abbreviated as a suffix.

I 26 BYP ROAD                  <->    INTERSTATE 26 BYPASS RD
Bypass is not abbreviated as it is part of the street name.

I 44 FRONTAGE ROAD             <->    INTERSTATE 44 FRONTAGE RD
Road is abbreviated as a suffix.

RD 5A                          <->    ROAD 5A
Road is not abbreviated if it is part of the street name.

RT 88                          <->    ROUTE 88
The word route is only abbreviated if it is a suffix, but not part of the street name.

RTE 95                         <->    ROUTE 95
RANCH RD 620                   <->    RANCH ROAD 620
ST HIGHWAY 303                 <->    STATE HIGHWAY 303
STATE HWY 60                   <->    STATE HIGHWAY 60
SR 220                         <->    STATE ROAD 220
ST RD 86                       <->    STATE ROAD 86
SR MM                          <->    STATE ROUTE MM
ST RT 175                      <->    STATE ROUTE 175
STATE RTE 260                  <->    STATE ROUTE 260
TOWNSHIP RD 20                 <->    TOWNSHIP ROAD 20
Road is not abbreviated as it is part of the street name.

TSR 45                         <->    TOWNSHIP ROAD 45
The word township is never abbreviated.

US 41 SW                       <->    US HIGHWAY 41 SW
US HWY 44                      <->    US HIGHWAY 44
KENTUCKY 440                   <->    KY HIGHWAY 440
KENTUCKY HIGHWAY 189           <->    KY HIGHWAY 189

State names that are part of street names may be abbreviated following Appendix D.
KY 1207                        <->    KY HIGHWAY 1207
KY HWY 75                      <->    KY HIGHWAY 75
KY ST HWY 1                    <->    KY STATE HIGHWAY 1
The word state is not abbreviated if part of a street name.

KENTUCKY STATE HIGHWAY 625     <->   KY STATE HIGHWAY 625

*/

// gluedInterstate matches I/IH immediately followed by a route designator (e.g. I10, IH280).
var gluedInterstate = regexp.MustCompile(`^(IH?)(\d+[A-Z]*)$`)

// routeID matches a highway/route designator token (digits with optional letter suffix, or letter-only).
var routeID = regexp.MustCompile(`^(\d+[A-Z]*|[A-Z]+)$`)

// digitRouteID matches a digit-bearing route designator (e.g. 440, 60E, 5A). Letter-only tokens are excluded.
var digitRouteID = regexp.MustCompile(`^\d+[A-Z]*$`)

// multiWordStateNames are full US state/possession names (sorted longest-first for greedy match).
// Built from region keys that contain a space and are not already two-letter codes.
var multiWordStateNames = []string{
	"DISTRICT OF COLUMBIA",
	"FEDERATED STATES OF MICRONESIA",
	"NORTHERN MARIANA ISLANDS",
	"NEW HAMPSHIRE",
	"NEW JERSEY",
	"NEW MEXICO",
	"NEW YORK",
	"NORTH CAROLINA",
	"NORTH DAKOTA",
	"RHODE ISLAND",
	"SOUTH CAROLINA",
	"SOUTH DAKOTA",
	"WEST VIRGINIA",
	"AMERICAN SAMOA",
	"MARSHALL ISLANDS",
	"PUERTO RICO",
	"VIRGIN ISLANDS",
}

// singleWordStateNames are full single-token US state names (not abbreviations).
var singleWordStateNames = map[string]bool{
	"ALABAMA": true, "ALASKA": true, "ARIZONA": true, "ARKANSAS": true,
	"CALIFORNIA": true, "COLORADO": true, "CONNECTICUT": true, "DELAWARE": true,
	"FLORIDA": true, "GEORGIA": true, "GUAM": true, "HAWAII": true,
	"IDAHO": true, "ILLINOIS": true, "INDIANA": true, "IOWA": true,
	"KANSAS": true, "KENTUCKY": true, "LOUISIANA": true, "MAINE": true,
	"MARYLAND": true, "MASSACHUSETTS": true, "MICHIGAN": true, "MINNESOTA": true,
	"MISSISSIPPI": true, "MISSOURI": true, "MONTANA": true, "NEBRASKA": true,
	"NEVADA": true, "OHIO": true, "OKLAHOMA": true, "OREGON": true,
	"PALAU": true, "PENNSYLVANIA": true, "TENNESSEE": true, "TEXAS": true,
	"UTAH": true, "VERMONT": true, "VIRGINIA": true, "WASHINGTON": true,
	"WISCONSIN": true, "WYOMING": true, "MICRONESIA": true,
}

// usStateAbbrevs are the two-letter US state/possession codes we treat as highway state prefixes.
// Excludes ambiguous codes that double as highway vocabulary (FM = Farm to Market).
var usStateAbbrevs = map[string]bool{
	"AL": true, "AK": true, "AS": true, "AZ": true, "AR": true, "CA": true,
	"CO": true, "CT": true, "DE": true, "DC": true, "FL": true, "GA": true,
	"GU": true, "HI": true, "ID": true, "IL": true, "IN": true, "IA": true,
	"KS": true, "KY": true, "LA": true, "ME": true, "MH": true, "MD": true,
	"MA": true, "MI": true, "MN": true, "MS": true, "MO": true, "MT": true,
	"NE": true, "NV": true, "NH": true, "NJ": true, "NM": true, "NY": true,
	"NC": true, "ND": true, "MP": true, "OH": true, "OK": true, "OR": true,
	"PW": true, "PA": true, "PR": true, "RI": true, "SC": true, "SD": true,
	"TN": true, "TX": true, "UT": true, "VT": true, "VI": true, "VA": true,
	"WA": true, "WV": true, "WI": true, "WY": true,
}

// NormalizeStreetName normalizes highway-style primary street names per Project US@.
// Input is the street name portion only (not full address). Returns uppercase.
// Names that match no highway rule are returned uppercased with collapsed whitespace.
// An error is returned only when the input is empty after trim.
func NormalizeStreetName(name string) (string, error) {
	s := collapseSpace(strings.ToUpper(strings.TrimSpace(name)))
	if s == "" {
		return "", fmt.Errorf("empty street name")
	}

	// Split glued interstate forms (I10, IH280) into separate tokens before field split.
	tokens := strings.Fields(expandGluedInterstate(s))

	if out, ok := normalizeTokens(tokens); ok {
		return out, nil
	}

	// No highway rule matched: pass through uppercased / collapsed form.
	return s, nil
}

func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// expandGluedInterstate rewrites tokens like I10 / IH280 to "I 10" / "IH 280".
func expandGluedInterstate(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		if m := gluedInterstate.FindStringSubmatch(p); m != nil {
			parts[i] = m[1] + " " + m[2]
		}
	}
	return strings.Join(parts, " ")
}

// normalizeTokens attempts ordered highway rewrites. ok is false if no rule applies.
func normalizeTokens(tokens []string) (string, bool) {
	if len(tokens) == 0 {
		return "", false
	}

	// Optional leading state name/abbrev, then remaining highway form.
	if abbrev, rest, ok := splitLeadingState(tokens); ok && len(rest) > 0 {
		if core, ok := matchHighwayCore(rest); ok {
			return join(append([]string{abbrev}, core...)), true
		}
		// State + bare route number → ST HIGHWAY N
		// Require a digit-bearing designator so non-highway names like OKLAHOMA AVE,
		// WASHINGTON BLVD, CA MAIN pass through. Letter-only routes are only for SR MM.
		if isDigitRouteID(rest[0]) {
			core := append([]string{"HIGHWAY", rest[0]}, normalizeSuffixes(rest[1:])...)
			return join(append([]string{abbrev}, core...)), true
		}
	}

	if core, ok := matchHighwayCore(tokens); ok {
		return join(core), true
	}
	return "", false
}

// splitLeadingState peels a full state name or two-letter US abbrev from the front of tokens.
func splitLeadingState(tokens []string) (abbrev string, rest []string, ok bool) {
	// Multi-word full names first (longest match).
	joined := join(tokens)
	for _, name := range multiWordStateNames {
		if joined == name {
			return "", nil, false // state name alone is not a highway form
		}
		prefix := name + " "
		if strings.HasPrefix(joined, prefix) {
			ab, err := region.NormalizeRegion(name)
			if err != nil {
				break
			}
			// Count how many tokens the name consumed.
			n := len(strings.Fields(name))
			return ab, tokens[n:], true
		}
	}

	// Single-word full state name.
	if singleWordStateNames[tokens[0]] {
		ab, err := region.NormalizeRegion(tokens[0])
		if err != nil {
			return "", nil, false
		}
		return ab, tokens[1:], true
	}

	// Two-letter state abbreviation (exclude FM — farm-to-market).
	if len(tokens[0]) == 2 && usStateAbbrevs[tokens[0]] {
		return tokens[0], tokens[1:], true
	}

	return "", nil, false
}

// matchHighwayCore matches highway vocabulary from the start of tokens (no leading state).
// Returns the full normalized token list including trailing suffixes when matched.
func matchHighwayCore(tokens []string) ([]string, bool) {
	if len(tokens) == 0 {
		return nil, false
	}

	// --- Farm to Market / FM ---
	// FARM TO MARKET N
	if len(tokens) >= 4 && tokens[0] == "FARM" && tokens[1] == "TO" && tokens[2] == "MARKET" && isRouteID(tokens[3]) {
		return withRouteAndSuffixes([]string{"FM"}, tokens[3:]), true
	}
	// HWY/HIGHWAY FM N
	if len(tokens) >= 3 && isHwy(tokens[0]) && tokens[1] == "FM" && isRouteID(tokens[2]) {
		return withRouteAndSuffixes([]string{"FM"}, tokens[2:]), true
	}
	// FM N
	if len(tokens) >= 2 && tokens[0] == "FM" && isRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"FM"}, tokens[1:]), true
	}

	// --- Interstate ---
	// INTERSTATE [HWY|HIGHWAY] N ... (digit-bearing route IDs only)
	if tokens[0] == "INTERSTATE" {
		i := 1
		if i < len(tokens) && isHwy(tokens[i]) {
			i++
		}
		if i < len(tokens) && isDigitRouteID(tokens[i]) {
			return withRouteAndSuffixes([]string{"INTERSTATE"}, tokens[i:]), true
		}
	}
	// I N ...  or  IH N ... (digit-bearing; "I STREET" must pass through)
	if (tokens[0] == "I" || tokens[0] == "IH") && len(tokens) >= 2 && isDigitRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"INTERSTATE"}, tokens[1:]), true
	}

	// --- US highway ---
	// US [HWY|HIGHWAY] N ... (digit-bearing; "US GRANT" must pass through)
	if tokens[0] == "US" && len(tokens) >= 2 {
		i := 1
		if i < len(tokens) && isHwy(tokens[i]) {
			i++
		}
		if i < len(tokens) && isDigitRouteID(tokens[i]) {
			return withRouteAndSuffixes([]string{"US", "HIGHWAY"}, tokens[i:]), true
		}
	}

	// --- County ---
	// CR N
	if tokens[0] == "CR" && len(tokens) >= 2 && isRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"COUNTY", "ROAD"}, tokens[1:]), true
	}
	// COUNTY|CNTY (HWY|HIGHWAY|RD|ROAD) N
	if isCounty(tokens[0]) && len(tokens) >= 3 {
		kind := highwayKind(tokens[1])
		if kind != "" && isRouteID(tokens[2]) {
			return withRouteAndSuffixes([]string{"COUNTY", kind}, tokens[2:]), true
		}
	}

	// --- Township ---
	// TSR N
	if tokens[0] == "TSR" && len(tokens) >= 2 && isRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"TOWNSHIP", "ROAD"}, tokens[1:]), true
	}
	// TOWNSHIP (RD|ROAD) N
	if tokens[0] == "TOWNSHIP" && len(tokens) >= 3 && isRoad(tokens[1]) && isRouteID(tokens[2]) {
		return withRouteAndSuffixes([]string{"TOWNSHIP", "ROAD"}, tokens[2:]), true
	}

	// --- Ranch ---
	// RANCH (RD|ROAD) N
	if tokens[0] == "RANCH" && len(tokens) >= 3 && isRoad(tokens[1]) && isRouteID(tokens[2]) {
		return withRouteAndSuffixes([]string{"RANCH", "ROAD"}, tokens[2:]), true
	}

	// --- State highway / road / route ---
	// STATE|ST (HWY|HIGHWAY|RD|ROAD|RT|RTE|ROUTE) N
	if isStateWord(tokens[0]) && len(tokens) >= 3 {
		kind := stateHighwayKind(tokens[1])
		if kind != "" && isRouteID(tokens[2]) {
			return withRouteAndSuffixes([]string{"STATE", kind}, tokens[2:]), true
		}
	}
	// SR N  — STATE ROAD when route has a digit, STATE ROUTE when letter-only (SR MM)
	if tokens[0] == "SR" && len(tokens) >= 2 && isRouteID(tokens[1]) {
		kind := "ROAD"
		if isLetterOnly(tokens[1]) {
			kind = "ROUTE"
		}
		return withRouteAndSuffixes([]string{"STATE", kind}, tokens[1:]), true
	}

	// --- Bare highway / road / route / expressway ---
	// Digit-bearing route IDs only (letter-only is reserved for SR MM).
	// HWY|HIGHWAY N
	if isHwy(tokens[0]) && len(tokens) >= 2 && isDigitRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"HIGHWAY"}, tokens[1:]), true
	}
	// RD|ROAD N  (road is part of the street name)
	if isRoad(tokens[0]) && len(tokens) >= 2 && isDigitRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"ROAD"}, tokens[1:]), true
	}
	// RT|RTE|ROUTE N
	if isRoute(tokens[0]) && len(tokens) >= 2 && isDigitRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"ROUTE"}, tokens[1:]), true
	}
	// EXPRESSWAY N
	if tokens[0] == "EXPRESSWAY" && len(tokens) >= 2 && isDigitRouteID(tokens[1]) {
		return withRouteAndSuffixes([]string{"EXPRESSWAY"}, tokens[1:]), true
	}

	return nil, false
}

// withRouteAndSuffixes emits [prefix..., routeID, ...normalized suffixes].
// tokens must start with the route ID.
func withRouteAndSuffixes(prefix []string, tokens []string) []string {
	out := make([]string, 0, len(prefix)+len(tokens))
	out = append(out, prefix...)
	out = append(out, tokens[0])
	if len(tokens) > 1 {
		out = append(out, normalizeSuffixes(tokens[1:])...)
	}
	return out
}

// normalizeSuffixes applies Project US@ suffix rules after the route number:
//   - BYPASS / BYP alone → BYP
//   - BYP|BYPASS + ROAD|RD → BYPASS RD  (bypass is part of the name)
//   - FRONTAGE + ROAD|RD → FRONTAGE RD
//   - trailing ROAD|RD → RD
func normalizeSuffixes(tokens []string) []string {
	if len(tokens) == 0 {
		return nil
	}
	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch {
		case isBypass(tok):
			// If followed by ROAD, bypass is part of the street name (full form) and ROAD is the suffix.
			if i+1 < len(tokens) && isRoad(tokens[i+1]) {
				out = append(out, "BYPASS", "RD")
				i++ // consume ROAD
				continue
			}
			// Lone bypass is a street suffix.
			out = append(out, "BYP")
		case tok == "FRONTAGE":
			out = append(out, "FRONTAGE")
			if i+1 < len(tokens) && isRoad(tokens[i+1]) {
				out = append(out, "RD")
				i++
			}
		case isRoad(tok):
			// Trailing road as suffix after the highway number.
			out = append(out, "RD")
		default:
			out = append(out, tok)
		}
	}
	return out
}

func isHwy(t string) bool {
	return t == "HWY" || t == "HIGHWAY"
}

func isRoad(t string) bool {
	return t == "RD" || t == "ROAD"
}

func isRoute(t string) bool {
	return t == "RT" || t == "RTE" || t == "ROUTE"
}

func isBypass(t string) bool {
	return t == "BYP" || t == "BYPASS" || t == "BYPA" || t == "BYPAS" || t == "BYPS"
}

func isCounty(t string) bool {
	return t == "COUNTY" || t == "CNTY"
}

func isStateWord(t string) bool {
	return t == "STATE" || t == "ST"
}

func highwayKind(t string) string {
	switch {
	case isHwy(t):
		return "HIGHWAY"
	case isRoad(t):
		return "ROAD"
	default:
		return ""
	}
}

// stateHighwayKind maps STATE/ST + second token to HIGHWAY / ROAD / ROUTE.
func stateHighwayKind(t string) string {
	switch {
	case isHwy(t):
		return "HIGHWAY"
	case isRoad(t):
		return "ROAD"
	case isRoute(t):
		return "ROUTE"
	default:
		return ""
	}
}

func isRouteID(t string) bool {
	if t == "" {
		return false
	}
	// Reject known highway vocabulary that can appear where a route id is expected.
	switch t {
	case "HWY", "HIGHWAY", "RD", "ROAD", "RT", "RTE", "ROUTE",
		"BYP", "BYPASS", "FRONTAGE", "COUNTY", "CNTY", "STATE", "ST",
		"TOWNSHIP", "RANCH", "US", "FM", "INTERSTATE", "EXPRESSWAY",
		"SR", "CR", "TSR", "IH", "I":
		return false
	}
	return routeID.MatchString(t)
}

// isDigitRouteID is true for digit-bearing route designators (optional trailing letters).
// Used for state + bare-route rewrites so letter-only street tokens are not treated as routes.
func isDigitRouteID(t string) bool {
	if t == "" {
		return false
	}
	return digitRouteID.MatchString(t)
}

func isLetterOnly(t string) bool {
	if t == "" {
		return false
	}
	for _, r := range t {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func join(parts []string) string {
	return strings.Join(parts, " ")
}
