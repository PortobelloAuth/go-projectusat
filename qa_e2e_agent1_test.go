package goprojectusat

import (
	"strings"
	"testing"
)

// TestQAE2E_ParseNormalizeFormat is the Agent-1 happy-path E2E suite:
// free-text → Parse → Normalize → Format, plus structured Normalize alone.
func TestQAE2E_ParseNormalizeFormat(t *testing.T) {
	type tc struct {
		name    string
		raw     string
		wantFmt string
		// optional field checks after Normalize (empty = skip)
		wantBusiness string
		wantPrimary  string
		wantPredir   string
		wantStreet   string
		wantSuffix   string
		wantPostdir  string
		wantSecDes   string
		wantSecNum   string
		wantCity     string
		wantRegion   string
		wantPostal   string
		checkFields  bool
	}

	cases := []tc{
		// --- Civilian multi-line street + last line ---
		{
			name:        "civilian multi-line simple",
			raw:         "123 Main Street\nSpringfield IL 62701",
			wantFmt:     "123 MAIN ST\nSPRINGFIELD IL 62701",
			wantPrimary: "123",
			wantStreet:  "MAIN",
			wantSuffix:  "ST",
			wantCity:    "SPRINGFIELD",
			wantRegion:  "IL",
			wantPostal:  "62701",
			checkFields: true,
		},
		{
			name:    "civilian multi-line full components trailing secondary",
			raw:     "123 North Main Street Apt 4\nSpringfield IL 62701",
			wantFmt: "123 N MAIN ST APT 4\nSPRINGFIELD IL 62701",
			wantPrimary: "123", wantPredir: "N", wantStreet: "MAIN", wantSuffix: "ST",
			wantSecDes: "APT", wantSecNum: "4",
			wantCity: "SPRINGFIELD", wantRegion: "IL", wantPostal: "62701",
			checkFields: true,
		},
		{
			name:    "civilian multi-word city ZIP+4",
			raw:     "100 Oak Avenue\nSalt Lake City UT 84105-1234",
			wantFmt: "100 OAK AVE\nSALT LAKE CITY UT 84105-1234",
			wantCity: "SALT LAKE CITY", wantRegion: "UT", wantPostal: "84105-1234",
			checkFields: true,
		},
		{
			name:    "civilian comma-separated street and last line",
			raw:     "456 Elm Road, Chicago IL 60601",
			wantFmt: "456 ELM RD\nCHICAGO IL 60601",
		},
		{
			name:    "civilian district of columbia region",
			raw:     "1600 Pennsylvania Avenue\nWashington District of Columbia 20500",
			wantFmt: "1600 PENNSYLVANIA AVE\nWASHINGTON DC 20500",
			wantRegion: "DC", checkFields: true,
		},

		// --- Directionals (single + multi-token) ---
		{
			name:    "single predirectional north",
			raw:     "101 North Main Street\nSpringfield IL 62701",
			wantFmt: "101 N MAIN ST\nSPRINGFIELD IL 62701",
			wantPredir: "N", checkFields: true,
		},
		{
			name:    "single postdirectional southwest abbreviated",
			raw:     "100 Main Street SW\nMiami FL 33101",
			wantFmt: "100 MAIN ST SW\nMIAMI FL 33101",
			wantPostdir: "SW", checkFields: true,
		},
		{
			name:    "multi-token pre and post directionals",
			raw:     "1011 South West Main Street North East Apt 12\nSpringfield IL 62701",
			wantFmt: "1011 SW MAIN ST NE APT 12\nSPRINGFIELD IL 62701",
			wantPredir: "SW", wantPostdir: "NE",
			wantSecDes: "APT", wantSecNum: "12",
			checkFields: true,
		},
		{
			name:    "multi-token predirectional N E",
			raw:     "3000 N E Main Street\nSpringfield IL 62701",
			wantFmt: "3000 NE MAIN ST\nSPRINGFIELD IL 62701",
			wantPredir: "NE", checkFields: true,
		},
		{
			name:    "northeast compound word",
			raw:     "101 Northeast Main Street\nSpringfield IL 62701",
			wantFmt: "101 NE MAIN ST\nSPRINGFIELD IL 62701",
			wantPredir: "NE", checkFields: true,
		},

		// --- Secondary trailing and leading ---
		{
			name:    "trailing suite",
			raw:     "200 Market Street Suite 500\nSan Francisco CA 94105",
			wantFmt: "200 MARKET ST STE 500\nSAN FRANCISCO CA 94105",
			wantSecDes: "STE", wantSecNum: "500", checkFields: true,
		},
		{
			name:    "trailing hash glued",
			raw:     "100 Main Street #12\nMiami FL 33101",
			wantFmt: "100 MAIN ST # 12\nMIAMI FL 33101",
			wantSecDes: "#", wantSecNum: "12", checkFields: true,
		},
		{
			name:    "leading apartment secondary",
			raw:     "Apartment 3200 152 South Tech Drive\nMiami FL 33101",
			wantFmt: "152 S TECH DR APT 3200\nMIAMI FL 33101",
			wantPrimary: "152", wantPredir: "S", wantStreet: "TECH", wantSuffix: "DR",
			wantSecDes: "APT", wantSecNum: "3200", checkFields: true,
		},
		{
			name:    "leading hash with primary",
			raw:     "#3200 152 South Tech Drive\nMiami FL 33101",
			wantFmt: "152 S TECH DR # 3200\nMIAMI FL 33101",
			wantSecDes: "#", wantSecNum: "3200", checkFields: true,
		},
		{
			name:    "leading unit with trailing upper",
			raw:     "Unit 3200 152 Tech Drive Upper\nMiami FL 33101",
			wantFmt: "152 TECH DR UNIT 3200 UPPR\nMIAMI FL 33101",
			wantSecDes: "UNIT", wantSecNum: "3200 UPPR", checkFields: true,
		},

		// --- Highway forms ---
		{
			name:    "state highway TN",
			raw:     "3324 TN HIGHWAY 431\nSomewhere KY 40000",
			wantFmt: "3324 TN HIGHWAY 431\nSOMEWHERE KY 40000",
			wantPrimary: "3324", wantStreet: "TN HIGHWAY 431", wantSuffix: "",
			checkFields: true,
		},
		{
			name:    "US hwy free-text rewrite",
			raw:     "100 US Hwy 41\nMiami FL 33101",
			wantFmt: "100 US HIGHWAY 41\nMIAMI FL 33101",
			wantPrimary: "100", wantStreet: "US HIGHWAY 41", checkFields: true,
		},
		{
			name:    "interstate form",
			raw:     "500 I10\nHouston TX 77001",
			wantFmt: "500 INTERSTATE 10\nHOUSTON TX 77001",
			wantStreet: "INTERSTATE 10", checkFields: true,
		},
		{
			name:    "county road form",
			raw:     "12 CR 1185\nAustin TX 78701",
			wantFmt: "12 COUNTY ROAD 1185\nAUSTIN TX 78701",
			wantStreet: "COUNTY ROAD 1185", checkFields: true,
		},

		// --- Rural route, PO Box ---
		{
			name:    "rural route spelled",
			raw:     "Rural Route 91 Box A7\nSpringfield IL 62701",
			wantFmt: "RR 91 BOX A7\nSPRINGFIELD IL 62701",
			wantStreet: "RR 91 BOX A7", wantPrimary: "", wantSuffix: "",
			checkFields: true,
		},
		{
			name:    "rural route RFD hash",
			raw:     "RFD 61 #87b\nSpringfield IL 62701",
			wantFmt: "RR 61 BOX 87B\nSPRINGFIELD IL 62701",
			wantStreet: "RR 61 BOX 87B", checkFields: true,
		},
		{
			name:    "rural route glued RR0061",
			raw:     "RR0061 #87b\nSpringfield IL 62701",
			wantFmt: "RR 61 BOX 87B\nSPRINGFIELD IL 62701",
		},
		{
			name:    "PO Box standard",
			raw:     "PO Box 11890\nSpringfield IL 62701",
			wantFmt: "PO BOX 11890\nSPRINGFIELD IL 62701",
			wantStreet: "PO BOX 11890", wantPrimary: "", checkFields: true,
		},
		{
			name:    "post office box spelled",
			raw:     "Post Office Box 55\nSpringfield IL 62701",
			wantFmt: "PO BOX 55\nSPRINGFIELD IL 62701",
			wantStreet: "PO BOX 55", checkFields: true,
		},
		{
			name:    "RD highway not rural route",
			raw:     "RD 5A\nSomewhere KY 40000",
			wantFmt: "ROAD 5A\nSOMEWHERE KY 40000",
			wantStreet: "ROAD 5A", checkFields: true,
		},

		// --- Military multi-line and comma ---
		{
			name:    "military PSC multi-line",
			raw:     "PSC 3 BOX 4120\nAPO AE 09021-0002",
			wantFmt: "PSC 3 BOX 4120\nAPO AE 09021-0002",
			wantStreet: "PSC 3 BOX 4120",
			wantCity: "APO", wantRegion: "AE", wantPostal: "09021-0002",
			checkFields: true,
		},
		{
			name:    "military UNIT comma-separated",
			raw:     "UNIT 2050 BOX 4190, APO AP 96278-2050",
			wantFmt: "UNIT 2050 BOX 4190\nAPO AP 96278-2050",
			wantStreet: "UNIT 2050 BOX 4190",
			wantCity: "APO", wantRegion: "AP", wantPostal: "96278-2050",
			checkFields: true,
		},
		{
			name:    "military FPO multi-line",
			raw:     "UNIT 100 BOX 1\nFPO AE 09501",
			wantFmt: "UNIT 100 BOX 1\nFPO AE 09501",
			wantCity: "FPO", wantRegion: "AE", checkFields: true,
		},
		{
			name:    "military DPO multi-line",
			raw:     "PSC 1 BOX 2\nDPO AA 34001",
			wantFmt: "PSC 1 BOX 2\nDPO AA 34001",
			wantCity: "DPO", wantRegion: "AA", checkFields: true,
		},

		// --- PR Calle/Avenida/Apartado ---
		{
			name:    "PR Calle trailing number",
			raw:     "Calle Luna 123\nSan Juan PR 00901",
			wantFmt: "123 CALLE LUNA\nSAN JUAN PR 00901",
			wantPrimary: "123", wantStreet: "CALLE LUNA", wantSuffix: "",
			checkFields: true,
		},
		{
			name:    "PR Calle US order",
			raw:     "123 Calle Luna\nSan Juan PR 00901",
			wantFmt: "123 CALLE LUNA\nSAN JUAN PR 00901",
		},
		{
			name:    "PR Avenida",
			raw:     "Avenida Ponce de Leon 456\nSan Juan PR 00907",
			wantFmt: "456 AVENIDA PONCE DE LEON\nSAN JUAN PR 00907",
			wantPrimary: "456", wantStreet: "AVENIDA PONCE DE LEON",
			checkFields: true,
		},
		{
			name:    "PR Apartamento secondary",
			raw:     "Calle Luna 123 Apartamento 4\nSan Juan PR 00901",
			wantFmt: "123 CALLE LUNA APT 4\nSAN JUAN PR 00901",
			wantSecDes: "APT", wantSecNum: "4", checkFields: true,
		},
		{
			name:    "PR URB secondary",
			raw:     "123 Calle Luna URB\nPonce PR 00716",
			wantFmt: "123 CALLE LUNA URB\nPONCE PR 00716",
			wantSecDes: "URB", checkFields: true,
		},
		{
			name:    "PR Apartado",
			raw:     "Apartado 11890\nSan Juan PR 00901",
			wantFmt: "PO BOX 11890\nSAN JUAN PR 00901",
			wantStreet: "PO BOX 11890", wantPrimary: "", checkFields: true,
		},
		{
			name:    "PR Apartado Postal",
			raw:     "Apartado Postal 55\nSan Juan PR 00901",
			wantFmt: "PO BOX 55\nSAN JUAN PR 00901",
			wantStreet: "PO BOX 55", checkFields: true,
		},

		// --- Business multi-line and same-line ---
		{
			name:         "business multi-line",
			raw:          "Acme Corp\n123 Main Street\nSpringfield IL 62701",
			wantFmt:      "ACME CORP\n123 MAIN ST\nSPRINGFIELD IL 62701",
			wantBusiness: "ACME CORP",
			wantPrimary:  "123",
			wantStreet:   "MAIN",
			wantSuffix:   "ST",
			checkFields:  true,
		},
		{
			name:         "business same-line pre-street",
			raw:          "Williamson Medical Center 3000 Edward Curd Lane\nSpringfield IL 62701",
			wantFmt:      "WILLIAMSON MEDICAL CENTER\n3000 EDWARD CURD LN\nSPRINGFIELD IL 62701",
			wantBusiness: "WILLIAMSON MEDICAL CENTER",
			wantPrimary:  "3000",
			wantStreet:   "EDWARD CURD",
			wantSuffix:   "LN",
			checkFields:  true,
		},

		// --- Grid, fractional, multi-secondary ---
		{
			name:    "grid spelled directionals",
			raw:     "1016 East 1700 South\nSalt Lake City UT 84105",
			wantFmt: "1016 E 1700 S\nSALT LAKE CITY UT 84105",
			wantPrimary: "1016", wantPredir: "E", wantStreet: "1700", wantPostdir: "S",
			wantSuffix: "", checkFields: true,
		},
		{
			name:    "grid abbreviated directionals",
			raw:     "842 E 1700 S\nSalt Lake City UT 84105",
			wantFmt: "842 E 1700 S\nSALT LAKE CITY UT 84105",
		},
		{
			name:    "fractional primary half",
			raw:     "123 1/2 Main Street\nSpringfield IL 62701",
			wantFmt: "123 1/2 MAIN ST\nSPRINGFIELD IL 62701",
			wantPrimary: "123", wantStreet: "1/2 MAIN", wantSuffix: "ST",
			checkFields: true,
		},
		{
			name:    "fractional with predirectional after fraction",
			raw:     "45 1/4 North Oak Avenue\nSpringfield IL 62701",
			wantFmt: "45 N 1/4 OAK AVE\nSPRINGFIELD IL 62701",
		},
		{
			name:    "multi-secondary building room",
			raw:     "450 Jane Stanford Way Building 420 Room 120\nSpringfield IL 62701",
			wantFmt: "450 JANE STANFORD WY BLDG 420 RM 120\nSPRINGFIELD IL 62701",
			wantSecDes: "BLDG", wantSecNum: "420 RM 120", checkFields: true,
		},
		{
			name:    "multi-secondary suite floor",
			raw:     "100 Main Street Suite 200 Floor 3\nSpringfield IL 62701",
			wantFmt: "100 MAIN ST STE 200 FL 3\nSPRINGFIELD IL 62701",
			wantSecDes: "STE", wantSecNum: "200 FL 3", checkFields: true,
		},
		{
			name:    "hyphenated NYC primary",
			raw:     "112-10 Bronx Road\nBronx NY 10475",
			wantFmt: "112-10 BRONX RD\nBRONX NY 10475",
			wantPrimary: "112-10", checkFields: true,
		},

		// --- Canadian last line ---
		{
			name:    "canadian last line wellington",
			raw:     "10 Wellington Street\nOttawa ON K1A 0B1",
			wantFmt: "10 WELLINGTON ST\nOTTAWA ON K1A 0B1",
			wantPrimary: "10", wantStreet: "WELLINGTON", wantSuffix: "ST",
			wantCity: "OTTAWA", wantRegion: "ON", wantPostal: "K1A 0B1",
			checkFields: true,
		},
		{
			name:    "canadian last line multi-word city",
			raw:     "100 Main Street\nNiagara Falls ON L2E 6T2",
			wantFmt: "100 MAIN ST\nNIAGARA FALLS ON L2E 6T2",
			wantCity: "NIAGARA FALLS", wantRegion: "ON", wantPostal: "L2E 6T2",
			checkFields: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			parsed, err := Parse(c.raw)
			if err != nil {
				t.Fatalf("Parse(%q): %v", c.raw, err)
			}
			norm, err := Normalize(parsed)
			if err != nil {
				t.Fatalf("Normalize after Parse(%q): %v\nparsed=%+v", c.raw, err, parsed)
			}
			got := Format(norm)
			if got != c.wantFmt {
				t.Errorf("Format = %q\n want = %q\nparsed=%+v\nnorm=%+v", got, c.wantFmt, parsed, norm)
			}
			if !c.checkFields {
				return
			}
			if c.wantBusiness != "" && norm.BusinessName != c.wantBusiness {
				t.Errorf("BusinessName = %q, want %q", norm.BusinessName, c.wantBusiness)
			}
			if c.wantPrimary != "" || (c.checkFields && strings.Contains(c.name, "rural") || strings.Contains(c.name, "PO") || strings.Contains(c.name, "Apartado") || strings.Contains(c.name, "military")) {
				// For explicit empty primary cases we set wantPrimary "" with checkFields.
			}
			// Always check fields that were set in the case (including intentional empty).
			if c.wantPrimary != "" && norm.PrimaryNumber != c.wantPrimary {
				t.Errorf("PrimaryNumber = %q, want %q", norm.PrimaryNumber, c.wantPrimary)
			}
			// Explicit empty primary when documented in name for specials
			if (strings.HasPrefix(c.name, "rural") || strings.HasPrefix(c.name, "PO") ||
				strings.HasPrefix(c.name, "post office") || strings.HasPrefix(c.name, "PR Apartado") ||
				strings.HasPrefix(c.name, "military")) && c.wantPrimary == "" && c.wantStreet != "" {
				if norm.PrimaryNumber != "" {
					t.Errorf("PrimaryNumber = %q, want empty for special form", norm.PrimaryNumber)
				}
			}
			if c.wantPredir != "" && norm.Predirectional != c.wantPredir {
				t.Errorf("Predirectional = %q, want %q", norm.Predirectional, c.wantPredir)
			}
			if c.wantStreet != "" && norm.StreetName != c.wantStreet {
				t.Errorf("StreetName = %q, want %q", norm.StreetName, c.wantStreet)
			}
			if c.wantSuffix != "" && norm.StreetSuffix != c.wantSuffix {
				t.Errorf("StreetSuffix = %q, want %q", norm.StreetSuffix, c.wantSuffix)
			}
			if c.wantPostdir != "" && norm.Postdirectional != c.wantPostdir {
				t.Errorf("Postdirectional = %q, want %q", norm.Postdirectional, c.wantPostdir)
			}
			if c.wantSecDes != "" && norm.SecondaryDesignator != c.wantSecDes {
				t.Errorf("SecondaryDesignator = %q, want %q", norm.SecondaryDesignator, c.wantSecDes)
			}
			if c.wantSecNum != "" && norm.SecondaryNumber != c.wantSecNum {
				t.Errorf("SecondaryNumber = %q, want %q", norm.SecondaryNumber, c.wantSecNum)
			}
			if c.wantCity != "" && norm.City != c.wantCity {
				t.Errorf("City = %q, want %q", norm.City, c.wantCity)
			}
			if c.wantRegion != "" && norm.Region != c.wantRegion {
				t.Errorf("Region = %q, want %q", norm.Region, c.wantRegion)
			}
			if c.wantPostal != "" && norm.Postal != c.wantPostal {
				t.Errorf("Postal = %q, want %q", norm.Postal, c.wantPostal)
			}
		})
	}
}

// TestQAE2E_NormalizeStructured exercises Normalize alone on structured Address
// (no Parse), covering content form and NormalizeWithOptions smoke paths.
func TestQAE2E_NormalizeStructured(t *testing.T) {
	t.Run("content form full street", func(t *testing.T) {
		in := Address{
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
		got, err := Normalize(in)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		want := "123 N MAIN ST SW APT 4\nSPRINGFIELD IL 62701"
		if Format(got) != want {
			t.Errorf("Format = %q, want %q", Format(got), want)
		}
	})

	t.Run("highway structured US Hwy", func(t *testing.T) {
		in := Address{
			PrimaryNumber:   "100",
			StreetName:      "US Hwy 41",
			Postdirectional: "Southwest",
			City:            "Miami",
			Region:          "FL",
			Postal:          "33101-1234",
		}
		got, err := Normalize(in)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		if got.StreetName != "US HIGHWAY 41" {
			t.Errorf("StreetName = %q, want US HIGHWAY 41", got.StreetName)
		}
		if got.Postdirectional != "SW" {
			t.Errorf("Postdirectional = %q, want SW", got.Postdirectional)
		}
		if Format(got) != "100 US HIGHWAY 41 SW\nMIAMI FL 33101-1234" {
			t.Errorf("Format = %q", Format(got))
		}
	})

	t.Run("military street wholly in StreetName", func(t *testing.T) {
		in := Address{
			StreetName: "PSC 3 BOX 4120",
			City:       "APO",
			Region:     "AE",
			Postal:     "09021-0002",
		}
		got, err := Normalize(in)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		if Format(got) != "PSC 3 BOX 4120\nAPO AE 09021-0002" {
			t.Errorf("Format = %q", Format(got))
		}
	})

	t.Run("PR structured Spanish street type as suffix fallback", func(t *testing.T) {
		// When callers put Spanish type in StreetSuffix, Normalize maps via puertorico.
		in := Address{
			PrimaryNumber: "123",
			StreetName:    "Luna",
			StreetSuffix:  "Calle",
			City:          "San Juan",
			Region:        "PR",
			Postal:        "00901",
		}
		got, err := Normalize(in)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}
		// AbbreviateStreetType → CLL for content in suffix field path.
		if got.StreetSuffix != "CLL" {
			t.Errorf("StreetSuffix = %q, want CLL (PR AbbreviateStreetType)", got.StreetSuffix)
		}
		if got.Region != "PR" {
			t.Errorf("Region = %q", got.Region)
		}
	})

	t.Run("postal compact nine digit", func(t *testing.T) {
		got, err := Normalize(Address{
			StreetName: "Main", StreetSuffix: "ST",
			City: "X", Region: "IL", Postal: "627011234",
		})
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got.Postal != "62701-1234" {
			t.Errorf("Postal = %q, want 62701-1234", got.Postal)
		}
	})

	t.Run("canadian postal structured", func(t *testing.T) {
		got, err := Normalize(Address{
			PrimaryNumber: "10", StreetName: "Wellington", StreetSuffix: "Street",
			City: "Ottawa", Region: "Ontario", Postal: "k1a 0b1",
		})
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got.Region != "ON" {
			t.Errorf("Region = %q, want ON", got.Region)
		}
		if got.Postal != "K1A 0B1" {
			t.Errorf("Postal = %q, want K1A 0B1", got.Postal)
		}
		if Format(got) != "10 WELLINGTON ST\nOTTAWA ON K1A 0B1" {
			t.Errorf("Format = %q", Format(got))
		}
	})
}

// TestQAE2E_NormalizeWithOptions smoke-tests Fuzzy, SecondaryAsHash, DiacriticMode.
func TestQAE2E_NormalizeWithOptions(t *testing.T) {
	t.Run("Fuzzy region and suffix typos", func(t *testing.T) {
		in := Address{
			PrimaryNumber: "10",
			StreetName:    "Oak",
			StreetSuffix:  "Aveneu",
			City:          "Sacramento",
			Region:        "Californa",
			Postal:        "95814",
		}
		if _, err := Normalize(in); err == nil {
			t.Fatal("content Normalize should reject mild typos")
		}
		got, err := NormalizeWithOptions(in, Options{Fuzzy: true})
		if err != nil {
			t.Fatalf("Fuzzy: %v", err)
		}
		if got.Region != "CA" || got.StreetSuffix != "AVE" {
			t.Errorf("got Region=%q Suffix=%q, want CA / AVE", got.Region, got.StreetSuffix)
		}
		if Format(got) != "10 OAK AVE\nSACRAMENTO CA 95814" {
			t.Errorf("Format = %q", Format(got))
		}
	})

	t.Run("SecondaryAsHash", func(t *testing.T) {
		in := Address{
			PrimaryNumber: "123", StreetName: "Main", StreetSuffix: "Street",
			SecondaryDesignator: "Apartment", SecondaryNumber: "4",
			City: "Springfield", Region: "IL", Postal: "62701",
		}
		got, err := NormalizeWithOptions(in, Options{SecondaryAsHash: true})
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got.SecondaryDesignator != "#" {
			t.Errorf("SecondaryDesignator = %q, want #", got.SecondaryDesignator)
		}
		if Format(got) != "123 MAIN ST # 4\nSPRINGFIELD IL 62701" {
			t.Errorf("Format = %q", Format(got))
		}
		// Content form keeps APT.
		content, err := Normalize(in)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if content.SecondaryDesignator != "APT" {
			t.Errorf("content SecondaryDesignator = %q, want APT", content.SecondaryDesignator)
		}
	})

	t.Run("DiacriticMode preserve substitute transliterate", func(t *testing.T) {
		in := Address{
			StreetName: "José", StreetSuffix: "Street",
			City: "San José", BusinessName: "Café",
			Region: "CA", Postal: "95112",
		}
		pres, err := NormalizeWithOptions(in, Options{})
		if err != nil {
			t.Fatalf("preserve: %v", err)
		}
		if pres.StreetName != "JOSÉ" || pres.City != "SAN JOSÉ" || pres.BusinessName != "CAFÉ" {
			t.Errorf("preserve: Street=%q City=%q Biz=%q", pres.StreetName, pres.City, pres.BusinessName)
		}

		sub, err := NormalizeWithOptions(in, Options{DiacriticMode: "substitute"})
		if err != nil {
			t.Fatalf("substitute: %v", err)
		}
		if sub.StreetName != "JOSE" || sub.City != "SAN JOSE" || sub.BusinessName != "CAFE" {
			t.Errorf("substitute: Street=%q City=%q Biz=%q", sub.StreetName, sub.City, sub.BusinessName)
		}

		tr, err := NormalizeWithOptions(in, Options{DiacriticMode: "transliterate"})
		if err != nil {
			t.Fatalf("transliterate: %v", err)
		}
		if tr.StreetName != "JOSE" || tr.City != "SAN JOSE" || tr.BusinessName != "CAFE" {
			t.Errorf("transliterate: Street=%q City=%q Biz=%q", tr.StreetName, tr.City, tr.BusinessName)
		}
	})

	t.Run("combined Fuzzy and SecondaryAsHash", func(t *testing.T) {
		in := Address{
			PrimaryNumber: "1", StreetName: "Main", StreetSuffix: "Aveneu",
			SecondaryDesignator: "Suite", SecondaryNumber: "2",
			City: "X", Region: "Californa", Postal: "90001",
		}
		got, err := NormalizeWithOptions(in, Options{Fuzzy: true, SecondaryAsHash: true})
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got.Region != "CA" || got.StreetSuffix != "AVE" || got.SecondaryDesignator != "#" {
			t.Errorf("got Region=%q Suffix=%q Sec=%q", got.Region, got.StreetSuffix, got.SecondaryDesignator)
		}
	})
}

// TestQAE2E_BusinessMerge multi-line + same-line business merge behavior.
func TestQAE2E_BusinessMerge(t *testing.T) {
	// Multi-line firm + same-line prefix: README mergeBusinessName — same-line
	// prepended when multi-line also present: "PREFIX MULTI".
	raw := "Acme Holdings\nWilliamson Wing 3000 Edward Curd Lane\nSpringfield IL 62701"
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	norm, err := Normalize(parsed)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	// Expect both business sources combined and street components.
	if !strings.Contains(norm.BusinessName, "ACME") {
		t.Errorf("BusinessName missing multi-line firm: %q", norm.BusinessName)
	}
	if norm.PrimaryNumber != "3000" {
		t.Errorf("PrimaryNumber = %q, want 3000", norm.PrimaryNumber)
	}
	if norm.StreetName != "EDWARD CURD" || norm.StreetSuffix != "LN" {
		t.Errorf("street = name=%q suffix=%q", norm.StreetName, norm.StreetSuffix)
	}
	// Document actual merge result for punchlist review if surprising.
	t.Logf("BusinessName merge result: %q Format:\n%s", norm.BusinessName, Format(norm))
}

// TestQAE2E_ParseThenOptions composes Parse with exchange options.
func TestQAE2E_ParseThenOptions(t *testing.T) {
	raw := "123 North Main Street Apartment 4\nSpringfield IL 62701"
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Content
	content, err := Normalize(parsed)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if Format(content) != "123 N MAIN ST APT 4\nSPRINGFIELD IL 62701" {
		t.Errorf("content Format = %q", Format(content))
	}
	// Exchange: secondary as hash
	ex, err := NormalizeWithOptions(parsed, Options{SecondaryAsHash: true})
	if err != nil {
		t.Fatalf("NormalizeWithOptions: %v", err)
	}
	if Format(ex) != "123 N MAIN ST # 4\nSPRINGFIELD IL 62701" {
		t.Errorf("exchange Format = %q", Format(ex))
	}
}
