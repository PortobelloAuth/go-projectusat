package puertorico_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/puertorico"
)

var streetTypeMap = map[string]string{
	"AVENIDA":  "AVE",
	"CALLE":    "CLL",
	"CAMINITO": "CMT",
	"CAMINO":   "CAM",
	"CERRADA":  "CER",
	"CIRCULO":  "CIR",
	"ENTRADA":  "ENT",
	"PASEO":    "PSO",
	"PLACITA":  "PLA",
	"RANCHO":   "RCH",
	"VEREDA":   "VER",
	"VISTA":    "VIS",
}

func TestStreetTypeRoundTrip(t *testing.T) {
	for primary, short := range streetTypeMap {
		abrev, err := puertorico.AbbreviateStreetType(primary)
		if err != nil {
			t.Errorf("AbbreviateStreetType(%q): %v", primary, err)
		}
		if abrev != short {
			t.Errorf("AbbreviateStreetType(%q) = %q, want %q", primary, abrev, short)
		}

		full, err := puertorico.NormalizeStreetType(short)
		if err != nil {
			t.Errorf("NormalizeStreetType(%q): %v", short, err)
		}
		if full != primary {
			t.Errorf("NormalizeStreetType(%q) = %q, want %q", short, full, primary)
		}
	}
}

func TestStreetTypeStaySame(t *testing.T) {
	for primary, short := range streetTypeMap {
		abrev, err := puertorico.AbbreviateStreetType(short)
		if err != nil {
			t.Errorf("AbbreviateStreetType(%q): %v", short, err)
		}
		if abrev != short {
			t.Errorf("AbbreviateStreetType(%q) = %q, want %q", short, abrev, short)
		}

		full, err := puertorico.NormalizeStreetType(primary)
		if err != nil {
			t.Errorf("NormalizeStreetType(%q): %v", primary, err)
		}
		if full != primary {
			t.Errorf("NormalizeStreetType(%q) = %q, want %q", primary, full, primary)
		}
	}
}

func TestStreetTypeCaseAndWhitespace(t *testing.T) {
	cases := []struct {
		in          string
		wantPrimary string
		wantAbrev   string
	}{
		{"avenida", "AVENIDA", "AVE"},
		{" Ave ", "AVENIDA", "AVE"},
		{"calle", "CALLE", "CLL"},
		{"CLL", "CALLE", "CLL"},
		{"Camino", "CAMINO", "CAM"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := puertorico.NormalizeStreetType(tc.in)
			if err != nil {
				t.Fatalf("NormalizeStreetType(%q): %v", tc.in, err)
			}
			if got != tc.wantPrimary {
				t.Fatalf("NormalizeStreetType(%q) = %q, want %q", tc.in, got, tc.wantPrimary)
			}
			abrev, err := puertorico.AbbreviateStreetType(tc.in)
			if err != nil {
				t.Fatalf("AbbreviateStreetType(%q): %v", tc.in, err)
			}
			if abrev != tc.wantAbrev {
				t.Fatalf("AbbreviateStreetType(%q) = %q, want %q", tc.in, abrev, tc.wantAbrev)
			}
		})
	}
}

func TestStreetTypeUnknown(t *testing.T) {
	for _, in := range []string{"STREET", "ST", "Fake", ""} {
		t.Run("normalize/"+in, func(t *testing.T) {
			got, err := puertorico.NormalizeStreetType(in)
			if err == nil {
				t.Fatalf("NormalizeStreetType(%q) expected error", in)
			}
			if got != "" {
				t.Fatalf("NormalizeStreetType(%q) = %q, want empty", in, got)
			}
		})
		t.Run("abbreviate/"+in, func(t *testing.T) {
			got, err := puertorico.AbbreviateStreetType(in)
			if err == nil {
				t.Fatalf("AbbreviateStreetType(%q) expected error", in)
			}
			if got != "" {
				t.Fatalf("AbbreviateStreetType(%q) = %q, want empty", in, got)
			}
		})
	}
}

func TestNormalizeSecondary(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Apartamento", "APT"},
		{"APT", "APT"},
		{"apartamento", "APT"},
		{"Barriada", "BDA"},
		{"BDA", "BDA"},
		{"Building", "BLDG"},
		{"BLDG", "BLDG"},
		{"Bloque", "BL"},
		{"BL", "BL"},
		{"Barrio", "BO"},
		{"BO", "BO"},
		{"Carretera", "CARR"},
		{"CARR", "CARR"},
		{"Caserio", "CAS"},
		{"CAS", "CAS"},
		{"Condominio", "COND"},
		{"COND", "COND"},
		{"Cooperativa", "COOP"},
		{"COOP", "COOP"},
		{"Corporacion", "CORP"},
		{"CORP", "CORP"},
		{"Departamento", "DEPT"},
		{"DEPT", "DEPT"},
		{"Edificio", "EDIF"},
		{"EDIF", "EDIF"},
		{"Entrega General", "GEN DEL"},
		{"GEN DEL", "GEN DEL"},
		{"entrega  general", "GEN DEL"},
		{"Extencion", "EXT"},
		{"EXT", "EXT"},
		{"Hospital", "HOSP"},
		{"HOSP", "HOSP"},
		{"Industrial", "IND"},
		{"IND", "IND"},
		{"Jardines", "JARD"},
		{"JARD", "JARD"},
		{"Mansiones", "MANS"},
		{"MANS", "MANS"},
		{"Parcelas", "PARC"},
		{"PARC", "PARC"},
		{"Quebrada", "QBDA"},
		{"QBDA", "QBDA"},
		{"Reparto", "REPTO"},
		{"REPTO", "REPTO"},
		{"Residencial", "RES"},
		{"RES", "RES"},
		{"Sector", "SEC"},
		{"SEC", "SEC"},
		{"Terraza", "TERR"},
		{"TERR", "TERR"},
		{"Urbanization", "URB"},
		{"URB", "URB"},
		{"Villa", "VIL"},
		{"VIL", "VIL"},
		{" urb ", "URB"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := puertorico.NormalizeSecondary(tc.in)
			if err != nil {
				t.Fatalf("NormalizeSecondary(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeSecondary(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeSecondaryUnknown(t *testing.T) {
	for _, in := range []string{"WING", "SUITE", "Fake", ""} {
		t.Run(in, func(t *testing.T) {
			got, err := puertorico.NormalizeSecondary(in)
			if err == nil {
				t.Fatalf("NormalizeSecondary(%q) expected error", in)
			}
			if got != "" {
				t.Fatalf("NormalizeSecondary(%q) = %q, want empty", in, got)
			}
		})
	}
}
