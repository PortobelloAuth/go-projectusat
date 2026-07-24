package military

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

/*
"MILITARY ADDRESSES
Patient records containing addresses to Army/Air Post Offices (APOs), or Fleet Post Offices (FPOs) are
required to include the patient’s name and rank, per USPS Publication 28. Guidance for the patient’s
name and rank is out of scope for this document.
APO/FPO patient addresses MUST include the unit, the box number, the APO/FPO address, and the 9-
digit ZIP Code. City or country names MUST NOT be included in APO/FPO shipping addresses.
The Street Address Line for all APO/FPO military patient addresses MUST be standardized to
include the appropriate military address type with its assigned number and a box number. There
are five possible military address types: CMR (Consolidated Mail Room), OMC (Official Mail
Center), PSC (Postal Service Center), UMR (Unit Mail Room), and UNIT. The assigned number
and the box number MUST follow one of these acronyms.

Examples of standardized military address:
Army/Air Post Office (APO)
PSC 3 BOX 4120
APO AE 09021-0002
UNIT 2050 BOX 4190
APO AP 96278-2050
CMR 802 BOX 74
APO AE 09499-0074
Fleet Post Office (FPO)
UNIT 100100 BOX 4120
FPO AP 96691-0104
UNIT 4856 BOX 121
FPO AP 96667-3931
Diplomatic Post Office (DPO)
UNIT 8400 BOX 0000
DPO AE 09498-0048

Domestic Locations

Most domestic military addresses must have a conventional street style address. Domestic
Military addresses MUST use only the city name along with the approved two–character state
abbreviation and the ZIP Code or ZIP+4 Code.

Overseas Locations

Overseas military addresses MUST contain the APO or FPO designation along with a two–
character “state” abbreviation of AE, AP, or AA and the ZIP Code or ZIP+4 Code. AE is used for
armed forces in Europe, the Middle East, Africa, and Canada; AP is for the Pacific; and AA is for
the Americas excluding Canada.

DEPARTMENT OF STATE ADDRESSES
DPOs are postal facilities that operate at one of the Department of State's missions abroad as a branch
post office of the U.S. Postal Service (USPS). DPO patient addresses MUST include the unit, the box
number, the DPO address, and the 9-digit ZIP Code. City or country names MUST NOT be included in
DPO shipping addresses. Patient records containing addresses to DPOs are required to include the
patient’s name, per USPS Publication 28. Guidance for the patient’s name is out of scope for this
document.

Example:
UNIT 9900 BOX 0500
DPO AE 09701-0500
"
*/

// AddressType is a military mail facility designator (street-line prefix).
type AddressType string

const (
	AddressCMR  AddressType = "CMR"  // Consolidated Mail Room
	AddressOMC  AddressType = "OMC"  // Official Mail Center
	AddressPSC  AddressType = "PSC"  // Postal Service Center
	AddressUMR  AddressType = "UMR"  // Unit Mail Room
	AddressUNIT AddressType = "UNIT" // Unit
)

var validAddressTypes = map[string]AddressType{
	"CMR":  AddressCMR,
	"OMC":  AddressOMC,
	"PSC":  AddressPSC,
	"UMR":  AddressUMR,
	"UNIT": AddressUNIT,
}

var validCities = map[string]bool{
	"APO": true,
	"FPO": true,
	"DPO": true,
}

var validRegions = map[string]bool{
	"AE": true, // Armed Forces Europe, Middle East, Africa, Canada
	"AP": true, // Armed Forces Pacific
	"AA": true, // Armed Forces Americas (except Canada)
}

// assignedNumber matches the facility assigned number (digits, optional leading zeros).
var assignedNumber = regexp.MustCompile(`^\d+$`)

// boxNumber matches the box number (digits, optional leading zeros).
var boxNumber = regexp.MustCompile(`^\d+$`)

// postalCode matches ZIP (#####) or ZIP+4 (#####-####).
var postalCode = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

// NormalizeStreetLine normalizes a military street line to
// "{TYPE} {ASSIGNED} BOX {BOXNUM}" (uppercase, single spaces).
// TYPE is one of CMR, OMC, PSC, UMR, UNIT.
func NormalizeStreetLine(line string) (string, error) {
	s := textutil.CollapseSpace(strings.ToUpper(strings.TrimSpace(line)))
	if s == "" {
		return "", fmt.Errorf("empty military street line")
	}

	tokens := strings.Fields(s)
	// Exactly: TYPE ASSIGNED BOX BOXNUM
	if len(tokens) != 4 {
		return "", fmt.Errorf("invalid military street line: %q", line)
	}

	typ, ok := validAddressTypes[tokens[0]]
	if !ok {
		return "", fmt.Errorf("unknown military address type %q", tokens[0])
	}
	if !assignedNumber.MatchString(tokens[1]) {
		return "", fmt.Errorf("invalid assigned number %q", tokens[1])
	}
	if tokens[2] != "BOX" {
		return "", fmt.Errorf("expected BOX, got %q", tokens[2])
	}
	if !boxNumber.MatchString(tokens[3]) {
		return "", fmt.Errorf("invalid box number %q", tokens[3])
	}

	return string(typ) + " " + tokens[1] + " BOX " + tokens[3], nil
}

// NormalizeLastLine normalizes an overseas military last line
// "{APO|FPO|DPO} {AE|AP|AA} {ZIP|ZIP+4}".
// City or country names must not appear; extra tokens are rejected.
func NormalizeLastLine(line string) (city, region, postal string, err error) {
	s := textutil.CollapseSpace(strings.ToUpper(strings.TrimSpace(line)))
	if s == "" {
		return "", "", "", fmt.Errorf("empty military last line")
	}

	tokens := strings.Fields(s)
	// Exactly three tokens — extra city/country names are not allowed.
	if len(tokens) != 3 {
		return "", "", "", fmt.Errorf("invalid military last line: %q", line)
	}

	city = tokens[0]
	if !validCities[city] {
		return "", "", "", fmt.Errorf("invalid military city %q (want APO, FPO, or DPO)", city)
	}

	region = tokens[1]
	if !validRegions[region] {
		return "", "", "", fmt.Errorf("invalid military region %q (want AE, AP, or AA)", region)
	}

	postal = tokens[2]
	if !postalCode.MatchString(postal) {
		return "", "", "", fmt.Errorf("invalid postal code %q", postal)
	}

	return city, region, postal, nil
}

// Score returns how strongly token looks like a military street address type
// (CMR/OMC/PSC/UMR/UNIT) or military city (APO/FPO/DPO) or military region.
// 0 = not military vocabulary.
func Score(token string) (int, error) {
	token = strings.ToUpper(strings.TrimSpace(token))
	if token == "" {
		return 0, nil
	}
	if _, ok := validAddressTypes[token]; ok {
		return 100, nil
	}
	if validCities[token] {
		return 100, nil
	}
	if validRegions[token] {
		return 100, nil
	}
	return 0, nil
}
