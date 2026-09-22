package cityabbreviations_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/cityabbreviations"
)

var expansionMap = map[string]string{
	"ST":  "SAINT",
	"STE": "SAINTE",
	"MT":  "MOUNT",
	"FT":  "FORT",
}

func TestExpand(t *testing.T) {
	for short, full := range expansionMap {
		got, err := cityabbreviations.Expand(short)
		if err != nil {
			t.Errorf("%s", err)
		}
		if got != full {
			t.Errorf("Unexpected expansion %s for %s. Expected: %s", got, short, full)
		}
	}
}

func TestExpandIsCaseInsensitive(t *testing.T) {
	got, err := cityabbreviations.Expand("st")
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != "SAINT" {
		t.Errorf("Unexpected expansion %s for st. Expected: SAINT", got)
	}
}

func TestExpandUnknown(t *testing.T) {
	fake := "FAKE"
	got, err := cityabbreviations.Expand(fake)
	if err == nil {
		t.Errorf("Invalid city abbreviation '%s' should produce an error", fake)
	}
	if got != "" {
		t.Errorf("Invalid city abbreviation '%s' should produce an empty string. Got: '%s'", fake, got)
	}

	// A full form is not itself a key: Expand only answers for the
	// abbreviated spelling, since a full spelling never needs expanding.
	full := "SAINT"
	got, err = cityabbreviations.Expand(full)
	if err == nil {
		t.Errorf("Full spelling '%s' should not be expandable", full)
	}
	if got != "" {
		t.Errorf("Full spelling '%s' should produce an empty string. Got: '%s'", full, got)
	}
}
