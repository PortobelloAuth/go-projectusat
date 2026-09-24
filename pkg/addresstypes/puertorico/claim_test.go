package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
)

// area returns the value the best reading assigns to the Area, and whether
// there is a reading at all.
func area(source string) (string, bool) {
	claims := puertorico.Claims(token.Tokenize(source))
	if len(claims) == 0 {
		return "", false
	}
	if len(claims) > 1 {
		return "", false
	}

	return claims[0].Parts[0].Value, true
}

// The standard's own Puerto Rico example. The urbanization occupies the first
// line of the street address block, above the secondary and primary lines.
func TestTheStandardsUrbanizationExample(t *testing.T) {
	got, ok := area("URB HIGHLAND GDNS\nCOND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926")
	if !ok {
		t.Fatal("the standard's own example is not claimed")
	}
	if got != "URB HIGHLAND GDNS" {
		t.Errorf("area = %q, want %q", got, "URB HIGHLAND GDNS")
	}
}

func TestEverySpellingOfTheDesignatorClaimsTheSameArea(t *testing.T) {
	for _, designator := range []string{"URB", "Urb", "URBANIZACION", "Urbanización", "URBANIZATION"} {
		t.Run(designator, func(t *testing.T) {
			got, ok := area(designator + " LAS GLADIOLAS\n150 CALLE A\nSAN JUAN PR 00926")
			if !ok {
				t.Fatalf("%q does not open an urbanization", designator)
			}
			if got != "URB LAS GLADIOLAS" {
				t.Errorf("area = %q, want %q", got, "URB LAS GLADIOLAS")
			}
		})
	}
}

// The name is free text of whatever length its developer chose, and the line
// is the only thing that says where it ends.
//
// The name deliberately does not open with any of the pp.28-29 standalone
// urbanization names (see TestStandaloneUrbanizationExceptions below) — a
// name that did would be a different rule firing, not this one.
func TestTheWholeLineAfterTheDesignatorIsTheName(t *testing.T) {
	got, ok := area("URB LAS GLADIOLAS DE COUNTRY CLUB\n123 CALLE MAIN\nSAN JUAN PR 00926")
	if !ok {
		t.Fatal("a multi word development name is not claimed")
	}
	if got != "URB LAS GLADIOLAS DE COUNTRY CLUB" {
		t.Errorf("area = %q, want the whole line", got)
	}
}

// pp.28-29's Exceptions table: these 38 names stand alone and MUST NOT take a
// URB in front, whether the input carried one or not. The standard's own two
// worked examples, plus the Full-spelling and plural forms the table also
// carries.
func TestStandaloneUrbanizationExceptions(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "standard's example, no URB in the input",
			input: "EXT VISTA BELLA",
			want:  "EXT VISTA BELLA",
		},
		{
			name:  "standard's example, URB stripped",
			input: "URB EXT VISTA BELLA",
			want:  "EXT VISTA BELLA",
		},
		{
			name:  "second standard example, already abbreviated",
			input: "URB ALTS DE CANA",
			want:  "ALTS DE CANA",
		},
		{
			name:  "Full spelling abbreviates the same as Short",
			input: "URB EXTENSION VISTA BELLA",
			want:  "EXT VISTA BELLA",
		},
		{
			name:  "a plural (S) row that is unabbreviated in the table",
			input: "URB VISTAS DEL MAR",
			want:  "VISTAS DEL MAR",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := area(tc.input + "\n123 CALLE MAIN\nSAN JUAN PR 00926")
			if !ok {
				t.Fatalf("%q is not claimed as an urbanization", tc.input)
			}
			if got != tc.want {
				t.Errorf("area(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// #132 case 1: a primary address number may precede the designator on the
// same line, when the number itself opens the line — p.28's own example, a
// standalone urbanization name acting as the street name with the primary
// number that always leads a Puerto Rico street line still in front of it.
//
// The claim covers the designator through the name, same as every other
// urbanization claim; the primary number is a separate token this claim
// leaves unclaimed; assembling it with the urbanization into a full street
// line reading is a candidate.go concern this fix does not reach.
func TestAPrimaryNumberMayPrecedeTheDesignator(t *testing.T) {
	got, ok := area("A17 URB JARDINES FAGOTA\nPONCE PR 00731")
	if !ok {
		t.Fatal("the standard's A17 URB JARDINES FAGOTA example is not claimed")
	}
	if got != "JARD FAGOTA" {
		t.Errorf("area = %q, want %q", got, "JARD FAGOTA")
	}
}

// The exemption is for a primary number only, and only when it is the one
// thing ahead of the designator. Arbitrary text is still refused, whether or
// not a number happens to be the nearest token to the designator.
func TestOnlyABarePrimaryNumberMayPrecedeTheDesignator(t *testing.T) {
	for _, source := range []string{
		"ACME CORP URB HIGHLAND GDNS\n123 CALLE MAIN\nSAN JUAN PR 00926",
		"ACME A17 URB JARDINES FAGOTA\nPONCE PR 00731",
		"A17 B18 URB JARDINES FAGOTA\nPONCE PR 00731",
	} {
		t.Run(source, func(t *testing.T) {
			if claims := puertorico.Claims(token.Tokenize(source)); len(claims) != 0 {
				t.Errorf("got %d claims, want none: only a bare primary number may precede the designator", len(claims))
			}
		})
	}
}

// A claim never crosses a line break, so the street line that follows stays
// available to whatever reads it.
func TestTheClaimStopsAtTheLineBreak(t *testing.T) {
	tokens := token.Tokenize("URB HIGHLAND GDNS\n123 CALLE MAIN\nSAN JUAN PR 00926")
	claims := puertorico.Claims(tokens)
	if len(claims) != 1 {
		t.Fatalf("got %d claims, want 1", len(claims))
	}
	if claims[0].Start() != 0 || claims[0].End() != 3 {
		t.Errorf("extent = [%d,%d), want [0,3): the claim runs past its line",
			claims[0].Start(), claims[0].End())
	}
}

func TestNoUrbanizationIsNoClaims(t *testing.T) {
	for _, source := range []string{
		"123 MAIN ST\nDENVER CO 80202",
		"COND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926",
	} {
		t.Run(source, func(t *testing.T) {
			if claims := puertorico.Claims(token.Tokenize(source)); len(claims) != 0 {
				t.Errorf("got %d claims, want none", len(claims))
			}
		})
	}
}

// A designator with no name after it is a fragment of a pattern that did not
// match, not weak evidence of one.
func TestADesignatorAloneIsNotAnUrbanization(t *testing.T) {
	if claims := puertorico.Claims(token.Tokenize("URB\n123 CALLE MAIN\nSAN JUAN PR 00926")); len(claims) != 0 {
		t.Errorf("got %d claims for a bare designator, want none", len(claims))
	}
}

// The designator has to open its line. Claiming from partway along one would
// swallow whatever preceded it into a component that cannot contain it.
func TestADesignatorPartwayAlongALineIsNotClaimed(t *testing.T) {
	if claims := puertorico.Claims(token.Tokenize("ACME CORP URB HIGHLAND GDNS\n123 CALLE MAIN\nSAN JUAN PR 00926")); len(claims) != 0 {
		t.Errorf("got %d claims, want none: the designator does not open the line", len(claims))
	}
}

// An urbanization sits above the street line, so something has to follow it.
// On a single line there is nothing to say where the name stops, and the
// reading is declined rather than offered weakly.
func TestAnUrbanizationIsNeverTheLastLine(t *testing.T) {
	for _, source := range []string{
		"URB LAS GLADIOLAS 150 CALLE A SAN JUAN PR 00926",
		"123 CALLE MAIN\nURB LAS GLADIOLAS",
	} {
		t.Run(source, func(t *testing.T) {
			if claims := puertorico.Claims(token.Tokenize(source)); len(claims) != 0 {
				t.Errorf("got %d claims, want none", len(claims))
			}
		})
	}
}

// Every claim has to be usable by a parser: one part, in range, naming the
// Area, and never assigning a token twice.
func TestEveryClaimIsWellFormed(t *testing.T) {
	sources := []string{
		"URB HIGHLAND GDNS\nCOND LAS AMAPOLAS APT 103\n123 CALLE MAIN\nSAN JUAN PR 00926",
		"URB LAS GLADIOLAS\n150 CALLE A\nSAN JUAN PR 00926",
		"URB SANTA MARIA\nURB LAS FLORES\nSAN JUAN PR 00926",
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			tokens := token.Tokenize(source)
			for _, c := range puertorico.Claims(tokens) {
				if len(c.Parts) != 1 {
					t.Fatalf("claim has %d parts, want 1", len(c.Parts))
				}
				p := c.Parts[0]
				if p.Part != claim.PartArea {
					t.Errorf("part = %q, want %q", p.Part, claim.PartArea)
				}
				if p.Start < 0 || p.End() > len(tokens) {
					t.Errorf("part [%d,%d) is outside the %d tokens", p.Start, p.End(), len(tokens))
				}
				if p.Length < 2 {
					t.Errorf("part length = %d, want at least a designator and a name", p.Length)
				}
				if c.Confidence != claim.ConfidenceExact {
					t.Errorf("confidence = %d, want exact", c.Confidence)
				}
			}
		})
	}
}
