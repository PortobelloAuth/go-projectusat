package parse

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizerfn"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizeroptions"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
	"github.com/poetic-systems/zipcity"
)

// agreement is what the reference data had to say about a reading.
type agreement int

const (
	// unknown is the answer whenever the data was not consulted, could not be
	// consulted, or declined to answer. It is not evidence either way.
	unknown agreement = iota
	// agrees means the data may hold this pairing.
	agrees
	// contradicts means the data definitively does not hold it.
	contradicts
	// missingInZip means a street question split: the city-state filter
	// found the street, the ZIP filter did not. The street is real in the
	// city but not recorded for this ZIP — the ZIP is the suspect.
	missingInZip
	// missingInCity means a street question split the other way: the ZIP
	// filter found the street, the city-state filter did not. The street is
	// real for this ZIP but not recorded for this city name — the city is
	// the suspect. zipcity's city-street keys come from TIGER place names,
	// and a postal or GeoNames city name need not be one, so this direction
	// is probably not rare — which is exactly why it must not be folded
	// into agrees or contradicts.
	missingInCity
)

// questionAnswers is one candidate's answers to the two questions choose can
// ask: the city question (zip+city, or city+state when there is no ZIP —
// see cityAgreement) and the street question (see streetAgreement). unknown
// is each field's zero value, which already means "not evidence" throughout
// this package — see agreement's doc comment — so a question this reading
// never qualified to ask needs no separate flag.
type questionAnswers struct {
	city, street agreement
}

// gated reports qa's answers as score should see them: a contradicts stands
// only when the matching confirmed flag says some competing candidate's
// answer to that same question was agrees. Without that competitor, "zipcity
// never saw this" is indistinguishable from a gap in the data, so it must
// cost nothing, exactly like a question this reading never asked —
// go-projectusat#176.
func (qa questionAnswers) gated(cityConfirmed, streetConfirmed bool) []agreement {
	return []agreement{
		gateOne(qa.city, cityConfirmed),
		gateOne(qa.street, streetConfirmed),
	}
}

func gateOne(ans agreement, confirmed bool) agreement {
	if ans == contradicts && !confirmed {
		return unknown
	}
	return ans
}

// agreement asks zipcity every question this reading supports: the city
// question first (zip+city, or city+state, when there is no ZIP Code to pair
// the city with — see cityZipsAgreement), then the street question, folded
// from up to two zipcity calls into one answer — see streetAgreement.
//
// Every answer is evidence, and none of them are symmetric. zipcity answers
// from bloom filters built at a 0.005 false positive rate (the rate is set in
// zipcity's `internal/bloomgenerator`), so a false is definitive — the key was
// never added — while a true is a likelihood ratio of about 200:1 in favour of
// the pairing rather than a confirmation of it. Both move a candidate's score
// by one unit and neither settles it: querying several mutually exclusive
// readings that are all genuinely absent yields a spurious true about
// 1-0.995^k of the time, so a true must not be allowed to resolve a reading
// on its own, and must never be reported to a caller as verification.
//
// A contradiction is likewise not proof the address is wrong. zipcity is built
// from Census TIGER files with documented gaps, so a real address the Census
// missed lands here too. That is why every answer is one point of score
// rather than rejection, and why UseReferenceData is off by default — and,
// since go-projectusat#176, why choose only charges a contradicts when some
// competing candidate's agrees on the same question shows the gap is the
// address's and not the data's — see choose's gating of questionAnswers.
func (p *Parser) agreement(r reference, a *address.Address) questionAnswers {
	var qa questionAnswers

	if ans, ok := p.cityAgreement(r, a.Postal, a.City, a.Region); ok {
		qa.city = ans
	}
	if ans, ok := p.streetAgreement(r, a); ok {
		qa.street = ans
	}

	return qa
}

// reference answers each distinct zipcity question once per choice. The
// filters are static, so asking a key twice cannot change the false-positive
// odds; what asking once buys is that readings sharing a key are ranked on
// one answer even if a check ever failed unstably, and one call per key.
type reference map[string]bool

func (r reference) check(key string, query func() (bool, error)) (bool, error) {
	if present, ok := r[key]; ok {
		return present, nil
	}
	present, err := query()
	if err == nil {
		r[key] = present
	}
	return present, err
}

// zipCityAgreement asks whether the ZIP and city pair. It declines, rather
// than answering, when there is nothing to ask about or zipcity could not be
// consulted: an address the data cannot speak to is not thereby a better or a
// worse reading.
func (p *Parser) zipCityAgreement(r reference, postal, city string) (agreement, bool) {
	m := zip5.FindStringSubmatch(postal)
	if m == nil || city == "" {
		return unknown, false
	}

	present, err := r.check("zip city "+m[1]+" "+city, func() (bool, error) {
		return p.zipcity.CheckZipAndCity(m[1], city)
	})
	if err != nil {
		return unknown, false
	}
	return answerFor(present), true
}

// cityZipsAgreement asks, for a reading with a city and a two-letter region
// but no ZIP Code, whether the data has seen that city anywhere in that
// state: ZipsKnownFor yields a code for it or yields nothing. It is
// zipCityAgreement's question read the other way, and it is asked only when
// that one cannot be, since a city known for this ZIP Code is stronger
// evidence than a city known somewhere in the state.
//
// An empty answer is definitive in the sense a false from the zip-city
// filter is — the name was never seen for any code in the state, in GeoNames
// or in TIGER — and is charged the same one point, no more: zipcity's own
// caveat that the list is neither complete nor preferred-first holds, and a
// real city the data missed still parses. What the point buys is the split
// nothing else marks: with no ZIP Code and no comma, 123 MAIN ST WEST PALM
// BEACH FL reads as well with ST WEST PALM BEACH for its city as with WEST
// PALM BEACH, and only the data knows one of those is a place
// (addressparsers#17). Which of the city's codes the address belongs to is
// not asked here; that is the street question's business.
func (p *Parser) cityZipsAgreement(r reference, postal, city, region string) (agreement, bool) {
	if zip5.MatchString(postal) || city == "" || len(region) != 2 {
		return unknown, false
	}

	present, err := r.check("city zips "+region+" "+city, func() (bool, error) {
		for range p.zipcity.ZipsKnownFor(region, city) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return unknown, false
	}
	return answerFor(present), true
}

// streetAgreement asks zipcity about a reading's street, folding the answer
// — or the two answers, when the reading qualifies for both — into one
// agreement.
//
// ordinarystreet and puertorico are the two types whose StreetName names a
// place. pobox, ruralroute, military and generaldelivery each carry a fixed
// pseudo-name that describes their format rather than a street — "PO BOX",
// "RR 4", "PSC 3" — and zipcity was never built to answer whether those
// strings are streets. Asking it would manufacture a contradiction on every
// closed-form reading rather than evidence about one, so those types are
// left at zip+city alone.
//
// Step 4 of #17 is why puertorico is in that list and not only
// ordinarystreet. The two address types offer readings of the same tokens, so
// where one is asked the street question and the other is not, the answer is
// not evidence about the address — it is a free rung for whichever reading
// happened to be admitted. A17 CALLE 1 / SAN JUAN PR is the measured case:
// both readings hold identical Predirectional, StreetName, StreetSuffix and
// Postdirectional, streetForQuery renders "CALLE 1" for both, and zipcity
// answers that key true. With only ordinarystreet admitted the two readings
// tied on the grammar at rung 3 and the one point went to ordinarystreet
// alone, flipping the reading that reference data off returns. Admitting
// puertorico puts the same question to both, so the point cancels and the
// grammar decides — which is the rule the rest of choose already follows:
// reference data orders readings against each other, it does not get to
// decide which of them may be asked.
//
// An earlier revision of this function did the opposite — skipped the street
// question entirely when the last line put the address in Puerto Rico — on
// the belief that zipcity keys only TIGER's Appendix E short form, so the
// unabbreviated name the spec requires on output could never match. That is
// false against the pinned filter: CheckCityStateAndStreet("SAN JUAN", "PR",
// "CALLE 1") and "CALLE LOIZA" both answer true, and it is "CLL 1" / "CLL
// LOIZA" that answer false. Pub28FeatureName expands featname 156 to CALLE
// and ApplySpanishPrefixOverrides leaves it alone, so the long form is the
// key. A Puerto Rico street is answerable like any other and there is nothing
// here to suppress.
//
// MatchZipAndStreet and MatchCityStateAndStreet are different filters over
// different keys — one shard scoped by ZIP, the other by city and state —
// so their false positives are independent: about 0.00075^2 for both to be
// spurious, against 0.00075 for either alone. On the truth side the two are
// almost perfectly correlated, since a street TIGER recorded for a ZIP is
// almost always recorded for that ZIP's city too, so a second true adds
// little evidence on its own. What it buys is corroboration: asking both and
// folding the answers before either moves confidence squares the exposure to
// a spurious step instead of doubling it, which is what two separate agrees
// calls would cost. That is why this reading asks both questions whenever it
// has a ZIP, a city, and a two-letter region, and asks whichever one it can
// when it has only one of those — see foldStreetAnswers for the fold.
//
// Each call folds a match into present/absent through presentMatch, which is
// just Found(): an exact hit or a directional-variant hit (DECATUR RD found
// under N DECATUR RD) both count as present, regardless of whether the
// reading itself carries a directional. #34 (addressparsers#6) split on
// hasDirectional instead, on the belief that a variant hit under a reading
// that already carries a directional means TIGER's directional disagrees
// with the input's — a refutation rather than corroboration. #35 found that
// belief false: directionalVariants (zipcity.go) only ever fills whichever
// end, front or back, the reading left empty; it never swaps or drops a
// directional the reading already supplied. So a variant hit can never
// contradict a directional already present — it can only be evidence the
// reading is simply missing one, exactly like the no-directional case. The
// exact-only split also silently starved fully-directional streets (both
// Predirectional and Postdirectional set) of any corroboration at all, since
// there both ends are already filled and Variants is always empty — Found()
// removes that gap along with the false premise. The two-shard fold above
// still applies on top of this; only what counts as "present" in one shard
// changed.
func (p *Parser) streetAgreement(r reference, a *address.Address) (agreement, bool) {
	if civic, ok := a.Type.(address.DataDependentAddressType); !ok || !civic.IsCivicAddressType() {
		return unknown, false
	}
	if a.StreetName == "" {
		return unknown, false
	}
	street := streetForQuery(a)

	m := zip5.FindStringSubmatch(a.Postal)
	hasZip := m != nil
	hasCity := a.City != "" && len(a.Region) == 2
	if !hasZip && !hasCity {
		return unknown, false
	}

	var inZip, inCity bool
	if hasZip {
		present, err := r.check("zip street "+m[1]+" "+street, func() (bool, error) {
			match, err := p.zipcity.MatchZipAndStreet(m[1], street)
			return presentMatch(match), err
		})
		if err != nil {
			return unknown, false
		}
		inZip = present
	}
	if hasCity {
		present, err := r.check("city street "+a.City+" "+a.Region+" "+street, func() (bool, error) {
			match, err := p.zipcity.MatchCityStateAndStreet(a.City, a.Region, street)
			return presentMatch(match), err
		})
		if err != nil {
			return unknown, false
		}
		inCity = present
	}

	if hasZip && hasCity {
		return foldStreetAnswers(inZip, inCity), true
	}
	return answerFor(inZip || inCity), true
}

// presentMatch decides whether a zipcity.Match counts as present. It is just
// Found() — an exact hit or a directional-variant hit both count, whether or
// not the reading already carries a directional. directionalVariants only
// ever fills the end (front or back) the reading left empty; it never
// disputes or replaces one already there, so a variant hit is never evidence
// against a directional the reading supplies — it can only supply one the
// reading is missing. See addressparsers#35.
func presentMatch(match zipcity.Match) bool {
	return match.Found()
}

// foldStreetAnswers combines MatchZipAndStreet's and MatchCityStateAndStreet's
// answers into one street agreement: both true is agrees, both false is
// contradicts, and a split is neither — see streetAgreement and missingInZip
// / missingInCity for why the split case is recorded rather than resolved.
func foldStreetAnswers(zipPresent, cityPresent bool) agreement {
	switch {
	case zipPresent && cityPresent:
		return agrees
	case !zipPresent && !cityPresent:
		return contradicts
	case cityPresent:
		return missingInZip
	default:
		return missingInCity
	}
}

// streetForQuery renders a reading's street the way zipcity's filters are
// keyed: predirectional, name, suffix, postdirectional, in Pub 28's order —
// the same order address.Address.FormatStreetLine uses for these four fields,
// minus the primary number and the secondary unit, which are not part of a
// street.
//
// The four fields are normalized before they are joined, because the string
// that has to be real is the one the caller will hash, not the one the reading
// happens to be written in. The two are usually the same word for word, which
// is why asking the fields as written worked for as long as it did. They come
// apart when a reading's StreetName absorbs a suffix word or a directional:
// the normalizer spells those out inside a multi-word name, so the reading
// that offers ST NW as a name asks zipcity about E ST NW, which Washington DC
// has, and then emits E STREET NORTHWEST, which it does not. Scoring a reading
// on a rendering nobody will ever hash is how 100 EAST ST NW came to beat the
// reading that keeps EAST as the street name.
//
// It is the content normalizer rather than the matching one because the two
// differ, over these four fields, only in whether diacritics are substituted,
// and substituting them is a question about zipcity's keys that #1 over there
// has open. Preserving them is what the fields as written already did.
func streetForQuery(a *address.Address) string {
	street := &address.Address{
		Predirectional:  a.Predirectional,
		StreetName:      a.StreetName,
		StreetSuffix:    a.StreetSuffix,
		Postdirectional: a.Postdirectional,
	}
	// An unreadable vocabulary word is the normalizer declining to rewrite
	// the street, not a reason to decline to ask about it: the fields as
	// written are still the best rendering available.
	// NormalizeStreetLine renders the street parts the way Pub 28 expects it
	// and zipcity stores it.
	o := normalizeroptions.NewFnOptions(street, *normalizeroptions.Content())
	if normalized, err := normalizerfn.NormalizeStreetLine(street, o); err == nil {
		street = normalized
	}
	return textutil.JoinNonEmpty(" ", street.Predirectional, street.StreetName, street.StreetSuffix, street.Postdirectional)
}

// answerFor turns a bloom filter's boolean into the agreement it represents.
func answerFor(present bool) agreement {
	if present {
		return agrees
	}
	return contradicts
}
