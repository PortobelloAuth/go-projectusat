package region

import (
	"fmt"
	"strings"
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
	"DELEWARE":                       "DE",
	"DE":                             "DE",
	"DISTRICT OF COLUMBIA":           "DE",
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

func NormalizeRegion(r string) (string, error) {
	// TODO: clean out any punctuation
	capitalized := strings.ToUpper(r)
	abrev, ok := regionMap[capitalized]
	if !ok {
		return "", fmt.Errorf("Unrecognized state, possession, Canadian provice, or US Armed Forces region")
	}

	return abrev, nil
}
