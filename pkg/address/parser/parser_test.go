package parser_test

import (
	"slices"
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

func TestTokenize(t *testing.T) {
	addr := "43 E 200 N, NORTH SALT LAKE, UT"
	want := []parser.Token{
		{"43", 0, 1, 0, -1},
		{"E", 0, 1, 1, -1},
		{"200", 0, 1, 2, -1},
		{"N", 0, 1, 3, -1},
		{"NORTH", 0, 1, 4, 0},
		{"SALT", 0, 1, 5, -1},
		{"LAKE", 0, 1, 6, -1},
		{"UT", 0, 1, 7, 1},
	}

	out := parser.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	// no commas
	addr = "43 E 200 N NORTH SALT LAKE UT"
	want = []parser.Token{
		{"43", 0, 1, 0, -1},
		{"E", 0, 1, 1, -1},
		{"200", 0, 1, 2, -1},
		{"N", 0, 1, 3, -1},
		{"NORTH", 0, 1, 4, -1},
		{"SALT", 0, 1, 5, -1},
		{"LAKE", 0, 1, 6, -1},
		{"UT", 0, 1, 7, -1},
	}

	out = parser.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	addr = "3253 W 9200 S, West Jordan, UT  84088"
	want = []parser.Token{
		{"3253", 0, 1, 0, -1},
		{"W", 0, 1, 1, -1},
		{"9200", 0, 1, 2, -1},
		{"S", 0, 1, 3, -1},
		{"West", 0, 1, 4, 0},
		{"Jordan", 0, 1, 5, -1},
		{"UT", 0, 1, 6, 1},
		{"84088", 0, 1, 7, -1},
	}

	out = parser.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	addr = "3253 W 9200 S\nWest Jordan, UT  84088\nUSA"
	want = []parser.Token{
		{"3253", 0, 3, 0, -1},
		{"W", 0, 3, 1, -1},
		{"9200", 0, 3, 2, -1},
		{"S", 0, 3, 3, -1},
		{"West", 1, 3, 0, -1},
		{"Jordan", 1, 3, 1, -1},
		{"UT", 1, 3, 2, 0},
		{"84088", 1, 3, 3, -1},
		{"USA", 2, 3, 0, -1},
	}

	out = parser.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}
}
