package normalizer_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
)

func TestContentNormalizerStreetNameIdempotent(t *testing.T) {
	for _, name := range []string{"CONNECTICUT", "CONNECTICUT AVENUE", "CT AVENUE", "COURT", "NEW JERSEY", "CT"} {
		in := &address.Address{PrimaryNumber: "1", StreetName: name, StreetSuffix: "AVE", Postdirectional: "NW", City: "WASHINGTON", Region: "DC", Postal: "20008"}
		n := normalizer.NewContentNomalizer()
		first, err := n.Normalize(in)
		if err != nil {
			t.Fatalf("%q: %v", name, err)
		}
		second, err := n.Normalize(first)
		if err != nil {
			t.Fatalf("%q second: %v", name, err)
		}
		if first.StreetName != second.StreetName {
			t.Errorf("%q: first %q, second %q", name, first.StreetName, second.StreetName)
		}
	}
}
