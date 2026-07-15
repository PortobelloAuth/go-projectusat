package directionals_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
)

var directionMap = map[string]string{
	"NORTH":     "N",
	"EAST":      "E",
	"SOUTH":     "S",
	"WEST":      "W",
	"NORTHEAST": "NE",
	"SOUTHEAST": "SE",
	"NORTHWEST": "NW",
	"SOUTHWEST": "SW",
}

func TestDirectionalRoundTrip(t *testing.T) {
	for full, short := range directionMap {
		abrev, err := directionals.AbbreviateDirectional(full)
		if err != nil {
			t.Errorf("%s", err)
		}
		if abrev != short {
			t.Errorf("Unexpected directional abreviation %s for %s. Expected: %s", abrev, full, short)
		}

		primary, err := directionals.NormalizeDirectional(short)
		if err != nil {
			t.Errorf("%s", err)
		}
		if primary != full {
			t.Errorf("Unexpected directional primary text %s for %s. Expected: %s", primary, short, full)
		}
	}
}

func TestDirectionalStaySame(t *testing.T) {
	for full, short := range directionMap {
		abrev, err := directionals.AbbreviateDirectional(short)
		if err != nil {
			t.Errorf("%s", err)
		}
		if abrev != short {
			t.Errorf("Unexpected directional abreviation %s for %s. Expected: %s", abrev, short, short)
		}

		primary, err := directionals.NormalizeDirectional(full)
		if err != nil {
			t.Errorf("%s", err)
		}
		if primary != full {
			t.Errorf("Unexpected directional primary text %s for %s. Expected: %s", primary, full, full)
		}
	}
}

func TestDirectionalUnknown(t *testing.T) {
	fake := "Fake"
	abrev, err := directionals.AbbreviateDirectional(fake)
	if err == nil {
		t.Errorf("Invalid directional '%s' should produce an error", fake)
	}
	if abrev != "" {
		t.Errorf("Invalid directional '%s' should produce an empty string. Got: '%s'", fake, abrev)
	}

	unknown := "U"
	primary, err := directionals.NormalizeDirectional(unknown)
	if err == nil {
		t.Errorf("Invalid directional abreviation '%s' should produce an error", unknown)
	}
	if primary != "" {
		t.Errorf("Invalid directional abreviation '%s' should produce an empty string. Got: '%s'", unknown, abrev)
	}
}
