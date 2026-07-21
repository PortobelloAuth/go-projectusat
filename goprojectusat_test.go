package goprojectusat

import (
	"strings"
	"testing"
)

func TestNormalizeBasicStreet(t *testing.T) {
	// 123 Main Street, Apt 4, Springfield IL 62701
	in := Address{
		PrimaryNumber:       "123",
		StreetName:          "Main",
		StreetSuffix:        "Street",
		SecondaryDesignator: "Apartment",
		SecondaryNumber:     "4",
		City:                "Springfield",
		Region:              "Illinois",
		Postal:              "62701",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	want := Address{
		PrimaryNumber:       "123",
		StreetName:          "MAIN",
		StreetSuffix:        "ST",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "4",
		City:                "SPRINGFIELD",
		Region:              "IL",
		Postal:              "62701",
	}
	if got != want {
		t.Fatalf("Normalize = %+v, want %+v", got, want)
	}
}

func TestNormalizeDirectionalsAndHighway(t *testing.T) {
	in := Address{
		PrimaryNumber:    "100",
		Predirectional:   "North",
		StreetName:       "US Hwy 41",
		StreetSuffix:     "",
		Postdirectional:  "Southwest",
		City:             "Miami",
		Region:           "FL",
		Postal:           "33101-1234",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.Predirectional != "N" {
		t.Errorf("Predirectional = %q, want N", got.Predirectional)
	}
	if got.Postdirectional != "SW" {
		t.Errorf("Postdirectional = %q, want SW", got.Postdirectional)
	}
	if got.StreetName != "US HIGHWAY 41" {
		t.Errorf("StreetName = %q, want US HIGHWAY 41", got.StreetName)
	}
	if got.Postal != "33101-1234" {
		t.Errorf("Postal = %q, want 33101-1234", got.Postal)
	}
	if got.Region != "FL" {
		t.Errorf("Region = %q, want FL", got.Region)
	}
}

func TestNormalizeUnknownAndEmpty(t *testing.T) {
	in := Address{
		PrimaryNumber:       "UNKNOWN",
		Predirectional:      "",
		StreetName:          "Main",
		StreetSuffix:        "Ave",
		SecondaryDesignator: "unknown",
		City:                "Springfield",
		Region:              "IL",
		Postal:              "",
		Country:             "Unknown",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: unexpected error: %v", err)
	}
	if got.PrimaryNumber != "" {
		t.Errorf("PrimaryNumber UNKNOWN = %q, want blank", got.PrimaryNumber)
	}
	if got.SecondaryDesignator != "" {
		t.Errorf("SecondaryDesignator UNKNOWN = %q, want blank", got.SecondaryDesignator)
	}
	if got.Country != "" {
		t.Errorf("Country UNKNOWN = %q, want blank", got.Country)
	}
	if got.Predirectional != "" || got.Postdirectional != "" {
		t.Errorf("empty directionals should stay blank, got pre=%q post=%q",
			got.Predirectional, got.Postdirectional)
	}
	if got.StreetSuffix != "AVE" {
		t.Errorf("StreetSuffix = %q, want AVE", got.StreetSuffix)
	}

	// CollapseSpace must run before Upper/NormalizeUnknown so padded UNKNOWN blanks.
	padded, err := Normalize(Address{
		PrimaryNumber: " UNKNOWN ",
		StreetName:    "Main",
		StreetSuffix:  "ST",
		Region:        " UNKNOWN ",
		City:          "Springfield",
	})
	if err != nil {
		t.Fatalf("padded UNKNOWN should not error (region blanks, skips NormalizeRegion): %v", err)
	}
	if padded.PrimaryNumber != "" {
		t.Errorf("PrimaryNumber %q → %q, want blank", " UNKNOWN ", padded.PrimaryNumber)
	}
	if padded.Region != "" {
		t.Errorf("Region %q → %q, want blank", " UNKNOWN ", padded.Region)
	}
}

func TestNormalizePostalVariants(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"62701", "62701"},
		{"62701-1234", "62701-1234"},
		{"627011234", "62701-1234"},
		{"62701 1234", "62701-1234"},
		{"k1a 0b1", "K1A 0B1"},
		{"K1A  0B1", "K1A 0B1"},
		{"", ""},
		{"unknown", ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Normalize(Address{Postal: tc.in, Region: "IL", City: "X", StreetName: "Main", StreetSuffix: "ST"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Postal != tc.want {
				t.Fatalf("Postal %q → %q, want %q", tc.in, got.Postal, tc.want)
			}
		})
	}
}

func TestNormalizePreservesDiacritics(t *testing.T) {
	in := Address{
		StreetName:   "José",
		StreetSuffix: "Street",
		City:         "San José",
		Region:       "CA",
		Postal:       "95112",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.StreetName != "JOSÉ" {
		t.Errorf("StreetName = %q, want JOSÉ (diacritics preserved)", got.StreetName)
	}
	if got.City != "SAN JOSÉ" {
		t.Errorf("City = %q, want SAN JOSÉ", got.City)
	}
}

func TestNormalizeBusinessNameAndWhitespace(t *testing.T) {
	in := Address{
		BusinessName: "  Acme   Corp.  ",
		PrimaryNumber: "  112-10 ",
		StreetName:   "  Bronx  ",
		StreetSuffix: " Road ",
		City:         " Bronx ",
		Region:       " ny ",
		Postal:       " 10475 ",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.BusinessName != "ACME CORP." {
		// baseField only collapses + upper; does not strip period from free text
		t.Errorf("BusinessName = %q, want ACME CORP.", got.BusinessName)
	}
	if got.PrimaryNumber != "112-10" {
		t.Errorf("PrimaryNumber = %q, want 112-10", got.PrimaryNumber)
	}
	if got.StreetName != "BRONX" {
		t.Errorf("StreetName = %q, want BRONX", got.StreetName)
	}
	if got.StreetSuffix != "RD" {
		t.Errorf("StreetSuffix = %q, want RD", got.StreetSuffix)
	}
	if got.Region != "NY" {
		t.Errorf("Region = %q, want NY", got.Region)
	}
}

func TestNormalizeErrors(t *testing.T) {
	t.Run("bad region", func(t *testing.T) {
		_, err := Normalize(Address{Region: "Narnia", City: "X", StreetName: "Main", StreetSuffix: "ST"})
		if err == nil {
			t.Fatal("expected error for unrecognized region")
		}
		if !strings.Contains(err.Error(), "region") {
			t.Errorf("error %q should mention region", err)
		}
	})
	t.Run("bad predirectional", func(t *testing.T) {
		_, err := Normalize(Address{Predirectional: "Sideways", StreetName: "Main", StreetSuffix: "ST", Region: "IL"})
		if err == nil {
			t.Fatal("expected error for unrecognized predirectional")
		}
	})
	t.Run("bad street suffix", func(t *testing.T) {
		_, err := Normalize(Address{StreetName: "Main", StreetSuffix: "NotASuffix", Region: "IL"})
		if err == nil {
			t.Fatal("expected error for unrecognized street suffix")
		}
	})
	t.Run("bad secondary designator", func(t *testing.T) {
		_, err := Normalize(Address{
			StreetName: "Main", StreetSuffix: "ST", Region: "IL",
			SecondaryDesignator: "Wing",
		})
		if err == nil {
			t.Fatal("expected error for unrecognized secondary designator")
		}
	})
}

func TestNormalizeEmptyAddress(t *testing.T) {
	got, err := Normalize(Address{})
	if err != nil {
		t.Fatalf("empty Address should not error: %v", err)
	}
	if got != (Address{}) {
		t.Fatalf("got %+v, want empty Address", got)
	}
}

func TestFormatStreetLine(t *testing.T) {
	cases := []struct {
		name string
		in   Address
		want string
	}{
		{
			name: "full order PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM",
			in: Address{
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
			in: Address{
				PrimaryNumber: "100",
				StreetName:    "OAK",
				StreetSuffix:  "AVE",
			},
			want: "100 OAK AVE",
		},
		{
			name: "secondary without designator still emits number",
			in: Address{
				PrimaryNumber:   "50",
				StreetName:      "ELM",
				StreetSuffix:    "RD",
				SecondaryNumber: "2B",
			},
			want: "50 ELM RD 2B",
		},
		{
			name: "empty address",
			in:   Address{},
			want: "",
		},
		{
			name: "only street name",
			in:   Address{StreetName: "BROADWAY"},
			want: "BROADWAY",
		},
		{
			name: "highway style no suffix",
			in: Address{
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
			got := FormatStreetLine(tc.in)
			if got != tc.want {
				t.Fatalf("FormatStreetLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatLastLine(t *testing.T) {
	cases := []struct {
		name string
		in   Address
		want string
	}{
		{
			name: "city region postal",
			in: Address{
				City:   "SPRINGFIELD",
				Region: "IL",
				Postal: "62701",
			},
			want: "SPRINGFIELD IL 62701",
		},
		{
			name: "ZIP+4",
			in: Address{
				City:   "MIAMI",
				Region: "FL",
				Postal: "33101-1234",
			},
			want: "MIAMI FL 33101-1234",
		},
		{
			name: "omits blank city",
			in: Address{
				Region: "IL",
				Postal: "62701",
			},
			want: "IL 62701",
		},
		{
			name: "omits blank postal",
			in: Address{
				City:   "SPRINGFIELD",
				Region: "IL",
			},
			want: "SPRINGFIELD IL",
		},
		{
			name: "empty",
			in:   Address{},
			want: "",
		},
		{
			name: "canadian postal spaced",
			in: Address{
				City:   "OTTAWA",
				Region: "ON",
				Postal: "K1A 0B1",
			},
			want: "OTTAWA ON K1A 0B1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatLastLine(tc.in)
			if got != tc.want {
				t.Fatalf("FormatLastLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		name string
		in   Address
		want string
	}{
		{
			name: "business street last",
			in: Address{
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
			in: Address{
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
			in: Address{
				PrimaryNumber: "50",
				StreetName:    "ELM",
				StreetSuffix:  "RD",
			},
			want: "50 ELM RD",
		},
		{
			name: "last line only",
			in: Address{
				City:   "SPRINGFIELD",
				Region: "IL",
				Postal: "62701",
			},
			want: "SPRINGFIELD IL 62701",
		},
		{
			name: "business and last no street",
			in: Address{
				BusinessName: "ACME",
				City:         "CHICAGO",
				Region:       "IL",
				Postal:       "60601",
			},
			want: "ACME\nCHICAGO IL 60601",
		},
		{
			name: "empty address",
			in:   Address{},
			want: "",
		},
		{
			name: "business only",
			in:   Address{BusinessName: "ACME"},
			want: "ACME",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Format(tc.in)
			if got != tc.want {
				t.Fatalf("Format = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatAfterNormalize(t *testing.T) {
	// Integration: Normalize then Format yields content-form multiline address.
	in := Address{
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
	norm, err := Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	got := Format(norm)
	want := "ACME CORP\n123 N MAIN ST SW APT 4\nSPRINGFIELD IL 62701"
	if got != want {
		t.Fatalf("Format(Normalize(...)) = %q, want %q", got, want)
	}
}
