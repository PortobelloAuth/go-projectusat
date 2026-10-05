package puertorico_test

import (
	"testing"
)

// The standard's own Spanish designators, p. 29: each is rewritten as PO BOX,
// and the box number is kept.
func TestEveryPuertoRicoSpellingOfAPOBoxIsRewrittenAsPOBox(t *testing.T) {
	for _, tc := range []struct {
		designator string
		want       string
	}{
		{"APARTADO 2018", "PO BOX 2018"},
		{"Apartado 2018", "PO BOX 2018"},
		{"GPO BOX 1118", "PO BOX 1118"},
		{"gpo box 1118", "PO BOX 1118"},
		{"PO BOX S-1190", "PO BOX 1190"},
	} {
		t.Run(tc.designator, func(t *testing.T) {
			got, ok := street(t, "XYZ COMPANY\n"+tc.designator+"\nSAN JUAN PR 00907")
			if !ok {
				t.Fatalf("%q is not read as a Puerto Rico post office box", tc.designator)
			}
			if got != tc.want {
				t.Errorf("street line = %q, want %q", got, tc.want)
			}
		})
	}
}

// PO BOX and GPO BOX contain a BOX designator of their own. The box must be
// offered once, as APARTADO is, not once for each designator it contains.
func TestAMultiWordDesignatorIsOfferedOnceLikeASingleWordOne(t *testing.T) {
	baseline := len(candidates("XYZ COMPANY\nAPARTADO 2018\nSAN JUAN PR 00907"))

	for _, designator := range []string{"PO BOX 2018", "GPO BOX 2018"} {
		t.Run(designator, func(t *testing.T) {
			got := len(candidates("XYZ COMPANY\n" + designator + "\nSAN JUAN PR 00907"))
			if got != baseline {
				t.Errorf("%q offered %d candidates, want %d as for APARTADO", designator, got, baseline)
			}
		})
	}
}

// A designator with no box number after it is not a box, as on the mainland.
func TestASpanishDesignatorWithNoNumberIsNotABox(t *testing.T) {
	for _, c := range candidates("XYZ COMPANY\nAPARTADO\nSAN JUAN PR 00907") {
		if c.Address.StreetName == "PO BOX" {
			t.Errorf("APARTADO with no number was read as a box: %q", c.Address.FormatStreetLine())
		}
	}
}

// The Spanish designator is a Puerto Rico reading only. A mainland address is
// never offered APARTADO as a box, because the dialect gate sits in front of
// every Puerto Rico candidate.
func TestAMainlandAddressIsNotOfferedASpanishPOBox(t *testing.T) {
	if got := candidates("XYZ COMPANY\nAPARTADO 2018\nSPRINGFIELD IL 62701"); len(got) != 0 {
		t.Errorf("mainland address was offered %d Puerto Rico candidates", len(got))
	}
}
