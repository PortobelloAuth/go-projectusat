package puertorico

import (
	"fmt"
	"maps"
	"strings"
)

/*
Spanish        Sp. Abreviation    English
AVENIDA   <->  AVE                Avenue
CALLE     <->  CLL                Street
CAMINITO  <->  CMT                Little Road
CAMINO    <->  CAM                Road
CERRADA   <->  CER                Closed
CIRCULO   <->  CIR                Circle
ENTRADA   <->  ENT                Entrance
PASEO     <->  PSO                Path
PLACITA   <->  PLA                Little Plaza
RANCHO    <->  RCH                Ranch
VEREDA    <->  VER                Small Path
VISTA     <->  VIS                View
*/

/*
Spanish             English
Apartado       <->  PO Box
Buzon          <->  Box
Buzon Rural    <->  Rural Box
Ruta Rural     <->  Rural Route
Ruta Estrella  <->  Highway Contract
Edificio       <->  Building

NOTE: for Puerto Rico addresses, normalize to the Spanish word
*/

/*
Apartamento APT
Barriada BDA
Building BLDG
Bloque BL
Barrio BO
Carretera CARR
Caserio CAS
Condominio COND
Cooperativa COOP
Corporacion CORP
Departamento DEPT
Edificio EDIF
Entrega General GEN DEL
Extencion EXT
Hospital HOSP
Industrial IND
Jardines JARD
Mansiones MANS
Parcelas PARC
Quebrada QBDA
Reparto REPTO
Residencial RES
Sector SEC
Terraza TERR
Urbanization URB
Villa VIL
*/

// streetTypeMap maps Spanish primary street type -> abbreviation.
// Project US@ keeps Spanish forms (do not force English).
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

var streetTypeShortMap = maps.Collect(func(yield func(string, string) bool) {
	for primary, short := range streetTypeMap {
		if !yield(short, primary) {
			return
		}
	}
})

// secondaryMap maps Spanish/English primary secondary designator -> abbreviation.
// Per Project US@ secondary designators, Normalize returns the uppercase short form.
var secondaryMap = map[string]string{
	"APARTAMENTO":     "APT",
	"BARRIADA":        "BDA",
	"BUILDING":        "BLDG",
	"BLOQUE":          "BL",
	"BARRIO":          "BO",
	"CARRETERA":       "CARR",
	"CASERIO":         "CAS",
	"CONDOMINIO":      "COND",
	"COOPERATIVA":     "COOP",
	"CORPORACION":     "CORP",
	"DEPARTAMENTO":    "DEPT",
	"EDIFICIO":        "EDIF",
	"ENTREGA GENERAL": "GEN DEL",
	"EXTENCION":       "EXT",
	"HOSPITAL":        "HOSP",
	"INDUSTRIAL":      "IND",
	"JARDINES":        "JARD",
	"MANSIONES":       "MANS",
	"PARCELAS":        "PARC",
	"QUEBRADA":        "QBDA",
	"REPARTO":         "REPTO",
	"RESIDENCIAL":     "RES",
	"SECTOR":          "SEC",
	"TERRAZA":         "TERR",
	"URBANIZATION":    "URB",
	"URBANIZACION":    "URB", // Spanish spelling (one z)
	"VILLA":           "VIL",
}

var secondaryShortMap = maps.Collect(func(yield func(string, string) bool) {
	for primary, short := range secondaryMap {
		if !yield(short, primary) {
			return
		}
	}
})

// NormalizeStreetType maps a Spanish PR street type (primary or abbreviation)
// to the Spanish primary form. Example: "AVE" or "AVENIDA" -> "AVENIDA".
func NormalizeStreetType(s string) (string, error) {
	capitalized := strings.ToUpper(strings.TrimSpace(s))

	if primary, ok := streetTypeShortMap[capitalized]; ok {
		return primary, nil
	}
	if _, ok := streetTypeMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized street type")
}

// AbbreviateStreetType maps a Spanish PR street type (primary or abbreviation)
// to the Spanish abbreviation. Example: "AVE" or "AVENIDA" -> "AVE".
func AbbreviateStreetType(s string) (string, error) {
	capitalized := strings.ToUpper(strings.TrimSpace(s))

	if short, ok := streetTypeMap[capitalized]; ok {
		return short, nil
	}
	if _, ok := streetTypeShortMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized street type")
}

// NormalizeSecondary maps a Puerto Rico secondary designator (primary or short)
// to the uppercase abbreviation. Example: "Apartamento" or "APT" -> "APT".
func NormalizeSecondary(s string) (string, error) {
	capitalized := strings.ToUpper(strings.TrimSpace(s))
	// collapse internal whitespace for multi-word designators (e.g. GEN DEL)
	capitalized = strings.Join(strings.Fields(capitalized), " ")

	if short, ok := secondaryMap[capitalized]; ok {
		return short, nil
	}
	if _, ok := secondaryShortMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized secondary designator")
}

// Score returns how strongly token looks like a PR Spanish street type or secondary.
// 0 = not PR vocabulary. Street types score 100; secondaries score 90.
func Score(token string) (int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, nil
	}
	if _, err := NormalizeStreetType(token); err == nil {
		return 100, nil
	}
	if _, err := NormalizeSecondary(token); err == nil {
		return 90, nil
	}
	return 0, nil
}

// LooksLikePRPostal reports whether a postal code falls in Puerto Rico ZIP ranges
// (006xx–007xx, 009xx).
func LooksLikePRPostal(postal string) bool {
	compact := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, postal)
	if len(compact) < 5 {
		return false
	}
	prefix := compact[:3]
	switch prefix {
	case "006", "007", "009":
		return true
	default:
		return false
	}
}

// UsePRDialect reports whether PR Spanish vocabulary should be applied based on
// region code and/or postal ZIP range.
func UsePRDialect(regionCode, postal string) bool {
	if strings.EqualFold(strings.TrimSpace(regionCode), "PR") {
		return true
	}
	return LooksLikePRPostal(postal)
}

// TryAbbreviateStreetType is a high-level PR street-type lookup for normalize
// fallbacks. ok is false when the token is not PR vocabulary.
func TryAbbreviateStreetType(s string) (abbr string, ok bool) {
	abbr, err := AbbreviateStreetType(s)
	if err != nil {
		return "", false
	}
	return abbr, true
}

// TryNormalizeSecondary is a high-level PR secondary lookup for normalize fallbacks.
func TryNormalizeSecondary(s string) (abbr string, ok bool) {
	abbr, err := NormalizeSecondary(s)
	if err != nil {
		return "", false
	}
	return abbr, true
}
