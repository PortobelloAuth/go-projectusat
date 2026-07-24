package goprojectusat_test

import (
	"fmt"
	"testing"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
)

var csharpParityCases = []struct {
	in, want, group string
}{
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

func streetOut(t *testing.T, in string) (string, error) {
	t.Helper()
	p, err := goprojectusat.Parse(in + "\nSpringfield IL 62701")
	if err != nil {
		return "", err
	}
	n, err := goprojectusat.Normalize(p)
	if err != nil {
		return "", err
	}
	if n.BusinessName != "" {
		return n.BusinessName + " " + goprojectusat.FormatStreetLine(n), nil
	}
	return goprojectusat.FormatStreetLine(n), nil
}

func TestCSharpParityScore(t *testing.T) {
	ok, miss, errn := 0, 0, 0
	byGroup := map[string][3]int{}
	for _, tc := range csharpParityCases {
		g := byGroup[tc.group]
		got, err := streetOut(t, tc.in)
		if err != nil {
			errn++
			g[2]++
			byGroup[tc.group] = g
			t.Logf("ERR  [%s] %q → %v", tc.group, tc.in, err)
			continue
		}
		if got == tc.want {
			ok++
			g[0]++
		} else {
			miss++
			g[1]++
			t.Logf("MISS [%s] %q\n  got  %q\n  want %q", tc.group, tc.in, got, tc.want)
		}
		byGroup[tc.group] = g
	}
	total := ok + miss + errn
	pct := 100.0 * float64(ok) / float64(total)
	fmt.Printf("\n=== C# STREET PARITY SCORE ===\nOK=%d MISS=%d ERR=%d TOTAL=%d  →  %.1f%%\n", ok, miss, errn, total, pct)
	for _, g := range []string{"dir", "dir-name", "grid", "dbl-suf", "state", "state-part", "hwy", "po", "rr", "sec", "biz", "clean"} {
		if v, has := byGroup[g]; has {
			fmt.Printf("  %-12s %d/%d/%d\n", g, v[0], v[1], v[2])
		}
	}
	if miss > 0 || errn > 0 {
		t.Fatalf("parity not complete: OK=%d MISS=%d ERR=%d (%.1f%%)", ok, miss, errn, pct)
	}
}
