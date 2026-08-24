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
//
// The urbanization is deliberately absent. The standard puts it on a line of
// its own above the secondary address identifier, and this library carries it
// in Address.Area rather than as a secondary designator, so it has its own
// vocabulary in urbanization.go. Listing it in both places would make URB two
// things at once.
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

// prPostalPrefixes are the three-digit ZIP prefixes assigned to Puerto Rico.
// 008 is deliberately absent: it belongs to the US Virgin Islands, which does
// not use the Spanish vocabulary in this package.
var prPostalPrefixes = map[string]bool{
	"006": true,
	"007": true,
	"009": true,
}

// LooksLikePRPostal reports whether a postal code falls in a Puerto Rico ZIP
// range. Non-digits are ignored, so "00926" and "00926-1234" both match.
func LooksLikePRPostal(postal string) bool {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, postal)

	if len(digits) < 5 {
		return false
	}

	return prPostalPrefixes[digits[:3]]
}

// UsePRDialect reports whether Puerto Rico Spanish vocabulary applies to an
// address. A region code of PR is authoritative; a Puerto Rico ZIP range
// engages the dialect on its own, so an address that arrives without a region
// is still read in Spanish.
func UsePRDialect(regionCode, postal string) bool {
	if strings.EqualFold(strings.TrimSpace(regionCode), "PR") {
		return true
	}

	return LooksLikePRPostal(postal)
}
