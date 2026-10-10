package parse_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/zipcityembedded/parse"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ordinarystreet"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
	"github.com/poetic-systems/zipcity"
)

// Every address below is either published in the Project US@ specification or
// invented. None is a real person's address.

// The parser must be usable through go-projectusat's own seam. This is the
// whole point of the package, so it is asserted at compile time rather than
// left to a caller to discover.
var _ parser.ParsingFunc = (*parse.Parser)(nil)

func TestItReadsTheSpecialAddressFormats(t *testing.T) {
	cases := []struct {
		name       string
		source     string
		streetName string
		primary    string
		city       string
		region     string
		postal     string
	}{
		{
			name:       "post office box",
			source:     "PO BOX 11890\nWEST JORDAN UT 84088",
			streetName: "PO BOX", primary: "11890",
			city: "WEST JORDAN", region: "UT", postal: "84088",
		},
		{
			name:       "rural route",
			source:     "RR 4 BOX 125\nWEST JORDAN UT 84088",
			streetName: "RR 4", primary: "BOX 125",
			city: "WEST JORDAN", region: "UT", postal: "84088",
		},
		{
			name:       "military",
			source:     "PSC 3 BOX 4120\nAPO AE 09021",
			streetName: "PSC 3", primary: "BOX 4120",
			city: "APO", region: "AE", postal: "09021",
		},
	}

	p := parse.New()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, err := p.Parse(c.source)
			if err != nil {
				t.Fatalf("parsing: %v", err)
			}
			if a.StreetName != c.streetName || a.PrimaryNumber != c.primary {
				t.Errorf("street line = %q %q, want %q %q",
					a.StreetName, a.PrimaryNumber, c.streetName, c.primary)
			}
			if a.City != c.city || a.Region != c.region || a.Postal != c.postal {
				t.Errorf("last line = %q %q %q, want %q %q %q",
					a.City, a.Region, a.Postal, c.city, c.region, c.postal)
			}
			if a.Type == nil {
				t.Error("the winning candidate must carry the address type that read it")
			}
		})
	}
}

// The ordinary street line is assembled from the shared vocabularies rather
// than recognized, so what matters is that every element comes out in its
// place: the formatted line is the whole decomposition in one string.
func TestItReadsTheOrdinaryStreetLine(t *testing.T) {
	cases := []struct {
		source string
		street string
	}{
		{"123 MAIN ST\nWEST JORDAN UT 84088", "123 MAIN ST"},
		{"123 N MAIN ST APT 4\nWEST JORDAN UT 84088", "123 N MAIN ST APT 4"},
		{"1600 PENNSYLVANIA AVE NW\nWASHINGTON DC 20500", "1600 PENNSYLVANIA AVE NW"},
		{"GENERAL DELIVERY\nFAIRHAVEN MA 02719", "GENERAL DELIVERY"},
		// A private mailbox only comes out if privatemailbox is among the
		// vocabularies consulted; without its Detail claim the line reads
		// as a street named MAIN STREET PMB 4545.
		{"123 MAIN STREET PMB 4545\nHERNDON VA 22071", "123 MAIN ST PMB 4545"},
		{"PMB 234\n123 MAIN ST\nHERNDON VA 22071", "123 MAIN ST PMB 234"},
	}

	p := parse.New()
	for _, c := range cases {
		t.Run(c.street, func(t *testing.T) {
			a, err := p.Parse(c.source)
			if err != nil {
				t.Fatalf("parsing: %v", err)
			}
			if got := a.FormatStreetLine(); got != c.street {
				t.Errorf("street line = %q, want %q", got, c.street)
			}
		})
	}
}

func TestEmptyInputHasNoReading(t *testing.T) {
	p := parse.New()

	for _, source := range []string{"", "   ", "\n"} {
		if _, err := p.Parse(source); !errors.Is(err, parse.ErrNoReading) {
			t.Errorf("source %q: want ErrNoReading, got %v", source, err)
		}
	}
}

// Reference data demotes a reading; it never discards one. A ZIP and city that
// the data does not pair still parse, because zipcity's gaps are real and a
// patient at an address the Census missed must not be dropped.
func TestReferenceDataDemotesButDoesNotReject(t *testing.T) {
	source := "PO BOX 11890\nNOT A REAL MUNICIPALITY UT 84088"

	withData := parse.New()
	a, err := withData.Parse(source)
	if err != nil {
		t.Fatalf("a contradicted reading must still parse: %v", err)
	}
	if a.StreetName != "PO BOX" || a.PrimaryNumber != "11890" {
		t.Errorf("street line = %q %q, want %q %q",
			a.StreetName, a.PrimaryNumber, "PO BOX", "11890")
	}
}

func TestTheErrorCarriesNoPartOfTheInput(t *testing.T) {
	p := parse.New()

	// No last line, so no address type can read it.
	source := "123 INVENTED LANE\nNOT A REAL MUNICIPALITY"
	_, err := p.Parse(source)
	if err == nil {
		t.Fatal("want a rejection")
	}
	for _, part := range []string{"INVENTED", "NOT A REAL MUNICIPALITY"} {
		if strings.Contains(err.Error(), part) {
			t.Errorf("the error text leaks %q: %v", part, err)
		}
	}
}

// The case #7 was opened for: a reading rated Exact that the data contradicts
// must not beat a reading rated Strong that the data agrees with. The old
// scale stepped Confidence itself and stopped short of Exact, so the two tied
// at Strong and the grammar-rating tie-break handed it to the reading the data
// had just spoken against. Uncapped, Exact contradicted is 3-1=2 and Strong
// agreed is 2+1=3.
func TestAgreementLiftsAPresentReadingPastAnAbsentOne(t *testing.T) {
	exactContradicted := parse.Score(claim.ConfidenceExact, []parse.Agreement{parse.Contradicts})
	strongAgreed := parse.Score(claim.ConfidenceStrong, []parse.Agreement{parse.Agrees})
	if exactContradicted >= strongAgreed {
		t.Errorf("Score(Exact, Contradicts) = %d, Score(Strong, Agrees) = %d; want the agreed reading ahead",
			exactContradicted, strongAgreed)
	}
}

// A single agreement moves a reading by one rung, which is enough to pass a
// reading one rung above it that the data spoke against — that inversion is
// the whole point of asking — but not enough to pass a reading two rungs
// above it that nothing was asked about at all. The data orders readings
// against each other; it does not manufacture a gap the grammar never gave.
func TestAgreementCannotCarryAReadingAcrossATwoRungGap(t *testing.T) {
	likelyAgreed := parse.Score(claim.ConfidenceLikely, []parse.Agreement{parse.Agrees})
	exactUntouched := parse.Score(claim.ConfidenceExact, nil)
	if likelyAgreed >= exactUntouched {
		t.Errorf("Score(Likely, Agrees) = %d, Score(Exact, nil) = %d; want the untouched exact reading ahead",
			likelyAgreed, exactUntouched)
	}
}

// Two readings of one street line that share a zipcity key ask it once and
// get back the same answer, so agreement shifts both of their scores by the
// same amount. Whatever gap the grammar put between two rungs, applying the
// same answers to both must leave the sign of that gap alone — that is the
// fact that lets choose retire the grammar-rating tie-break and rely on the
// score's own ordering instead.
func TestSharedAnswersPreserveTheGrammarsGap(t *testing.T) {
	rungs := []claim.Confidence{
		claim.ConfidenceWeak, claim.ConfidenceLikely, claim.ConfidenceStrong, claim.ConfidenceExact,
	}
	for _, answers := range [][]parse.Agreement{{parse.Agrees}, {parse.Contradicts}} {
		for _, lo := range rungs {
			for _, hi := range rungs {
				if parse.Rung(lo) >= parse.Rung(hi) {
					continue
				}
				loScore := parse.Score(lo, answers)
				hiScore := parse.Score(hi, answers)
				if loScore >= hiScore {
					t.Errorf("answers %v: Score(%d) = %d, Score(%d) = %d; want the higher rung still ahead",
						answers, lo, loScore, hi, hiScore)
				}
			}
		}
	}
}

// missingInZip, missingInCity, and unknown are not evidence — see agreement's
// doc comment for why a split street answer or a declined question must not
// read as either an agreement or a contradiction — so none of them may move
// a reading off the rung the grammar gave it.
func TestSplitsAndDeclinesAddNothing(t *testing.T) {
	rungs := []claim.Confidence{
		claim.ConfidenceWeak, claim.ConfidenceLikely, claim.ConfidenceStrong, claim.ConfidenceExact,
	}
	for _, ans := range []parse.Agreement{parse.MissingInZip, parse.MissingInCity, parse.Unknown} {
		for _, c := range rungs {
			if got, want := parse.Score(c, []parse.Agreement{ans}), parse.Rung(c); got != want {
				t.Errorf("Score(%d, %v) = %d, want %d (the bare rung)", c, ans, got, want)
			}
		}
	}
}

// choose builds one reference per call so that two readings asking the same
// key cost zipcity a single query. A query that keeps a counter is the only
// way to see that from outside the package.
func TestReferenceAsksEachKeyOnce(t *testing.T) {
	r := parse.Reference{}
	calls := 0
	query := func() (bool, error) {
		calls++
		return true, nil
	}

	first, err := parse.Check(r, "key", query)
	if err != nil {
		t.Fatalf("first check: %v", err)
	}
	second, err := parse.Check(r, "key", query)
	if err != nil {
		t.Fatalf("second check: %v", err)
	}
	if calls != 1 {
		t.Errorf("query ran %d times, want 1", calls)
	}
	if first != second {
		t.Errorf("first = %v, second = %v, want the cached answer both times", first, second)
	}
}

// An error means zipcity was not actually consulted, so it must not be
// cached — caching it would silently turn "we don't know" into a permanent
// false for the rest of the call. check's own logic only stores the answer
// when err is nil; this pins that down from outside the package.
func TestReferenceDoesNotCacheAnError(t *testing.T) {
	r := parse.Reference{}
	calls := 0
	flaky := func() (bool, error) {
		calls++
		if calls == 1 {
			return false, errors.New("zipcity unavailable")
		}
		return true, nil
	}

	if _, err := parse.Check(r, "key", flaky); err == nil {
		t.Fatal("want the first call's error")
	}
	if _, err := parse.Check(r, "key", flaky); err != nil {
		t.Fatalf("second check: %v", err)
	}
	if calls != 2 {
		t.Errorf("query ran %d times, want 2 — an erroring query must not be cached", calls)
	}
}

// A candidate with a ZIP, a city, and a two-letter region asks both
// MatchZipAndStreet and MatchCityStateAndStreet and folds the two answers —
// see foldStreetAnswers. W 9200 S is real in zipcity's West Jordan data on
// both shards (verified by probe), so the fold lands on Agrees.
func TestStreetAgreementAsksZipAndStreetWhenAZipIsPresent(t *testing.T) {
	a := &address.Address{
		Type:            &ordinarystreet.OrdinaryStreetAddress{},
		Predirectional:  "W",
		StreetName:      "9200",
		Postdirectional: "S",
		City:            "WEST JORDAN",
		Region:          "UT",
		Postal:          "84088",
	}

	ans, ok := parse.StreetAgreement(a)
	if !ok {
		t.Fatal("want a question asked")
	}
	if ans != parse.Agrees {
		t.Errorf("StreetAgreement = %v, want Agrees", ans)
	}
}

// A candidate with no ZIP falls back to MatchCityStateAndStreet rather than
// asking nothing — the choice CONTRIBUTING calls out as worth deciding
// explicitly. Pleasant Hill Rd is real in zipcity's Pleasant Hill, CA
// city-street data (verified by probe); a made up street at the same city and
// state is absent, and the two together show the city-state path is what
// answered, not a decline that happened to look like one.
func TestStreetAgreementFallsBackToCityAndStateWithNoZip(t *testing.T) {
	real := &address.Address{
		Type:         &ordinarystreet.OrdinaryStreetAddress{},
		StreetName:   "PLEASANT HILL",
		StreetSuffix: "RD",
		City:         "PLEASANT HILL",
		Region:       "CA",
	}
	if ans, ok := parse.StreetAgreement(real); !ok || ans != parse.Agrees {
		t.Errorf("StreetAgreement(real street) = %v, %v, want Agrees, true", ans, ok)
	}

	absent := &address.Address{
		Type:         &ordinarystreet.OrdinaryStreetAddress{},
		StreetName:   "ZQXVBORK",
		StreetSuffix: "LN",
		City:         "PLEASANT HILL",
		Region:       "CA",
	}
	if ans, ok := parse.StreetAgreement(absent); !ok || ans != parse.Contradicts {
		t.Errorf("StreetAgreement(absent street) = %v, %v, want Contradicts, true", ans, ok)
	}
}

// With no ZIP Code to pair the city with, the city is asked about on its own:
// has the state ever had a place by that name. It is asked only then — a ZIP
// Code present means zipCityAgreement's stronger question is the one to ask
// — and only when there is a two-letter region to ask it of.
func TestCityZipsAgreementAsksTheStateWhenThereIsNoZip(t *testing.T) {
	real := &address.Address{City: "WEST PALM BEACH", Region: "FL"}
	if ans, ok := parse.CityZipsAgreement(real); !ok || ans != parse.Agrees {
		t.Errorf("CityZipsAgreement(real city) = %v, %v, want Agrees, true", ans, ok)
	}

	absent := &address.Address{City: "ST WEST PALM BEACH", Region: "FL"}
	if ans, ok := parse.CityZipsAgreement(absent); !ok || ans != parse.Contradicts {
		t.Errorf("CityZipsAgreement(absent city) = %v, %v, want Contradicts, true", ans, ok)
	}

	for name, a := range map[string]*address.Address{
		"with a ZIP":     {City: "WEST PALM BEACH", Region: "FL", Postal: "33401"},
		"without region": {City: "WEST PALM BEACH"},
		"without city":   {Region: "FL"},
	} {
		if ans, ok := parse.CityZipsAgreement(a); ok {
			t.Errorf("CityZipsAgreement(%s) = %v, true; want it to decline", name, ans)
		}
	}
}

// Without a ZIP Code or a comma nothing in the grammar says where the street
// ends and the city begins: 123 NORTH PARK ST PAUL MN reads as well with PARK
// ST PAUL for its city as with ST PAUL, and lastline offers both. The data
// knows only one of them is a place in Minnesota, and that is what splits
// the line (addressparsers#17). The same holds for ST WEST PALM BEACH against
// the two real places behind it — which of those two is not the data's
// call here, since both are real.
func TestTheDataSplitsAnUnmarkedCityWithNoZip(t *testing.T) {
	withData := parse.New()

	a, err := withData.Parse("123 NORTH PARK ST PAUL MN")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.City != "ST PAUL" {
		t.Errorf("City = %q, want ST PAUL", a.City)
	}

	a, err = withData.Parse("123 MAIN ST WEST PALM BEACH FL")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.City != "WEST PALM BEACH" && a.City != "PALM BEACH" {
		t.Errorf("City = %q, want WEST PALM BEACH or PALM BEACH", a.City)
	}
}

// The ambiguity 123 NORTH PARK ST PAUL MN carries is twin: PARK is a street
// name and a street suffix, and ST is STREET and SAINT, so NORTH PARK in
// ST PAUL and NORTH PARK ST in PAUL are both grammatical readings of the same
// tokens. Asking the state settles the unmarked form towards
// ST PAUL, since Minnesota has never had a place called PAUL — but it settles
// only the form nothing else marks. Spelling the suffix out, breaking the
// line, or naming a state the city is real in each still reaches the other
// reading, so the data narrows the ambiguity rather than deciding it.
//
// The street the settled form decomposes to is NORTH with a suffix of PARK,
// not a name of NORTH PARK. go-projectusat a4da5b1 stopped charging a name
// that is nothing but a placed directional, because the standard's NORTH AVE
// is a directional street name (p.17), and PARK is a Pub 28 suffix, so
// NORTH PARK is read the way NORTH AVE is. That is a decomposition, not a
// hash: all four rows here and the line-broken form all normalize to
// 123 NORTH PARK / ST PAUL MN either way. It is this branch that makes them
// agree — before it the line-broken form already read NORTH + PARK while the
// one-line form read NORTH PARK. Which decomposition is right cannot be had
// from the string.
func TestBothReadingsOfTheSaintStreetAmbiguityAreReachable(t *testing.T) {
	withData := parse.New()

	for _, tc := range []struct {
		why                     string
		in                      string
		pre, name, suffix, city string
	}{
		{"nothing marks the split, so the data does", "123 NORTH PARK ST PAUL MN", "", "NORTH", "PARK", "ST PAUL"},
		{"a spelled out suffix cannot be SAINT", "123 NORTH PARK STREET, PAUL, MN", "N", "PARK", "ST", "PAUL"},
		{"a line break marks the split itself", "123 NORTH PARK ST\nPAUL, MN", "N", "PARK", "ST", "PAUL"},
		{"Paul is a real place in Idaho", "123 NORTH PARK ST, PAUL, ID", "N", "PARK", "ST", "PAUL"},
	} {
		a, err := withData.Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q): %v", tc.in, err)
			continue
		}
		got := []string{a.Predirectional, a.StreetName, a.StreetSuffix, a.City}
		want := []string{tc.pre, tc.name, tc.suffix, tc.city}
		if !slices.Equal(got, want) {
			t.Errorf("Parse(%q) — %s\n got  pre/name/suffix/city = %q\n want                   = %q", tc.in, tc.why, got, want)
		}
	}
}

// Intake data is not consistent about commas, and a comma must not change the
// hash (#20). It used to: with the comma the last line is marked and the
// street reading ordinarystreet rates highest wins, and without it the city
// claim comes in a rung lower and CandidateAddress.Confidence — the minimum
// over the accepted claims — flattens every reading of the street line to the
// city's own confidence, so enumeration order picked the street instead.
func TestACommaDoesNotChangeTheStreet(t *testing.T) {
	withData := parse.New()

	for _, marked := range []string{
		"123 OCEAN BOULEVARD, WEST PALM BEACH, FL",
		"123 MAIN STREET, WEST PALM BEACH, FL",
		"3253 W 9200 S, WEST JORDAN, UT",
		"123 NORTH PARK ST, ST PAUL, MN",
	} {
		unmarked := strings.ReplaceAll(marked, ",", "")

		a, err := withData.Parse(marked)
		if err != nil {
			t.Errorf("Parse(%q): %v", marked, err)
			continue
		}
		b, err := withData.Parse(unmarked)
		if err != nil {
			t.Errorf("Parse(%q): %v", unmarked, err)
			continue
		}

		got := []string{b.Predirectional, b.StreetName, b.StreetSuffix, b.Postdirectional, b.City}
		want := []string{a.Predirectional, a.StreetName, a.StreetSuffix, a.Postdirectional, a.City}
		if !slices.Equal(got, want) {
			t.Errorf("Parse(%q)\n got  pre/name/suffix/post/city = %q\n want (from %q)            = %q", unmarked, got, marked, want)
		}
	}
}

// StreetConfidence reads the street line's own rating off the claims the
// reading accepted, which is what the candidate's minimum destroyed. A
// reading with no street line has nothing to rate and reports what it is
// already held at, so that this tie-break neither lifts a post office box
// above a street nor drops it below one.
func TestStreetConfidenceReadsTheStreetAndNothingElse(t *testing.T) {
	street := &address.CandidateAddress{
		Confidence: claim.ConfidenceLikely,
		Claims: []claim.Claim{
			{Confidence: claim.ConfidenceExact, Parts: []claim.ClaimPart{{Part: claim.PartStreetName}}},
			{Confidence: claim.ConfidenceStrong, Parts: []claim.ClaimPart{{Part: claim.PartStreetSuffix}}},
			{Confidence: claim.ConfidenceWeak, Parts: []claim.ClaimPart{{Part: claim.PartCity}}},
		},
	}
	if got := parse.StreetConfidence(street); got != claim.ConfidenceStrong {
		t.Errorf("StreetConfidence(street) = %v, want %v — the city must not be counted", got, claim.ConfidenceStrong)
	}

	box := &address.CandidateAddress{
		Confidence: claim.ConfidenceExact,
		Claims:     []claim.Claim{{Confidence: claim.ConfidenceWeak, Parts: []claim.ClaimPart{{Part: claim.PartCity}}}},
	}
	if got := parse.StreetConfidence(box); got != claim.ConfidenceExact {
		t.Errorf("StreetConfidence(no street) = %v, want %v", got, claim.ConfidenceExact)
	}
}

// The closed forms carry a fixed pseudo street name — "PO BOX" is not a
// street zipcity was ever asked about — so StreetAgreement must decline
// rather than manufacture a contradiction on every one of them.
func TestStreetAgreementDeclinesForClosedForms(t *testing.T) {
	a := &address.Address{
		Type:          &pobox.POBoxAddress{},
		StreetName:    "PO BOX",
		PrimaryNumber: "11890",
		City:          "WEST JORDAN",
		Region:        "UT",
		Postal:        "84088",
	}

	if _, ok := parse.StreetAgreement(a); ok {
		t.Error("want the question declined for a closed form")
	}
}

// A reading with no street name at all — an ordinarystreet candidate the
// vocabularies left empty — has nothing to ask zipcity about.
func TestStreetAgreementDeclinesWithNoStreetName(t *testing.T) {
	a := &address.Address{
		Type:   &ordinarystreet.OrdinaryStreetAddress{},
		City:   "WEST JORDAN",
		Region: "UT",
		Postal: "84088",
	}

	if _, ok := parse.StreetAgreement(a); ok {
		t.Error("want the question declined with no street name")
	}
}

// Step 4 of #17: the puertorico reading's StreetName names a place exactly as
// the ordinarystreet reading's does, so it must be asked the street question
// rather than declined. Declining it is what let the ordinarystreet reading of
// the same tokens collect a rung unopposed — see streetAgreement's comment.
// CALLE 1 is a real San Juan street in the pinned filter, and the long form is
// the key, so the honest answer here is agrees for both readings.
func TestStreetAgreementAnswersForPuertoRico(t *testing.T) {
	pr := &address.Address{
		Type:       &puertorico.PuertoRicoAddress{},
		StreetName: "CALLE 1",
		City:       "SAN JUAN",
		Region:     "PR",
	}
	ordinary := &address.Address{
		Type:       &ordinarystreet.OrdinaryStreetAddress{},
		StreetName: "CALLE 1",
		City:       "SAN JUAN",
		Region:     "PR",
	}

	got, ok := parse.StreetAgreement(pr)
	if !ok {
		t.Fatal("want the street question asked of a puertorico reading")
	}
	if got != parse.Agrees {
		t.Errorf("puertorico reading: want %v, got %v", parse.Agrees, got)
	}

	// The point of asking is that the two readings of one street line get the
	// same answer, so neither is ranked above the other by a question only one
	// of them was asked.
	same, ok := parse.StreetAgreement(ordinary)
	if !ok || same != got {
		t.Errorf("ordinarystreet reading of the same street: want (%v, true), got (%v, %v)", got, same, ok)
	}
}

// A street the data contradicts steps the reading down but the address still
// parses — the same guarantee TestReferenceDataDemotesButDoesNotReject makes
// for zip+city, now exercised for the street question. ZQXVBORK is invented
// and West Jordan, UT 84088 is real, so only the street half of agreement can
// be contradicting here.
//
// The street carries no suffix on purpose. ordinarystreet always offers a
// second reading beside "name ZQXVBORK" that absorbs a would-be suffix into
// the name instead — see candidate.go's streetConfidence — and that second
// reading renders to the identical zipcity query string as the first, so a
// zip+city agreement (West Jordan pairs with 84088) can promote it to tie
// the properly split reading before the street contradiction demotes both by
// the same step. A bare name has no suffix to absorb, so there is only one
// reading and no tie for the demotion to land ambiguously on.
func TestStreetContradictionDemotesButStillParses(t *testing.T) {
	source := "123 ZQXVBORK\nWEST JORDAN UT 84088"

	withData, err := parse.New().Parse(source)
	if err != nil {
		t.Fatalf("a street-contradicted reading must still parse: %v", err)
	}
	if withData.StreetName != "ZQXVBORK" || withData.PrimaryNumber != "123" {
		t.Errorf("street = %q %q, want %q %q", withData.PrimaryNumber, withData.StreetName, "123", "ZQXVBORK")
	}
}

// Two readings of one street line that differ only in which field holds a
// word ask the data the same question, so agreement lifts both by the same
// amount and the grammar's gap between them survives. Under the old capped
// scale they landed together at the ceiling and the winner was candidate
// order; this pins the end-to-end result now that the score carries the gap.

// TODO: determine if we still need this test. It cannot work the way it is built
// because this parser does not haev the `UseReferenceData` option that
// `addressparsers` does.

// func TestAgreementCannotLiftAnAbsorbedReadingPastTheSplitOne(t *testing.T) {
// 	source := "1600 PENNSYLVANIA AVE NW\nWASHINGTON DC 20500"

// 	for _, opts := range []parse.Options{{}, {UseReferenceData: true}} {
// 		a, err := parse.New(opts).Parse(source)
// 		if err != nil {
// 			t.Fatalf("parsing with %+v: %v", opts, err)
// 		}
// 		if a.StreetName != "PENNSYLVANIA" || a.StreetSuffix != "AVE" || a.Postdirectional != "NW" {
// 			t.Errorf("with %+v: name=%q suffix=%q post=%q, want PENNSYLVANIA AVE NW split into its fields",
// 				opts, a.StreetName, a.StreetSuffix, a.Postdirectional)
// 		}
// 	}
// }

// foldStreetAnswers is the whole fold: both true is agrees, both false is
// contradicts, and the two split combinations each name the side that was
// absent.
func TestFoldStreetAnswers(t *testing.T) {
	cases := []struct {
		zipPresent, cityPresent bool
		want                    parse.Agreement
	}{
		{true, true, parse.Agrees},
		{false, false, parse.Contradicts},
		{false, true, parse.MissingInZip},
		{true, false, parse.MissingInCity},
	}
	for _, tc := range cases {
		if got := parse.FoldStreetAnswers(tc.zipPresent, tc.cityPresent); got != tc.want {
			t.Errorf("FoldStreetAnswers(%v, %v) = %v, want %v",
				tc.zipPresent, tc.cityPresent, got, tc.want)
		}
	}
}

// presentMatch is just Found(), regardless of whether the reading carries a
// directional. #34 (addressparsers#6) split on that, reasoning a variant hit
// under a directional-carrying reading meant TIGER disagreed with the
// input's directional. #35 found the split rested on a false premise:
// directionalVariants only ever fills an end the reading left empty, never
// disputes one already there, so a variant hit is never a refutation —
// Found() is safe unconditionally.
func TestPresentMatchIsFound(t *testing.T) {
	exact := zipcity.Match{Exact: true}
	variant := zipcity.Match{Variants: []string{"N DECATUR RD"}}
	absent := zipcity.Match{}

	cases := []struct {
		name  string
		match zipcity.Match
		want  bool
	}{
		{"exact hit", exact, true},
		{"variant hit", variant, true},
		{"absent", absent, false},
	}
	for _, tc := range cases {
		if got := parse.PresentMatch(tc.match); got != tc.want {
			t.Errorf("%s: PresentMatch = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A reading with a ZIP, a city, and a two-letter region asks both street
// questions rather than just one. 1600 Pennsylvania Ave NW is real in
// zipcity's Washington, DC data on both the ZIP shard and the city-state
// shard (verified by probe), so both questions come back true and the fold
// lands on Agrees.
func TestStreetAgreementAsksBothQuestionsWhenItCan(t *testing.T) {
	a := &address.Address{
		Type:            &ordinarystreet.OrdinaryStreetAddress{},
		StreetName:      "PENNSYLVANIA",
		StreetSuffix:    "AVE",
		Postdirectional: "NW",
		City:            "WASHINGTON",
		Region:          "DC",
		Postal:          "20500",
	}

	ans, ok := parse.StreetAgreement(a)
	if !ok {
		t.Fatal("want a question asked")
	}
	if ans != parse.Agrees {
		t.Errorf("StreetAgreement = %v, want Agrees", ans)
	}
}

// A split street answer takes no step, in either direction: the fold's
// missingInZip and missingInCity cases must leave a reading exactly where it
// would sit with reference data off.
//
// Wisconsin Ave NW is real in zipcity's Washington, DC city-state data, but
// not for ZIP 20500 (verified by probe): Wisconsin Avenue runs through upper
// northwest DC, nowhere near the White House ZIP. That is a genuine split —
// the street is real, the pairing with this particular ZIP is not — rather
// than a manufactured one.

// TODO: determine if we still need this test. It cannot work the way it is built
// because this parser does not haev the `UseReferenceData` option that
// `addressparsers` does.

// func TestASplitStreetAnswerTakesNoStep(t *testing.T) {
// 	a := &address.Address{
// 		Type:            &ordinarystreet.OrdinaryStreetAddress{},
// 		StreetName:      "WISCONSIN",
// 		StreetSuffix:    "AVE",
// 		Postdirectional: "NW",
// 		City:            "WASHINGTON",
// 		Region:          "DC",
// 		Postal:          "20500",
// 	}
// 	if ans, ok := parse.StreetAgreement(a); !ok || ans != parse.MissingInZip {
// 		t.Fatalf("StreetAgreement = %v, %v, want MissingInZip, true", ans, ok)
// 	}

// 	source := "1600 WISCONSIN AVE NW\nWASHINGTON DC 20500"
// 	plain, err := parse.New(parse.Options{}).Parse(source)
// 	if err != nil {
// 		t.Fatalf("parsing without reference data: %v", err)
// 	}
// 	withData, err := parse.New(parse.Options{UseReferenceData: true}).Parse(source)
// 	if err != nil {
// 		t.Fatalf("parsing with reference data: %v", err)
// 	}
// 	if !plain.Equals(withData) {
// 		t.Errorf("a split street answer moved a reading:\n without = %+v\n with    = %+v", plain, withData)
// 	}
// }

// The string zipcity answers about has to be the string the caller will hash.
// A reading whose StreetName absorbs the suffix word and the directional —
// E as a predirectional, ST NW as the name — is written E ST NW and emitted
// E STREET NORTHWEST, because the normalizer spells both out inside a
// multi-word name. Asking about the form it is written in collects evidence
// for an address that is never produced.
func TestStreetForQueryAsksTheStreetThatWillBeEmitted(t *testing.T) {
	absorbed := &address.Address{
		Type:           &ordinarystreet.OrdinaryStreetAddress{},
		Predirectional: "E",
		StreetName:     "ST NW",
	}
	if got, want := parse.StreetForQuery(absorbed), "E STREET NORTHWEST"; got != want {
		t.Errorf("StreetForQuery(absorbed) = %q, want %q", got, want)
	}

	// A reading that puts each word in its own field is already written the
	// way it is emitted, so normalizing changes nothing — which is why asking
	// the fields as written was right for every reading but the absorbed one.
	split := &address.Address{
		Type:            &ordinarystreet.OrdinaryStreetAddress{},
		StreetName:      "E",
		StreetSuffix:    "ST",
		Postdirectional: "NW",
	}
	if got, want := parse.StreetForQuery(split), "E ST NW"; got != want {
		t.Errorf("StreetForQuery(split) = %q, want %q", got, want)
	}
}

// amadsen on go-projectusat#112: a valid directional street name stays
// written out, and a valid alphabetic one stays a letter. Washington DC has
// E ST NW and no EAST ST NW, so the absorbed reading used to collect the
// agreement E ST NW earned and win with it. With both readings asked about
// the street they would emit, neither is a street zipcity has, nothing is
// promoted, and the grammar keeps EAST as the name.
func TestADirectionalStreetNameIsNotTradedForAnAbsorbedReading(t *testing.T) {
	p := parse.New()

	directional, err := p.Parse("100 EAST ST NW, WASHINGTON, DC 20004")
	if err != nil {
		t.Fatalf("Parse(EAST) error = %v", err)
	}
	if directional.StreetName != "EAST" {
		t.Errorf("StreetName = %q, want EAST (pre=%q suf=%q post=%q)",
			directional.StreetName, directional.Predirectional,
			directional.StreetSuffix, directional.Postdirectional)
	}

	// The alphabetic street the data does have is unchanged: it still agrees
	// and still wins, so this is not the demotion of one reading but the two
	// being asked about themselves.
	alphabetic, err := p.Parse("100 E ST NW, WASHINGTON, DC 20004")
	if err != nil {
		t.Fatalf("Parse(E) error = %v", err)
	}
	if alphabetic.StreetName != "E" || alphabetic.StreetSuffix != "ST" || alphabetic.Postdirectional != "NW" {
		t.Errorf("got name=%q suffix=%q post=%q, want E / ST / NW",
			alphabetic.StreetName, alphabetic.StreetSuffix, alphabetic.Postdirectional)
	}
}

// Step 6 of #17: a secondary designator written ahead of the street belongs
// after it. Publication 28 §213 puts the designator between the street and
// the city, but intake data writes it first constantly, and the grammar reads
// a street line left to right — so the whole run becomes one street name and
// no candidate in the pile holds the right answer. See leadingSecondary.
func TestALeadingSecondaryDesignatorMovesAfterTheStreet(t *testing.T) {
	cases := []struct {
		source string
		street string
	}{
		{"Apartment 3200 152 South Tech Dr\nTAMPA FL 33602", "152 S TECH DR APT 3200"},
		// #3200 is not a primary number: §213.2 makes # the secondary
		// designator of unspecified type, so the number ordinarystreet found
		// is the unit, in the one place the standard does not put it.
		{"#3200 152 South Tech Dr\nTAMPA FL 33602", "152 S TECH DR # 3200"},
		{"PMB 456 152 South Tech Dr\nTAMPA FL 33602", "152 S TECH DR PMB 456"},
	}

	for _, c := range cases {
		t.Run(c.street, func(t *testing.T) {
			for _, ref := range []bool{false, true} {
				a, err := parse.New().Parse(c.source)
				if err != nil {
					t.Fatalf("reference data %v: %v", ref, err)
				}
				if got := a.FormatStreetLine(); got != c.street {
					t.Errorf("reference data %v: street line = %q, want %q", ref, got, c.street)
				}
			}
		})
	}
}

// The guards are what keep step 6 from firing on an address that means what
// it says: a reading that already placed a secondary, a closed form that read
// its own line whole, and a designator word with no number after it — which
// at the start of a line is far more often a street name.
func TestALeadingSecondaryLeavesAnAddressThatMeansWhatItSaysAlone(t *testing.T) {
	cases := []struct {
		name   string
		source string
		street string
	}{
		{"already placed", "123 N MAIN ST APT 4\nWEST JORDAN UT 84088", "123 N MAIN ST APT 4"},
		{"closed form", "PO BOX 11890\nWEST JORDAN UT 84088", "PO BOX 11890"},
		{"unnumbered designator word", "Key West Ave\nROCKVILLE MD 20850", "AVE KEY WEST"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, err := parse.New().Parse(c.source)
			if err != nil {
				t.Fatalf("parsing: %v", err)
			}
			if got := a.FormatStreetLine(); got != c.street {
				t.Errorf("street line = %q, want %q", got, c.street)
			}
			if a.BusinessName != "" {
				t.Errorf("BusinessName = %q, want empty", a.BusinessName)
			}
		})
	}
}
