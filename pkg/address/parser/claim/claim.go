// Package claim defines the contract a vocabulary package uses to report what
// it believes a run of tokens means.
//
// It is deliberately a leaf: it describes claims in terms of token indices and
// depends on nothing else in the library, so any package can produce claims and
// any package can consume them without an import cycle. Tokens supply the
// coordinate system a claim is expressed in; this package supplies the meaning.
package claim

import "cmp"

// Part names an address component a claim can be made against. The values
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
// what it claims. The scale is shared across packages, so a caller can compare
// a region claim against a street suffix claim on the same tokens and get a
// meaningful answer. Use the named values rather than bare numbers.
//
// Confidence belongs to the whole claim, not to its parts. What a vocabulary
// is sure or unsure of is the pattern it matched: "PO BOX 11890" is exact and
// "DRAWER 11890" is contested, and neither statement can be attributed to one
// token in the pattern rather than the other.
type Confidence int

const (
	// ConfidenceExact: the tokens can only be read this way within the
	// vocabulary. A two-letter postal abbreviation, a ZIP code, a known
	// secondary unit designator followed by its number.
	ConfidenceExact Confidence = 100

	// ConfidenceStrong: a canonical name or a documented alias, where the
	// vocabulary is sure of the reading but the tokens could carry ordinary
	// English meaning elsewhere in the address.
	ConfidenceStrong Confidence = 90

	// ConfidenceLikely: a real match that the vocabulary knows is contested —
	// something else in the library can legitimately claim the same tokens.
	ConfidenceLikely Confidence = 75

	// ConfidenceWeak: the vocabulary's own rule is satisfied by the shape of
	// the tokens alone, with no table entry confirming them. A five digit
	// number is a well formed ZIP whether or not it is an assigned one.
	//
	// This is the floor. A reading with neither a table entry nor a rule
	// behind it is not weak evidence, it is no evidence, and is not claimed.
	ConfidenceWeak Confidence = 50
)

// ClaimPart assigns a run of tokens within a Claim to a single address
// component. It is not meaningful on its own: what makes a ClaimPart true is
// the Claim it belongs to, and taking one out of that context is what the
// grouping exists to prevent.
type ClaimPart struct {
	// Start is the index into the []Token the claim was made against.
	Start int

	// Length is how many consecutive tokens this part covers. Always at least
	// 1. Multi-token parts are ordinary: in "RR 4 BOX 125" the street name is
	// "RR 4" and the primary number is "BOX 125", two tokens each.
	Length int

	// Part is the address component being assigned.
	Part Part

	// Value is the normalized text this part would put in the component, so a
	// caller that accepts the claim does not have to repeat the lookup.
	//
	// Value is what the part says; the tokens are what it covers. These are
	// not the same thing and a caller must not reconstruct one from the other.
	// Usually they differ only in spelling — "RFD ROUTE 4" is three tokens
	// valued "RR 4" — but a part may also cover tokens that contribute nothing
	// to its value at all.
	//
	// That happens where the standard says text should not be on a line and a
	// vocabulary owns the whole line: a rural route claims "RR 2 BOX 18 BRYAN
	// DAIRY RD" with a primary number valued "BOX 18" covering all of "BOX 18
	// BRYAN DAIRY RD". The trailing tokens are absorbed deliberately. Leaving
	// them unclaimed would offer them to another vocabulary as a street name,
	// which is the reading the standard rules out, and the contract has no way
	// to say "covered, and contributing nothing" — every covered token belongs
	// to some part.
	//
	// So absorption is how a claim says those tokens are spoken for. A parser
	// that accepts such a claim writes Value into the component and discards
	// the extra tokens; it must not treat them as unassigned. Where the extent
	// is uncertain the vocabulary offers both readings, and the absorbing one
	// is the weaker of the two.
	Value string
}

// End returns the index one past the last token this part covers.
func (p ClaimPart) End() int {
	return p.Start + p.Length
}

// Claim is one vocabulary package's reading of a run of tokens, and the unit a
// parser accepts or rejects.
//
// A claim assigns one or more parts, and they stand or fall together. A rural
// route address is a street name and a primary address number — "RR 4 BOX 125"
// is route 4, box 125 on it — and there is no coherent state in which a parser
// keeps the box without the route. A vocabulary that reported those separately
// would be inviting exactly that outcome, so the indivisibility is in the type
// rather than in a rule the parser is trusted to follow.
//
// A package returns every reading that could be correct, and does not attempt
// to choose between them. Deciding which claims survive is the parser's job: a
// claim is evidence, not an assignment.
//
// "Could be correct" is the boundary, and it is narrower than "anything the
// package can think of". A reading belongs in the result when something in the
// vocabulary supports it — a table entry, an alias, or a pattern the package
// owns. A reading the package can construct but has no basis for is not a low
// confidence claim; it is not a claim. Confidence ranks the readings that
// qualify, it is not a way to admit ones that do not.
//
// A pattern may span several tokens and consult what follows: "SOUTH DAKOTA"
// is one region claim of length 2, and a numbered designator claims its number
// with it, because a designator without its number is not weaker evidence of an
// address, it is misleading evidence. What a vocabulary may not do is decide
// which of its competing readings wins.
//
// Because a package may return more than one Claim over the same tokens, and
// because claims from different packages may overlap, the full set of claims
// for an address describes every interpretation of it that the library can
// see. That is what lets the parser offer alternative interpretations rather
// than only its best guess.
type Claim struct {
	// Confidence is how strongly this reading is held. See the constants.
	Confidence Confidence

	// Parts are the components this reading assigns. A claim with no parts
	// asserts nothing and covers no tokens.
	//
	// Two parts of the same claim must not cover the same token: within one
	// reading a token belongs to exactly one component. Parts of different
	// claims overlapping is the normal case, and is what Overlaps reports.
	Parts []ClaimPart
}

// Start returns the index of the first token this claim covers, or 0 if it
// covers none.
func (c Claim) Start() int {
	start := 0
	for i, p := range c.Parts {
		if i == 0 || p.Start < start {
			start = p.Start
		}
	}

	return start
}

// End returns the index one past the last token this claim covers, or 0 if it
// covers none.
func (c Claim) End() int {
	end := 0
	for i, p := range c.Parts {
		if i == 0 || p.End() > end {
			end = p.End()
		}
	}

	return end
}

// Length is how many tokens lie within the claim's extent.
//
// The extent is derived from the parts rather than stored alongside them, so a
// claim cannot assert a span wider than what it actually assigns. Deriving it
// also means the parts need not be given in token order.
func (c Claim) Length() int {
	return c.End() - c.Start()
}

// Overlaps reports whether two claims cover any of the same tokens. Overlapping
// claims are competing readings: at most one can be accepted.
//
// A claim with no parts covers no tokens and so overlaps nothing, including
// itself. The interval arithmetic alone would report an empty span as
// overlapping every span that contains its Start, which would make a claim
// that asserts nothing look like a competing reading of real tokens.
//
// Comparison is on the full extent. If a claim's parts ever leave a gap — none
// of the patterns in the library currently do — a claim sitting in that gap
// counts as competing with it. That is the conservative reading, and the one
// that keeps the parser from accepting two claims that disagree.
func (c Claim) Overlaps(other Claim) bool {
	if len(c.Parts) == 0 || len(other.Parts) == 0 {
		return false
	}

	return c.Start() < other.End() && other.Start() < c.End()
}

// Compare orders two competing readings of the same tokens, best first.
//
// Confidence decides it. Where two readings are held equally strongly, the
// longer one wins: "UNIT 2050" and "UNIT 2050 BOX 4190" are both military
// street lines and both exact, and the complete one is the reading meant. A
// vocabulary that matched more of the address on the same evidence has
// explained more of it.
//
// Length only ever breaks a tie, which is what keeps it from overriding the
// absorption rule documented on ClaimPart.Value. Where a vocabulary offers
// both a plain reading and a longer one that swallows trailing tokens, the
// contract requires the absorbing reading to be the weaker of the two —
// ruralroute offers the four token pattern as ConfidenceExact and the whole
// line as ConfidenceLikely — so confidence separates them before extent is
// consulted, and the greedier reading does not win by being greedy.
//
// Compare does not check Overlaps. Ordering claims that do not compete is
// harmless and useful, but only overlapping claims are alternatives, and
// deciding which of them to accept is still the parser's job.
//
// The result is negative if c is the better reading, positive if other is, and
// zero if neither is preferred. Zero means genuinely tied rather than
// identical: two vocabularies can reach the same confidence over the same
// extent and still disagree about what the tokens mean, and a parser must not
// read a tie as agreement.
//
// The signature is the one slices.SortFunc wants, so
//
//	slices.SortFunc(claims, Claim.Compare)
//
// sorts a set of readings best first.
func (c Claim) Compare(other Claim) int {
	if result := cmp.Compare(other.Confidence, c.Confidence); result != 0 {
		return result
	}

	return cmp.Compare(other.Length(), c.Length())
}
