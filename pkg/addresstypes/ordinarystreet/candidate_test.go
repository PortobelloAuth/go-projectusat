package ordinarystreet_test

import (
	"sort"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ordinarystreet"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

// Every address here is invented, or is a street name with no premise number
// attached. None of them identifies a residence. See CONTRIBUTING §5.

// vocabulary is the claim pool this package reads. It is assembled from the
// real vocabularies rather than from hand-written claims because the point of
// every test below is how this package behaves against what those packages
// actually say — that ROUTE is a suffix, that a state name is also offered as a
// street name, that a lone letter is offered as both directionals.
func vocabulary(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim
	for _, f := range []func([]token.Token) []claim.Claim{
		region.Claims, postalcode.Claims, country.Claims,
		highways.Claims, streetsuffixes.Claims, directionals.Claims, secondaryunit.Claims,
	} {
		claims = append(claims, f(tokens)...)
	}

	return claims
}

// bestReading returns the single highest-confidence candidate for a source, and
// fails if the reading this package prefers is not unique. A tie is a real
// failure and not a detail of the test: the caller picks between candidates by
// confidence, so two readings at the top means the caller is choosing by
// nothing.
func bestReading(t *testing.T, source string) *addressReading {
	t.Helper()

	tokens := token.Tokenize(source)
	claims := vocabulary(tokens)

	lines := lastline.LineClaims(tokens, claims)
	if len(lines) == 0 {
		t.Fatalf("no last line found")
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].Claim.Compare(lines[j].Claim) < 0 })

	candidates := ordinarystreet.Candidates(tokens, claims, lines[0])
	if len(candidates) == 0 {
		t.Fatalf("no candidates")
	}

	best := candidates[0]
	ties := 0
	for _, c := range candidates {
		switch {
		case c.Confidence > best.Confidence:
			best, ties = c, 0
		case c.Confidence == best.Confidence && c != best:
			ties++
		}
	}
	if ties > 0 {
		t.Fatalf("%d readings tied at confidence %d; the best reading is not unique", ties+1, best.Confidence)
	}

	a := best.Address

	return &addressReading{
		confidence: best.Confidence,
		number:     a.PrimaryNumber,
		pre:        a.Predirectional,
		name:       a.StreetName,
		suffix:     a.StreetSuffix,
		post:       a.Postdirectional,
		designator: a.SecondaryDesignator,
		secondary:  a.SecondaryNumber,
		formatted:  a.FormatStreetLine(),
	}
}

type addressReading struct {
	confidence claim.Confidence
	number     string
	pre        string
	name       string
	suffix     string
	post       string
	designator string
	secondary  string
	formatted  string
}

func TestBestReadingDecomposesTheStreetLine(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		want   addressReading
	}{
		{
			// The plainest case, and the one that shows why a name that
			// swallows a claim is demoted: "MAIN ST" as a name is also offered,
			// and loses to the reading that explains the suffix.
			name:   "suffix outranks the name that swallows it",
			source: "123 MAIN ST\nDENVER CO 80201",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "123", name: "MAIN", suffix: "ST",
				formatted: "123 MAIN ST",
			},
		},
		{
			// region offers PENNSYLVANIA as a possible street name at
			// ConfidenceLikely. Corroboration must not drag the reading down to
			// that: a vocabulary agreeing is not a vocabulary objecting.
			name:   "a street named after a state is not penalized for it",
			source: "1600 PENNSYLVANIA AVE NW\nWASHINGTON DC 20500",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "1600", name: "PENNSYLVANIA", suffix: "AVE", post: "NW",
				formatted: "1600 PENNSYLVANIA AVE NW",
			},
		},
		{
			name:   "predirectional, suffix and secondary unit all come out",
			source: "123 N MAIN ST APT 4\nDENVER CO 80201",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "123", pre: "N", name: "MAIN", suffix: "ST",
				designator: "APT", secondary: "4",
				formatted: "123 N MAIN ST APT 4",
			},
		},
		{
			// PARK is a Pub 28 suffix sitting in the middle of the name. The
			// reading that places DR as the suffix cannot also place PARK
			// there, so it is not a reading that declined to place anything —
			// charging it demoted all four readings of this address to the same
			// confidence and left no best one at all.
			name:   "a suffix word inside the name is not a declined slot",
			source: "123 W FOX PARK DR\nDENVER CO 80201",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "123", pre: "W", name: "FOX PARK", suffix: "DR",
				formatted: "123 W FOX PARK DR",
			},
		},
		{
			// The delivery address written across two lines. Reading the unit
			// line as the street line discarded 123 MAIN ST and reported a
			// street named APT 4.
			name:   "a secondary unit on its own line under the street",
			source: "123 MAIN ST\nAPT 4\nDENVER CO 80201",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "123", name: "MAIN", suffix: "ST",
				designator: "APT", secondary: "4",
				formatted: "123 MAIN ST APT 4",
			},
		},
		{
			// Wisconsin grid numbers are alphanumeric. #55 settled that these
			// must be supported, so the primary number is any leading token
			// carrying a digit and not a run of digits.
			name:   "alphanumeric grid primary number",
			source: "N6W23001 BLUEMOUND RD\nWAUKESHA WI 53188",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "N6W23001", name: "BLUEMOUND", suffix: "RD",
				formatted: "N6W23001 BLUEMOUND RD",
			},
		},
		{
			name:   "a fraction extends the primary number",
			source: "123 1/2 MAIN ST\nDENVER CO 80201",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "123 1/2", name: "MAIN", suffix: "ST",
				formatted: "123 1/2 MAIN ST",
			},
		},
		{
			// The Utah grid address from the zipcity README. Both directionals
			// are read and the numeric name sits between them.
			name:   "grid address with a directional on each side",
			source: "3253 W 9200 S\nWEST JORDAN UT 84088",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "3253", pre: "W", name: "9200", post: "S",
				formatted: "3253 W 9200 S",
			},
		},
		{
			// ROUTE is a Pub 28 suffix, so this name absorbs one — but highways
			// claims the whole run as the street name, and it knows what this
			// package does not. This is the demotion of highways to a
			// vocabulary working end to end: highways names the street and this
			// type builds the address around it. See #56.
			name:   "a corroborated highway name keeps the tokens it absorbs",
			source: "123 STATE ROUTE 9\nALBANY NY 12084",
			want: addressReading{
				confidence: claim.ConfidenceStrong,
				number:     "123", name: "STATE ROUTE 9",
				formatted: "123 STATE ROUTE 9",
			},
		},
		{
			// A street name with no premise number. It is a real reading and it
			// is a weaker one: nothing distinguishes it from a business name or
			// the tail of the line above.
			name:   "a street line with no primary number is read but held lower",
			source: "BROADWAY\nDENVER CO 80201",
			want: addressReading{
				confidence: claim.ConfidenceLikely,
				name:       "BROADWAY",
				formatted:  "BROADWAY",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := bestReading(t, tc.source)
			if *got != tc.want {
				t.Errorf("best reading\n got %+v\nwant %+v", *got, tc.want)
			}
		})
	}
}

// A line above the street line is not this package's to explain. It falls out
// as leftover, and lastline.LineClaim.Candidate charges a confidence step for
// it — which is what keeps an address carrying an unread line below an
// otherwise identical one that reads whole.
func TestALineAboveTheStreetLineCostsAConfidenceStep(t *testing.T) {
	plain := bestReading(t, "123 MAIN ST\nDENVER CO 80201")
	withExtra := bestReading(t, "ACME WIDGETS INC\n123 MAIN ST\nDENVER CO 80201")

	if withExtra.name != plain.name || withExtra.suffix != plain.suffix || withExtra.number != plain.number {
		t.Fatalf("the street line should read the same either way\n got %+v\nwant %+v", *withExtra, *plain)
	}

	if withExtra.confidence >= plain.confidence {
		t.Errorf("unread line above the street line was not charged for: got %d, plain reads %d",
			withExtra.confidence, plain.confidence)
	}
}

// Candidates is called with a last line, and the street line is what sits
// ahead of it. An address that is nothing but its last line has no street line
// and this package has nothing to say about it.
func TestNoStreetLineYieldsNoCandidates(t *testing.T) {
	tokens := token.Tokenize("DENVER CO 80201")
	claims := vocabulary(tokens)

	lines := lastline.LineClaims(tokens, claims)
	if len(lines) == 0 {
		t.Fatalf("no last line found")
	}

	for _, line := range lines {
		if line.Span.Start != 0 {
			continue
		}
		if got := ordinarystreet.Candidates(tokens, claims, line); got != nil {
			t.Errorf("expected no candidates with nothing ahead of the last line, got %d", len(got))
		}
	}
}

// FormatStreetLine is the whole of the AddressType interface, and this type
// implements it by delegating to the ordering address.Address already owns.
// Reaching that branch takes a copy with no Type, and getting the copy wrong
// would recurse forever or drop the fields — so it is worth its own test rather
// than only being exercised through the readings above.
func TestFormatStreetLineDelegatesToTheOrdinaryOrdering(t *testing.T) {
	a := address.Address{
		PrimaryNumber:       "123",
		Predirectional:      "N",
		StreetName:          "MAIN",
		StreetSuffix:        "ST",
		Postdirectional:     "SW",
		SecondaryDesignator: "APT",
		SecondaryNumber:     "4",
	}

	ordinary := a
	ordinary.Type = &ordinarystreet.OrdinaryStreetAddress{}

	const want = "123 N MAIN ST SW APT 4"
	if got := a.FormatStreetLine(); got != want {
		t.Errorf("untyped address formatted %q, want %q", got, want)
	}
	if got := ordinary.FormatStreetLine(); got != want {
		t.Errorf("ordinarystreet formatted %q, want %q", got, want)
	}
}
