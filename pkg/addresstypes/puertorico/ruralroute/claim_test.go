package ruralroute

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// claimed renders the single claim Claims makes over source, as
// "streetname|primarynumber", with the token span each part covers.
func claimed(t *testing.T, source string) (string, int, int) {
	t.Helper()

	claims := Claims(token.Tokenize(source))
	if len(claims) != 1 {
		t.Fatalf("Claims(%q) made %d claims, want 1", source, len(claims))
	}

	c := claims[0]
	if len(c.Parts) != 2 {
		t.Fatalf("Claims(%q) made a claim of %d parts, want 2", source, len(c.Parts))
	}

	return c.Parts[0].Value + "|" + c.Parts[1].Value, c.Start(), c.End()
}

// TestClaimsSplitsTheRouteFromTheBox checks the halves of the pattern and the
// tokens each covers, which a rendered street line cannot show. The split is
// what makes the route a street name and the box a primary address number.
func TestClaimsSplitsTheRouteFromTheBox(t *testing.T) {
	cases := []struct {
		source     string
		want       string
		start, end int
	}{
		{"RUTA RURAL 3 BUZON 12000", "RR 3|BOX 12000", 0, 5},
		{"Ruta Estrella 1 Buzón 18", "HC 1|BOX 18", 0, 5},
		{"RFD ROUTE 4 BZN 1725", "RR 4|BOX 1725", 0, 5},
		{"RFD 1 Bzn 17-A", "RR 1|BOX 17A", 0, 4},
		{"HC 03 Bzn 1050", "HC 3|BOX 1050", 0, 4},
	}

	for _, c := range cases {
		t.Run(c.source, func(t *testing.T) {
			got, start, end := claimed(t, c.source)
			if got != c.want {
				t.Errorf("Claims(%q) = %q, want %q", c.source, got, c.want)
			}
			if start != c.start || end != c.end {
				t.Errorf("Claims(%q) covers [%d,%d), want [%d,%d)", c.source, start, end, c.start, c.end)
			}
		})
	}
}

// TestClaimsLeavesTheSectorLineUnread is p. 30's rule that a sector name
// written with a route MUST be eliminated. The claim stops at the end of the
// route's own line, so the sector is stranded rather than absorbed — see
// Claims, and token.LineEnd for why a claim may not reach across the break.
func TestClaimsLeavesTheSectorLineUnread(t *testing.T) {
	source := "RUTA RURAL 3 BUZON 12000\nSECTOR EL BRINCO"

	got, start, end := claimed(t, source)
	if want := "RR 3|BOX 12000"; got != want {
		t.Errorf("Claims(%q) = %q, want %q", source, got, want)
	}
	if start != 0 || end != 5 {
		t.Errorf("Claims(%q) covers [%d,%d), want [0,5) — the route line alone", source, start, end)
	}
}

// TestClaimsFindsNoRouteIn keeps the vocabulary from claiming lines that
// belong to another reading.
func TestClaimsFindsNoRouteIn(t *testing.T) {
	for _, source := range []string{
		"1234 CALLE AURORA",
		"URB LOS OLMOS",
		"PO BOX 1190",
		"SECTOR EL BRINCO",
		"RUTA RURAL 3",
		// The route and the box are on separate lines, so neither half is a
		// route. A claim never spans a line break.
		"RUTA RURAL 3\nBUZON 12000",
	} {
		if claims := Claims(token.Tokenize(source)); len(claims) != 0 {
			t.Errorf("Claims(%q) made %d claims, want none", source, len(claims))
		}
	}
}
