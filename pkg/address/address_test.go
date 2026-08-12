package address_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
)

func TestFormatStreetLine(t *testing.T) {
	cases := []struct {
		name string
		in   address.Address
		want string
	}{
		{
			name: "full order PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM",
			in: address.Address{
				PrimaryNumber:       "123",
				Predirectional:      "N",
				StreetName:          "MAIN",
				StreetSuffix:        "ST",
				Postdirectional:     "SW",
				SecondaryDesignator: "APT",
				SecondaryNumber:     "4",
			},
			want: "123 N MAIN ST SW APT 4",
		},
		{
			name: "omits blank elements",
			in: address.Address{
				PrimaryNumber: "100",
				StreetName:    "OAK",
				StreetSuffix:  "AVE",
			},
			want: "100 OAK AVE",
		},
		{
			name: "secondary without designator still emits number",
			in: address.Address{
				PrimaryNumber:   "50",
				StreetName:      "ELM",
				StreetSuffix:    "RD",
				SecondaryNumber: "2B",
			},
			want: "50 ELM RD 2B",
		},
		{
			name: "empty address",
			in:   address.Address{},
			want: "",
		},
		{
			name: "only street name",
			in:   address.Address{StreetName: "BROADWAY"},
			want: "BROADWAY",
		},
		{
			name: "highway style no suffix",
			in: address.Address{
				PrimaryNumber:   "100",
				Predirectional:  "N",
				StreetName:      "US HIGHWAY 41",
				Postdirectional: "SW",
			},
			want: "100 N US HIGHWAY 41 SW",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.FormatStreetLine()
			if got != tc.want {
				t.Fatalf("FormatStreetLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatLastLine(t *testing.T) {
	cases := []struct {
		name string
		in   address.Address
		want string
	}{
		{
			name: "city region postal",
			in: address.Address{
				City:   "SPRINGFIELD",
				Region: "IL",
				Postal: "62701",
			},
			want: "SPRINGFIELD IL 62701",
		},
		{
			name: "ZIP+4",
			in: address.Address{
				City:   "MIAMI",
				Region: "FL",
				Postal: "33101-1234",
			},
			want: "MIAMI FL 33101-1234",
		},
		{
			name: "omits blank city",
			in: address.Address{
				Region: "IL",
				Postal: "62701",
			},
			want: "IL 62701",
		},
		{
			name: "omits blank postal",
			in: address.Address{
				City:   "SPRINGFIELD",
				Region: "IL",
			},
			want: "SPRINGFIELD IL",
		},
		{
			name: "empty",
			in:   address.Address{},
			want: "",
		},
		{
			name: "canadian postal spaced",
			in: address.Address{
				City:   "OTTAWA",
				Region: "ON",
				Postal: "K1A 0B1",
			},
			want: "OTTAWA ON K1A 0B1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.FormatLastLine()
			if got != tc.want {
				t.Fatalf("FormatLastLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		name string
		in   address.Address
		want string
	}{
		{
			name: "business street last",
			in: address.Address{
				BusinessName:        "ACME CORP",
				PrimaryNumber:       "123",
				StreetName:          "MAIN",
				StreetSuffix:        "ST",
				SecondaryDesignator: "APT",
				SecondaryNumber:     "4",
				City:                "SPRINGFIELD",
				Region:              "IL",
				Postal:              "62701",
			},
			want: "ACME CORP\n123 MAIN ST APT 4\nSPRINGFIELD IL 62701",
		},
		{
			name: "no business line",
			in: address.Address{
				PrimaryNumber: "100",
				StreetName:    "OAK",
				StreetSuffix:  "AVE",
				City:          "MIAMI",
				Region:        "FL",
				Postal:        "33101",
			},
			want: "100 OAK AVE\nMIAMI FL 33101",
		},
		{
			name: "street only omits empty last",
			in: address.Address{
				PrimaryNumber: "50",
				StreetName:    "ELM",
				StreetSuffix:  "RD",
			},
			want: "50 ELM RD",
		},
		{
			name: "last line only",
			in: address.Address{
				City:   "SPRINGFIELD",
				Region: "IL",
				Postal: "62701",
			},
			want: "SPRINGFIELD IL 62701",
		},
		{
			name: "business and last no street",
			in: address.Address{
				BusinessName: "ACME",
				City:         "CHICAGO",
				Region:       "IL",
				Postal:       "60601",
			},
			want: "ACME\nCHICAGO IL 60601",
		},
		{
			name: "empty address",
			in:   address.Address{},
			want: "",
		},
		{
			name: "business only",
			in:   address.Address{BusinessName: "ACME"},
			want: "ACME",
		},
		{
			name: "rural route address type",
			in: address.Address{
				Type:          &ruralroute.RuralRouteAddress{},
				PrimaryNumber: "BOX 125",
				StreetName:    "RR 4",
				City:          "CUMBERLAND",
				Region:        "IA",
				Postal:        "50843",
			},
			want: "RR 4 BOX 125\nCUMBERLAND IA 50843",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.Format()
			if got != tc.want {
				t.Fatalf("Format = %q, want %q", got, tc.want)
			}
		})
	}
}
