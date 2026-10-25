package parse

import (
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

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
	// The ZIP holds the street and the city name does not: the street is read
	// right and the city is the suspect, so there is nothing to slide past.
	if ans, ok := p.streetAgreement(r, best.Address); ok && (ans == agrees || ans == missingInCity) {
		return
	}

	for start, end := 1, token.LineEnd(tokens, 0); start < end; start++ {
		window := p.readLine(tokens[start:], false)
		if window == nil || !recognizable(window.Address) {
			continue
		}
		if ans, ok := p.streetAgreement(r, window.Address); !ok || ans != agrees {
			continue
		}
		if window.Address.BusinessName == "" {
			window.Address.BusinessName = token.Join(tokens[:start])
		}
		*best = *window
		return
	}
}

// spellOut spells out a conflicted abbreviation in the street name — MT, ST,
// CT, NE, DR — where the name data settles which word it stands for, re-reads
// the tokens with that word written in full, and returns the tokens it read.
//
// go-projectusat#235. MT TREASURE is MONTANA TREASURE and MT VERNON is MOUNT
// VERNON, and nothing in the token tells the two apart, so no rule in the
// normalizer can be both right and idempotent. zipcity's conflicted
// abbreviation filter can: it lists the spellings the data has seen the word
// stand for beside its neighbors. A word is rewritten only where exactly one
// listed spelling is longer than the word itself, so this only ever spells an
// abbreviation out — never MONTANA into MOUNT — and a word the data has seen
// either way, or not at all, is left as written for the normalizer, which
// keeps the abbreviation.
//
// Every word of the street name is asked about, each beside its neighbors in
// the name: MT TREASURE is MONTANA TREASURE, ROCKY MT is ROCKY MOUNT, and
// SPRING ST NW is SPRING STREET NORTHWEST. A word the data has seen as the
// suffix or directional it is, or not at all, has no single longer reading and
// is left alone.
//
// The re-read is kept unless the street the data recognized as written is
// lost by spelling it out.
func (p *Parser) spellOut(tokens []token.Token, best *address.CandidateAddress) []token.Token {
	spelled := slices.Clone(tokens)
	changed := false
	name := covered(best, claim.PartStreetName)
	for _, i := range name {
		var prev, next string
		if slices.Contains(name, i-1) {
			prev = strings.ToUpper(tokens[i-1].Text)
		}
		if slices.Contains(name, i+1) {
			next = strings.ToUpper(tokens[i+1].Text)
		}
		word := strings.ToUpper(tokens[i].Text)
		readings, err := p.zipcity.ConflictedAbbreviationReadings(word, prev, next)
		if err != nil {
			return tokens
		}
		// Readings come back abbreviation first. A word that is among them
		// without leading them is already a spelling (COURT of CT COURT
		// CONNECTICUT), not one to rewrite. A word the data has only seen
		// spelled out is absent from them (MT beside ZION: MOUNT), and that is
		// the abbreviation.
		if i := slices.Index(readings, word); i > 0 {
			continue
		}
		longer := slices.DeleteFunc(readings, func(r string) bool { return len(r) <= len(word) })
		if len(longer) == 1 {
			spelled[i].Text = longer[0]
			changed = true
		}
	}
	if !changed {
		return tokens
	}

	reread := p.readLine(spelled, true)
	if reread == nil {
		return tokens
	}
	r := reference{}
	if p.agrees(r, best.Address) && !p.agrees(r, reread.Address) {
		return tokens
	}
	*best = *reread
	return spelled
}

// fusedNumber is a number with letters run onto its end: 400E, 199HENRY,
// 133ST, 133RDST; fusedPredirectional is one whose letters are a directional:
// 400E. fusedDirectional is a directional run onto the front of a number:
// E135TH, E115.
var (
	fusedNumber         = regexp.MustCompile(`^(\d+)([A-Z]+)$`)
	fusedPredirectional = regexp.MustCompile(`^(\d+)(NE|NW|SE|SW|N|S|E|W)$`)
	fusedDirectional    = regexp.MustCompile(`^(NE|NW|SE|SW|N|S|E|W)(\d+)([A-Z]*)$`)
)

// unfuse splits a directional, a street name or a suffix off a number it was
// typed against, where the street data holds the street that results, and
// returns the tokens it read.
//
// go-projectusat#231, #232 and #233. 400E 30ST is 400 E 30TH ST, 199HENRY ST
// is 199 HENRY ST, and E135TH ST is E 135TH ST, but 123A is a primary number
// as written and nothing in the token says which is which. The data does: a
// split is taken only where zipcity recognizes the street it results in. Data
// promotes and never rejects, so where the data has nothing to say the tokens
// stand.
//
// Where the street as typed is already one the data has, only a directional
// run onto the primary number is split off, and only into a predirectional
// the data has on that street: 400E 30TH ST is 400 E 30TH ST because E 30TH
// ST exists. The data does not track primary numbers, so it has no evidence
// for 400E and some for E.
//
// Each fused token in the primary number or street name offers its splits,
// and the combinations are re-read with the most splits first, up to
// maxUnfused re-reads.
func (p *Parser) unfuse(tokens []token.Token, best *address.CandidateAddress) []token.Token {
	if !recognizable(best.Address) {
		return tokens
	}
	r := reference{}
	typed := p.agrees(r, best.Address)
	primary := covered(best, claim.PartPrimaryNumber)

	var choices []fused
	var predirectional string
	name := covered(best, claim.PartStreetName)
	for _, i := range covered(best, claim.PartPrimaryNumber, claim.PartStreetName) {
		words := splits(tokens[i].Text)
		if slices.Contains(name, i) {
			words = append(words, ordinals(tokens[i].Text)...)
		}
		if typed {
			words = nil
			if m := fusedPredirectional.FindStringSubmatch(tokens[i].Text); m != nil && slices.Contains(primary, i) {
				words, predirectional = [][]string{{m[1], m[2]}}, m[2]
			}
		}
		if len(words) > 0 {
			choices = append(choices, fused{i, words})
		}
	}
	// A primary number led by a directional is the Wisconsin grid form (N165;
	// see grid), so it is not split; only a street name has a directional
	// fused onto it by mistake (E135TH).
	choices = slices.DeleteFunc(choices, func(f fused) bool {
		return fusedDirectional.MatchString(tokens[f.at].Text) && !slices.Contains(name, f.at)
	})

	// Combinations grow as a power of the fused tokens; past a few there is no
	// typo left to recover, only a run of codes.
	if len(choices) > maxFused {
		return tokens
	}

	// Every combination of one split or none per token, the most splits first.
	var combinations [][]int
	var walk func(picked []int)
	walk = func(picked []int) {
		if len(picked) == len(choices) {
			combinations = append(combinations, slices.Clone(picked))
			return
		}
		for k := range len(choices[len(picked)].words) + 1 {
			walk(append(picked, k))
		}
	}
	walk(nil)
	sort.SliceStable(combinations, func(a, b int) bool {
		return splitCount(combinations[a], choices) > splitCount(combinations[b], choices)
	})

	reads := 0
	for _, c := range combinations {
		if splitCount(c, choices) == 0 || reads == maxUnfused {
			break
		}
		reads++
		split := tokens
		for n := len(choices) - 1; n >= 0; n-- {
			if c[n] < len(choices[n].words) {
				split = splice(split, choices[n].at, choices[n].words[c[n]])
			}
		}
		reread := p.readLine(split, true)
		if reread == nil || !recognizable(reread.Address) || !p.held(r, reread.Address) {
			continue
		}
		if typed && reread.Address.Predirectional != predirectional {
			continue
		}
		*best = *reread
		return split
	}
	return tokens
}

// respace closes up or opens one space inside a street name, where the street
// data holds the street that results better than it holds the street as
// typed, and returns the tokens it read.
//
// go-projectusat#231. BLUEJAY DR and BLUE JAY DR are one street, and so are
// EASTCHESTER RD and EAST CHESTER RD, but the data holds each in one form and
// nothing in the tokens says which. So, like unfuse, a reading is taken only
// where the data holds the street it results in (#231: normalize toward what
// the data holds). A predirectional is a word of the street like any other
// here: EAST of EAST CHESTER may be the front of EASTCHESTER.
//
// Every space a word could be split at, and every space between two words, is
// offered, so a word of ten letters is asked about eight times, and a false
// positive somewhere among them is not rare. What keeps one from deciding is
// that a reading must be held exactly for its ZIP Code, not by a directional
// variant (each Match asks up to seventeen keys; see zipcity's Match), and
// held better than the street as typed (see holding),
// and better than every other respacing: two readings held alike are the data
// unable to say, and the tokens stand. And the space is one inside the name
// (see inside), so a respacing that makes a suffix or a directional of part of
// a word is not one: BLACKPOOL COURT is not a street named BLACKP OOL COURT,
// nor SOUTH AVE the street SO UTH AVE, however the filters answer.
func (p *Parser) respace(tokens []token.Token, best *address.CandidateAddress) []token.Token {
	if !recognizable(best.Address) {
		return tokens
	}
	r := reference{}
	typed := p.holding(r, best.Address)
	if typed == heldBoth {
		return tokens
	}
	street := covered(best, claim.PartPredirectional, claim.PartStreetName)

	var found []token.Token
	var reading *address.CandidateAddress
	top, tied := typed, false
	for _, respaced := range respacings(tokens, street) {
		reread := p.readLine(respaced, true)
		if reread == nil || !recognizable(reread.Address) || !inside(best.Address, reread.Address) {
			continue
		}
		h := p.holding(r, reread.Address)
		if !p.exactInZip(r, reread.Address) {
			h = heldNot
		}
		switch {
		case h > top:
			found, reading, top, tied = respaced, reread, h, false
		case h == top && found != nil:
			tied = true
		}
	}
	if found == nil || tied {
		return tokens
	}
	*best = *reading
	return found
}

// exactInZip reports whether the ZIP Code's street filter holds a reading's
// street as written, without trying its directional variants.
func (p *Parser) exactInZip(r reference, a *address.Address) bool {
	m := zip5.FindStringSubmatch(a.Postal)
	if m == nil {
		return false
	}
	street := streetForQuery(a)
	present, err := r.check("zip street exact "+m[1]+" "+street, func() (bool, error) {
		return p.zipcity.CheckZipAndStreet(m[1], street)
	})
	return err == nil && present
}

// inside reports whether a respacing kept within the street name: the suffix
// and postdirectional are as typed, and so is the predirectional, unless the
// name took it in (EAST CHESTER RD is EASTCHESTER RD).
func inside(typed, respaced *address.Address) bool {
	return respaced.StreetSuffix == typed.StreetSuffix &&
		respaced.Postdirectional == typed.Postdirectional &&
		(respaced.Predirectional == typed.Predirectional || respaced.Predirectional == "")
}

// holding is how well the data holds a reading's street: for its ZIP Code and
// city name both, for its ZIP Code alone (see held), or not at all.
type holding int

const (
	heldNot holding = iota
	heldZip
	heldBoth
)

func (p *Parser) holding(r reference, a *address.Address) holding {
	switch ans, ok := p.streetAgreement(r, a); {
	case ok && ans == agrees:
		return heldBoth
	case ok && ans == missingInCity:
		return heldZip
	}
	return heldNot
}

// word is a run of letters respace may split or join; numbers are unfuse's.
var word = regexp.MustCompile(`^[A-Z]+$`)

// respacings lists the tokens with one space closed up between two adjacent
// words of the street, or one opened inside a word of it, leaving at least
// two letters either side.
func respacings(tokens []token.Token, street []int) [][]token.Token {
	var out [][]token.Token
	for k, i := range street {
		text := strings.ToUpper(tokens[i].Text)
		if !word.MatchString(text) {
			continue
		}
		for cut := 2; cut <= len(text)-2; cut++ {
			out = append(out, splice(tokens, i, []string{text[:cut], text[cut:]}))
		}
		if k+1 < len(street) && street[k+1] == i+1 && tokens[i+1].Line == tokens[i].Line {
			if next := strings.ToUpper(tokens[i+1].Text); word.MatchString(next) {
				out = append(out, join(tokens, i, text+next))
			}
		}
	}
	return out
}

// join replaces the tokens at i and i+1 with one token, text, renumbering the
// positions after it on that line; it is splice run the other way.
func join(tokens []token.Token, i int, text string) []token.Token {
	out := slices.Clone(tokens[:i+1])
	out[i].Text = text
	for _, t := range tokens[i+2:] {
		if t.Line == tokens[i].Line {
			t.Position--
		}
		out = append(out, t)
	}
	return out
}

// held reports whether the data holds a reading's street for its ZIP Code:
// agrees, or the ZIP holds it and only the city name is the suspect, the case
// streetWindow already leaves alone. NY, NY 10030 and NEW YORK, NY 10030 are
// one W 143RD ST, and a split the ZIP confirms is not one the city name can
// veto, or the same street settles one way under one city name and another
// under the other.
func (p *Parser) held(r reference, a *address.Address) bool {
	ans, ok := p.streetAgreement(r, a)
	return ok && (ans == agrees || ans == missingInCity)
}

// maxUnfused caps the re-reads unfuse makes for one address, and maxFused the
// fused tokens it will consider.
const (
	maxUnfused = 8
	maxFused   = 3
)

// fused is a token unfuse may split, at its index, with the ways to split it.
type fused struct {
	at    int
	words [][]string
}

// splitCount is how many tokens a combination splits; index len(words) is the
// token left as typed.
func splitCount(c []int, choices []fused) int {
	n := 0
	for i, k := range c {
		if k < len(choices[i].words) {
			n++
		}
	}
	return n
}

// splits lists the ways a fused token can be written as two, or none where it
// is not fused. An ordinal (133RD, 1ST) is not fused.
func splits(text string) [][]string {
	if m := fusedDirectional.FindStringSubmatch(text); m != nil {
		if m[3] == "" {
			return [][]string{{m[1], m[2]}, {m[1], ordinal(m[2])}}
		}
		return [][]string{{m[1], m[2] + m[3]}}
	}
	m := fusedNumber.FindStringSubmatch(text)
	if m == nil || ordinal(m[1]) == text {
		return nil
	}
	if th := ordinal(m[1]); strings.HasPrefix(text, th) {
		return [][]string{{th, text[len(th):]}}
	}
	return [][]string{{m[1], m[2]}, {ordinal(m[1]), m[2]}}
}

// bareNumber is a street name word that is only digits: the 121 of E 121 ST.
var bareNumber = regexp.MustCompile(`^\d+$`)

// ordinals lists the readings of a street name word that settle a number to
// the ordinal the data holds, or none where the word is not a number.
//
// go-projectusat#231 and #233. E 121 ST, E 121ST and E121ST are one street,
// and the data holds it as E 121ST ST and in no other form. A bare number is
// written as its ordinal: 121 is 121ST. And an ordinal whose own suffix is ST
// may be the ordinal with the street suffix run onto it, since 121ST and 121
// ST are written alike: 121ST is 121ST ST. A directional run onto the front
// (E121ST) is offered both with the split splits already makes. Nothing in
// the token says which reading is meant, so, like every split here, one is
// taken only where the data holds the street it results in (#231: normalize
// toward what the data holds).
func ordinals(text string) [][]string {
	if bareNumber.MatchString(text) {
		return [][]string{{ordinal(text)}}
	}
	d := ""
	if m := fusedDirectional.FindStringSubmatch(text); m != nil {
		d, text = m[1], m[2]+m[3]
	}
	m := fusedNumber.FindStringSubmatch(text)
	if m == nil || m[2] != "ST" || ordinal(m[1]) != text {
		return nil
	}
	if d != "" {
		return [][]string{{d, text, "ST"}}
	}
	return [][]string{{text, "ST"}}
}

// ordinal writes a number as the ordinal street names use: 1ST, 2ND, 13TH.
func ordinal(n string) string {
	tens, ones := "", n[len(n)-1:]
	if len(n) > 1 {
		tens = n[len(n)-2 : len(n)-1]
	}
	switch {
	case tens == "1":
		return n + "TH"
	case ones == "1":
		return n + "ST"
	case ones == "2":
		return n + "ND"
	case ones == "3":
		return n + "RD"
	}
	return n + "TH"
}

// splice replaces the token at i with one token per word, on the same line and
// after the same comma, renumbering the positions after it on that line.
func splice(tokens []token.Token, i int, words []string) []token.Token {
	out := slices.Clone(tokens[:i])
	for k, w := range words {
		t := tokens[i]
		t.Text, t.Position = w, t.Position+k
		if k > 0 {
			t.FollowsComma = -1
		}
		out = append(out, t)
	}
	for _, t := range tokens[i+1:] {
		if t.Line == tokens[i].Line {
			t.Position += len(words) - 1
		}
		out = append(out, t)
	}
	return out
}

// gridHalf is one half of the Wisconsin grid form: a directional and a number,
// N165 or W2123.
var gridHalf = regexp.MustCompile(`^(?i)([NSEW])\d+$`)

// grid joins the Wisconsin grid form written as two tokens, N165 W2123, into
// the one token Project US@ and Pub 28 write, N165W2123. The pair must lead
// its line and run on both axes, one half N or S and the other E or W.
func grid(tokens []token.Token) []token.Token {
	for i := 0; i+1 < len(tokens); i++ {
		a, b := tokens[i], tokens[i+1]
		ma, mb := gridHalf.FindStringSubmatch(a.Text), gridHalf.FindStringSubmatch(b.Text)
		if a.Position != 0 || b.Line != a.Line || b.FollowsComma != -1 || ma == nil || mb == nil ||
			strings.ContainsAny(ma[1], "NSns") == strings.ContainsAny(mb[1], "NSns") {
			continue
		}
		out := slices.Clone(tokens[:i+1])
		out[i].Text = a.Text + b.Text
		for _, t := range tokens[i+2:] {
			if t.Line == a.Line {
				t.Position--
			}
			out = append(out, t)
		}
		return out
	}
	return tokens
}

// covered lists, in order, the token indices a reading's claims assign to any
// of parts.
func covered(best *address.CandidateAddress, parts ...claim.Part) []int {
	var at []int
	for _, c := range best.Claims {
		for _, part := range c.Parts {
			if !slices.Contains(parts, part.Part) {
				continue
			}
			for i := part.Start; i < part.Start+part.Length; i++ {
				if !slices.Contains(at, i) {
					at = append(at, i)
				}
			}
		}
	}
	slices.Sort(at)
	return at
}

// heads lists the first token index of each claim part a reading assigns to
// part.
func heads(best *address.CandidateAddress, part claim.Part) []int {
	var at []int
	for _, c := range best.Claims {
		for _, cp := range c.Parts {
			if cp.Part == part {
				at = append(at, cp.Start)
			}
		}
	}
	return at
}
