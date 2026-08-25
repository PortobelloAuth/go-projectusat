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

	// Highway contract routes. Each row is the rural route case directly
	// above its family, rewritten with an HC designator, because the two
	// share every rule but the designator itself.
	{"HC 4 BOX 12", "HC 4 BOX 12"},
	{"HCR 4 BOX 12", "HC 4 BOX 12"},
	{"Highway Contract Route 4 Box 12", "HC 4 BOX 12"},
	{"HIGHWAY CONTRACT 4 BOX 12", "HC 4 BOX 12"},
	{"HC04 BOX 0012", "HC 4 BOX 12"},
	{"HC No. 4 # 12", "HC 4 BOX 12"},
	{"HC 4 BOX 12 Bryan Dairy Rd", "HC 4 BOX 12"},
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

// The spellings this package deliberately does not accept yet, and why. See the
// deviation note in ruralroute.go: each is a plausible highway contract route
// that no wording in this repository authorizes, so recognizing it would be a
// guess. A failure here means the vocabulary grew without the note being
// revisited.
func TestUnconfirmedHighwayContractSpellings(t *testing.T) {
	cases := []struct {
		in     string
		reason string
	}{
		{"STAR ROUTE 4 BOX 12", "likely, but the standard's list of spellings is not quoted here"},
		{"RUTA ESTRELLA 4 BOX 12", "Puerto Rico vocabulary, which belongs with the rest of that table"},
	}

	for _, tc := range cases {
		if out, err := ruralroute.Normalize(tc.in); err == nil {
			t.Errorf("Normalize(%q) = %q, want an error: %s", tc.in, out, tc.reason)
		}
	}
}

// A route number and a box number are both required, for HC exactly as for RR.
// A bare designator says nothing a parser can use.
func TestAHighwayContractDesignatorAloneIsNotARoute(t *testing.T) {
	for _, in := range []string{"HC", "HC 4", "HC BOX 12", "Highway Contract Route"} {
		if out, err := ruralroute.Normalize(in); err == nil {
			t.Errorf("Normalize(%q) = %q, want an error", in, out)
		}
	}
}

// HC is two letters that occur inside ordinary words. The pattern is anchored
// and needs a route and a box after the designator, which is what keeps a
// street name from being mistaken for a route.
func TestOrdinaryStreetNamesContainingTheDesignatorAreNotRoutes(t *testing.T) {
	for _, in := range []string{"Beechcroft Rd", "123 Highway View Dr", "HCR"} {
		if out, err := ruralroute.Normalize(in); err == nil {
			t.Errorf("Normalize(%q) = %q, want an error", in, out)
		}
	}
}

// A highway contract route formats through the same street line as a rural
// route, which is the reason it did not need an address type of its own.
func TestFormatStreetLineFormatsAHighwayContractRoute(t *testing.T) {
	a := &address.Address{
		StreetName:          "HC 4",
		PrimaryNumber:       "BOX 125",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "2",
	}

	if got := (&ruralroute.RuralRouteAddress{}).FormatStreetLine(a); got != "HC 4 BOX 125 APT 2" {
		t.Errorf("FormatStreetLine() = %q, want %q", got, "HC 4 BOX 125 APT 2")
	}
}

// The standard's CMRA section shows "PMB 234" over "RR 1 BOX 12", so a rural
// route carries a private mailbox number. Detail qualifies the secondary
// number and follows it, as it does in the ordinary street line.
func TestFormatStreetLineRendersTheDetail(t *testing.T) {
	a := &address.Address{
		StreetName:    "RR 1",
		PrimaryNumber: "BOX 12",
		Detail:        "PMB 234",
	}

	want := "RR 1 BOX 12 PMB 234"
	if got := (&ruralroute.RuralRouteAddress{}).FormatStreetLine(a); got != want {
		t.Errorf("FormatStreetLine() = %q, want %q", got, want)
	}

	withUnit := &address.Address{
		StreetName:          "RR 1",
		PrimaryNumber:       "BOX 12",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "2",
		Detail:              "PMB 234",
	}

	want = "RR 1 BOX 12 APT 2 PMB 234"
	if got := (&ruralroute.RuralRouteAddress{}).FormatStreetLine(withUnit); got != want {
		t.Errorf("FormatStreetLine() = %q, want %q", got, want)
	}
}
