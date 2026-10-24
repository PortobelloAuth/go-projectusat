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
// Every street, number and ZIP Code here is synthetic, modeled on the shape
// of a failure bucket. No corpus address is reproduced. Each class carries
// controls: the neighboring spelling the issue says parses fine, so a
// regression on the working form shows up next to the failing one.
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
// (WantFields). Both are the author's best reading of what Project US@ asks
// for, not ground truth reviewed by the maintainers: a row marked
// "UNSURE" in the table below is one the author could not settle and wants a
// reviewer to rule on. A city abbreviation is expected to pass through as
// written, the way "NY, NY 11201" does today; whether NY should become NEW
// YORK is #230's open question and is not asserted.
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
	// issue's own repro table (synthetic streets).
	{"abbrev-city", "227 MADISON ST", "NY", "10002", "227", "", "MADISON", "ST"},
	{"abbrev-city", "227 MADISON ST", "BK", "10002", "227", "", "MADISON", "ST"},
	{"abbrev-city", "2297 PITKIN AVE", "BK", "11207", "2297", "", "PITKIN", "AVE"},
	// More of the same shape. The abbreviation does not agree
	// with the ZIP Code, and the primary number is dropped today.
	{"abbrev-city", "227 ELM ST", "NY", "10002", "227", "", "ELM", "ST"},
	{"abbrev-city", "18 LARCH PL", "NY", "10009", "18", "", "LARCH", "PL"},
	{"abbrev-city", "5 WAVERLY AVE", "NY", "10013", "5", "", "WAVERLY", "AVE"},
	{"abbrev-city", "640 PARKER ST", "NY", "10029", "640", "", "PARKER", "ST"},
	{"abbrev-city", "1021 GROVE RD", "NY", "10035", "1021", "", "GROVE", "RD"},
	{"abbrev-city", "77 CEDAR LN", "NY", "10002", "77", "", "CEDAR", "LN"},
	{"abbrev-city", "2297 ASHLAND AVE", "BK", "11207", "2297", "", "ASHLAND", "AVE"},
	{"abbrev-city", "34 SUMNER ST", "BK", "11216", "34", "", "SUMNER", "ST"},
	{"abbrev-city", "910 HOLLY BLVD", "BK", "11236", "910", "", "HOLLY", "BLVD"},
	{"abbrev-city", "441 REMSEN ST", "BK", "11212", "441", "", "REMSEN", "ST"},
	{"abbrev-city", "60 OSBORN CT", "BK", "11225", "60", "", "OSBORN", "CT"},
	{"abbrev-city", "1500 TILDEN AVE", "BK", "11226", "1500", "", "TILDEN", "AVE"},
	{"abbrev-city", "812 FOX ST", "BX", "10459", "812", "", "FOX", "ST"},
	{"abbrev-city", "3010 BRIAR PL", "BX", "10461", "3010", "", "BRIAR", "PL"},
	{"abbrev-city", "255 MORRIS AVE", "BX", "10451", "255", "", "MORRIS", "AVE"},
	{"abbrev-city", "47 TIFFANY DR", "BX", "10474", "47", "", "TIFFANY", "DR"},
	{"abbrev-city", "1188 WESTCHESTER AVE", "BX", "10459", "1188", "", "WESTCHESTER", "AVE"},
	{"abbrev-city", "92 LORING AVE", "BKLYN", "11208", "92", "", "LORING", "AVE"},
	{"abbrev-city", "738 BEDFORD AVE", "BKLYN", "11205", "738", "", "BEDFORD", "AVE"},
	{"abbrev-city", "15 CLARK ST", "BKLYN", "11201", "15", "", "CLARK", "ST"},
	{"abbrev-city", "2410 NOSTRAND AVE", "BKLYN", "11210", "2410", "", "NOSTRAND", "AVE"},
	{"abbrev-city", "166 HOYT ST", "BKLYN", "11231", "166", "", "HOYT", "ST"},
	{"abbrev-city", "43 ROOSEVELT AVE", "FLUSHING", "11373", "43", "", "ROOSEVELT", "AVE"},
	{"abbrev-city", "19 PRINCE CT", "FLUSHING", "11368", "19", "", "PRINCE", "CT"},
	{"abbrev-city", "86 COLDEN ST", "FLUSHING", "11354", "86", "", "COLDEN", "ST"},
	{"abbrev-city", "301 MYRTLE AVE", "GLENDALE", "11373", "301", "", "MYRTLE", "AVE"},
	{"abbrev-city", "58 CYPRESS AVE", "GLENDALE", "11378", "58", "", "CYPRESS", "AVE"},
	{"abbrev-city", "7 ONDERDONK AVE", "GLENDALE", "11237", "7", "", "ONDERDONK", "AVE"},
	// Two spellings of the abbreviation the issue's table shows with the same
	// shape: a directional and a suffix on the same street line.
	{"abbrev-city", "221 E CARROLL ST", "BK", "11213", "221", "E", "CARROLL", "ST"},
	{"abbrev-city", "14 W LINDEN AVE", "NY", "10027", "14", "W", "LINDEN", "AVE"},
	// abbrev-city controls: same shape, a city that agrees with the ZIP Code
	// (or an abbreviation that is known to parse). These pass today.
	{"abbrev-city control", "227 ELM ST", "NEW YORK", "10002", "227", "", "ELM", "ST"},
	{"abbrev-city control", "227 ELM ST", "NY", "11201", "227", "", "ELM", "ST"},
	{"abbrev-city control", "2297 ASHLAND AVE", "BROOKLYN", "11207", "2297", "", "ASHLAND", "AVE"},
	{"abbrev-city control", "812 FOX ST", "BRONX", "10459", "812", "", "FOX", "ST"},
	{"abbrev-city control", "43 ROOSEVELT AVE", "FLUSHING", "11355", "43", "", "ROOSEVELT", "AVE"},
	{"abbrev-city", "301 MYRTLE AVE", "GLENDALE", "11385", "301", "", "MYRTLE", "AVE"}, // UNSURE: written as a control (GLENDALE with its own ZIP) but it fails today with city "GLENDALE NY"; confirm 11385 is a GLENDALE ZIP in the reference data

	// fused-suffix (#231): a numeric or ordinal street name fused to its
	// suffix.
	{"fused-suffix", "250 133RDST", "NEW YORK", "10027", "250", "", "133RD", "ST"},
	{"fused-suffix", "41 52NDST", "NEW YORK", "10019", "41", "", "52ND", "ST"},
	{"fused-suffix", "9 81STST", "NEW YORK", "10024", "9", "", "81ST", "ST"},
	{"fused-suffix", "630 194TH ST", "BRONX", "10458", "630", "", "194TH", "ST"},
	{"fused-suffix", "250 133ST", "NEW YORK", "10027", "250", "", "133RD", "ST"},    // UNSURE: name could stay 133 (issue: "133RD (or 133)")
	{"fused-suffix", "630 194ST", "BRONX", "10458", "630", "", "194TH", "ST"},       // UNSURE: ordinal form of 194, same question as 133ST
	{"fused-suffix", "12 30ST", "NEW YORK", "10016", "12", "", "30TH", "ST"},        // UNSURE: ordinal form of 30
	{"fused-suffix", "205 E 133ST", "NEW YORK", "10027", "205", "E", "133RD", "ST"}, // UNSURE: as 133ST
	{"fused-suffix", "18 W 104ST", "NEW YORK", "10025", "18", "W", "104TH", "ST"},   // UNSURE: as 133ST
	{"fused-suffix", "320 N 7ST", "BROOKLYN", "11211", "320", "N", "7TH", "ST"},     // UNSURE: as 133ST
	{"fused-suffix", "88 E 167ST", "BRONX", "10452", "88", "E", "167TH", "ST"},      // UNSURE: as 133ST
	{"fused-suffix", "2 W 222ST", "BRONX", "10463", "2", "W", "222ND", "ST"},        // UNSURE: as 133ST
	{"fused-suffix", "1501 3RDAVE", "NEW YORK", "10128", "1501", "", "3RD", "AVE"},  // UNSURE: fused AVE is not in #231's table, extrapolated
	// fused-suffix controls.
	{"fused-suffix control", "250 133RD ST", "NEW YORK", "10027", "250", "", "133RD", "ST"},
	{"fused-suffix control", "205 E 133RD ST", "NEW YORK", "10027", "205", "E", "133RD", "ST"},
	{"fused-suffix control", "630 194TH ST", "BRONX", "10458", "630", "", "194TH", "ST"},
	{"fused-suffix control", "12 30TH ST", "NEW YORK", "10016", "12", "", "30TH", "ST"},

	// digit-letter (#232): a leading token of digits and letters read whole
	// as the primary number.
	{"digit-letter", "400E 30ST", "NEW YORK", "10016", "400", "E", "30TH", "ST"},
	{"digit-letter", "400E 30 ST", "NEW YORK", "10016", "400", "E", "30", "ST"},
	{"digit-letter", "199HENRY ST", "BROOKLYN", "11201", "199", "", "HENRY", "ST"},
	{"digit-letter", "52W 71 ST", "NEW YORK", "10023", "52", "W", "71", "ST"}, // UNSURE: name 71 vs 71ST; the normalizer does not ordinalize a bare number today
	{"digit-letter", "310N 6TH ST", "BROOKLYN", "11211", "310", "N", "6TH", "ST"},
	{"digit-letter", "88S LAKE ST", "BROOKLYN", "11201", "88", "S", "LAKE", "ST"}, // UNSURE: S could be the start of a name; no data distinguishes it from a directional
	{"digit-letter", "27WILLOW ST", "BROOKLYN", "11201", "27", "", "WILLOW", "ST"},
	{"digit-letter", "1204PARK AVE", "NEW YORK", "10128", "1204", "", "PARK", "AVE"},
	{"digit-letter", "63MAPLE AVE", "BRONX", "10453", "63", "", "MAPLE", "AVE"},
	{"digit-letter", "905E 167 ST", "BRONX", "10459", "905", "E", "167", "ST"},   // UNSURE: name 167 vs 167TH, as 71
	{"digit-letter", "14W 120ST", "NEW YORK", "10027", "14", "W", "120TH", "ST"}, // UNSURE: combines digit-letter with the fused suffix (#231)
	{"digit-letter", "5E 4TH ST", "NEW YORK", "10003", "5", "E", "4TH", "ST"},
	// digit-letter controls: a spaced directional, and the legitimate
	// digits-plus-letter primary number that the issue says must survive.
	{"digit-letter control", "400 E 30 ST", "NEW YORK", "10016", "400", "E", "30", "ST"},
	{"digit-letter control", "199 HENRY ST", "BROOKLYN", "11201", "199", "", "HENRY", "ST"},
	{"digit-letter control", "123A MAIN ST", "BROOKLYN", "11201", "123A", "", "MAIN", "ST"},
	{"digit-letter control", "45B ELM ST", "BROOKLYN", "11201", "45B", "", "ELM", "ST"},
	{"digit-letter control", "52 W 71 ST", "NEW YORK", "10023", "52", "W", "71", "ST"},

	// fused-directional (#233): a directional fused to a street number.
	{"fused-directional", "60 E135TH ST", "NEW YORK", "10037", "60", "E", "135TH", "ST"},
	{"fused-directional", "70 E115 ST", "NEW YORK", "10029", "70", "E", "115TH", "ST"}, // UNSURE: 115TH vs 115 (issue: "115TH (or 115)"); inconsistent with the spaced control below, passes today
	{"fused-directional", "8 W72ND ST", "NEW YORK", "10023", "8", "W", "72ND", "ST"},
	{"fused-directional", "315 W125 ST", "NEW YORK", "10027", "315", "W", "125TH", "ST"},
	{"fused-directional", "41 E87TH ST", "NEW YORK", "10128", "41", "E", "87TH", "ST"},
	{"fused-directional", "260 N6TH ST", "BROOKLYN", "11211", "260", "N", "6TH", "ST"},
	{"fused-directional", "19 S1ST ST", "BROOKLYN", "11211", "19", "S", "1ST", "ST"},
	{"fused-directional", "712 E161 ST", "BRONX", "10451", "712", "E", "161ST", "ST"},      // UNSURE: 161 ordinalizes to 161ST
	{"fused-directional", "33 W231 ST", "BRONX", "10463", "33", "W", "231ST", "ST"},        // UNSURE: 231 ordinalizes to 231ST
	{"fused-directional", "6 E135TH STREET", "NEW YORK", "10037", "6", "E", "135TH", "ST"}, // UNSURE: written-out suffix on a fused directional, extrapolated
	{"fused-directional", "52 E135", "NEW YORK", "10037", "52", "E", "135TH", ""},          // UNSURE: no suffix at all; may be a bare name, may be an error
	{"fused-directional", "10 E135ST", "NEW YORK", "10037", "10", "E", "135TH", "ST"},      // UNSURE: combines #233 with the fused suffix (#231)
	// fused-directional controls.
	{"fused-directional control", "60 E 135TH ST", "NEW YORK", "10037", "60", "E", "135TH", "ST"},
	{"fused-directional control", "70 E 115 ST", "NEW YORK", "10029", "70", "E", "115", "ST"}, // UNSURE: the normalizer does not ordinalize a bare number (cf. "E 30 ST"), unlike E115 ST above, which today yields 115TH
	{"fused-directional control", "8 W 72ND ST", "NEW YORK", "10023", "8", "W", "72ND", "ST"},
	{"fused-directional control", "260 N 6TH ST", "BROOKLYN", "11211", "260", "N", "6TH", "ST"},
	{"fused-directional control", "55 EAST ST", "BROOKLYN", "11222", "55", "", "EAST", "ST"}, // UNSURE: EAST ST is the directional-vs-name ambiguity of #110; included as a negative control against over-eager splitting
}

func buildHteCases() []Case {
	cases := make([]Case, 0, len(hteRows))
	for _, r := range hteRows {
		street := strings.Join(nonEmpty(r.num, r.pre, r.name, r.suf), " ")
		cases = append(cases, Case{
			Source: "hte",
			Note:   "cms-hte L3: " + r.class,
			Input:  r.line + "\n" + r.city + ", NY " + r.zip,
			Want:   street + "\n" + r.city + " NY " + r.zip,
			WantFields: &address.Address{
				Type:          &ordinarystreet.OrdinaryStreetAddress{},
				PrimaryNumber: r.num, Predirectional: r.pre,
				StreetName: r.name, StreetSuffix: r.suf,
				City: r.city, Region: "NY", Postal: r.zip,
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
