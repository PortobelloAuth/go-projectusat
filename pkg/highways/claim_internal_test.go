package highways

import (
	"reflect"
	"strings"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

// claimsReference is Claims as it was before the performance work on #194:
// the compound-directional check ahead of the recognizer, and no
// leading-word guard. Claims must return exactly what it returns.
func claimsReference(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for start := range tokens {
		span := min(maxSpan, token.LineEnd(tokens, start)-start)
		for length := span; length >= 1; length-- {
			if splitsCompoundDirectional(tokens, start, start+length) {
				continue
			}

			text := token.Join(tokens[start : start+length])

			normalized, err := NormalizeStreetName(text)
			if err != nil {
				continue
			}
			if outsidePuertoRico(normalized, tokens[start+length:]) {
				continue
			}

			claims = append(claims, claim.Claim{
				Confidence: claim.ConfidenceStrong,
				Parts: []claim.ClaimPart{{
					Start:  start,
					Length: length,
					Part:   claim.PartStreetName,
					Value:  normalized,
				}},
			})
		}
	}

	return claims
}

// claimsInputs are the standard's highway examples, the compound-directional
// cases the order of Claims' filters matters for, every leading word the
// recognizer accepts (alone, lower-cased, and glued), state prefixes, and
// ordinary address lines.
var claimsInputs = []string{
	"COUNTY HIGHWAY 140", "COUNTY HWY 60E", "CNTY HWY 20", "COUNTY RD 441",
	"CR 1185", "CNTY RD 33", "CA COUNTY RD 150", "CALIFORNIA COUNTY ROAD 555",
	"EXPRESSWAY 55", "FARM TO MARKET 1200", "farm to market 1200", "FM 187",
	"HWY FM 1320", "HWY 64", "HWY 11 BYPASS", "HWY 66 FRONTAGE ROAD",
	"HIGHWAY 3 BYP ROAD", "I10", "i10", "IH280", "I-10", "INTERSTATE HWY 680",
	"I 55 BYPASS", "I 26 BYP ROAD", "I 44 FRONTAGE ROAD", "RD 5A", "RT 88",
	"RTE 95", "RANCH RD 620", "ST HIGHWAY 303", "STATE HWY 60", "SR 220",
	"ST RD 86", "SR MM", "ST RT 175", "STATE RTE 260", "TOWNSHIP RD 20",
	"TSR 45", "US 41 SW", "US HWY 44", "KENTUCKY 440", "KENTUCKY HIGHWAY 189",
	"KY 1207", "KY HWY 75", "KY ST HWY 1", "KENTUCKY STATE HIGHWAY 625",
	"NEW YORK 17", "NORTH CAROLINA HIGHWAY 12", "FEDERATED STATES OF MICRONESIA 5",
	"DISTRICT OF COLUMBIA ROUTE 1", "MICRONESIA 4", "RD 39.4",
	"COUNTY ROAD N EAST", "COUNTY ROAD N", "COUNTY ROAD N\nEAST LANSING MI",
	"COUNTY ROAD S W", "HWY 9 N E", "SR N W", "COUNTY ROAD", "CNTY HWY",
	"123 COUNTY ROAD 150 NORTH, AUSTIN TX 78701",
	"4500 FM 1960 RD W STE 200\nHOUSTON TX 77069",
	"PO BOX 12\nKY 1207 OWENSBORO KY 42301", "123 MAIN ST", "OKLAHOMA AVE",
	"I STREET", "US GRANT", "CA MAIN", "WASHINGTON BLVD", "ROUTE", "", " ",
	" HWY 64", "HWY 64", "  HWY 64", "\vI10", "ıH280", "ſR 220",
	"\xffHWY 64", "HWY \xff 64", "ST", "S", "N", "EAST",
	"519 PR 462 LEONARD TX 75452", "123 PR 2\nSAN JUAN PR 00901",
}

func TestClaimsMatchesReference(t *testing.T) {
	for _, in := range claimsInputs {
		for _, s := range []string{in, strings.ToLower(in)} {
			checkClaims(t, s)
		}
	}
}

func FuzzClaimsMatchesReference(f *testing.F) {
	for _, in := range claimsInputs {
		f.Add(in)
	}
	f.Fuzz(checkClaims)
}

// checkClaims fails when Claims and claimsReference disagree on the tokens
// of s.
func checkClaims(t *testing.T, s string) {
	tokens := token.Tokenize(s)
	want := claimsReference(tokens)
	got := Claims(tokens)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Claims(Tokenize(%q)) = %+v; reference = %+v", s, got, want)
	}
}
