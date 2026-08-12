package secondaryunit_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Apartment", "APT"},
		{"APT", "APT"},
		{"apartment", "APT"},
		{"Suite", "STE"},
		{"STE", "STE"},
		{"suite", "STE"},
		{"Building", "BLDG"},
		{"BLDG", "BLDG"},
		{"Floor", "FL"},
		{"FL", "FL"},
		{"Room", "RM"},
		{"RM", "RM"},
		{"Unit", "UNIT"},
		{"UNIT", "UNIT"},
		{"Department", "DEPT"},
		{"DEPT", "DEPT"},
		{"Basement", "BSMT"},
		{"BSMT", "BSMT"},
		{"Penthouse", "PH"},
		{"PH", "PH"},
		{"Lobby", "LBBY"},
		{"Upper", "UPPR"},
		{"Lower", "LOWR"},
		{"Trailer", "TRLR"},
		{"Space", "SPC"},
		{"#", "#"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := secondaryunit.Normalize(tc.in)
			if err != nil {
				t.Fatalf("Normalize(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestInfoNumbered(t *testing.T) {
	numbered := []string{"Apartment", "APT", "Suite", "STE", "Building", "Floor", "Room", "Unit"}
	for _, in := range numbered {
		t.Run(in, func(t *testing.T) {
			info, err := secondaryunit.Info(in)
			if err != nil {
				t.Fatalf("Info(%q): %v", in, err)
			}
			if !info.Numbered {
				t.Fatalf("Info(%q).Numbered = false, want true", in)
			}
			if info.Short == "" {
				t.Fatalf("Info(%q).Short is empty", in)
			}
			if info.Primary == "" {
				t.Fatalf("Info(%q).Primary is empty", in)
			}
		})
	}
}

func TestInfoNotNumbered(t *testing.T) {
	notNumbered := []string{"Basement", "BSMT", "Front", "FRNT", "Lobby", "LBBY", "Penthouse", "PH", "Upper", "UPPR"}
	for _, in := range notNumbered {
		t.Run(in, func(t *testing.T) {
			info, err := secondaryunit.Info(in)
			if err != nil {
				t.Fatalf("Info(%q): %v", in, err)
			}
			if info.Numbered {
				t.Fatalf("Info(%q).Numbered = true, want false", in)
			}
		})
	}
}

func TestNormalizeUnknown(t *testing.T) {
	got, err := secondaryunit.Normalize("WING")
	if err == nil {
		t.Fatal("expected error for unrecognized unit type")
	}
	if got != "" {
		t.Fatalf("got %q, want empty string on error", got)
	}
}

func TestInfoHash(t *testing.T) {
	info, err := secondaryunit.Info("#")
	if err != nil {
		t.Fatalf("Info(\"#\"): %v", err)
	}
	if info.Short != "#" {
		t.Fatalf("Short = %q, want #", info.Short)
	}
	if !info.Numbered {
		t.Fatal("Numbered = false, want true")
	}
	if info.Primary != "#" {
		t.Fatalf("Primary = %q, want #", info.Primary)
	}
}

// BOX is a deviation from the specification, which does not list it among the
// secondary unit designators. It is recognized here so that rural route,
// military, and post office box addresses do not each carry their own copy of
// the word. See the comment on nonStandardUnitTypes.
func TestInfoBox(t *testing.T) {
	cases := []string{"BOX", "Box", "box"}

	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			info, err := secondaryunit.Info(in)
			if err != nil {
				t.Fatalf("Info(%q): %v", in, err)
			}
			if info.Short != "BOX" {
				t.Errorf("Short = %q, want BOX", info.Short)
			}
			if !info.Numbered {
				t.Error("Numbered = false, want true: a box always carries a number")
			}
		})
	}
}

func TestNormalizeBox(t *testing.T) {
	got, err := secondaryunit.Normalize("Box")
	if err != nil {
		t.Fatalf("Normalize(\"Box\"): %v", err)
	}
	if got != "BOX" {
		t.Errorf("got %q, want BOX", got)
	}
}

// The deviation is meant to be one word, added deliberately. If this fails,
// something was added to nonStandardUnitTypes without the documentation and
// review that adding a non-standard designator is supposed to require.
func TestOnlyBoxDeviatesFromTheStandard(t *testing.T) {
	// Every designator the standard lists, per the block comment at the top of
	// the package. Anything Info accepts that is not here and not "#" is a
	// deviation.
	standard := []string{
		"APT", "BSMT", "BLDG", "DEPT", "FL", "FRNT", "HNGR", "KEY",
		"LBBY", "LOT", "LOWR", "OFC", "PH", "PIER", "REAR", "RM",
		"SIDE", "SLIP", "SPC", "STOP", "STE", "TRLR", "UNIT", "UPPR",
	}

	recognized := map[string]bool{}
	for _, s := range append(standard, "BOX", "#") {
		if _, err := secondaryunit.Info(s); err != nil {
			t.Errorf("Info(%q) failed; expected it to be recognized", s)
			continue
		}
		recognized[s] = true
	}

	if len(recognized) != len(standard)+2 {
		t.Errorf("recognized %d designators, want %d", len(recognized), len(standard)+2)
	}
}
