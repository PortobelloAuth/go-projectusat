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

type RegionInfo struct {
	Primary            string
	Short              string
	Alt                []string
	PossibleStreetName bool
}

var usStatesAndPossessions = []RegionInfo{
	{Primary: "ALABAMA", Short: "AL", Alt: []string{"ALABAMA", "AL"}, PossibleStreetName: true},
	{"ALASKA", "AK", []string{"ALASKA", "AK"}, true},
	{"AMERICAN SAMOA", "AS", []string{"AMERICAN SAMOA", "AS"}, false},
	{"ARIZONA", "AZ", []string{"ARIZONA", "AZ"}, true},
	{"ARKANSAS", "AR", []string{"ARKANSAS", "AR"}, true},
	{"CALIFORNIA", "CA", []string{"CALIFORNIA", "CA"}, true},
	{"COLORADO", "CO", []string{"COLORADO", "CO"}, true},
	{"CONNECTICUT", "CT", []string{"CONNECTICUT", "CT", "CONN"}, true},
	{"DELAWARE", "DE", []string{"DELAWARE", "DELEWARE", "DE"}, true},
	{"DISTRICT OF COLUMBIA", "DC", []string{"DISTRICT OF COLUMBIA", "DC"}, true},
	{
		Primary: "FEDERATED STATES OF MICRONESIA",
		Short:   "FM",
		Alt: []string{
			"FEDERATED STATES OF MICRONESIA",
			"MICRONESIA",
			"FM",
		},
		PossibleStreetName: false,
	},
	{"FLORIDA", "FL", []string{"FLORIDA", "FL"}, true},
	{"GEORGIA", "GA", []string{"GEORGIA", "GA"}, true},
	{"GUAM", "GU", []string{"GUAM", "GU"}, true},
	{"HAWAII", "HI", []string{"HAWAII", "HI"}, true},
	{"IDAHO", "ID", []string{"IDAHO", "ID"}, true},
	{"ILLINOIS", "IL", []string{"ILLINOIS", "IL"}, true},
	{"INDIANA", "IN", []string{"INDIANA", "IN"}, true},
	{"IOWA", "IA", []string{"IOWA", "IA"}, true},
	{"KANSAS", "KS", []string{"KANSAS", "KS"}, true},
	{"KENTUCKY", "KY", []string{"KENTUCKY", "KY"}, true},
	{"LOUISIANA", "LA", []string{"LOUISIANA", "LA"}, true},
	{"MAINE", "ME", []string{"MAINE", "ME"}, true},
	{
		Primary: "MARSHALL ISLANDS",
		Short:   "MH",
		Alt: []string{
			"MARSHALL ISLANDS",
			"MARSHALL IS",
			"MARSHALL ISL",
			"MARSHALL ISLS",
			"MARSHALL ISS",
			"MARSHALL ISLD",
			"MH",
		},
		PossibleStreetName: true,
	},
	{"MARYLAND", "MD", []string{"MARYLAND", "MD"}, true},
	{"MASSACHUSETTS", "MA", []string{"MASSACHUSETTS", "MA", "MASS"}, true},
	{"MICHIGAN", "MI", []string{"MICHIGAN", "MI"}, true},
	{"MINNESOTA", "MN", []string{"MINNESOTA", "MN", "MINN"}, true},
	{"MISSISSIPPI", "MS", []string{"MISSISSIPPI", "MS"}, true},
	{"MISSOURI", "MO", []string{"MISSOURI", "MO"}, true},
	{"MONTANA", "MT", []string{"MONTANA", "MT"}, true},
	{"NEBRASKA", "NE", []string{"NEBRASKA", "NE"}, true},
	{"NEVADA", "NV", []string{"NEVADA", "NV"}, true},
	{"NEW HAMPSHIRE", "NH", []string{"NEW HAMPSHIRE", "NH"}, true},
	{"NEW JERSEY", "NJ", []string{"NEW JERSEY", "NJ"}, true},
	{"NEW MEXICO", "NM", []string{"NEW MEXICO", "NM"}, true},
	{"NEW YORK", "NY", []string{"NEW YORK", "NY"}, true},
	{
		Primary: "NORTH CAROLINA",
		Short:   "NC",
		Alt: []string{
			"NORTH CAROLINA",
			"N CAROLINA",
			"NC",
		},
		PossibleStreetName: true,
	},
	{
		Primary: "NORTH DAKOTA",
		Short:   "ND",
		Alt: []string{
			"NORTH DAKOTA",
			"N DAKOTA",
			"ND",
		},
		PossibleStreetName: true,
	},
	{
		Primary: "NORTHERN MARIANA ISLANDS",
		Short:   "MP",
		Alt: []string{
			"NORTHERN MARIANA ISLANDS",
			"NORTHERN MARIANA IS",
			"NORTHERN MARIANA ISL",
			"NORTHERN MARIANA ISLS",
			"NORTHERN MARIANA ISS",
			"NORTHERN MARIANA ISLD",
			"N MARIANA ISLANDS",
			"N MARIANA IS",
			"N MARIANA ISL",
			"N MARIANA ISLS",
			"N MARIANA ISS",
			"N MARIANA ISLD",
			"MP",
		},
		PossibleStreetName: true,
	},
	{"OHIO", "OH", []string{"OHIO", "OH"}, true},
	{"OKLAHOMA", "OK", []string{"OKLAHOMA", "OK"}, true},
	{"OREGON", "OR", []string{"OREGON", "OR"}, true},
	{"PALAU", "PW", []string{"PALAU", "PW"}, true},
	{"PENNSYLVANIA", "PA", []string{"PENNSYLVANIA", "PENN", "PA"}, true},
	{"PUERTO RICO", "PR", []string{"PUERTO RICO", "PR"}, true},
	{
		Primary: "RHODE ISLAND",
		Short:   "RI",
		Alt: []string{
			"RHODE ISLAND",
			"RHODE IS",
			"RHODE ISL",
			"RHODE ISLD",
			"RI",
		},
		PossibleStreetName: true,
	},
	{
		Primary: "SOUTH CAROLINA",
		Short:   "SC",
		Alt: []string{
			"SOUTH CAROLINA",
			"S CAROLINA",
			"SC",
		},
		PossibleStreetName: true,
	},
	{
		Primary: "SOUTH DAKOTA",
		Short:   "SD",
		Alt: []string{
			"SOUTH DAKOTA",
			"S DAKOTA",
			"SD",
		},
		PossibleStreetName: true,
	},
	{"TENNESSEE", "TN", []string{"TENNESSEE", "TENN", "TN"}, true},
	{"TEXAS", "TX", []string{"TEXAS", "TX"}, true},
	{"UTAH", "UT", []string{"UTAH", "UT"}, true},
	{"VERMONT", "VT", []string{"VERMONT", "VT"}, true},
	{
		Primary: "VIRGIN ISLANDS",
		Short:   "VI",
		Alt: []string{
			"VIRGIN ISLANDS",
			"VIRGIN IS",
			"VIRGIN ISL",
			"VIRGIN ISLS",
			"VIRGIN ISS",
			"VIRGIN ISLD",
			"US VIRGIN ISLANDS",
			"US VIRGIN IS",
			"US VIRGIN ISL",
			"US VIRGIN ISLS",
			"US VIRGIN ISS",
			"US VIRGIN ISLD",
			"USVI",
			"VIS",
			"USA VI",
			"VI USA",
			"VI",
		},
		PossibleStreetName: true,
	},
	{"VIRGINIA", "VA", []string{"VIRGINIA", "VA"}, true},
	{"WASHINGTON", "WA", []string{"WASHINGTON", "WA"}, true},
	{"WEST VIRGINIA", "WV", []string{"WEST VIRGINIA", "W VIRGINIA", "WV"}, true},
	{"WISCONSIN", "WI", []string{"WISCONSIN", "WI"}, true},
	{"WYOMING", "WY", []string{"WYOMING", "WY"}, true},
}

var canadianProvincesAndTerritories = []RegionInfo{
	{"ALBERTA", "AB", []string{"ALBERTA", "AB"}, true},
	{"BRITISH COLUMBIA", "BC", []string{"BRITISH COLUMBIA", "BC"}, true},
	{"MANITOBA", "MB", []string{"MANITOBA", "MB"}, true},
	{"NEW BRUNSWICK", "NB", []string{"NEW BRUNSWICK", "NB"}, true},
	{
		Primary: "NEWFOUNDLAND AND LABRADOR",
		Short:   "NL",
		Alt: []string{
			"NEWFOUNDLAND AND LABRADOR",
			"NEWFOUNDLAND",
			"LABRADOR",
			"NL",
		},
		PossibleStreetName: false,
	},
	{
		Primary: "NORTHWEST TERRITORIES",
		Short:   "NT",
		Alt: []string{
			"NORTHWEST TERRITORIES",
			"NORTHWEST TERR",
			"NW TERRITORIES",
			"NW TERR",
			"NT",
		},
	},
	{"NOVA SCOTIA", "NS", []string{"NOVA SCOTIA", "NS"}, true},
	{
		Primary: "NUNAVAT TERRITORIES",
		Short:   "NU",
		Alt: []string{
			"NUNAVAT TERRITORY",
			"NUNAVAT TERR",
			"NU",
		},
	},
	{"ONTARIO", "ON", []string{"ONTARIO", "ON"}, true},
	{
		Primary: "PRINCE EDWARD ISLAND",
		Short:   "PE",
		Alt: []string{
			"PRINCE EDWARD ISLAND",
			"PRINCE EDWARD IS",
			"PRINCE EDWARD ISL",
			"PRINCE EDWARD ISLD",
			"PE",
		},
	},
	{"QUEBEC", "QC", []string{"QUEBEC", "QC"}, true},
	{"SASKATCHEWAN", "SK", []string{"SASKATCHEWAN", "SK"}, true},
	{
		Primary: "YUKON TERRITORY",
		Short:   "YT",
		Alt: []string{
			"YUKON TERRITORY",
			"YUKON TERR",
			"YUKON",
			"YT",
		},
		PossibleStreetName: true,
	},
}

var usMillitaryRegions = []RegionInfo{
	{
		Primary: "ARMED FORCES EUROPE THE MIDDLE EAST AND CANADA",
		Short:   "AE",
		Alt: []string{
			"ARMED FORCES EUROPE THE MIDDLE EAST AND CANADA",
			"ARMED FORCES EUROPE",
			"AE",
		},
		PossibleStreetName: false,
	},
	{"ARMED FORCES PACIFIC", "AP", []string{"ARMED FORCES PACIFIC", "AP"}, false},
	{
		Primary: "ARMED FORCES AMERICA",
		Short:   "AA",
		Alt: []string{
			"ARMED FORCES AMERICAS",
			"ARMED FORCES AMERICA",
			"AA",
		},
		PossibleStreetName: false,
	},
}

var regionMap = maps.Collect(func(yield func(string, RegionInfo) bool) {
	for _, v := range usStatesAndPossessions {
		for _, a := range v.Alt {
			if !yield(a, v) {
				return
			}
		}
	}

	for _, v := range canadianProvincesAndTerritories {
		for _, a := range v.Alt {
			if !yield(a, v) {
				return
			}
		}
	}

	for _, v := range usMillitaryRegions {
		for _, a := range v.Alt {
			if !yield(a, v) {
				return
			}
		}
	}
})

var regionKeys = slices.Collect(maps.Keys(regionMap))
var alphaspace = regexp.MustCompile("[^a-zA-Z ]+")

func normalizeRegion(r string, fuzzy bool) (string, error) {
	info, err := Info(r, fuzzy)
	if err != nil {
		return "", err
	}

	return info.Short, nil
}

func Info(r string, fuzzy bool) (*RegionInfo, error) {
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
	info, ok := regionMap[rkey]
	if !ok {
		return nil, fmt.Errorf("Unrecognized state, possession, Canadian provice, or US Armed Forces region")
	}
	fmt.Printf("r: %s info: %v\n", r, info)

	return &RegionInfo{
		Primary:            info.Primary,
		Short:              info.Short,
		Alt:                slices.Clone(info.Alt),
		PossibleStreetName: info.PossibleStreetName,
	}, nil
}

func NormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, false)
}

func FuzzyNormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, true)
}
