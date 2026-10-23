package goprojectusat_test

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizeroptions"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
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
	{"3M Corporation 100 Main Street", "3M CORP 100 MAIN ST", "biz"},
	{"UCENT Building 847 North 49th Street", "UCENT BLDG 847 N 49TH ST", "biz"},
	{"UCENT Building Suite 480 411 N Central Ave", "UCENT BLDG 411 N CENTRAL AVE STE 480", "biz"},
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

var otherCases = []NormalizeTestCase{
	{"ACME SUITE 3200 152 South Tech Dr\nPhoenix, Arizona 85001", "ACME\n152 S TECH DR STE 3200\nPHOENIX AZ 85001", "other"},
	{"84 Lumber Suite 3200 152 South Tech Dr\nPhoenix, Arizona 85001", "84 LUMBER\n152 S TECH DR STE 3200\nPHOENIX AZ 85001", "other"},
	{"BLDG 4 STE 200 1403 MAIN ST OGDEN UT 84401", "1403 MAIN ST BLDG 4 STE 200\nOGDEN UT 84401", "other"},
	{"84 LUMBER BLDG 4 STE 200 1403 MAIN ST OGDEN UT 84401", "84 LUMBER\n1403 MAIN ST BLDG 4 STE 200\nOGDEN UT 84401", "other"},
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

	for _, v := range otherCases {
		if !yield(NormalizeTestCase{in: v.in, want: v.want, group: fmt.Sprintf("other - %s", v.group)}) {
			return
		}
	}
})

func TestNormalize(t *testing.T) {
	usat, err := goprojectusat.New()
	if err != nil {
		t.Fatalf("New() error: %s", err)
	}
	singlelineusat, err := goprojectusat.New(goprojectusat.WithSingleLineFormatting())
	if err != nil {
		t.Fatalf("New(WithSingleLineFormatting()) error: %s", err)
	}
	for _, tc := range cases {
		// NOTE: we are using goprojectusat.WithSingleLineFormatting() because several of
		// the test cases are written with single line formatted results and changing that
		// is its own task.
		u := singlelineusat
		if strings.Contains(tc.want, "\n") {
			u = usat
		}
		got, err := u.Normalize(tc.in)
		if err != nil {
			t.Errorf("USAt.Normalize() - %q error: %s", tc.in, err)
		}

		if got != tc.want {
			t.Errorf("USAt.Normalize() = %q, want %q, %s", got, tc.want, tc.group)
		}
	}
}

// legacyNormalize is the body of the package-level Normalize as it was before
// USAt existed (go-projectusat#195), kept verbatim so the tests below compare
// USAt against the old behaviour rather than against itself.
func legacyNormalize(source string, opts ...goprojectusat.USAtNormalizeOption) (string, error) {
	popts := &parser.AddressParsingOptions{}
	nopts := &normalizeroptions.AddressNormalizationOptions{}
	fopts := &address.FormatOptions{}
	for _, fn := range opts {
		err := fn(popts, nopts, fopts)
		if err != nil {
			return "", fmt.Errorf("Error setting normailzation options: %w", err)
		}
	}
	p, err := parser.New(*popts)
	if err != nil {
		return "", fmt.Errorf("Unable to instantiate parser: %w", err)
	}
	addr, err := p.Parse(source)
	if err != nil {
		return "", fmt.Errorf("Unable to parse address: %w", err)
	}
	n := normalizer.NewNomalizer(*nopts)
	addr, err = n.Normalize(addr)
	if err != nil {
		return "", fmt.Errorf("Unable to normalize address: %w", err)
	}
	return addr.Format(*fopts)
}

var optionSets = map[string][]goprojectusat.Option{
	"none":                  nil,
	"content":               {goprojectusat.WithContentNormalization()},
	"matching":              {goprojectusat.WithMatchingNormalization()},
	"single-line":           {goprojectusat.WithSingleLineFormatting()},
	"content+single-line":   {goprojectusat.WithContentNormalization(), goprojectusat.WithSingleLineFormatting()},
	"matching+single-line":  {goprojectusat.WithMatchingNormalization(), goprojectusat.WithSingleLineFormatting()},
	"fuzzy+hash":            {goprojectusat.WithFuzzyNormalization(), goprojectusat.WithSecondaryAsHash()},
	"substitute-diacritics": {goprojectusat.WithDiacriticNormalization(diacritics.SubstituteDiacritics), goprojectusat.WithDiacriticFormatting(diacritics.SubstituteDiacritics)},
	"matching-then-content": {goprojectusat.WithMatchingNormalization(), goprojectusat.WithContentNormalization()},
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TestUSAtMatchesLegacyNormalize runs every parsertest case through the
// pre-#195 pipeline, a USAt, and the package-level Normalize, under each
// option set, and requires identical output and identical errors.
func TestUSAtMatchesLegacyNormalize(t *testing.T) {
	for name, opts := range optionSets {
		t.Run(name, func(t *testing.T) {
			u, err := goprojectusat.New(opts...)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			for _, c := range parsertest.Cases {
				want, wantErr := legacyNormalize(c.Input, opts...)
				got, gotErr := u.Normalize(c.Input)
				if got != want || errString(gotErr) != errString(wantErr) {
					t.Errorf("USAt.Normalize(%q)\n got  %q, %v\n want %q, %v", c.Input, got, gotErr, want, wantErr)
				}
			}
		})
	}
}

// TestUSAtNormalizeAddressMatchesNormalize checks that NormalizeAddress
// (go-projectusat#206), applied after Parse, agrees with the combined
// Normalize on the same input and options — it is the same normalizer, only
// reachable a part at a time.
func TestUSAtNormalizeAddressMatchesNormalize(t *testing.T) {
	u, err := goprojectusat.New(goprojectusat.WithFuzzyNormalization())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, c := range parsertest.Cases {
		want, wantErr := u.Normalize(c.Input)

		parsed, parseErr := u.Parse(c.Input)
		if errString(parseErr) != errString(wantErr) && parseErr != nil {
			continue // Normalize's error came from normalization, not parsing; not this test's case.
		}
		if parseErr != nil {
			if errString(parseErr) != errString(wantErr) {
				t.Errorf("Parse(%q) err = %v, want %v", c.Input, parseErr, wantErr)
			}
			continue
		}

		normalized, normErr := u.NormalizeAddress(parsed)
		got := ""
		gotErr := normErr
		if normErr == nil {
			got, gotErr = normalized.Format()
		}
		if got != want || errString(gotErr) != errString(wantErr) {
			t.Errorf("Parse+NormalizeAddress(%q)\n got  %q, %v\n want %q, %v", c.Input, got, gotErr, want, wantErr)
		}
	}
}

// TestUSAtNormalizeGluedPunctuationDesignator is Aaron's go-projectusat#207
// repro: "Apt." keeps its period glued to the word (the tokenizer only drops
// a token that is punctuation alone), and that period must not stop the
// designator from being recognized. The Puerto Rico diacritics case from his
// comment is excluded here — that one is #205's open architectural question,
// not this bug.
func TestUSAtNormalizeGluedPunctuationDesignator(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{
			input:  "123 Main Street Apt 23\nNorfolk, Virginia 23510",
			expect: "123 MAIN ST # 23\nNORFOLK VA 23510",
		},
		{
			input:  "123 Main Street Apt. 23\nNorfolk, Virginia 23510",
			expect: "123 MAIN ST # 23\nNORFOLK VA 23510",
		},
	}

	usat, err := goprojectusat.New(
		goprojectusat.WithDiacriticNormalization(diacritics.SubstituteDiacritics),
		goprojectusat.WithSecondaryAsHash(),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, c := range cases {
		got, err := usat.Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) error: %v", c.input, err)
			continue
		}
		if got != c.expect {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.expect)
		}
	}
}

// TestUSAtConcurrentUse shares one USAt across goroutines; run with -race.
func TestUSAtConcurrentUse(t *testing.T) {
	u, err := goprojectusat.New(goprojectusat.WithMatchingNormalization())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	want := make([]string, len(parsertest.Cases))
	for i, c := range parsertest.Cases {
		want[i], _ = u.Normalize(c.Input)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for i, c := range parsertest.Cases {
				if got, _ := u.Normalize(c.Input); got != want[i] {
					t.Errorf("concurrent Normalize(%q) = %q, want %q", c.Input, got, want[i])
				}
			}
		})
	}
	wg.Wait()
}

func ExampleNew() {
	usat, err := goprojectusat.New(goprojectusat.WithMatchingNormalization(), goprojectusat.WithSingleLineFormatting())
	if err != nil {
		panic(err)
	}
	for _, in := range []string{
		"152 South Tech Dr Apartment 3200, Salt Lake City, UT 84101",
		"PO Box 11890, Phoenix, Arizona 85001",
	} {
		out, err := usat.Normalize(in)
		if err != nil {
			panic(err)
		}
		fmt.Println(out)
	}
	// Output:
	// 152 S TECH DR # 3200 SALT LAKE CITY UT 84101
	// PO BOX 11890 PHOENIX AZ 85001
}

func ExampleUSAt_Parse() {
	usat, err := goprojectusat.New()
	if err != nil {
		panic(err)
	}
	addr, err := usat.Parse("Acme Widgets, 152 South Tech Drive Suite 3200, Salt Lake City, Utah 84101")
	if err != nil {
		panic(err)
	}
	fmt.Println("BusinessName:", addr.BusinessName)
	fmt.Println("PrimaryNumber:", addr.PrimaryNumber)
	fmt.Println("Predirectional:", addr.Predirectional)
	fmt.Println("StreetName:", addr.StreetName)
	fmt.Println("StreetSuffix:", addr.StreetSuffix)
	fmt.Println("SecondaryDesignator:", addr.SecondaryDesignator)
	fmt.Println("SecondaryNumber:", addr.SecondaryNumber)
	fmt.Println("City:", addr.City)
	fmt.Println("Region:", addr.Region)
	fmt.Println("Postal:", addr.Postal)

	// Output:
	// BusinessName: ACME WIDGETS
	// PrimaryNumber: 152
	// Predirectional: S
	// StreetName: TECH
	// StreetSuffix: DR
	// SecondaryDesignator: STE
	// SecondaryNumber: 3200
	// City: SALT LAKE CITY
	// Region: UT
	// Postal: 84101
}

// Parse and NormalizeAddress together are Normalize without the formatting,
// leaving the address in fields. Here WithMatchingNormalization is applied by
// NormalizeAddress, not by Parse, and replaces the designator with #.
func ExampleUSAt_NormalizeAddress() {
	usat, err := goprojectusat.New(goprojectusat.WithMatchingNormalization())
	if err != nil {
		panic(err)
	}
	parsed, err := usat.Parse("152 South Tech Drive Suite 3200, Salt Lake City, UT 84101")
	if err != nil {
		panic(err)
	}
	normalized, err := usat.NormalizeAddress(parsed)
	if err != nil {
		panic(err)
	}
	fmt.Println("parsed:", parsed.SecondaryDesignator, parsed.SecondaryNumber)
	fmt.Println("normalized:", normalized.SecondaryDesignator, normalized.SecondaryNumber)

	// Output:
	// parsed: STE 3200
	// normalized: # 3200
}

// An address need not come from Parse. One that holds only some of its
// fields normalizes just those.
func ExampleUSAt_NormalizeAddress_fields() {
	usat, err := goprojectusat.New()
	if err != nil {
		panic(err)
	}
	normalized, err := usat.NormalizeAddress(&address.Address{City: "salt lake city", Region: "utah"})
	if err != nil {
		panic(err)
	}
	fmt.Println(normalized.City)
	fmt.Println(normalized.Region)

	// Output:
	// SALT LAKE CITY
	// UT
}
