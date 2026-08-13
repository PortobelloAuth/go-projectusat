package ruralroute_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
)

var cases = []struct {
	In   string
	Want string
}{
	{"RURAL ROUTE 91 BOX A7", "RR 91 BOX A7"},
	{"Rural Route 91 Box A7", "RR 91 BOX A7"},
	{"RFD 82 BOX 12", "RR 82 BOX 12"},
	{"RD 51 # 25", "RR 51 BOX 25"},
	{"RFD Route 4 #87a", "RR 4 BOX 87A"},
	{"RR 2 BOX 18 Bryan Dairy Rd", "RR 2 BOX 18"},
	{"RR03 BOX 98D", "RR 3 BOX 98D"},
	{"RR03 BOX 0098D", "RR 3 BOX 98D"},
	{"RURAL ROUTE 91 BOX #A7", "RR 91 BOX A7"},
	{"Rural Route 91 Box Num. A7", "RR 91 BOX A7"},
	{"RFD 82 BOX NUMBER 12", "RR 82 BOX 12"},
	{"RD #51 # 25", "RR 51 BOX 25"},
	{"RFD Route Num. 4 #87a", "RR 4 BOX 87A"},
	{"RR No. 2 BOX 18 Bryan Dairy Rd", "RR 2 BOX 18"},
}

func TestNormalize(t *testing.T) {
	for _, tc := range cases {
		out, err := ruralroute.Normalize(tc.In)
		if err != nil {
			t.Errorf("%s", err)
		}
		if out != tc.Want {
			t.Errorf("Unexpected normalized rural route text %s for %s. Expected: %s", out, tc.In, tc.Want)
		}
	}
}

func TestNotRuralRoute(t *testing.T) {
	notruralroute := "Main St"
	out, err := ruralroute.Normalize(notruralroute)
	if err == nil {
		t.Errorf("Expected error for non-rural route: %s got: %s", notruralroute, out)
	}
}

// Directionals are deliberately absent: the standard's rural route order is
// STREET PRIMARY SEC SECNUM, and additional designations do not apply.
func TestFormatStreetLineOmitsDirectionals(t *testing.T) {
	a := &address.Address{
		Predirectional:      "N",
		StreetName:          "RR 4",
		PrimaryNumber:       "BOX 125",
		Postdirectional:     "W",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "2",
	}

	if got := (&ruralroute.RuralRouteAddress{}).FormatStreetLine(a); got != "RR 4 BOX 125 APT 2" {
		t.Errorf("FormatStreetLine() = %q, want %q", got, "RR 4 BOX 125 APT 2")
	}
}
