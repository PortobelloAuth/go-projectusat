package parsertest

import (
	"fmt"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
)

// Result is one Case run through a parser.
type Result struct {
	Case Case
	Got  string
	Err  error
}

// Settled reports whether the case has ground truth to be scored against.
func (r Result) Settled() bool {
	return r.Case.Want != ""
}

// Pass reports whether the case's expectation was met. An unsettled case
// never passes and never fails; see Case.Want.
func (r Result) Pass() bool {
	return r.Settled() && r.Err == nil && r.Got == r.Case.Want
}

// Run scores p against cases by running each Input through
// goprojectusat.Normalize, with content normalization, and comparing the
// result to Want. It returns one Result per case, in the order given, so a
// caller can report by Source rather than only a total.
func Run(p parser.ParsingFunc, cases []Case) []Result {
	opts := []goprojectusat.USAtNormalizeOption{
		goprojectusat.WithCustomAddressParser(p),
		goprojectusat.WithContentNormalization(),
	}

	results := make([]Result, len(cases))
	for i, c := range cases {
		got, err := goprojectusat.Normalize(c.Input, opts...)
		results[i] = Result{Case: c, Got: got, Err: err}
	}
	return results
}

// CountPass returns how many of results passed, how many were settled enough
// to be scored at all, and the total. Reporting settled separately keeps an
// open question from reading as a parser's failure.
func CountPass(results []Result) (pass, settled, total int) {
	for _, r := range results {
		if r.Settled() {
			settled++
		}
		if r.Pass() {
			pass++
		}
	}
	return pass, settled, len(results)
}

// FieldResult is one field of an Address, comparing what a Case's
// WantFields asserted against what Parse returned for it. Want and Got are
// the field's rendering as a string even where the underlying field is not
// one (see the Type field, compared by which AddressType implementation is
// present rather than by value, the same way Address.Equals compares it).
type FieldResult struct {
	Field string
	Want  string
	Got   string
}

// Match reports whether the field matched what the case asserted.
func (f FieldResult) Match() bool {
	return f.Want == f.Got
}

// typeName renders an AddressType the way Address.Equals compares one: by
// which implementation is present, not by value.
func typeName(t address.AddressType) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%T", t)
}

// FieldsResult is one Case with a WantFields assertion, run through a
// parser's Parse and compared field by field. It is Result's counterpart
// for the decomposition rather than the rendering; see Case.WantFields for
// why the two are scored separately.
type FieldsResult struct {
	Case Case
	Got  *address.Address
	Err  error
}

// Settled reports whether the case has a decomposition to be scored
// against. Mirrors Result.Settled for WantFields instead of Want.
func (r FieldsResult) Settled() bool {
	return r.Case.WantFields != nil
}

// Fields compares every field of Case.WantFields against Got, including the
// fields WantFields leaves empty — an unasserted empty field is how a wrong
// decomposition sneaks through unnoticed, which is the reason this type
// exists (go-projectusat#123). It returns nil for an unsettled case, the
// same way there is nothing to compare an empty Want against.
//
// The fields are listed out explicitly rather than walked by reflection,
// mirroring Address.Equals: a field added to Address needs a line added
// here too, the same tradeoff Equals already made, and reflection would be
// a second way to answer the question Equals answers rather than a shared
// one — the DRY violation would be in the knowledge of how to compare an
// Address, not in the characters of the comparison.
func (r FieldsResult) Fields() []FieldResult {
	if !r.Settled() {
		return nil
	}
	got := r.Got
	if got == nil {
		got = &address.Address{}
	}
	want := r.Case.WantFields
	return []FieldResult{
		{Field: "Type", Want: typeName(want.Type), Got: typeName(got.Type)},
		{Field: "BusinessName", Want: want.BusinessName, Got: got.BusinessName},
		{Field: "Area", Want: want.Area, Got: got.Area},
		{Field: "PrimaryNumber", Want: want.PrimaryNumber, Got: got.PrimaryNumber},
		{Field: "Predirectional", Want: want.Predirectional, Got: got.Predirectional},
		{Field: "StreetName", Want: want.StreetName, Got: got.StreetName},
		{Field: "StreetSuffix", Want: want.StreetSuffix, Got: got.StreetSuffix},
		{Field: "Postdirectional", Want: want.Postdirectional, Got: got.Postdirectional},
		{Field: "SecondaryDesignator", Want: want.SecondaryDesignator, Got: got.SecondaryDesignator},
		{Field: "SecondaryNumber", Want: want.SecondaryNumber, Got: got.SecondaryNumber},
		{Field: "Detail", Want: want.Detail, Got: got.Detail},
		{Field: "City", Want: want.City, Got: got.City},
		{Field: "Region", Want: want.Region, Got: got.Region},
		{Field: "Postal", Want: want.Postal, Got: got.Postal},
		{Field: "Country", Want: want.Country, Got: got.Country},
	}
}

// Pass reports whether every field matched. An unsettled case, or one Parse
// returned an error for, never passes and never fails.
func (r FieldsResult) Pass() bool {
	if !r.Settled() || r.Err != nil {
		return false
	}
	for _, f := range r.Fields() {
		if !f.Match() {
			return false
		}
	}
	return true
}

// RunFields scores p against cases with a WantFields assertion. Unlike Run,
// which renders through goprojectusat.Normalize and can only compare the
// finished string, RunFields calls p.Parse directly and compares the
// returned *address.Address field by field — the fields are already there
// on every parse, Run just never looks at them. It returns one FieldsResult
// per case, in the order given, mirroring Run.
func RunFields(p parser.ParsingFunc, cases []Case) []FieldsResult {
	results := make([]FieldsResult, len(cases))
	for i, c := range cases {
		got, err := p.Parse(c.Input)
		results[i] = FieldsResult{Case: c, Got: got, Err: err}
	}
	return results
}

// CountPassFields is CountPass for RunFields' results: how many passed, how
// many were settled enough to be scored, and the total.
func CountPassFields(results []FieldsResult) (pass, settled, total int) {
	for _, r := range results {
		if r.Settled() {
			settled++
		}
		if r.Pass() {
			pass++
		}
	}
	return pass, settled, len(results)
}
