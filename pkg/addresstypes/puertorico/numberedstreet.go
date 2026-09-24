package puertorico

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// numberedStreetNameFields is how many tokens the street name of a numbered
// street line covers: the street type and the number that names the street,
// "CALLE 19". Everything after them is the primary address number.
//
// It is a named constant rather than a 2 in streetLine because a claim over
// such a line has to say where the street name stops and the number begins,
// and deriving that from the returned strings would be reconstructing a part's
// extent from its value — the thing claim.ClaimPart.Value says not to do.
const numberedStreetNameFields = 2

// blockIdentifiers are the words p. 27 says a developer MUST NOT include.
//
//	Due to the amount of numbers within a block and a house number in Puerto
//	Rico addresses, many identifiers are commonly used to separate address
//	elements, including BLOQUE, NUM, NO, CASA, LOTE, or a # sign. These
//	identifiers MUST NOT be included in patient addresses.
//
// BLQ is not in that sentence; it is in the standard's own example row on the
// same page, "CALLE 19 BLQ 199 Casa 31". NUM covers NÚM because the spellings
// are folded before they are looked up, the same way ruralroute folds BUZÓN.
//
// The table is here rather than in addresstables, where the route words and
// the street types live, because these words have no standard form to hold: a
// Spelling/Standard row says what a word becomes, and what these become is
// nothing. BLOQUE also already has a row in that package's Secondaries, where
// it is a designator that is kept and abbreviated BL — a second row saying to
// delete it would make one word two things in one table, which is the reason
// puertorico.go gives for keeping URB out of secondaryMap.
//
// The order is the standard's, with BLQ beside the spelling it abbreviates.
// Nothing depends on it: every pattern derived from this table matches a whole
// field, so no alternative can win by being a prefix of another.
var blockIdentifiers = []string{"BLOQUE", "BLQ", "NUM", "NO", "CASA", "LOTE", "#"}

// numericForm matches the number that names a numbered street, and the numbers
// the identifiers introduce. Plain digits, deliberately: it is what separates a
// numbered street from a named one, and a named street's trailing token is not
// evidence of a house number.
var numericForm = regexp.MustCompile(`^[0-9]+$`)

// gluedIdentifier matches an identifier written onto the number it introduces,
// which the standard's own "Núm.18" does once the punctuation between them is
// gone. p. 30 states the equivalent rule for routes — a developer MUST have a
// space between the designator and its number — and the same input habit shows
// up here, so it is read rather than refused.
var gluedIdentifier = regexp.MustCompile(`^(` + identifierGroup() + `)([0-9]+)$`)

// identifierOnly matches an identifier standing on its own, ahead of the
// number it introduces.
var identifierOnly = regexp.MustCompile(`^(?:` + identifierGroup() + `)$`)

// identifierGroup renders blockIdentifiers as an alternation, so the
// vocabulary is written down once and the patterns are derived from it.
func identifierGroup() string {
	quoted := make([]string, len(blockIdentifiers))
	for i, word := range blockIdentifiers {
		quoted[i] = regexp.QuoteMeta(word)
	}

	return strings.Join(quoted, "|")
}

// unspellable matches everything a spelling may not contain once it is folded
// and capitalized. The # sign survives because p. 27 lists it as one of the
// identifiers; the hyphen and the period do not, because they are the
// punctuation "C-19" and "Núm.18" carry and the standard requires gone.
var unspellable = regexp.MustCompile(`[^0-9A-Z#]+`)

// NormalizeNumberedStreetLine reads the form pp. 26-27 print in their
// Incorrect Form column for a numbered street, and returns the primary address
// number and street name of the Correct Form beside it.
//
// The shape is the standard's own, and it is the reverse of the one
// NormalizeStreetLine reads: the street comes first and its house number
// follows, "CALLE 1 A17" for "A17 CALLE 1". Three rules on those two pages
// describe it.
//
//	Numbered streets MUST always contain the word CALLE. This avoids
//	misinterpretation between numbered streets and house numbers in patient
//	addresses.
//
//	House numbers may have fractional or alphabetic modifiers. Developers MUST
//	place the house number before the street name. When placing alphanumeric
//	house numbers prior to the street name, developers MUST NOT use hyphens to
//	separate the letter from the number.
//
//	When addresses contain block numbers and house numbers, developers MUST
//	use a hyphen to separate the block number from the house number.
//
// So the street name is a street type and a number, and what follows is either
// one house number — A17, B113, C-19 — or numbers introduced by the
// identifiers of blockIdentifiers, which come out joined by the hyphen the
// last rule asks for. Both halves of the primary number are read by
// normalizePrimaryNumber and the shapes it already documents; nothing new is
// written down about what a house number looks like.
//
// The number that names the street is what makes this reading safe to offer.
// A named street's trailing token — "CALLE AMAPOLA A17" — is as plausibly part
// of the name as a house number, and this package has nothing to settle that
// with, so it is refused rather than guessed. p. 26's "1510 CALLE 3 NO" is
// refused for the mirror reason: NO there is the Spanish directional for
// Northwest, and it introduces no number.
//
// Two deviations from the standard's printed examples, documented per
// CONTRIBUTING §2.
//
// First, p. 26 prints "13 CALLE 191" as the Correct Form of "CALLE 191 B113".
// That drops the B1 of B113, which no rule on either page licenses — the rules
// move the house number and remove a hyphen from it, they do not shorten it.
// This returns "B113", following the rule over the example the way #119 chose
// the rule over p. 30's printed "HC 1 BOX 1050". The corpus still carries the
// standard's row, so that one example is expected to fail until its Want is
// corrected, the way #125 corrected the other one.
//
// Second, the standard requires the hyphen for "up to three-digit numeric
// block numbers" and says nothing about a longer one. The hyphen is written
// whatever the block number's length: a four-digit block is not a case the
// standard rules on, and the alternative — joining 1234 and 18 into 123418 —
// would invent a house number that is in neither input nor standard.
//
// An error means these are not the tokens of a Puerto Rico numbered street
// line, per CONTRIBUTING §1.6.
func NormalizeNumberedStreetLine(s string) (primaryNumber, streetName string, err error) {
	fields := strings.Fields(strings.ToUpper(strings.TrimSpace(s)))

	// A type, the street's number, and at least one field of house number.
	if len(fields) < numberedStreetNameFields+1 {
		return "", "", fmt.Errorf("Not a Puerto Rico numbered street line")
	}

	streetType, err := NormalizeStreetType(fields[0])
	if err != nil {
		return "", "", err
	}

	if !numericForm.MatchString(fields[1]) {
		return "", "", fmt.Errorf("Not a Puerto Rico numbered street line")
	}

	number, ok := houseNumber(fields[numberedStreetNameFields:])
	if !ok {
		return "", "", fmt.Errorf("Not a Puerto Rico numbered street line")
	}

	return number, streetType + " " + fields[1], nil
}

// houseNumber reads the house number that trails a numbered street.
//
// A single field is the ordinary case and is the primary address number
// itself, so normalizePrimaryNumber answers it — including the hyphen "C-19"
// carries, which that function already knows is punctuation. Anything else has
// to be identifiers and the numbers they introduce, and the standard allows
// two of those: a block and a house.
func houseNumber(fields []string) (string, bool) {
	if len(fields) == 1 {
		if number, ok := normalizePrimaryNumber(fields[0]); ok {
			return number, true
		}
	}

	var numbers []string
	for len(fields) > 0 {
		number, rest, ok := identifiedNumber(fields)
		if !ok || len(numbers) == 2 {
			return "", false
		}

		numbers = append(numbers, number)
		fields = rest
	}

	if len(numbers) == 0 {
		return "", false
	}

	return strings.Join(numbers, "-"), true
}

// identifiedNumber reads one identifier and the number it introduces, and
// returns the fields left after them.
//
// The identifier is required. A bare number with nothing saying what it counts
// is the shape this reading must not accept: it is what would let "CALLE 19
// 199 31" or a stray token from a line above become a block and a house.
func identifiedNumber(fields []string) (number string, rest []string, ok bool) {
	head, err := spelling(fields[0])
	if err != nil {
		return "", nil, false
	}

	if glued := gluedIdentifier.FindStringSubmatch(head); glued != nil {
		return glued[2], fields[1:], true
	}

	if !identifierOnly.MatchString(head) || len(fields) < 2 {
		return "", nil, false
	}

	next, err := spelling(fields[1])
	if err != nil || !numericForm.MatchString(next) {
		return "", nil, false
	}

	return next, fields[2:], true
}

// spelling reduces one field to the form the identifier vocabulary is written
// in: diacritics folded, because the standard writes its own example as
// "Núm.18", and punctuation removed, because that is what separates the
// identifier from its number there.
func spelling(field string) (string, error) {
	folded, err := diacritics.Substitute(field)
	if err != nil {
		return "", err
	}

	return unspellable.ReplaceAllString(strings.ToUpper(folded), ""), nil
}
