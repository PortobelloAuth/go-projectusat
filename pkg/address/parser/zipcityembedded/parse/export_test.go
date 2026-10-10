package parse

import "github.com/PortobelloAuth/go-projectusat/pkg/address"

// Score and Rung are the ranking arithmetic, exposed for tests.
//
// Parse returns only an address, so the score that decided between readings
// is never observable through the public API — it is internal by design, not
// an oversight. Testing the arithmetic directly is the only way to pin down
// that a single agreement cannot carry a reading across a real grammar gap.
var (
	Score = score
	Rung  = rung
)

// Agreement and the functions that produce one are exposed for the same
// reason: which questions agreement asks, and which single street question a
// reading qualifies for, is discarded the same way the score is. Testing
// zipCityAgreement and streetAgreement directly is the only way to pin down
// that a candidate with no ZIP asks MatchCityStateAndStreet rather than
// silently asking nothing, without depending on which reading choose happens
// to rank first.
type Agreement = agreement

const (
	Unknown       = unknown
	Agrees        = agrees
	Contradicts   = contradicts
	MissingInZip  = missingInZip
	MissingInCity = missingInCity
)

// Reference is the memoisation cache choose builds once per call, exposed so
// a test can ask the same question through it twice and confirm zipcity only
// saw it once.
type Reference = reference

// Check calls reference's unexported check method, exposed for the same
// reason Reference is.
func Check(r Reference, key string, query func() (bool, error)) (bool, error) {
	return r.check(key, query)
}

// ZipCityAgreement and StreetAgreement wrap a fresh Reference, since the
// existing tests exercise one reading at a time and have no cache of their
// own to pass in.
func (p *Parser) ZipCityAgreement(a *address.Address) (agreement, bool) {
	return p.zipCityAgreement(reference{}, a.Postal, a.City)
}
func (p *Parser) CityZipsAgreement(a *address.Address) (agreement, bool) {
	return p.cityZipsAgreement(reference{}, a.Postal, a.City, a.Region)
}
func (p *Parser) StreetAgreement(a *address.Address) (agreement, bool) {
	return p.streetAgreement(reference{}, a)
}

var (
	StreetForQuery    = streetForQuery
	StreetConfidence  = streetConfidence
	FoldStreetAnswers = foldStreetAnswers
	PresentMatch      = presentMatch
)
