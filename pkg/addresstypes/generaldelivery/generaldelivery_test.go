package generaldelivery_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/generaldelivery"
)

func TestEverySpellingBecomesTheSpelledOutOne(t *testing.T) {
	// The standard requires the words GENERAL DELIVERY, all uppercase, spelled
	// out, whatever the record arrived as.
	cases := []string{
		"GENERAL DELIVERY",
		"general delivery",
		"General Delivery",
		"  GENERAL   DELIVERY  ",
		"GENERAL DEL",
		"GEN DELIVERY",
		"GEN DEL",
		"Gen. Del.",
	}

	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			got, err := generaldelivery.Normalize(in)
			if err != nil {
				t.Fatalf("Normalize(%q) error = %v, want nil", in, err)
			}

			if got != "GENERAL DELIVERY" {
				t.Errorf("Normalize(%q) = %q, want %q", in, got, "GENERAL DELIVERY")
			}
		})
	}
}

func TestSomethingThatMerelyStartsTheSameWayIsNotGeneralDelivery(t *testing.T) {
	// The phrase is the whole street address line, so there is nothing that may
	// follow it and nothing it may follow. Accepting these would drop the extra
	// tokens, and a dropped token is how two different addresses come to look
	// alike.
	cases := []string{
		"GENERAL DELIVERY 5",
		"123 GENERAL DELIVERY",
		"GENERAL DELIVERY LN",
		"GENERAL",
		"DELIVERY",
		"DEL",
		"GENERAL DELIVERANCE",
		"ENTREGA GENERAL",
		"",
	}

	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			got, err := generaldelivery.Normalize(in)
			if err == nil {
				t.Fatalf("Normalize(%q) = %q, want an error", in, got)
			}
		})
	}
}

func TestTheErrorSaysNothingAboutTheAddress(t *testing.T) {
	// CONTRIBUTING §5: an error is the value most likely to reach a log, a crash
	// report or a bug tracker, so it must not carry any part of the address.
	_, err := generaldelivery.Normalize("742 EVERGREEN TER APT 4")

	if err == nil {
		t.Fatal("Normalize error = nil, want an error")
	}

	for _, leaked := range []string{"742", "EVERGREEN", "TER", "APT", "4"} {
		if contains(err.Error(), leaked) {
			t.Errorf("error %q contains %q from the input", err.Error(), leaked)
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}

	return false
}

func TestTheStreetLineIsTheSpelledOutPhraseAndNothingElse(t *testing.T) {
	// Every other field is dropped: this shape has no number, no unit and no
	// suffix for a value to have come from.
	gd := &generaldelivery.GeneralDeliveryAddress{}

	got := gd.FormatStreetLine(&address.Address{
		StreetName:          "GEN DEL",
		PrimaryNumber:       "742",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "4",
		StreetSuffix:        "TER",
		Detail:              "PMB 12",
	})

	if got != "GENERAL DELIVERY" {
		t.Errorf("FormatStreetLine = %q, want %q", got, "GENERAL DELIVERY")
	}
}

func TestAStreetThatIsNotGeneralDeliveryRendersAsNothing(t *testing.T) {
	// Emitting GENERAL DELIVERY over the top of some other street would state
	// something the address never said.
	gd := &generaldelivery.GeneralDeliveryAddress{}

	got := gd.FormatStreetLine(&address.Address{
		PrimaryNumber: "742",
		StreetName:    "EVERGREEN",
		StreetSuffix:  "TER",
	})

	if got != "" {
		t.Errorf("FormatStreetLine = %q, want the empty string", got)
	}
}

func TestNormalizeAddsTheDashNineNineNineNineAddOnToABareZip(t *testing.T) {
	// The standard (p.22) says every general delivery record SHOULD carry the
	// -9999 add-on. A bare five digit ZIP is completed with it; addressparsers
	// PR#32 proposed doing this in parse.go and Aaron (CTO) redirected it here,
	// to the type's own Normalize, per go-projectusat#138's seam.
	cases := []struct {
		name   string
		postal string
		want   string
	}{
		{"bare five digit ZIP gets the add-on", "33602", "33602-9999"},
		{"an existing ZIP+4 is left alone", "33602-1234", "33602-1234"},
		{"no ZIP at all stays ZIP-less", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gd := &generaldelivery.GeneralDeliveryAddress{}
			in := &address.Address{
				Type:       gd,
				StreetName: "GENERAL DELIVERY",
				Postal:     c.postal,
			}

			got, err := gd.Normalize(in, normalizer.AddressNormalizationOptions{})
			if err != nil {
				t.Fatalf("Normalize(%q) error = %v, want nil", c.postal, err)
			}

			if got.Postal != c.want {
				t.Errorf("Normalize(%q).Postal = %q, want %q", c.postal, got.Postal, c.want)
			}
		})
	}
}

func TestNormalizeRejectsAStreetNameThatIsNotGeneralDelivery(t *testing.T) {
	// Normalize defers to this package's own recognizer for the street line;
	// a line that recognizer rejects is not a general delivery address, and
	// there is nothing here to fill in the -9999 add-on for.
	gd := &generaldelivery.GeneralDeliveryAddress{}
	in := &address.Address{
		Type:       gd,
		StreetName: "EVERGREEN TER",
		Postal:     "33602",
	}

	if _, err := gd.Normalize(in, normalizer.AddressNormalizationOptions{}); err == nil {
		t.Fatal("Normalize error = nil, want an error")
	}
}

func TestNormalizeRejectsAnAddressOfTheWrongType(t *testing.T) {
	// The type assertion guard matches pobox.Normalize: an address built for a
	// different AddressType did not come from this package and Normalize
	// should refuse to guess at it.
	gd := &generaldelivery.GeneralDeliveryAddress{}
	in := &address.Address{
		StreetName: "GENERAL DELIVERY",
		Postal:     "33602",
	}

	if _, err := gd.Normalize(in, normalizer.AddressNormalizationOptions{}); err == nil {
		t.Fatal("Normalize error = nil, want an error")
	}
}

func TestNormalizeRunsTheTypeSpecificStreetNameNormalization(t *testing.T) {
	// An accepted abbreviated spelling like GEN DEL should come back spelled
	// out, confirming Normalize actually ran this package's Normalize()
	// rather than passing the street name through untouched.
	gd := &generaldelivery.GeneralDeliveryAddress{}
	in := &address.Address{
		Type:       gd,
		StreetName: "GEN DEL",
		Postal:     "33602",
	}

	got, err := gd.Normalize(in, normalizer.AddressNormalizationOptions{})
	if err != nil {
		t.Fatalf("Normalize error = %v, want nil", err)
	}

	if got.StreetName != "GENERAL DELIVERY" {
		t.Errorf("Normalize StreetName = %q, want %q", got.StreetName, "GENERAL DELIVERY")
	}
}
