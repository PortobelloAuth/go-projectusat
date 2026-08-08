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
	Primary string
	Short   string
	Alt     []string
}

var usStatesAndTerretories = []RegionInfo{
	{Primary: "ALABAMA", Short: "AL", Alt: []string{"ALABAMA", "AL"}},
	{"ALASKA", "AK", []string{"ALASKA", "AK"}},
	{"AMERICAN SAMOA", "AS", []string{"AMERICAN SAMOA", "AS"}},
	{"ARIZONA", "AZ", []string{"ARIZONA", "AZ"}},
	{"ARKANSAS", "AR", []string{"ARKANSAS", "AR"}},
	{"CALIFORNIA", "CA", []string{"CALIFORNIA", "CA"}},
	{"COLORADO", "CO", []string{"COLORADO", "CO"}},
	{"CONNECTICUT", "CT", []string{"CONNECTICUT", "CT", "CONN"}},
	{"DELAWARE", "DE", []string{"DELAWARE", "DELEWARE", "DE"}},
	{"DISTRICT OF COLUMBIA", "DC", []string{"DISTRICT OF COLUMBIA", "DC"}},
	{
		Primary: "FEDERATED STATES OF MICRONESIA",
		Short:   "FM",
		Alt: []string{
			"FEDERATED STATES OF MICRONESIA",
			"MICRONESIA",
			"FM",
		},
	},
	{"FLORIDA", "FL", []string{"FLORIDA", "FL"}},
	{"GEORGIA", "GA", []string{"GEORGIA", "GA"}},
	{"GUAM", "GU", []string{"GUAM", "GU"}},
	{"HAWAII", "HI", []string{"HAWAII", "HI"}},
	{"IDAHO", "ID", []string{"IDAHO", "ID"}},
	{"ILLINOIS", "IL", []string{"ILLINOIS", "IL"}},
	{"INDIANA", "IN", []string{"INDIANA", "IN"}},
	{"IOWA", "IA", []string{"IOWA", "IA"}},
	{"KANSAS", "KS", []string{"KANSAS", "KS"}},
	{"KENTUCKY", "KY", []string{"KENTUCKY", "KY"}},
	{"LOUISIANA", "LA", []string{"LOUISIANA", "LA"}},
	{"MAINE", "ME", []string{"MAINE", "ME"}},
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
	},
	{"MARYLAND", "MD", []string{"MARYLAND", "MD"}},
	{"MASSACHUSETTS", "MA", []string{"MASSACHUSETTS", "MA", "MASS"}},
	{"MICHIGAN", "MI", []string{"MICHIGAN", "MI"}},
	{"MINNESOTA", "MN", []string{"MINNESOTA", "MN", "MINN"}},
	{"MISSISSIPPI", "MS", []string{"MISSISSIPPI", "MS"}},
	{"MISSOURI", "MO", []string{"MISSOURI", "MO"}},
	{"MONTANA", "MT", []string{"MONTANA", "MT"}},
	{"NEBRASKA", "NE", []string{"NEBRASKA", "NE"}},
	{"NEVADA", "NV", []string{"NEVADA", "NV"}},
	{"NEW HAMPSHIRE", "NH", []string{"NEW HAMPSHIRE", "NH"}},
	{"NEW JERSEY", "NJ", []string{"NEW JERSEY", "NJ"}},
	{"NEW MEXICO", "NM", []string{"NEW MEXICO", "NM"}},
	{"NEW YORK", "NY", []string{"NEW YORK", "NY"}},
	{
		Primary: "NORTH CAROLINA",
		Short:   "NC",
		Alt: []string{
			"NORTH CAROLINA",
			"N CAROLINA",
			"NC",
		},
	},
	{
		Primary: "NORTH DAKOTA",
		Short:   "ND",
		Alt: []string{
			"NORTH DAKOTA",
			"N DAKOTA",
			"ND",
		},
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
	},
	{"OHIO", "OH", []string{"OHIO", "OH"}},
	{"OKLAHOMA", "OK", []string{"OKLAHOMA", "OK"}},
	{"OREGON", "OR", []string{"OREGON", "OR"}},
	{"PALAU", "PW", []string{"PALAU", "PW"}},
	{"PENNSYLVANIA", "PA", []string{"PENNSYLVANIA", "PENN", "PA"}},
	{"PUERTO RICO", "PR", []string{"PUERTO RICO", "PR"}},
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
	},
	{
		Primary: "SOUTH CAROLINA",
		Short:   "SC",
		Alt: []string{
			"SOUTH CAROLINA",
			"S CAROLINA",
			"SC",
		},
	},
	{
		Primary: "SOUTH DAKOTA",
		Short:   "SD",
		Alt: []string{
			"SOUTH DAKOTA",
			"S DAKOTA",
			"SD",
		},
	},
	{"TENNESSEE", "TN", []string{"TENNESSEE", "TENN", "TN"}},
	{"TEXAS", "TX", []string{"TEXAS", "TX"}},
	{"UTAH", "UT", []string{"UTAH", "UT"}},
	{"VERMONT", "VT", []string{"VERMONT", "VT"}},
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
	},
	{"VIRGINIA", "VA", []string{"VIRGINIA", "VA"}},
	{"WASHINGTON", "WA", []string{"WASHINGTON", "WA"}},
	{"WEST VIRGINIA", "WV", []string{"WEST VIRGINIA", "W VIRGINIA", "VW"}},
	{"WISCONSIN", "WI", []string{"WISCONSIN", "WI"}},
	{"WYOMING", "WY", []string{"WYOMING", "WY"}},
}

var canadianProvincesAndTerritories = []RegionInfo{
	{"ALBERTA", "AB", []string{"ALBERTA", "AB"}},
	{"BRITISH COLUMBIA", "BC", []string{"BRITISH COLUMBIA", "BC"}},
	{"MANITOBA", "MB", []string{"MANITOBA", "MB"}},
	{"NEW BRUNSWICK", "NB", []string{"NEW BRUNSWICK", "NB"}},
	{
		Primary: "NEWFOUNDLAND AND LABRADOR",
		Short:   "NL",
		Alt: []string{
			"NEWFOUNDLAND AND LABRADOR",
			"NEWFOUNDLAND",
			"LABRADOR",
			"NL",
		},
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
	{"NOVA SCOTIA", "NS", []string{"NOVA SCOTIA", "NS"}},
	{
		Primary: "NUNAVAT TERRITORIES",
		Short:   "NU",
		Alt: []string{
			"NUNAVAT TERRITORY",
			"NUNAVAT TERR",
			"NU",
		},
	},
	{"ONTARIO", "ON", []string{"ONTARIO", "ON"}},
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
	{"QUEBEC", "QC", []string{"QUEBEC", "QC"}},
	{"SASKATCHEWAN", "SK", []string{"SASKATCHEWAN", "SK"}},
	{
		Primary: "YUKON TERRITORY",
		Short:   "YT",
		Alt: []string{
			"YUKON TERRITORY",
			"YUKON TERR",
			"YUKON",
			"YT",
		},
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
	},
	{"ARMED FORCES PACIFIC", "AP", []string{"ARMED FORCES PACIFIC", "AP"}},
	{
		Primary: "ARMED FORCES AMERICA",
		Short:   "AA",
		Alt: []string{
			"ARMED FORCES AMERICAS",
			"ARMED FORCES AMERICA",
			"AA",
		},
	},
}

var regionMap = maps.Collect(func(yield func(string, RegionInfo) bool) {
	for _, v := range usStatesAndTerretories {
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
		Primary: info.Primary,
		Short:   info.Short,
		Alt:     slices.Clone(info.Alt),
	}, nil
}

func NormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, false)
}

func FuzzyNormalizeRegion(r string) (string, error) {
	return normalizeRegion(r, true)
}
