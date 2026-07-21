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
