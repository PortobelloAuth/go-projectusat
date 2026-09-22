package parsertest

import "fmt"

// historicalCase is one row of the parity suite go-projectusat carried in
// its own goprojectusat_test.go from fbaa970 (the custom-parser option)
// onward, run through every custom parser plugged into Normalize until
// b2c9194 cut that back to one stub case so this repository would not need
// a live libpostal service. group is the label the original table used.
type historicalCase struct {
	in, want, group string
}

// csharpParity is derived from https://github.com/ica-carealign/project-us-normalizer
// (MIT licensed, per that project's ProjectUsNormalizer.csproj at 00e40d7).
var csharpParity = []historicalCase{
	{"1011 South West Main Thing St North East Apt 12", "1011 SW MAIN THING ST NE APT 12", "dir"},
	{"3002 NORTH EAST MAIN STREET", "3002 NE MAIN ST", "dir"},
	{"3009 NORTHEAST MAIN STREET", "3009 NE MAIN ST", "dir"},
	{"3402 MAIN STREET NORTH EAST", "3402 MAIN ST NE", "dir"},
	{"1016 East 1700 South", "1016 E 1700 S", "grid"},
	{"1005 south ave east", "1005 SOUTH AVE E", "dir-name"},
	{"1001 AVE E", "1001 AVENUE E", "dir-name"},
	{"1014 BAY W DRIVE", "1014 BAY WEST DR", "dir-name"},
	{"1015 NORTH AVENUE", "1015 NORTH AVE", "dir-name"},
	{"2000 main avenue drive", "2000 MAIN AVENUE DR", "dbl-suf"},
	{"2002 Main Pky Ave", "2002 MAIN PARKWAY AVE", "dbl-suf"},
	{"2009 Church Court Way", "2009 CHURCH COURT WAY", "dbl-suf"},
	{"8000 OK avenue", "8000 OKLAHOMA AVE", "state"},
	{"8004 CT Drive", "8004 CONNECTICUT DR", "state"},
	{"8006 CT CT", "8006 CONNECTICUT CT", "state"},
	{"8011 WY WY", "8011 WYOMING WAY", "state"},
	{"8100 Montana Treasure Avenue", "8100 MT TREASURE AVE", "state-part"},
	{"8103 South Carolina county road 22", "8103 SC COUNTY ROAD 22", "state-part"},
	{"8105 TN 431", "8105 TN HIGHWAY 431", "hwy"},
	{"8007 EAST KENTUCKY KEY", "8007 E KENTUCKY KY", "state"},
	{"9013 I 10", "9013 INTERSTATE 10", "hwy"},
	{"9020 US 41", "9020 US HIGHWAY 41", "hwy"},
	{"9038 SR 220", "9038 STATE ROAD 220", "hwy"},
	{"9047 RT 88", "9047 ROUTE 88", "hwy"},
	{"9052 I10", "9052 INTERSTATE 10", "hwy"},
	{"9062 farm to market 1200", "9062 FM 1200", "hwy"},
	{"9004 CR 20 NE", "9004 COUNTY ROAD 20 NE", "hwy"},
	{"9011 HWY 66 FRONTAGE ROAD", "9011 HIGHWAY 66 FRONTAGE RD", "hwy"},
	{"9028 rd 39.4", "9028 ROAD 39.4", "grid"},
	{"Post office Box G", "PO BOX G", "po"},
	{"PO Box 11890", "PO BOX 11890", "po"},
	{"POB 11890", "PO BOX 11890", "po"},
	{"Rural Route 91 Box A7", "RR 91 BOX A7", "rr"},
	{"RFD 61 #87b", "RR 61 BOX 87B", "rr"},
	{"RFD Route 61 Box 87b", "RR 61 BOX 87B", "rr"},
	{"RR0061 #87b", "RR 61 BOX 87B", "rr"},
	{"RR0061#87b", "RR 61 BOX 87B", "rr"},
	{"152 South Tech Dr Apartment 3200", "152 S TECH DR APT 3200", "sec"},
	{"Apartment 3200 152 South Tech Dr", "152 S TECH DR APT 3200", "sec"},
	{"#3200 South Tech Dr", "S TECH DR # 3200", "sec"},
	{"#3200 152 South Tech Dr", "152 S TECH DR # 3200", "sec"},
	{"Unit 3200 152 Tech Dr Room 12", "152 TECH DR UNIT 3200 RM 12", "sec"},
	{"Unit 3200 152 Tech Dr Upper", "152 TECH DR UNIT 3200 UPPR", "sec"},
	{"450 Jane Stanford Way Building 420 Room 120", "450 JANE STANFORD WAY BLDG 420 RM 120", "sec"},
	{"100 Main Street Southwest # 12", "100 MAIN ST SW # 12", "sec"},
	{"Williamson Medical Center 3000 Edward Curd Lane", "WILLIAMSON MEDICAL CENTER 3000 EDWARD CURD LN", "biz"},
	{"Center of Hope 110 East 7th Street", "CENTER OF HOPE 110 E 7TH ST", "biz"},
	{"3M Corporation 100 Main Street", "3M CORPORATION 100 MAIN ST", "biz"},
	{"UCENT Building 847 North 49th Street", "UCENT BUILDING 847 N 49TH ST", "biz"},
	{"UCENT Building Suite 480 411 N Central Ave", "UCENT BUILDING 411 N CENTRAL AVE STE 480", "biz"},
	{"4000 12TH. Street", "4000 12TH ST", "clean"},
	{"4007 West Main' rd", "4007 W MAIN RD", "clean"},
	{"4008 @ West Main STREET", "4008 W MAIN ST", "clean"},
	{"3005 N.E. MAIN STREET", "3005 NE MAIN ST", "dir"},
	{"3010 NORTH-EAST MAIN STREET", "3010 NE MAIN ST", "dir"},
}

var gridHistorical = []historicalCase{
	// Post-directional followed by a City with a directional prefix
	{"43 E 200 N, NORTH SALT LAKE, UT", "43 E 200 N NORTH SALT LAKE UT", "directional city"},
	{"43 E 200 N NORTH SALT LAKE UT", "43 E 200 N NORTH SALT LAKE UT", "directional city"},
	{"3253 W 9200 S, West Jordan, UT 84088", "3253 W 9200 S WEST JORDAN UT 84088", "directional city"},
}

var saintHistorical = []historicalCase{
	{"915 2ND ST N SAINT CLOUD MN 56301", "915 2ND ST N SAINT CLOUD MN 56301", "city"},
	{"915 2ND ST N ST CLOUD MN 56301", "915 2ND ST N SAINT CLOUD MN 56301", "city"},
	{"435 S SAINT CLAIR ST TOLEDO OH 43601", "435 S SAINT CLAIR ST TOLEDO OH 43601", "street"},
	{"435 S ST CLAIR ST TOLEDO OH 43601", "435 S SAINT CLAIR ST TOLEDO OH 43601", "street"},
}

// HistoricalCases is go-projectusat's own pre-parsertest parity suite: the
// cases the project used, through the Normalize API, to show what a plugged-
// in parser did and did not get right before this package existed.
var HistoricalCases = buildHistoricalCases()

// buildHistoricalCases turns each historicalCase into a Case, completing the
// csharpParity rows with a last line the way buildSpecCases completes a spec
// fragment: those rows are street lines on their own, and a parser that
// admits no address type without a city, region and ZIP Code reads none of
// them. That one fact, not 62 separate bugs, is why HistoricalCases scored
// nothing at all before amadsen raised it on #112.
//
// The last line is a property of the set rather than of the row, and it is
// deliberately not derived from hasLastLine: that predicate wants a
// five-digit ZIP Code, and a gridHistorical row such as
// "43 E 200 N, NORTH SALT LAKE, UT" has a city and a region without one, so
// asking hasLastLine would give it a second last line.
func buildHistoricalCases() []Case {
	groups := []struct {
		set      []historicalCase
		lastLine string
	}{
		{csharpParity, mainlandLastLine},
		{gridHistorical, ""},
		{saintHistorical, ""},
	}

	var cases []Case
	for _, g := range groups {
		for _, hc := range g.set {
			cases = append(cases, Case{
				Source: "go-projectusat",
				Note:   fmt.Sprintf("parity - %s", hc.group),
				Input:  hc.in + g.lastLine,
				Want:   hc.want + g.lastLine,
			})
		}
	}
	return cases
}
