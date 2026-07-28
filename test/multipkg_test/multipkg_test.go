package multipkg_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
)

func TestFormatAfterNormalize(t *testing.T) {
	// Integration: Normalize then Format yields content-form multiline address.
	in := &address.Address{
		BusinessName:        "Acme Corp",
		PrimaryNumber:       "123",
		Predirectional:      "North",
		StreetName:          "Main",
		StreetSuffix:        "Street",
		Postdirectional:     "Southwest",
		SecondaryDesignator: "Apartment",
		SecondaryNumber:     "4",
		City:                "Springfield",
		Region:              "Illinois",
		Postal:              "62701",
	}
	cn := normalizer.NewContentNomalizer()
	norm, err := cn.Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	got := norm.Format()
	want := "ACME CORP\n123 N MAIN ST SW APT 4\nSPRINGFIELD IL 62701"
	if got != want {
		t.Fatalf("Format(Normalize(...)) = %q, want %q", got, want)
	}
}
