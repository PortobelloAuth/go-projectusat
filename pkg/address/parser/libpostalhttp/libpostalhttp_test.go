package libpostalhttp_test

import (
	"testing"
	"time"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/libpostalhttp"
)

func TestParse(t *testing.T) {
	cases := []struct {
		In   string
		Want address.Address
	}{
		// Post-directional followed by a City with a directional prefix
		{
			In: "43 E 200 N, NORTH SALT LAKE, UT",
			Want: address.Address{
				PrimaryNumber:       "43",
				Predirectional:      "E",
				StreetName:          "200",
				StreetSuffix:        "",
				Postdirectional:     "N",
				SecondaryDesignator: "",
				SecondaryNumber:     "",
				City:                "NORTH SALT LAKE",
				Region:              "UT",
				Postal:              "",
				Country:             "",
			},
		},
		// 3253 W 9200 S, West Jordan, UT 84088
		{
			In: "3253 W 9200 S, West Jordan, UT 84088",
			Want: address.Address{
				PrimaryNumber:       "3253",
				Predirectional:      "W",
				StreetName:          "9200",
				StreetSuffix:        "",
				Postdirectional:     "S",
				SecondaryDesignator: "",
				SecondaryNumber:     "",
				City:                "WEST JORDAN",
				Region:              "UT",
				Postal:              "84088",
				Country:             "",
			},
		},
		// 3590 S Jordan Pkwy W, South Jordan, UT 84095
		// This address was collected from Google Maps. The street
		// name is actually "South Jordan Parkway" which runs east/west
		// (and is also known as 10600 S and 10400 S depending on what
		// segment of the road you are referencing) but some
		// pre/postdirectional logic seems to have moved the "West"
		// predirectional to a postdirectional and treated "South"
		// from the city name in the street name as a predirectional.
		// Oddly, "West" as a postdirectional feels reasonable to me.
		// I don't know if it is "right". I suspect that it is the
		// maps service handling the mix of directionals incorrectly
		//
		// 3590 W South Jordan Pkwy, South Jordan, Utah 84095
		// This is the expected construction and, according to a
		// Google AI summary referencing https://gis.utah.gov/_assets/911addressing.MvlRLWTr.pdf
		// the format "favored by modern mapping and emergency dispatch in Utah",
		// with the "West" postdirectional described as a legacy format.
		// The PDF's allowance of postdirectionals to distinguish road
		// segments doesn't actually seem to justify the Google Maps'
		// address construction or the AI summary's explaination.
		//
		// All of this is just to demonstrate that address parsing is challenging
		//
		// See https://gis.ny.gov/system/files/documents/2022/07/streetaddressparsing-cldxf.pdf
		// for various other difficult to parse street name constructions
	}

	p, err := libpostalhttp.NewService("http://localhost:4400", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Unable to create libpostal http service: %v", err)
	}
	for _, tc := range cases {
		got, err := p.Parse(tc.In)
		if err != nil {
			t.Fatalf("Error parsing '%s': %s", tc.In, err)
		}

		if *got != tc.Want {
			t.Errorf("Unexpected result parsing '%s': %s expected: %s", tc.In, *got, tc.Want)
			if got.PrimaryNumber != tc.Want.PrimaryNumber {
				t.Errorf("Primary Number did not match: %s expected: %s", got.PrimaryNumber, tc.Want.PrimaryNumber)
			}
			if got.Predirectional != tc.Want.Predirectional {
				t.Errorf("Predirectional did not match: %s expected: %s", got.Predirectional, tc.Want.Predirectional)
			}
			if got.StreetName != tc.Want.StreetName {
				t.Errorf("Street Name did not match: %s expected: %s", got.StreetName, tc.Want.StreetName)
			}
			if got.StreetSuffix != tc.Want.StreetSuffix {
				t.Errorf("Street Suffix did not match: %s expected: %s", got.StreetSuffix, tc.Want.StreetSuffix)
			}
			if got.Postdirectional != tc.Want.Postdirectional {
				t.Errorf("Postdirectional did not match: %s expected: %s", got.Postdirectional, tc.Want.Postdirectional)
			}
			if got.SecondaryDesignator != tc.Want.SecondaryDesignator {
				t.Errorf("Secondary Designator did not match: %s expected: %s", got.SecondaryDesignator, tc.Want.SecondaryDesignator)
			}
			if got.SecondaryNumber != tc.Want.SecondaryNumber {
				t.Errorf("Secondary Number did not match: %s expected: %s", got.SecondaryNumber, tc.Want.SecondaryNumber)
			}
			if got.City != tc.Want.City {
				t.Errorf("City did not match: %s expected: %s", got.City, tc.Want.City)
			}
			if got.Region != tc.Want.Region {
				t.Errorf("Region did not match: %s expected: %s", got.Region, tc.Want.Region)
			}
			if got.Postal != tc.Want.Postal {
				t.Errorf("Postal code did not match: %s expected: %s", got.Postal, tc.Want.Postal)
			}
			if got.Country != tc.Want.Country {
				t.Errorf("Country did not match: %s expected: %s", got.Country, tc.Want.Country)
			}
		}
	}
}
