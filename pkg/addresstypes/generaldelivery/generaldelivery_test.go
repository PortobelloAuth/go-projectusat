package generaldelivery_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
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
