package region

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hbollon/go-edlib"
)

/*
State or Possession Postal Abbreviations

Alabama AL
Alaska AK
American Samoa AS
Arizona AZ
Arkansas AR
California CA
Colorado CO
Connecticut CT
Delaware DE
District of Columbia DC
Federated States of Micronesia FM
Florida FL
Georgia GA
Guam GU
Hawaii HI
Idaho ID
Illinois IL
Indiana IN
Iowa IA
Kansas KS
Kentucky KY
Louisiana LA
Maine ME
Marshall Islands MH
Maryland MD
Massachusetts MA
Michigan MI
Minnesota MN
Mississippi MS
Missouri MO
Montana MT
Nebraska NE
Nevada NV
New Hampshire NH
New Jersey NJ
New Mexico NM
New York NY
North Carolina NC
North Dakota ND
Northern Mariana Islands MP
Ohio OH
Oklahoma OK
Oregon OR
Palau PW
Pennsylvania PA
Puerto Rico PR
Rhode Island RI
South Carolina SC
South Dakota SD
Tennessee TN
Texas TX
Utah UT
Vermont VT
Virgin Islands VI
Virginia VA
Washington WA
West Virginia WV
Wisconsin WI
Wyoming WY

Canadian Province/Territory Postal Service Abbreviations

Alberta                     AB
British Columbia            BC
Manitoba                    MB
New Brunswick               NB
Newfoundland and Labrador   NL
Northwest Territories       NT
Nova Scotia                 NS
Nunavat Territory           NU
Ontario                     ON
Prince Edward Island        PE
Quebec                      QC
Saskatchewan                SK
Yukon Territory             YT

Military "State" Abbreviations

Armed Forces Europe, the Middle East, and Canada   AE
Armed Forces Pacific                               AP
Armed Forces Americas (except Canada)              AA
*/

var regionMap = map[string]string{
	"ALABAMA":                        "AL",
	"AL":                             "AL",
	"ALASKA":                         "AK",
	"AK":                             "AK",
	"AMERICAN SAMOA":                 "AS",
	"AS":                             "AS",
	"ARIZONA":                        "AZ",
	"AZ":                             "AZ",
	"ARKANSAS":                       "AR",
	"AR":                             "AR",
	"CALIFORNIA":                     "CA",
	"CA":                             "CA",
	"COLORADO":                       "CO",
	"CO":                             "CO",
	"CONNECTICUT":                    "CT",
	"CT":                             "CT",
	"DELAWARE":                       "DE",
	"DELEWARE":                       "DE", // common misspelling
	"DE":                             "DE",
	"DISTRICT OF COLUMBIA":           "DC",
	"DC":                             "DC",
	"FEDERATED STATES OF MICRONESIA": "FM",
	"MICRONESIA":                     "FM",
	"FM":                             "FM",
	"FLORIDA":                        "FL",
	"FL":                             "FL",
	"GEORGIA":                        "GA",
	"GA":                             "GA",
	"GUAM":                           "GU",
	"GU":                             "GU",
	"HAWAII":                         "HI",
	"HI":                             "HI",
	"IDAHO":                          "ID",
	"ID":                             "ID",
	"ILLINOIS":                       "IL",
	"IL":                             "IL",
	"INDIANA":                        "IN",
	"IN":                             "IN",
	"IOWA":                           "IA",
	"IA":                             "IA",
	"KANSAS":                         "KS",
	"KS":                             "KS",
	"KENTUCKY":                       "KY",
	"KY":                             "KY",
	"LOUISIANA":                      "LA",
	"LA":                             "LA",
	"MAINE":                          "ME",
	"ME":                             "ME",
	"MARSHALL ISLANDS":               "MH",
	"MARSHALL IS":                    "MH",
	"MARSHALL ISL":                   "MH",
	"MARSHALL ISLS":                  "MH",
	"MARSHALL ISS":                   "MH",
	"MARSHALL ISLD":                  "MH",
	"MH":                             "MH",
	"MARYLAND":                       "MD",
	"MD":                             "MD",
	"MASSACHUSETTS":                  "MA",
	"MA":                             "MA",
	"MICHIGAN":                       "MI",
	"MI":                             "MI",
	"MINNESOTA":                      "MN",
	"MN":                             "MN",
	"MISSISSIPPI":                    "MS",
	"MS":                             "MS",
	"MISSOURI":                       "MO",
	"MO":                             "MO",
	"MONTANA":                        "MT",
	"MT":                             "MT",
	"NEBRASKA":                       "NE",
	"NE":                             "NE",
	"NEVADA":                         "NV",
	"NV":                             "NV",
	"NEW HAMPSHIRE":                  "NH",
	"NH":                             "NH",
	"NEW JERSEY":                     "NJ",
	"NJ":                             "NJ",
	"NEW MEXICO":                     "NM",
	"NM":                             "NM",
	"NEW YORK":                       "NY",
	"NY":                             "NY",
	"NORTH CAROLINA":                 "NC",
	"N CAROLINA":                     "NC",
	"NC":                             "NC",
	"NORTH DAKOTA":                   "ND",
	"N DAKOTA":                       "ND",
	"ND":                             "ND",
	"NORTHERN MARIANA ISLANDS":       "MP",
	"NORTHERN MARIANA IS":            "MP",
	"NORTHERN MARIANA ISL":           "MP",
	"NORTHERN MARIANA ISLS":          "MP",
	"NORTHERN MARIANA ISS":           "MP",
	"NORTHERN MARIANA ISLD":          "MP",
	"N MARIANA ISLANDS":              "MP",
	"N MARIANA IS":                   "MP",
	"N MARIANA ISL":                  "MP",
	"N MARIANA ISLS":                 "MP",
	"N MARIANA ISS":                  "MP",
	"N MARIANA ISLD":                 "MP",
	"MP":                             "MP",
	"OHIO":                           "OH",
	"OH":                             "OH",
	"OKLAHOMA":                       "OK",
	"OK":                             "OK",
	"OREGON":                         "OR",
	"OR":                             "OR",
	"PALAU":                          "PW",
	"PW":                             "PW",
	"PENNSYLVANIA":                   "PA",
	"PA":                             "PA",
	"PUERTO RICO":                    "PR",
	"PR":                             "PR",
	"RHODE ISLAND":                   "RI",
	"RHODE IS":                       "RI",
	"RHODE ISL":                      "RI",
	"RHODE ISLD":                     "RI",
	"RI":                             "RI",
	"SOUTH CAROLINA":                 "SC",
	"S CAROLINA":                     "SC",
	"SC":                             "SC",
	"SOUTH DAKOTA":                   "SD",
	"S DAKOTA":                       "SD",
	"SD":                             "SD",
	"TENNESSEE":                      "TN",
	"TN":                             "TN",
	"TEXAS":                          "TX",
	"TX":                             "TX",
	"UTAH":                           "UT",
	"UT":                             "UT",
	"VERMONT":                        "VT",
	"VT":                             "VT",
	"VIRGIN ISLANDS":                 "VI",
	"VIRGIN IS":                      "VI",
	"VIRGIN ISL":                     "VI",
	"VIRGIN ISLS":                    "VI",
	"VIRGIN ISS":                     "VI",
	"VIRGIN ISLD":                    "VI",
	"US VIRGIN ISLANDS":              "VI",
	"US VIRGIN IS":                   "VI",
	"US VIRGIN ISL":                  "VI",
	"US VIRGIN ISLS":                 "VI",
	"US VIRGIN ISS":                  "VI",
	"US VIRGIN ISLD":                 "VI",
	"USVI":                           "VI",
	"VIS":                            "VI",
	"USA VI":                         "VI",
	"VI USA":                         "VI",
	"VI":                             "VI",
	"VIRGINIA":                       "VA",
	"VA":                             "VA",
	"WASHINGTON":                     "WA",
	"WA":                             "WA",
	"WEST VIRGINIA":                  "WV",
	"W VIRGINIA":                     "WV",
	"WV":                             "WV",
	"WISCONSIN":                      "WI",
	"WI":                             "WI",
	"WYOMING":                        "WY",
	"WY":                             "WY",

	"ALBERTA":                   "AB",
	"AB":                        "AB",
	"BRITISH COLUMBIA":          "BC",
	"BC":                        "BC",
	"MANITOBA":                  "MB",
	"MB":                        "MB",
	"NEW BRUNSWICK":             "NB",
	"NB":                        "NB",
	"NEWFOUNDLAND AND LABRADOR": "NL",
	"NEWFOUNDLAND":              "NL",
	"LABRADOR":                  "NL",
	"NL":                        "NL",
	"NORTHWEST TERRITORIES":     "NT",
	"NORTHWEST TERR":            "NT",
	"NW TERRITORIES":            "NT",
	"NW TERR":                   "NT",
	"NT":                        "NT",
	"NOVA SCOTIA":               "NS",
	"NS":                        "NS",
	"NUNAVAT TERRITORY":         "NU",
	"NUNAVAT TERR":              "NU",
	"NU":                        "NU",
	"ONTARIO":                   "ON",
	"ON":                        "ON",
	"PRINCE EDWARD ISLAND":      "PE",
	"PRINCE EDWARD IS":          "PE",
	"PRINCE EDWARD ISL":         "PE",
	"PRINCE EDWARD ISLD":        "PE",
	"PE":                        "PE",
	"QUEBEC":                    "QC",
	"QC":                        "QC",
	"SASKATCHEWAN":              "SK",
	"SK":                        "SK",
	"YUKON TERRITORY":           "YT",
	"YUKON TERR":                "YT",
	"YUKON":                     "YT",
	"YT":                        "YT",

	"ARMED FORCES EUROPE THE MIDDLE EAST AND CANADA": "AE",
	"ARMED FORCES EUROPE":                            "AE",
	"AE":                                             "AE",
	"ARMED FORCES PACIFIC":                           "AP",
	"AP":                                             "AP",
	"ARMED FORCES AMERICA":                           "AA",
	"ARMED FORCES AMERICAS":                          "AA",
	"AA":                                             "AA",
}

var regionKeys = slices.Collect(maps.Keys(regionMap))
var alphaspace = regexp.MustCompile("[^a-zA-Z ]+")

func normalizeRegion(r string, fuzzy bool) (string, error) {
	// clean out any punctuation
	clean := alphaspace.ReplaceAllString(r, "")
	// capitalize
	capitalized := strings.ToUpper(clean)
	// if requested, fuzzy match keys
	rkey := capitalized
	if fuzzy && len(capitalized) > 3 {
		matched, err := edlib.FuzzySearchThreshold(capitalized, regionKeys, 0.7, edlib.DamerauLevenshtein)
		if err != nil {
			// TODO: figure out how to let the user control logging in this library
			// log warn "Unable to fuzzy match supplied region string"
			matched = capitalized
		} else {
			rkey = matched
		}
	}

	// look up the abbreviation
	abrev, ok := regionMap[rkey]
	if !ok {
		return "", fmt.Errorf("Unrecognized state, possession, Canadian provice, or US Armed Forces region")
	}

	return abrev, nil
}

func NormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, false)
}

func FuzzyNormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, true)
}

// Score returns how strongly token looks like a region (state/province/military).
// 0 means not a region; higher is more confident. Exact two-letter codes and full
// names score highest; multi-word names should be scored via ScorePhrase.
func Score(token string) (int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, nil
	}
	if _, err := NormalizeRegion(token); err != nil {
		return 0, nil
	}
	u := strings.ToUpper(alphaspace.ReplaceAllString(token, ""))
	// Two-letter postal code: strongest single-token region signal.
	if len(u) == 2 {
		return 100, nil
	}
	// Full name or alias.
	if len(strings.Fields(u)) == 1 {
		return 90, nil
	}
	return 85, nil
}

// ScorePhrase scores a multi-token region candidate (e.g. "SOUTH CAROLINA").
func ScorePhrase(phrase string) (int, error) {
	return Score(phrase)
}

// usStateFullNames maps US state/possession codes to fully spelled primary names.
// Excludes military AE/AP/AA and Canadian provinces.
var usStateFullNames = map[string]string{
	"AL": "ALABAMA", "AK": "ALASKA", "AS": "AMERICAN SAMOA", "AZ": "ARIZONA",
	"AR": "ARKANSAS", "CA": "CALIFORNIA", "CO": "COLORADO", "CT": "CONNECTICUT",
	"DE": "DELAWARE", "DC": "DISTRICT OF COLUMBIA", "FM": "FEDERATED STATES OF MICRONESIA",
	"FL": "FLORIDA", "GA": "GEORGIA", "GU": "GUAM", "HI": "HAWAII", "ID": "IDAHO",
	"IL": "ILLINOIS", "IN": "INDIANA", "IA": "IOWA", "KS": "KANSAS", "KY": "KENTUCKY",
	"LA": "LOUISIANA", "ME": "MAINE", "MH": "MARSHALL ISLANDS", "MD": "MARYLAND",
	"MA": "MASSACHUSETTS", "MI": "MICHIGAN", "MN": "MINNESOTA", "MS": "MISSISSIPPI",
	"MO": "MISSOURI", "MT": "MONTANA", "NE": "NEBRASKA", "NV": "NEVADA",
	"NH": "NEW HAMPSHIRE", "NJ": "NEW JERSEY", "NM": "NEW MEXICO", "NY": "NEW YORK",
	"NC": "NORTH CAROLINA", "ND": "NORTH DAKOTA", "MP": "NORTHERN MARIANA ISLANDS",
	"OH": "OHIO", "OK": "OKLAHOMA", "OR": "OREGON", "PW": "PALAU", "PA": "PENNSYLVANIA",
	"PR": "PUERTO RICO", "RI": "RHODE ISLAND", "SC": "SOUTH CAROLINA", "SD": "SOUTH DAKOTA",
	"TN": "TENNESSEE", "TX": "TEXAS", "UT": "UTAH", "VT": "VERMONT", "VI": "VIRGIN ISLANDS",
	"VA": "VIRGINIA", "WA": "WASHINGTON", "WV": "WEST VIRGINIA", "WI": "WISCONSIN",
	"WY": "WYOMING",
}

// FullName returns the fully spelled US state/possession name for a two-letter code.
func FullName(code string) (string, bool) {
	full, ok := usStateFullNames[strings.ToUpper(strings.TrimSpace(code))]
	return full, ok
}

// IsUSStateOrPossession reports whether code is a US state/possession (not CA province/military).
func IsUSStateOrPossession(code string) bool {
	_, ok := usStateFullNames[strings.ToUpper(strings.TrimSpace(code))]
	return ok
}

// LeadingStateMatch peels a leading full US state/possession name from tokens.
// Returns the two-letter code, number of tokens consumed, and whether a match was found.
// Prefer longer matches (up to 4 tokens). Does not match two-letter codes alone when
// requireFullName is true.
func LeadingStateMatch(tokens []string, requireResidual bool) (code string, n int, ok bool) {
	if len(tokens) == 0 {
		return "", 0, false
	}
	maxN := 4
	if maxN > len(tokens) {
		maxN = len(tokens)
	}
	if requireResidual && maxN > len(tokens)-1 {
		maxN = len(tokens) - 1
	}
	for n := maxN; n >= 1; n-- {
		if n == 1 && len(tokens[0]) <= 2 {
			// Already abbreviated; only accept when not requiring a full name residual rewrite.
			if requireResidual {
				continue
			}
		}
		candidate := strings.Join(tokens[:n], " ")
		abbr, err := NormalizeRegion(candidate)
		if err != nil {
			continue
		}
		if !IsUSStateOrPossession(abbr) {
			continue
		}
		if full, ok := FullName(abbr); ok {
			// Prefer multi-word full names when n>=2
			if n >= 2 || len(strings.Fields(full)) == 1 || len(tokens[0]) > 2 {
				return abbr, n, true
			}
		}
	}
	return "", 0, false
}

// MultiWordStateNames returns full multi-token US state/possession names, longest first.
func MultiWordStateNames() []string {
	var out []string
	for code, full := range usStateFullNames {
		_ = code
		if len(strings.Fields(full)) >= 2 {
			out = append(out, full)
		}
	}
	// longest first for greedy prefix match
	slices.SortFunc(out, func(a, b string) int {
		if len(b) != len(a) {
			return len(b) - len(a)
		}
		return strings.Compare(a, b)
	})
	return out
}

// SingleWordStateNames returns a set of single-token full US state names.
func SingleWordStateNames() map[string]bool {
	out := make(map[string]bool)
	for _, full := range usStateFullNames {
		if len(strings.Fields(full)) == 1 {
			out[full] = true
		}
	}
	return out
}

// USStateAbbrevs returns two-letter US state/possession codes (excludes FM highway clash optional).
func USStateAbbrevs() map[string]bool {
	out := make(map[string]bool, len(usStateFullNames))
	for code := range usStateFullNames {
		out[code] = true
	}
	return out
}
