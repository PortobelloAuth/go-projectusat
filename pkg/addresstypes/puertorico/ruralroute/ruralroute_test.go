package ruralroute

import "testing"

// TestNormalizeStandardizesTheSpecExamples runs every Incorrect/Correct pair
// the standard prints for Puerto Rico routes, pp. 30-31.
func TestNormalizeStandardizesTheSpecExamples(t *testing.T) {
	cases := []struct {
		page int
		in   string
		want string
	}{
		{30, "RR03 BOX 9800", "RR 3 BOX 9800"},
		{30, "RFD ROUTE 4 BZN 1725", "RR 4 BOX 1725"},
		{30, "RUTA RURAL 3 BUZON 12000", "RR 3 BOX 12000"},
		{30, "RFD 1 Bzn 17-A", "RR 1 BOX 17A"},
		{31, "Ruta Estrella 1 Buzón 18", "HC 1 BOX 18"},
		// p. 31 prints "HC 1 BOX 1050" as the Correct Form for this row, which
		// cannot be standardized from HC 03. The leading-zero rule on the same
		// page gives HC 3, and only one of the two can be implemented. See
		// #119, and Aaron's ruling to pin the implementable side.
		{31, "HC 03 Bzn 1050", "HC 3 BOX 1050"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := Normalize(c.in)
			if err != nil {
				t.Fatalf("p. %d: Normalize(%q) errored: %v", c.page, c.in, err)
			}
			if got != c.want {
				t.Errorf("p. %d: Normalize(%q) = %q, want %q", c.page, c.in, got, c.want)
			}
		})
	}
}

// TestNormalizeDropsWhatFollowsThePattern covers p. 30's sector examples where
// the sector was written on the route's own line rather than below it. The
// standard requires the information to be eliminated either way.
func TestNormalizeDropsWhatFollowsThePattern(t *testing.T) {
	got, err := Normalize("RR 2 BOX 1980 SECTOR EL BRINCO")
	if err != nil {
		t.Fatalf("Normalize errored: %v", err)
	}
	if want := "RR 2 BOX 1980"; got != want {
		t.Errorf("Normalize = %q, want %q", got, want)
	}
}

// TestNormalizeRefusesWhatIsNotARoute keeps the recognizer honest: every one
// of these is a real address line that the route vocabulary must not take.
func TestNormalizeRefusesWhatIsNotARoute(t *testing.T) {
	for _, in := range []string{
		"BUZON 12000",       // a box with no route
		"RUTA RURAL 3",      // a route with no box
		"1234 CALLE AURORA", // a Puerto Rico street line
		"URB LOS OLMOS",     // an urbanization line
		"PO BOX 1190",       // a post office box
		"SECTOR EL BRINCO",  // the sector name on its own
		"123 RD BOX 4",      // RD as a street suffix, not a designator
	} {
		if got, err := Normalize(in); err == nil {
			t.Errorf("Normalize(%q) = %q, want an error", in, got)
		}
	}
}
