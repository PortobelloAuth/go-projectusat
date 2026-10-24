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

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
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
type Parser struct {
	zipcity *zipcity.ZipCity
}

// New returns a Parser.
func New() (*Parser, error) {
	zc, err := zipcity.New()
	if err != nil {
		return nil, err
	}

	return &Parser{
		zipcity: zc,
	}, nil
}

// Parse reads source into a structured address, returning ErrNoReading when no
// address type recognizes it.
//
// The stages are go-projectusat's, in the order that library defines: tokenize,
// let every vocabulary claim what it recognizes, read the last line from those
// claims, then ask each address type for its reading of the whole. Only the
// final choice among readings belongs to this package.
func (p *Parser) Parse(source string) (*address.Address, error) {
	tokens := grid(token.Tokenize(source))
	if len(tokens) == 0 {
		return nil, ErrNoReading
	}

	best := p.readLine(tokens, true)
	if best == nil {
		return nil, ErrNoReading
	}
	tokens = p.spellOut(tokens, best)
	tokens = p.unfuse(tokens, best)
	p.streetWindow(tokens, best)
	return best.Address, nil
}

// readLine is read with the two layers that rewrite a chosen reading rather
// than rank it: the firm name above the address, and a secondary designator
// written ahead of the street. It is separate from read because step 7 asks
// for a whole reading of a shorter run of tokens, not for a raw candidate.
func (p *Parser) readLine(tokens []token.Token, lastlinepatch bool) *address.CandidateAddress {
	best, claims := p.read(tokens, lastlinepatch)
	if best == nil || best.Address == nil {
		return nil
	}
	if firm := firmLine(tokens, best); firm != "" {
		best.Address.BusinessName = firm
	}
	p.leadingSecondary(tokens, claims, best, lastlinepatch)
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
func (p *Parser) read(tokens []token.Token, lastlinepatch bool) (*address.CandidateAddress, []claim.Claim) {
	var claims []claim.Claim
	for _, vocabulary := range vocabularies {
		claims = append(claims, vocabulary(tokens)...)
	}

	r := reference{}
	var candidates []*address.CandidateAddress
	lines := p.lastLines(r, tokens, claims)
	lowconfidence := true
	for _, line := range lines {
		// If all line claims are ConfidenceWeak it is a signal that we may have a street-line-only
		// address that still parsed a last line, perhaps because it had the postdirectional "NE",
		// which is also a region ("NEBRASKA".) In such low confidence scenarios, we'll try adding
		// a fake last line too and then score the results against each other.
		//
		// One of our tests has a street-line-only address with a postdirectional of NE. It will
		// look like a city in Nebraska, but the city is the number 20.
		if line.Claim.Confidence > claim.ConfidenceWeak {
			lowconfidence = false
		}
		for _, addressType := range addressTypes {
			candidates = append(candidates, addressType(tokens, claims, line)...)
		}
	}

	// Only add a last line for addresses that are not already multi-line
	if tokens[0].OfLines == 1 && (len(candidates) == 0 || lowconfidence) && lastlinepatch {
		if len(lines) > 0 {
			// add a candidate with just the last line parts
			candidates = append(candidates, lines[0].Candidate(nil, 0, []claim.Claim{}))
			ll := candidates[0]
			// all partial address candidates are weak(?)
			ll.Confidence = claim.ConfidenceWeak
		}

		// add a fake last line in LASTLINE, DC
		end := len(tokens)
		line := lastline.LineClaim{
			Span: lastline.Span{Start: end, Length: 0},
			Claim: claim.Claim{
				Confidence: claim.ConfidenceWeak,
				Parts: []claim.ClaimPart{
					{
						Start:  end,
						Length: 0,
						Part:   claim.PartCity,
						Value:  "LASTLINE",
					},
					{
						Start:  end,
						Length: 0,
						Part:   claim.PartRegion,
						Value:  "DC",
					},
				},
			},
			Rejected: []claim.Claim{},
			Leftover: []lastline.Span{},
		}
		for _, addressType := range addressTypes {
			candidates = append(candidates, addressType(tokens, claims, line)...)
		}
	}

	if len(candidates) == 0 {
		return nil, claims
	}
	chosen := p.choose(r, candidates)
	// clean up the fake last line, if it was used
	if chosen.Address.City == "LASTLINE" && chosen.Address.Region == "DC" {
		chosen.Address.City = ""
		chosen.Address.Region = ""
	}
	return chosen, claims
}
