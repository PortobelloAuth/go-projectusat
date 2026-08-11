package goprojectusat_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/libpostalhttp"
)

type NormalizeTestCase struct {
	in, want, group string
}

// cSharp parity test cases derrived from https://github.com/ica-carealign/project-us-normalizer.
// https://github.com/ica-carealign/project-us-normalizer/blob/00e40d7c3ee3145655d14c4262a37640446ab491/src/ProjectUsNormalizer.csproj#L17
// indicates that https://github.com/ica-carealign/project-us-normalizer is MIT licensed.
var csharpParityCases = []NormalizeTestCase{
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

var gridCases = []NormalizeTestCase{
	// Post-directional followed by a City with a directional prefix
	{"43 E 200 N, NORTH SALT LAKE, UT", "43 E 200 N NORTH SALT LAKE UT", "directional city"},
	{"43 E 200 N NORTH SALT LAKE UT", "43 E 200 N NORTH SALT LAKE UT", "directional city"},
	{"3253 W 9200 S, West Jordan, UT 84088", "3253 W 9200 S WEST JORDAN UT 84088", "directional city"},
}

// TODO: make sure we have multiline versions of all these test addresses and results
var saintCases = []NormalizeTestCase{
	{"915 2ND ST N SAINT CLOUD MN 56301", "915 2ND ST N SAINT CLOUD MN 56301", "city"},
	{"915 2ND ST N ST CLOUD MN 56301", "915 2ND ST N SAINT CLOUD MN 56301", "city"},
	{"435 S SAINT CLAIR ST TOLEDO OH 43601", "435 S SAINT CLAIR ST TOLEDO OH 43601", "street"},
	{"435 S ST CLAIR ST TOLEDO OH 43601", "435 S SAINT CLAIR ST TOLEDO OH 43601", "street"},
}

var cases = slices.Collect(func(yield func(NormalizeTestCase) bool) {
	for _, v := range csharpParityCases {
		if !yield(NormalizeTestCase{in: v.in, want: v.want, group: fmt.Sprintf("parity - %s", v.group)}) {
			return
		}
	}

	for _, v := range gridCases {
		if !yield(NormalizeTestCase{in: v.in, want: v.want, group: fmt.Sprintf("grid - %s", v.group)}) {
			return
		}
	}

	for _, v := range saintCases {
		if !yield(NormalizeTestCase{in: v.in, want: v.want, group: fmt.Sprintf("saint - %s", v.group)}) {
			return
		}
	}
})

func TestNormalizeWithCustomParser(t *testing.T) {
	lph, err := libpostalhttp.NewService("http://127.0.0.1:4400/", 30*time.Millisecond)
	if err != nil {
		t.Fatalf("Unable to create libpostalhttp parser: %v", err)
	}
	for _, tc := range cases {
		got, err := goprojectusat.Normalize(
			tc.in,
			goprojectusat.WithCustomAddressParser(parser.ParsingFn(lph.Parse)),
		)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}

		if got != tc.want {
			t.Fatalf("Format(Normalize(...)) = %q, want %q, %s", got, tc.want, tc.group)
		}
	}
}

func TestNormalize(t *testing.T) {
	for _, tc := range cases {
		got, err := goprojectusat.Normalize(tc.in)
		if err != nil {
			t.Fatalf("Normalize: %v", err)
		}

		if got != tc.want {
			t.Fatalf("Format(Normalize(...)) = %q, want %q, %s", got, tc.want, tc.group)
		}
	}
}
