// Package parse is an address parser built from go-projectusat's vocabularies,
// consulting zipcity where the grammar alone cannot settle a reading.
//
// go-projectusat implements the standard: it tokenizes, every vocabulary says
// what tokens mean, and every address type offers its reading of the whole. All
// of that is exported and composable, so this package assembles those parts
// rather than reimplementing them. What it adds is the step the standard cannot
// specify — choosing among readings that are all grammatically valid — because
// that choice is data dependent and the data lives outside the standard.
//
// The resulting Parser satisfies parser.ParsingFunc, so it drops into
// go-projectusat through AddressParsingOptions.CustomParser.
package parse

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/generaldelivery"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ordinarystreet"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/puertorico"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/ruralroute"
	"github.com/PortobelloAuth/go-projectusat/pkg/country"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/lastline"
	"github.com/PortobelloAuth/go-projectusat/pkg/postalcode"
	"github.com/PortobelloAuth/go-projectusat/pkg/privatemailbox"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
	"github.com/poetic-systems/zipcity"
)

// vocabularies are the Claims functions consulted for every address, in no
// significant order: each says what it recognizes and none of them rank against
// the others. Listing them once here is what keeps "which vocabularies run" a
// single fact rather than a sequence of calls to keep in step.
var vocabularies = []func([]token.Token) []claim.Claim{
	country.Claims,
	region.Claims,
	postalcode.Claims,
	directionals.Claims,
	streetsuffixes.Claims,
	secondaryunit.Claims,
	privatemailbox.Claims,
	highways.Claims,
	pobox.Claims,
	puertorico.Claims,
	military.Claims,
	ruralroute.Claims,
	generaldelivery.Claims,
}

// addressTypes are the Candidates functions, each offering that type's reading
// of the whole address or offering none.
//
// ordinarystreet is the catchall and offers every reading it can assemble, so
// most of what choose ranks comes from it. It never rates a reading Exact and
// the closed forms rate their own lines Exact, which is what keeps a post
// office box a post office box and not a street named PO BOX: catchall means
// lowest precedence where a closed form also reads the tokens.
//
// puertorico reads the Spanish street line, and offers a reading only where
// the last line puts the address in Puerto Rico. It needs no ranking against
// the others for that reason: a mainland address never sees it, and a Puerto
// Rico address sees it alongside whatever ordinarystreet made of the same
// tokens.
var addressTypes = []func([]token.Token, []claim.Claim, lastline.LineClaim) []*address.CandidateAddress{
	pobox.Candidates,
	puertorico.Candidates,
	military.Candidates,
	ruralroute.Candidates,
	generaldelivery.Candidates,
	ordinarystreet.Candidates,
}

// ErrNoReading reports that no address type offered a reading of the source.
//
// It carries no part of the input. Addresses handled here may be protected
// health information, and an error is the value most likely to reach a log, a
// crash report, or a bug tracker.
var ErrNoReading = fmt.Errorf("parse: no address type offered a reading")

// zip5 matches the five digit form zipcity requires, narrowing a ZIP+4.
var zip5 = regexp.MustCompile(`^(\d{5})(?:-?\d{4})?$`)

// Parser reads a free text address.
type Parser struct{}

// New returns a Parser.
func New() *Parser { return &Parser{} }

// Parse reads source into a structured address, returning ErrNoReading when no
// address type recognizes it.
//
// The stages are go-projectusat's, in the order that library defines: tokenize,
// let every vocabulary claim what it recognizes, read the last line from those
// claims, then ask each address type for its reading of the whole. Only the
// final choice among readings belongs to this package.
func (p *Parser) Parse(source string) (*address.Address, error) {
	tokens := token.Tokenize(source)
	if len(tokens) == 0 {
		return nil, ErrNoReading
	}

	best := p.readLine(tokens)
	if best == nil {
		return nil, ErrNoReading
	}
	p.streetWindow(tokens, best)
	return best.Address, nil
}

// readLine is read with the two layers that rewrite a chosen reading rather
// than rank it: the firm name above the address, and a secondary designator
// written ahead of the street. It is separate from read because step 7 asks
// for a whole reading of a shorter run of tokens, not for a raw candidate.
func (p *Parser) readLine(tokens []token.Token) *address.CandidateAddress {
	best, claims := p.read(tokens)
	if best == nil || best.Address == nil {
		return nil
	}
	if firm := firmLine(tokens, best); firm != "" {
		best.Address.BusinessName = firm
	}
	p.leadingSecondary(tokens, claims, best)
	return best
}

// read is Parse with the tokenizing already done: every vocabulary claims, the
// data narrows the last line, every address type offers its reading, and
// choose picks. It is a function rather than the body of Parse because step 6
// re-reads a shorter run of the same tokens — see leadingSecondary — and the
// thing it needs to re-run is exactly this and not the tokenizer.
//
// Every last line reading the data leaves standing is carried forward rather
// than the best one chosen here. A weaker last line can still be the one the
// winning address type reads, and picking on the grammar alone would hide that
// reading from the choice — which is why what narrows the set is the data and
// nothing else. See lastLines.
func (p *Parser) read(tokens []token.Token) (*address.CandidateAddress, []claim.Claim) {
	var claims []claim.Claim
	for _, vocabulary := range vocabularies {
		claims = append(claims, vocabulary(tokens)...)
	}

	r := reference{}
	var candidates []*address.CandidateAddress
	for _, line := range p.lastLines(r, tokens, claims) {
		for _, addressType := range addressTypes {
			candidates = append(candidates, addressType(tokens, claims, line)...)
		}
	}
	if len(candidates) == 0 {
		return nil, claims
	}
	return p.choose(r, candidates), claims
}

// firmLine reports the topmost physical line as a business name, or "" where
// none applies.
//
// Step 3ii of #17: everything before the last line is address lines, and the
// topmost of them is the firm. ordinarystreet (and every other address type)
// reads that topmost line as leftover rather than street, exactly as
// documented — nothing downstream should try to place tokens a reading left
// unaccounted for back into that same reading. But a firm name is not part of
// the reading; it is a fact about the input the reading was never going to
// carry, which is why Address has its own field for it. So this runs after
// choose, not inside it: it looks for a leftover run that is the whole first
// line and nothing more, which is the one shape leftover tokens can take
// without implicating the reading choose already made.
//
// A two line input (street, last line) has nothing above the street to be a
// firm, so this only fires from three lines up.
func firmLine(tokens []token.Token, best *address.CandidateAddress) string {
	if len(tokens) == 0 || tokens[len(tokens)-1].Line < 2 {
		return ""
	}

	var topLine int
	for _, t := range tokens {
		if t.Line != 0 {
			break
		}
		topLine++
	}
	if topLine == 0 {
		return ""
	}

	for _, span := range best.Leftover {
		if span.Start == 0 && span.Length == topLine {
			return token.Join(tokens[:topLine])
		}
	}
	return ""
}

// streetWindow slides the left edge of the street right, one token at a time,
// and takes the first window the data recognizes, rewriting best in place.
//
// Step 7 of #17. The right edge is already fixed — the last line closed it in
// step 2 and step 6 closed it again at the secondary designator — so the only
// thing left undecided about the street is where it begins. The grammar
// anchors that at the first token of the address lines and offers no other
// reading: every candidate ordinarystreet builds for 411 N CENTRAL AVE
// preceded by two other words puts those two words inside the street name.
// Measured; it is the same shape leadingSecondary found and the reason both
// of these are re-reads rather than tie-breaks.
//
// Data promotes and never rejects, which here means three things. It runs
// only with reference data on. It runs only when the grammar's own reading is
// not one zipcity recognizes, so a street the data has seen is never slid off
// in favour of a shorter one it has also seen. And it takes a window only on
// an agrees — a contradiction, a split answer, or a question zipcity declined
// all leave the grammar's reading exactly where it was. Where the data has
// nothing to say, and against the TIGER gaps that are the normal case for a
// small ZIP Code, nothing here fires at all and the grammar decides.
//
// What the window slides past is not discarded. The tokens ahead of the
// street are the shape firmLine already reads as a firm name, so they land in
// BusinessName — but only where the reading has not already found one, since
// a firm on its own line above is the better claim on that field.
func (p *Parser) streetWindow(tokens []token.Token, best *address.CandidateAddress) {
	if !recognizable(best.Address) {
		return
	}
	r := reference{}
	if ans, ok := streetAgreement(r, best.Address); ok && ans == agrees {
		return
	}

	for start, end := 1, token.LineEnd(tokens, 0); start < end; start++ {
		window := p.readLine(tokens[start:])
		if window == nil || !recognizable(window.Address) {
			continue
		}
		if ans, ok := streetAgreement(r, window.Address); !ok || ans != agrees {
			continue
		}
		if window.Address.BusinessName == "" {
			window.Address.BusinessName = token.Join(tokens[:start])
		}
		*best = *window
		return
	}
}

// recognizable reports whether a reading's street names a place zipcity could
// have seen. It is streetAgreement's own type guard, asked ahead of the
// question so a window is never adopted on a closed form's pseudo-name — see
// streetAgreement for why "PO BOX" is not a street.
func recognizable(a *address.Address) bool {
	civic, ok := a.Type.(address.DataDependentAddressType)
	return ok && civic.IsCivicAddressType() && a.StreetName != ""
}

// leadingSecondary moves a secondary designator written ahead of the street
// to where the standard puts it, rewriting best in place.
//
// Step 6 of #17: the secondary designator and its details sit between the
// street and the city, and finding them is what fixes the right edge of the
// street window step 7 slides. APARTMENT 3200 152 SOUTH TECH DR writes the
// designator on the wrong side of the number, which Publication 28 §213 does
// not allow on a standardized record but real intake data does constantly.
//
// The grammar cannot recover from it. ordinarystreet reads a street line left
// to right, so a leading designator is not a designator to it — the whole run
// becomes one street name, the primary number is never found, and there is no
// candidate in the pile holding the right answer for choose to pick. Measured:
// APARTMENT 3200 152 SOUTH TECH DR offers exactly two readings and both keep
// APARTMENT in the name. So this is a re-read rather than a tie-break, which
// is why it runs here beside firmLine rather than inside choose.
//
// The guards are what keep it from firing on an address that means what it
// says. Only the catchall type is touched: a post office box, a rural route
// and the rest have each read their own line whole and there is nothing about
// them for a designator to be ahead of. Only a reading with no primary number
// of its own, which is the symptom the leading designator causes. Only a
// numbered designator — APT 3200, # 3200, PMB 456 — since a bare designator
// word at the start of a line is far more often a street name (KEY WEST is
// the vocabulary's own example). And only when the shorter re-read finds a
// street and claims no secondary of its own, so a designator is never stacked
// on a reading that already has one.
func (p *Parser) leadingSecondary(tokens []token.Token, claims []claim.Claim, best *address.CandidateAddress) {
	a := best.Address
	if _, ok := a.Type.(*ordinarystreet.OrdinaryStreetAddress); !ok {
		return
	}
	// Detail is excluded here, not folded into the rest.Address.SecondaryDesignator
	// check below: a leading designator stacked on a reading whose own tail
	// already claimed Detail (a trailing PMB) is a private-mailbox conflict
	// this function has no standing to resolve. SecondaryDesignator is not
	// excluded here on purpose — UNIT 3200 152 TECH DR ROOM 12 reads with no
	// primary number of its own (the documented symptom) and ROOM 12 already
	// in SecondaryDesignator, because the whole leading run was swallowed into
	// StreetName. primaryWithin below is what actually tests for "no primary
	// number of its own"; this field is the wrong proxy for that and bailing
	// on it here hid the stacked-designator case entirely.
	if a.Detail != "" {
		return
	}

	unit, width := leadingUnit(claims)
	if width == 0 || width >= len(tokens) || tokens[width-1].Line != tokens[0].Line {
		return
	}
	if !primaryWithin(best, width) {
		return
	}

	rest, _ := p.read(tokens[width:])
	if rest == nil || rest.Address == nil || rest.Address.StreetName == "" {
		return
	}
	if rest.Address.Detail != "" {
		return
	}

	// The re-read's own tail can already hold a secondary designator of its
	// own — UNIT 3200 152 TECH DR ROOM 12 re-reads as 152 TECH DR with ROOM
	// 12 already in SecondaryDesignator/SecondaryNumber. The standard still
	// has one designator slot, so the leading designator does not displace
	// it; it folds in ahead of it as a compound SecondaryNumber, the same
	// shape FormatStreetLine already joins with single spaces. The leading
	// designator's own word becomes the slot's designator, since it is the
	// one that sits closest to the street per Publication 28 §213.
	if rest.Address.SecondaryDesignator != "" {
		rest.Address.SecondaryNumber = textutil.JoinNonEmpty(" ",
			unit.number, rest.Address.SecondaryDesignator, rest.Address.SecondaryNumber)
		rest.Address.SecondaryDesignator = unit.designator
	} else {
		rest.Address.SecondaryDesignator = unit.designator
		rest.Address.SecondaryNumber = unit.number
		rest.Address.Detail = unit.detail
	}
	rest.Address.BusinessName = a.BusinessName
	*best = *rest
}

// primaryWithin reports whether a reading spent nothing outside the leading
// run on its primary address number — either it found none, or the tokens it
// found one in are the very tokens the designator wants back.
//
// #3200 152 SOUTH TECH DR is why this is a span test and not "has no primary
// number". ordinarystreet reads #3200 as the primary number and 152 SOUTH
// TECH DR as the street, which is a complete account of the line and looks
// nothing like the symptom the guard above is watching for. But Publication
// 28 has no primary number written with a pound sign — §213.2 makes # the
// secondary designator of unspecified type, and secondaryunit claims that
// token Exact as exactly that — so the number the reading found is the unit,
// in the one place the standard does not put it. Peeling it moves it; nothing
// is lost, because the run it came out of is the run being reassigned.
func primaryWithin(best *address.CandidateAddress, width int) bool {
	if best.Address.PrimaryNumber == "" {
		return true
	}
	for _, held := range best.Claims {
		for _, part := range held.Parts {
			if part.Part == claim.PartPrimaryNumber && part.End() > width {
				return false
			}
		}
	}
	return true
}

// unitParts is a numbered secondary designator read off one claim, in the
// fields Address keeps it in.
type unitParts struct {
	designator string
	number     string
	detail     string
}

// leadingUnit is the widest numbered secondary designator or private mailbox
// starting at the first token, and how many tokens it covers, or a zero width
// when there is none.
//
// Widest first because the vocabularies offer every extent they can support
// and the longer reading is the one that accounts for more of the input;
// secondaryunit's own numbered claim is two tokens where privatemailbox's can
// be three. A designator claimed without a number is not offered here at all
// — see leadingSecondary for why.
func leadingUnit(claims []claim.Claim) (unitParts, int) {
	var best unitParts
	width := 0

	for _, held := range claims {
		if held.Start() != 0 || held.End() <= width {
			continue
		}
		var unit unitParts
		ok := true
		for _, part := range held.Parts {
			switch part.Part {
			case claim.PartSecondaryDesignator:
				unit.designator = part.Value
			case claim.PartSecondaryNumber:
				unit.number = part.Value
			case claim.PartDetail:
				unit.detail = part.Value
			default:
				ok = false
			}
		}
		if !ok || (unit.number == "" && unit.detail == "") {
			continue
		}
		best, width = unit, held.End()
	}

	return best, width
}

// lastLines is the readings of the last line the data leaves standing.
//
// lastline offers every reading the grammar supports, and where an address
// writes no comma that is most of the tokens ahead of the region: 123 NORTH
// PARK ST, PAUL, MN 55102 is offered with PAUL for its city, with ST PAUL,
// with PARK ST PAUL, and with no city at all. The grammar cannot narrow that
// and does not try. The data can, and this is the only place it can do it
// before the street question is asked — see #17 step 2, and go-projectusat#61
// for why asking afterwards is too late.
//
// So: when the data confirms any reading's city, the readings it does not
// confirm fall away, and when it confirms none they all stand and the grammar
// decides. That is the promotion rule from #17 — a window the data has seen
// beats one it has not, and where it has seen none nothing is discarded — and
// it is why a real city the Census missed still parses.
//
// Nothing about a surviving reading is rewritten. Confidence in particular is
// left exactly as lastline wrote it, so the comma in 123 MAIN ST WEST, PALM
// BEACH, FL still separates PALM BEACH from WEST PALM BEACH when the data has
// seen both: corroborating a reading the data likes would lift the unmarked
// one to the marked one's confidence and erase the only thing telling the pair
// apart. Measured; it is not a hypothetical.
func (p *Parser) lastLines(r reference, tokens []token.Token, claims []claim.Claim) []lastline.LineClaim {
	lines := lastline.LineClaims(tokens, claims)

	var confirmed []lastline.LineClaim
	for _, line := range lines {
		ans, ok := cityAgreement(r, linePart(line, claim.PartPostal), linePart(line, claim.PartCity), linePart(line, claim.PartRegion))
		if ok && ans == agrees {
			confirmed = append(confirmed, line)
		}
	}
	if len(confirmed) == 0 {
		return lines
	}
	return confirmed
}

// linePart reports what a last line reading says a part is, or "" when the
// reading does not claim that part at all — a {Region} {Postal Code} reading
// has no city, and that is a reading rather than a gap.
func linePart(line lastline.LineClaim, part claim.Part) string {
	for _, p := range line.Claim.Parts {
		if p.Part == part {
			return p.Value
		}
	}
	return ""
}

// cityAgreement asks the strongest city question the fields support: whether
// the ZIP Code and the city pair, or — when there is no ZIP Code to pair the
// city with — whether the state has that city at all. A city known for this
// ZIP Code is stronger evidence than a city known somewhere in the state, so
// the second is asked only when the first cannot be.
func cityAgreement(r reference, postal, city, region string) (agreement, bool) {
	if ans, ok := zipCityAgreement(r, postal, city); ok {
		return ans, true
	}
	return cityZipsAgreement(r, postal, city, region)
}

// choose ranks candidates and returns the best, or nil when none survive.
//
// The score decides, coverage breaks ties — between two equally scored
// readings the one stranding fewer tokens is the better account of the input
// — and the street's own grammar breaks what is left, see streetConfidence.
// The score is this package's own integer — the grammar's rung (see rung)
// plus one for every question reference data agrees with and minus one for
// every question it contradicts, uncapped in either direction — see score.
// The data orders readings against each other and never rates one:
// CandidateAddress.Confidence is left exactly as the grammar wrote it, and
// Parse returns the Address alone, so no score this package computes is ever
// visible to a caller.
//
// A single agreement lifts a reading by one rung, which is enough to let it
// pass a reading one rung above it that the data contradicts — that
// inversion is the point of asking at all — but not enough to pass a reading
// two rungs above it that nothing was asked about. Two readings of one street
// line can differ only in which field holds a word — PENNSYLVANIA AVE as a
// name against PENNSYLVANIA with AVE as its suffix — and reference data
// cannot see that: both ask it the same key and get the same answer, so both
// scores shift by the same amount. The grammar's own gap between them
// survives the shift, and where the candidate's minimum had already closed
// that gap it is streetConfidence, below, that reopens it.
func (p *Parser) choose(r reference, candidates []*address.CandidateAddress) *address.CandidateAddress {
	type ranked struct {
		candidate *address.CandidateAddress
		answers   questionAnswers
		score     int
	}

	scored := make([]ranked, 0, len(candidates))
	for _, c := range candidates {
		if c == nil || c.Address == nil {
			continue
		}

		answers := p.agreement(r, c.Address)
		scored = append(scored, ranked{candidate: c, answers: answers})
	}
	if len(scored) == 0 {
		return nil
	}

	// A contradicts answer counts against a candidate only when some
	// competing candidate's answer to that same question is agrees —
	// go-projectusat#176. zipcity's gaps are real, so a question nothing in
	// the pile could confirm is exactly as uninformative as one never asked:
	// a Pub 28 placeholder street like KRYTON TN or a made-up one like
	// ZQXVBORK must not lose to a candidate that merely declined to ask,
	// which is what charging an unconfirmed contradicts always did before.
	// Gating is per question kind rather than per literal zipcity key, since
	// candidates in one pile are alternate readings of the same input and
	// rarely share a query string even when they are asking the same
	// question in spirit. Two candidates that both contradict the same
	// unconfirmed question (TestStreetContradictionDemotesButStillParses)
	// are ungated by the same amount, so the tie between them survives
	// exactly as it did when both were charged -1.
	var cityConfirmed, streetConfirmed bool
	for _, s := range scored {
		if s.answers.city == agrees {
			cityConfirmed = true
		}
		if s.answers.street == agrees {
			streetConfirmed = true
		}
	}
	for i := range scored {
		scored[i].score = score(scored[i].candidate.Confidence, scored[i].answers.gated(cityConfirmed, streetConfirmed))
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if a, b := len(scored[i].candidate.Leftover), len(scored[j].candidate.Leftover); a != b {
			return a < b
		}
		return streetConfidence(scored[i].candidate) > streetConfidence(scored[j].candidate)
	})
	return scored[0].candidate
}

// streetConfidence is how strongly the street line alone is held: the minimum
// over the claims carrying a street field, or the candidate's own confidence
// where a reading has no street line to hold — a post office box is as
// strongly held as it already says it is, and this tie-break must not rate it
// again.
//
// It exists because CandidateAddress.Confidence is the minimum over every
// accepted claim, and the minimum is routinely decided by a claim the tied
// readings share. 123 OCEAN BOULEVARD WEST PALM BEACH FL is read both as
// OCEAN BOULEVARD and as OCEAN with the suffix BLVD, and ordinarystreet does
// rate the second a rung above the first — but the city on that unmarked last
// line is held one rung lower than either, so both candidates come out at the
// city's confidence and the grammar's own gap is invisible to the sort. The
// same minimum flattens 3253 W 9200 S WEST JORDAN UT into a street named
// W 9200 S.
//
// So this is not a second opinion about the address; it is the first one,
// read off the part the tied readings actually disagree about. It is the last
// key rather than the second, and both of the keys above it matter: reference
// data still decides, and coverage still decides, which is what keeps a
// reading that strands the whole street line from winning on the confidence
// it is left with by not reading one — measured, two of the standard's own
// fragments do exactly that when this key is asked first.
func streetConfidence(c *address.CandidateAddress) claim.Confidence {
	found := false
	lowest := claim.ConfidenceExact

	for _, held := range c.Claims {
		if !holdsStreet(held) {
			continue
		}
		if !found || held.Confidence < lowest {
			found, lowest = true, held.Confidence
		}
	}
	if !found {
		return c.Confidence
	}
	return lowest
}

// holdsStreet reports whether a claim carries any of the four fields a street
// name is made of — the same four streetForQuery renders, and for the same
// reason: they are the street, and the primary number and secondary unit are
// what sits on it.
func holdsStreet(held claim.Claim) bool {
	for _, part := range held.Parts {
		switch part.Part {
		case claim.PartPredirectional, claim.PartStreetName, claim.PartStreetSuffix, claim.PartPostdirectional:
			return true
		}
	}
	return false
}

// rung maps a grammar confidence onto this package's own 0..3 scale, in the
// order go-projectusat's claim package defines them: Weak, Likely, Strong,
// Exact.
func rung(c claim.Confidence) int {
	switch {
	case c >= claim.ConfidenceExact:
		return 3
	case c >= claim.ConfidenceStrong:
		return 2
	case c >= claim.ConfidenceLikely:
		return 1
	default:
		return 0
	}
}

// score is this package's ranking of one reading: the grammar's rung plus
// one for every agrees and minus one for every contradicts, uncapped in
// either direction. missingInZip, missingInCity, and unknown add nothing —
// see agreement's doc comment for why a split or a decline is not evidence.
//
// It is uncapped because there is nothing here left for a cap to protect.
// The old scale stepped claim.Confidence itself and stopped short of
// ConfidenceExact so that agreement could never manufacture certainty, but
// that cap also meant a reading rated Exact that the data contradicted
// (stepped down to Strong) tied a reading rated Strong that the data agreed
// with (capped at Strong) — and the tie-break on the grammar's own rating
// then picked the one the data had just spoken against. Scoring on this
// package's own integer instead of on Confidence removes the seam: nothing
// is capped, so the ordering the data is asked to make is the one it gets.
func score(c claim.Confidence, answers []agreement) int {
	s := rung(c)
	for _, ans := range answers {
		switch ans {
		case agrees:
			s++
		case contradicts:
			s--
		}
	}
	return s
}

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

	if ans, ok := cityAgreement(r, a.Postal, a.City, a.Region); ok {
		qa.city = ans
	}
	if ans, ok := streetAgreement(r, a); ok {
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
func zipCityAgreement(r reference, postal, city string) (agreement, bool) {
	m := zip5.FindStringSubmatch(postal)
	if m == nil || city == "" {
		return unknown, false
	}

	present, err := r.check("zip city "+m[1]+" "+city, func() (bool, error) {
		return zipcity.CheckZipAndCity(m[1], city)
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
func cityZipsAgreement(r reference, postal, city, region string) (agreement, bool) {
	if zip5.MatchString(postal) || city == "" || len(region) != 2 {
		return unknown, false
	}

	present, err := r.check("city zips "+region+" "+city, func() (bool, error) {
		for range zipcity.ZipsKnownFor(region, city) {
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
// so their false positives are independent: about 0.005^2 for both to be
// spurious, against 0.005 for either alone. On the truth side the two are
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
func streetAgreement(r reference, a *address.Address) (agreement, bool) {
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
			match, err := zipcity.MatchZipAndStreet(m[1], street)
			return presentMatch(match), err
		})
		if err != nil {
			return unknown, false
		}
		inZip = present
	}
	if hasCity {
		present, err := r.check("city street "+a.City+" "+a.Region+" "+street, func() (bool, error) {
			match, err := zipcity.MatchCityStateAndStreet(a.City, a.Region, street)
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
	if normalized, err := queryNormalizer.Normalize(street); err == nil {
		street = normalized
	}
	return textutil.JoinNonEmpty(" ", street.Predirectional, street.StreetName, street.StreetSuffix, street.Postdirectional)
}

// queryNormalizer renders a street the way the caller will write it. It holds
// no state between calls.
var queryNormalizer = normalizer.NewContentNomalizer()

// answerFor turns a bloom filter's boolean into the agreement it represents.
func answerFor(present bool) agreement {
	if present {
		return agrees
	}
	return contradicts
}
