package normalizeroptions_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizeroptions"
)

// The original is a copy: normalizing the address must not change it.
func TestNewFnOptionsCopiesTheOriginal(t *testing.T) {
	a := &address.Address{StreetName: "MONTANA TREASURE"}
	o := normalizeroptions.NewFnOptions(a, *normalizeroptions.Matching())
	a.StreetName = "MT TREASURE"
	if o.OriginalAddress == a || o.OriginalAddress.StreetName != "MONTANA TREASURE" {
		t.Errorf("OriginalAddress = %+v, want a copy with StreetName MONTANA TREASURE", o.OriginalAddress)
	}
	if !o.Fuzzy {
		t.Error("options not carried: Fuzzy = false, want Matching's true")
	}
	if normalizeroptions.NewFnOptions(nil, *normalizeroptions.Content()).OriginalAddress != nil {
		t.Error("OriginalAddress for a nil address is not nil")
	}
}
