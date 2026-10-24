package parsertest

import (
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ordinarystreet"
)

// HteCases are representative addresses for the four failure classes the CMS
// HTE patient-matching L3 validation found in go-projectusat v0.1.0-alpha.1
// (cms-hte-validation, results/20261017T023051Z-L3-library-v0.1.0-alpha.1.md;
// go-projectusat#230, #231, #232, #233). It exists so a parser can track its
// progress on those classes with one number.
//
// Every number here is synthetic. Every street is a real one: the Want of
// each row is checked against the zipcity data for its ZIP Code (the street
// and the city are both in it), except the two negative controls marked in the
// table. Each class carries controls: the neighboring spelling the issue says
// parses fine, so a regression on the working form shows up next to the
// failing one.
//
// The classes (the Note on each case names its class):
//
//   - abbrev-city (#230): the last line's city is an abbreviation (NY, BK,
//     BX, BKLYN, or FLUSHING/GLENDALE with a ZIP Code that is not theirs) and
//     the primary number is dropped.
//   - fused-suffix (#231): a numeric or ordinal street name fused to its
//     suffix (133RDST, 133ST).
//   - digit-letter (#232): a leading digits-plus-letters token read whole as
//     the primary number (400E 30ST, 199HENRY ST).
//   - fused-directional (#233): a directional fused to a street number
//     (E135TH, E115).
//
// Each case pins both the rendering (Want) and the decomposition
// (WantFields). The rules behind the Wants, from Pub 28 and the maintainers'
// review: the city is spelled out, using the city zipcity knows for the ZIP
// Code (hteCity); an ordinal is added to a street number only when the data
// has only the ordinal form; a letter after a street number is a directional
// only when the data has the street with it, and stays on the number
// otherwise. A row marked "UNSURE" is one no rule settles and wants a
// reviewer to rule on.
var HteCases = buildHteCases()

// hteRow is one HteCases entry. The input is "<line> \n <city>, NY <zip>".
// The wanted street line is rendered from num, pre, name and suf, and the
// wanted last line is the city as written, NY, and the ZIP Code.
type hteRow struct {
	class string
	line  string
	city  string
	zip   string

	num, pre, name, suf string
}

var hteRows = []hteRow{
	// abbrev-city (#230): the failing shape. The first three rows are the
	// issue's own repro table, on real streets: each street is in the
	// zipcity data for its ZIP Code.
	{"abbrev-city", "227 MADISON ST", "NY", "10002", "227", "", "MADISON", "ST"},
	{"abbrev-city", "227 MADISON ST", "BK", "10002", "227", "", "MADISON", "ST"}, // BK contradicts the ZIP Code; wants the ZIP Code's own city
	{"abbrev-city", "2297 PITKIN AVE", "BK", "11207", "2297", "", "PITKIN", "AVE"},
	{"abbrev-city", "18 MONROE ST", "NY", "10002", "18", "", "MONROE", "ST"},
	{"abbrev-city", "5 MULBERRY ST", "NY", "10013", "5", "", "MULBERRY", "ST"},
	{"abbrev-city", "640 PARK AVE", "NY", "10029", "640", "", "PARK", "AVE"},
	{"abbrev-city", "1021 LEXINGTON AVE", "NY", "10029", "1021", "", "LEXINGTON", "AVE"},
	{"abbrev-city", "77 DELANCEY ST", "NY", "10002", "77", "", "DELANCEY", "ST"},
	{"abbrev-city", "910 LINDEN BLVD", "BK", "11207", "910", "", "LINDEN", "BLVD"},
	{"abbrev-city", "34 MADISON ST", "BK", "11216", "34", "", "MADISON", "ST"},
	{"abbrev-city", "910 ROCKAWAY PKWY", "BK", "11236", "910", "", "ROCKAWAY", "PKWY"},
	{"abbrev-city", "441 LIBERTY AVE", "BK", "11212", "441", "", "LIBERTY", "AVE"},
	{"abbrev-city", "60 CARROLL ST", "BK", "11225", "60", "", "CARROLL", "ST"},
	{"abbrev-city", "1500 NOSTRAND AVE", "BK", "11226", "1500", "", "NOSTRAND", "AVE"},
	{"abbrev-city", "812 SOUTHERN BLVD", "BX", "10459", "812", "", "SOUTHERN", "BLVD"},
	{"abbrev-city", "3010 WESTCHESTER AVE", "BX", "10461", "3010", "", "WESTCHESTER", "AVE"},
	{"abbrev-city", "255 MORRIS AVE", "BX", "10451", "255", "", "MORRIS", "AVE"},
	{"abbrev-city", "47 HUNTS POINT AVE", "BX", "10474", "47", "", "HUNTS POINT", "AVE"},
	{"abbrev-city", "1188 WESTCHESTER AVE", "BX", "10459", "1188", "", "WESTCHESTER", "AVE"},
	{"abbrev-city", "92 LIBERTY AVE", "BKLYN", "11208", "92", "", "LIBERTY", "AVE"},
	{"abbrev-city", "738 BEDFORD AVE", "BKLYN", "11205", "738", "", "BEDFORD", "AVE"},
	{"abbrev-city", "15 CLARK ST", "BKLYN", "11201", "15", "", "CLARK", "ST"},
	{"abbrev-city", "2410 NOSTRAND AVE", "BKLYN", "11210", "2410", "", "NOSTRAND", "AVE"},
	{"abbrev-city", "166 HENRY ST", "BKLYN", "11231", "166", "", "HENRY", "ST"},
	// A neighborhood name on a ZIP Code that is not its own (Pub 28 ch. 2
	// sec. 008 wants the city name spelled out; the ZIP Code's own city is the
	// one zipcity knows). 11354 is FLUSHING, so that row is a control below.
	{"abbrev-city", "43 JUNCTION BLVD", "FLUSHING", "11373", "43", "", "JUNCTION", "BLVD"},
	{"abbrev-city", "19 NORTHERN BLVD", "FLUSHING", "11368", "19", "", "NORTHERN", "BLVD"},
	{"abbrev-city", "301 GRAND AVE", "GLENDALE", "11373", "301", "", "GRAND", "AVE"},
	{"abbrev-city", "58 FRESH POND RD", "GLENDALE", "11378", "58", "", "FRESH POND", "RD"},
	{"abbrev-city", "7 MYRTLE AVE", "GLENDALE", "11237", "7", "", "MYRTLE", "AVE"},
	{"abbrev-city", "301 MYRTLE AVE", "GLENDALE", "11385", "301", "", "MYRTLE", "AVE"},
	// A directional and a suffix on the same street line.
	{"abbrev-city", "221 E 161ST ST", "BX", "10451", "221", "E", "161ST", "ST"},
	{"abbrev-city", "14 W 125TH ST", "NY", "10027", "14", "W", "125TH", "ST"},
	// abbrev-city controls: the city is already spelled out and agrees with
	// the ZIP Code. These pass today.
	{"abbrev-city control", "227 MADISON ST", "NEW YORK", "10002", "227", "", "MADISON", "ST"},
	{"abbrev-city control", "2297 PITKIN AVE", "BROOKLYN", "11207", "2297", "", "PITKIN", "AVE"},
	{"abbrev-city control", "812 SOUTHERN BLVD", "BRONX", "10459", "812", "", "SOUTHERN", "BLVD"},
	{"abbrev-city control", "86 MAIN ST", "FLUSHING", "11354", "86", "", "MAIN", "ST"},

	// fused-suffix (#231): a numeric or ordinal street name fused to its
	// suffix. The ordinal is added where the zipcity data has only the
	// ordinal form (E 167TH ST is in 10459, E 167 ST is not).
	{"fused-suffix", "250 133RDST", "NEW YORK", "10027", "250", "", "133RD", "ST"},
	{"fused-suffix", "41 52NDST", "NEW YORK", "10019", "41", "", "52ND", "ST"},
	{"fused-suffix", "9 81STST", "NEW YORK", "10024", "9", "", "81ST", "ST"},
	{"fused-suffix", "630 194TH ST", "BRONX", "10458", "630", "", "194TH", "ST"},
	{"fused-suffix", "250 133ST", "NEW YORK", "10027", "250", "", "133RD", "ST"},
	{"fused-suffix", "630 194ST", "BRONX", "10458", "630", "", "194TH", "ST"},
	{"fused-suffix", "12 30ST", "NEW YORK", "10016", "12", "", "30TH", "ST"},
	{"fused-suffix", "205 W 133ST", "NEW YORK", "10027", "205", "W", "133RD", "ST"},
	{"fused-suffix", "18 W 104ST", "NEW YORK", "10025", "18", "W", "104TH", "ST"},
	{"fused-suffix", "320 N 7ST", "BROOKLYN", "11211", "320", "N", "7TH", "ST"},
	{"fused-suffix", "88 E 167ST", "BRONX", "10459", "88", "E", "167TH", "ST"},
	{"fused-suffix", "2 E 222ST", "BRONX", "10469", "2", "E", "222ND", "ST"},
	{"fused-suffix", "1501 3RDAVE", "NEW YORK", "10128", "1501", "", "3RD", "AVE"}, // fused AVE is not in #231's table, it is an extrapolated test
	// fused-suffix controls.
	{"fused-suffix control", "250 133RD ST", "NEW YORK", "10027", "250", "", "133RD", "ST"},
	{"fused-suffix control", "205 W 133RD ST", "NEW YORK", "10027", "205", "W", "133RD", "ST"},
	{"fused-suffix control", "630 194TH ST", "BRONX", "10458", "630", "", "194TH", "ST"},
	{"fused-suffix control", "12 30TH ST", "NEW YORK", "10016", "12", "", "30TH", "ST"},

	// digit-letter (#232): a leading token of digits and letters read whole
	// as the primary number. A letter after the digits is a directional only
	// if the data has the street with it, otherwise it stays on the number.
	{"digit-letter", "400E 30ST", "NEW YORK", "10016", "400", "E", "30TH", "ST"},
	{"digit-letter", "400E 30 ST", "NEW YORK", "10016", "400", "E", "30TH", "ST"},
	{"digit-letter", "199HENRY ST", "BROOKLYN", "11201", "199", "", "HENRY", "ST"},
	{"digit-letter", "52W 71 ST", "NEW YORK", "10023", "52", "W", "71ST", "ST"},
	{"digit-letter", "310N 6TH ST", "BROOKLYN", "11211", "310", "N", "6TH", "ST"},
	{"digit-letter", "88S PORTLAND AVE", "BROOKLYN", "11217", "88", "S", "PORTLAND", "AVE"}, // S PORTLAND AVE is in the data
	{"digit-letter", "88S LAKE ST", "BROOKLYN", "11201", "88S", "", "LAKE", "ST"},           // no S LAKE ST in the data: the S stays on the number
	{"digit-letter", "27CLINTON ST", "BROOKLYN", "11201", "27", "", "CLINTON", "ST"},
	{"digit-letter", "1204PARK AVE", "NEW YORK", "10128", "1204", "", "PARK", "AVE"},
	{"digit-letter", "63MORRIS AVE", "BRONX", "10453", "63", "", "MORRIS", "AVE"},
	{"digit-letter", "905E 167 ST", "BRONX", "10459", "905", "E", "167TH", "ST"},
	{"digit-letter", "14W 120ST", "NEW YORK", "10027", "14", "W", "120TH", "ST"}, // combines digit-letter with the fused suffix (#231). This is the result we expect.
	{"digit-letter", "5E 4TH ST", "NEW YORK", "10003", "5", "E", "4TH", "ST"},
	// digit-letter controls: a spaced directional, and the legitimate
	// digits-plus-letter primary number that the issue says must survive.
	{"digit-letter control", "400 E 30 ST", "NEW YORK", "10016", "400", "E", "30TH", "ST"},
	{"digit-letter control", "199 HENRY ST", "BROOKLYN", "11201", "199", "", "HENRY", "ST"},
	{"digit-letter control", "123A CLINTON ST", "BROOKLYN", "11201", "123A", "", "CLINTON", "ST"},
	{"digit-letter control", "45B HENRY ST", "BROOKLYN", "11201", "45B", "", "HENRY", "ST"},
	{"digit-letter control", "52 W 71 ST", "NEW YORK", "10023", "52", "W", "71ST", "ST"},

	// fused-directional (#233): a directional fused to a street number.
	{"fused-directional", "60 E135TH ST", "NEW YORK", "10037", "60", "E", "135TH", "ST"},
	{"fused-directional", "70 E115 ST", "NEW YORK", "10029", "70", "E", "115TH", "ST"},
	{"fused-directional", "8 W72ND ST", "NEW YORK", "10023", "8", "W", "72ND", "ST"},
	{"fused-directional", "315 W125 ST", "NEW YORK", "10027", "315", "W", "125TH", "ST"},
	{"fused-directional", "41 E87TH ST", "NEW YORK", "10128", "41", "E", "87TH", "ST"},
	{"fused-directional", "260 N6TH ST", "BROOKLYN", "11211", "260", "N", "6TH", "ST"},
	{"fused-directional", "19 S1ST ST", "BROOKLYN", "11211", "19", "S", "1ST", "ST"},
	{"fused-directional", "712 E161 ST", "BRONX", "10451", "712", "E", "161ST", "ST"},
	{"fused-directional", "33 W231 ST", "BRONX", "10463", "33", "W", "231ST", "ST"},
	{"fused-directional", "6 E135TH STREET", "NEW YORK", "10037", "6", "E", "135TH", "ST"},
	{"fused-directional", "10 E135ST", "NEW YORK", "10037", "10", "E", "135TH", "ST"}, // combines #233 with the fused suffix (#231)
	// fused-directional controls.
	{"fused-directional control", "60 E 135TH ST", "NEW YORK", "10037", "60", "E", "135TH", "ST"},
	{"fused-directional control", "70 E 115 ST", "NEW YORK", "10029", "70", "E", "115TH", "ST"},
	{"fused-directional control", "8 W 72ND ST", "NEW YORK", "10023", "8", "W", "72ND", "ST"},
	{"fused-directional control", "260 N 6TH ST", "BROOKLYN", "11211", "260", "N", "6TH", "ST"},
	{"fused-directional control", "55 EAST ST", "BROOKLYN", "11222", "55", "", "EAST", "ST"}, // EAST ST is the directional-vs-name ambiguity of #110; a negative control against over-eager splitting
}

// hteCity is the city a row's last line should come out with when the input
// city is an abbreviation or a neighborhood: Pub 28 ch. 2 sec. 008 wants it
// spelled out, and zipcity's USPS locale data names the city for the ZIP Code
// (CitiesKnownFor lists it next to NEW YORK). Keyed by city and ZIP Code; a
// city not in the map is wanted as written.
var hteCity = map[string]string{
	"NY 10002": "NEW YORK", "NY 10013": "NEW YORK", "NY 10029": "NEW YORK", "NY 10027": "NEW YORK",
	"BK 10002": "NEW YORK", // When the abbreviation and the ZIP Code disagree the ZIP Code wins
	"BK 11207": "BROOKLYN", "BK 11216": "BROOKLYN", "BK 11236": "BROOKLYN",
	"BK 11212": "BROOKLYN", "BK 11225": "BROOKLYN", "BK 11226": "BROOKLYN",
	"BX 10459": "BRONX", "BX 10461": "BRONX", "BX 10451": "BRONX", "BX 10474": "BRONX",
	"BKLYN 11208": "BROOKLYN", "BKLYN 11205": "BROOKLYN", "BKLYN 11201": "BROOKLYN",
	"BKLYN 11210": "BROOKLYN", "BKLYN 11231": "BROOKLYN",
	"FLUSHING 11373": "ELMHURST", "FLUSHING 11368": "CORONA",
	"GLENDALE 11373": "ELMHURST", "GLENDALE 11378": "MASPETH",
	"GLENDALE 11237": "BROOKLYN", "GLENDALE 11385": "RIDGEWOOD",
}

func buildHteCases() []Case {
	cases := make([]Case, 0, len(hteRows))
	for _, r := range hteRows {
		street := strings.Join(nonEmpty(r.num, r.pre, r.name, r.suf), " ")
		wantCity := r.city
		if c, ok := hteCity[r.city+" "+r.zip]; ok {
			wantCity = c
		}
		cases = append(cases, Case{
			Source: "hte",
			Note:   "cms-hte L3: " + r.class,
			Input:  r.line + "\n" + r.city + ", NY " + r.zip,
			Want:   street + "\n" + wantCity + " NY " + r.zip,
			WantFields: &address.Address{
				Type:          &ordinarystreet.OrdinaryStreetAddress{},
				PrimaryNumber: r.num, Predirectional: r.pre,
				StreetName: r.name, StreetSuffix: r.suf,
				City: wantCity, Region: "NY", Postal: r.zip,
			},
		})
	}
	return cases
}

func nonEmpty(parts ...string) []string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
