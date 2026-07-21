package highways_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
)

func TestNormalizeStreetName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Brief table
		{"COUNTY HWY 60E", "COUNTY HIGHWAY 60E"},
		{"CNTY HWY 20", "COUNTY HIGHWAY 20"},
		{"COUNTY RD 441", "COUNTY ROAD 441"},
		{"CR 1185", "COUNTY ROAD 1185"},
		{"FARM TO MARKET 1200", "FM 1200"},
		{"HWY FM 1320", "FM 1320"},
		{"HWY 64", "HIGHWAY 64"},
		{"I10", "INTERSTATE 10"},
		{"IH280", "INTERSTATE 280"},
		{"INTERSTATE HWY 680", "INTERSTATE 680"},
		{"RT 88", "ROUTE 88"},
		{"SR 220", "STATE ROAD 220"},
		{"US HWY 44", "US HIGHWAY 44"},
		{"KENTUCKY 440", "KY HIGHWAY 440"},
		{"KY 1207", "KY HIGHWAY 1207"},

		// Additional examples from highways.go comments
		{"COUNTY HIGHWAY 140", "COUNTY HIGHWAY 140"},
		{"CNTY RD 33", "COUNTY ROAD 33"},
		{"CA COUNTY RD 150", "CA COUNTY ROAD 150"},
		{"CALIFORNIA COUNTY ROAD 555", "CA COUNTY ROAD 555"},
		{"EXPRESSWAY 55", "EXPRESSWAY 55"},
		{"FM 187", "FM 187"},
		{"HWY 11 BYPASS", "HIGHWAY 11 BYP"},
		{"HWY 66 FRONTAGE ROAD", "HIGHWAY 66 FRONTAGE RD"},
		{"HIGHWAY 3 BYP ROAD", "HIGHWAY 3 BYPASS RD"},
		{"I 55 BYPASS", "INTERSTATE 55 BYP"},
		{"I 26 BYP ROAD", "INTERSTATE 26 BYPASS RD"},
		{"I 44 FRONTAGE ROAD", "INTERSTATE 44 FRONTAGE RD"},
		{"RD 5A", "ROAD 5A"},
		{"RTE 95", "ROUTE 95"},
		{"RANCH RD 620", "RANCH ROAD 620"},
		{"ST HIGHWAY 303", "STATE HIGHWAY 303"},
		{"STATE HWY 60", "STATE HIGHWAY 60"},
		{"ST RD 86", "STATE ROAD 86"},
		{"SR MM", "STATE ROUTE MM"},
		{"ST RT 175", "STATE ROUTE 175"},
		{"STATE RTE 260", "STATE ROUTE 260"},
		{"TOWNSHIP RD 20", "TOWNSHIP ROAD 20"},
		{"TSR 45", "TOWNSHIP ROAD 45"},
		{"US 41 SW", "US HIGHWAY 41 SW"},
		{"KENTUCKY HIGHWAY 189", "KY HIGHWAY 189"},
		{"KY HWY 75", "KY HIGHWAY 75"},
		{"KY ST HWY 1", "KY STATE HIGHWAY 1"},
		{"KENTUCKY STATE HIGHWAY 625", "KY STATE HIGHWAY 625"},

		// Whitespace / case normalization
		{"  county   hwy  60e  ", "COUNTY HIGHWAY 60E"},
		{"farm to market 1200", "FM 1200"},

		// State name/code as street name portion only — not bare highway routes
		// (letter-only residual after peeling state must not become "… HIGHWAY …")
		{"OKLAHOMA AVE", "OKLAHOMA AVE"},
		{"WASHINGTON BLVD", "WASHINGTON BLVD"},
		{"CA MAIN", "CA MAIN"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := highways.NormalizeStreetName(tc.in)
			if err != nil {
				t.Fatalf("NormalizeStreetName(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeStreetName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeStreetNameEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "\t\n"} {
		got, err := highways.NormalizeStreetName(in)
		if err == nil {
			t.Fatalf("NormalizeStreetName(%q) expected error, got %q", in, got)
		}
		if got != "" {
			t.Fatalf("NormalizeStreetName(%q) = %q, want empty on error", in, got)
		}
	}
}

func TestNormalizeStreetNamePassthrough(t *testing.T) {
	// Ordinary free-text that matches no highway rule is uppercased and returned.
	got, err := highways.NormalizeStreetName("Main Street")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "MAIN STREET" {
		t.Fatalf("got %q, want %q", got, "MAIN STREET")
	}
}
