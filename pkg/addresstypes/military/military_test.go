package military_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/military"
)

func TestNormalizeStreetLine(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Brief / Publication 28 examples
		{"PSC 3 BOX 4120", "PSC 3 BOX 4120"},
		{"UNIT 2050 BOX 4190", "UNIT 2050 BOX 4190"},
		{"UNIT 100100 BOX 4120", "UNIT 100100 BOX 4120"},
		{"UNIT 8400 BOX 0000", "UNIT 8400 BOX 0000"},
		{"CMR 802 BOX 74", "CMR 802 BOX 74"},
		{"UNIT 4856 BOX 121", "UNIT 4856 BOX 121"},
		{"UNIT 9900 BOX 0500", "UNIT 9900 BOX 0500"},

		// All five address types
		{"OMC 12 BOX 99", "OMC 12 BOX 99"},
		{"UMR 5 BOX 1", "UMR 5 BOX 1"},

		// Case / whitespace
		{"  psc   3  box  4120  ", "PSC 3 BOX 4120"},
		{"unit 2050 box 4190", "UNIT 2050 BOX 4190"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := military.NormalizeStreetLine(tc.in)
			if err != nil {
				t.Fatalf("NormalizeStreetLine(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeStreetLine(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeStreetLineErrors(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"PSC 3",                // missing BOX + number
		"PSC BOX 4120",         // missing assigned number
		"FOO 3 BOX 4120",       // unknown type
		"PSC 3 BOX",            // missing box number
		"PSC 3 BOX 4120 EXTRA", // trailing junk
		"MAIN ST",              // not military
		"BOX 4120",             // no type
	}

	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			got, err := military.NormalizeStreetLine(in)
			if err == nil {
				t.Fatalf("NormalizeStreetLine(%q) expected error, got %q", in, got)
			}
			if got != "" {
				t.Fatalf("NormalizeStreetLine(%q) = %q, want empty on error", in, got)
			}
		})
	}
}

func TestNormalizeLastLine(t *testing.T) {
	cases := []struct {
		in             string
		city, reg, zip string
	}{
		// Brief / Publication 28 examples
		{"APO AE 09021-0002", "APO", "AE", "09021-0002"},
		{"APO AP 96278-2050", "APO", "AP", "96278-2050"},
		{"FPO AP 96691-0104", "FPO", "AP", "96691-0104"},
		{"FPO AP 96667-3931", "FPO", "AP", "96667-3931"},
		{"DPO AE 09498-0048", "DPO", "AE", "09498-0048"},
		{"DPO AE 09701-0500", "DPO", "AE", "09701-0500"},
		{"APO AE 09499-0074", "APO", "AE", "09499-0074"},

		// ZIP without +4
		{"APO AA 34002", "APO", "AA", "34002"},

		// Case / whitespace
		{"  apo  ae  09021-0002  ", "APO", "AE", "09021-0002"},
		{"fpo ap 96691-0104", "FPO", "AP", "96691-0104"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			city, reg, zip, err := military.NormalizeLastLine(tc.in)
			if err != nil {
				t.Fatalf("NormalizeLastLine(%q) unexpected error: %v", tc.in, err)
			}
			if city != tc.city || reg != tc.reg || zip != tc.zip {
				t.Fatalf("NormalizeLastLine(%q) = (%q, %q, %q), want (%q, %q, %q)",
					tc.in, city, reg, zip, tc.city, tc.reg, tc.zip)
			}
		})
	}
}

func TestNormalizeLastLineErrors(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"APO AE",                      // missing ZIP
		"APO 09021-0002",              // missing region
		"AE 09021-0002",               // missing city
		"NYC NY 10001",                // not military
		"APO AE GERMANY 09021-0002",   // country name not allowed
		"APO FRANKFURT AE 09021-0002", // city name not allowed
		"FRANKFURT AE 09021-0002",     // city must be APO/FPO/DPO
		"APO XX 09021-0002",           // invalid region
		"APO AE 9021",                 // bad ZIP length
		"APO AE 09021-000",            // bad ZIP+4
		"APO AE 090210002",            // missing hyphen in ZIP+4
		"APO AE 09021-0002 EXTRA",     // trailing junk
		"DPO CA 90210",                // domestic state not military region
	}

	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			city, reg, zip, err := military.NormalizeLastLine(in)
			if err == nil {
				t.Fatalf("NormalizeLastLine(%q) expected error, got (%q, %q, %q)", in, city, reg, zip)
			}
			if city != "" || reg != "" || zip != "" {
				t.Fatalf("NormalizeLastLine(%q) = (%q, %q, %q), want empty on error", in, city, reg, zip)
			}
		})
	}
}

func TestAddressTypeConstants(t *testing.T) {
	// Ensure exported AddressType values match USPS acronyms.
	want := map[military.AddressType]string{
		military.AddressCMR:  "CMR",
		military.AddressOMC:  "OMC",
		military.AddressPSC:  "PSC",
		military.AddressUMR:  "UMR",
		military.AddressUNIT: "UNIT",
	}
	for k, v := range want {
		if string(k) != v {
			t.Fatalf("AddressType %q = %q, want %q", k, string(k), v)
		}
	}
}
