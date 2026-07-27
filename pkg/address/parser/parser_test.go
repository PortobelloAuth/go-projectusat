package parser_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
)

// Distinguish E St from East St
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
				City:                "West Jordan",
				Region:              "UT",
				Postal:              "84088",
				Country:             "",
			},
		},
	}

	p := parser.New()
	for _, tc := range cases {
		got, err := p.Parse(tc.In)
		if err != nil {
			t.Fatalf("Error parsing '%s': %s", tc.In, err)
		}

		if *got != tc.Want {
			t.Errorf("Unexpected result parsing '%s': %s expected: %s", tc.In, *got, tc.Want)
		}
	}
}
