// Package claim defines the contract a vocabulary package uses to report what
// it believes a run of tokens means.
//
// It is deliberately a leaf: it describes claims in terms of token indices and
// depends on nothing else in the library, so any package can produce claims and
// any package can consume them without an import cycle. Tokens supply the
// coordinate system a claim is expressed in; this package supplies the meaning.
package claim

// Part names an address component a Claim can be made against. The values
// mirror the fields of address.Address; they live here rather than in that
// package so vocabulary packages can make claims without importing it.
type Part string

const (
	PartBusinessName        Part = "business name"
	PartPrimaryNumber       Part = "primary number"
	PartPredirectional      Part = "predirectional"
	PartStreetName          Part = "street name"
	PartStreetSuffix        Part = "street suffix"
	PartPostdirectional     Part = "postdirectional"
	PartSecondaryDesignator Part = "secondary designator"
	PartSecondaryNumber     Part = "secondary number"
	PartCity                Part = "city"
	PartRegion              Part = "region"
	PartPostal              Part = "postal"
	PartCountry             Part = "country"
)

// Confidence is how strongly a vocabulary package believes a run of tokens is
// the Part it claims. The scale is shared across packages, so a caller can
// compare a region claim against a street suffix claim on the same tokens and
// get a meaningful answer. Use the named values rather than bare numbers.
type Confidence int

const (
	// ConfidenceExact: the token can only be read this way within the
	// vocabulary. A two-letter postal abbreviation, a ZIP code, a known
	// secondary unit designator.
	ConfidenceExact Confidence = 100

	// ConfidenceStrong: a canonical name or a documented alias, where the
	// vocabulary is sure of the reading but the token could carry ordinary
	// English meaning elsewhere in the address.
	ConfidenceStrong Confidence = 90

	// ConfidenceLikely: a real match that the vocabulary knows is contested —
	// something else in the library can legitimately claim the same tokens.
	ConfidenceLikely Confidence = 75

	// ConfidenceWeak: a shape match only, with no entry in the vocabulary
	// behind it. An unrecognized alphanumeric that merely looks postal.
	ConfidenceWeak Confidence = 50
)

// Claim is one vocabulary package's reading of a run of tokens.
//
// A package returns every reading it can support, including readings it
// believes are unlikely, and does not attempt to choose between them or to
// consider what the neighbouring tokens are. Deciding which claims survive is
// the parser's job: a claim is evidence, not an assignment.
//
// Because a package may return more than one Claim over the same tokens, and
// because claims from different packages may overlap, the full set of claims
// for an address describes every interpretation of it that the library can
// see. That is what lets the parser offer alternative interpretations rather
// than only its best guess.
type Claim struct {
	// Start is the index into the []Token the claim was made against.
	Start int

	// Length is how many consecutive tokens the claim covers. Always at least
	// 1. Multi-token spans are ordinary: "SOUTH CAROLINA" is one region claim
	// of length 2, "FEDERATED STATES OF MICRONESIA" one of length 4.
	Length int

	// Part is the address component being claimed.
	Part Part

	// Confidence is how strongly this reading is held. See the constants.
	Confidence Confidence

	// Value is the normalized text this claim would put in the component, so a
	// caller that accepts the claim does not have to repeat the lookup.
	Value string
}

// End returns the index one past the last token the claim covers.
func (c Claim) End() int {
	return c.Start + c.Length
}

// Overlaps reports whether two claims cover any of the same tokens. Overlapping
// claims are competing readings: at most one can be accepted.
//
// A claim of zero Length covers no tokens and so overlaps nothing, including
// itself. Length is documented as always at least 1, but the interval
// arithmetic alone would report an empty span as overlapping every span that
// contains its Start, which would make a malformed claim look like a competing
// reading of real tokens.
func (c Claim) Overlaps(other Claim) bool {
	if c.Length < 1 || other.Length < 1 {
		return false
	}

	return c.Start < other.End() && other.Start < c.End()
}
