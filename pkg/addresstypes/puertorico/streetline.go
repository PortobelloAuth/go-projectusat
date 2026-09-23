package puertorico

import (
	"fmt"
	"regexp"
	"strings"
)

// primaryNumberForm matches the primary address number that opens a Puerto
// Rico street line.
//
// Three shapes appear in the standard's examples (pp. 26-27). A plain number,
// 1234 CALLE AURORA. A letter and a number, A17 CALLE AMAPOLA, which the
// standard also writes hyphenated, A-17 CALLE AMAPOLA, and may carry a
// trailing letter, B-17A CALLE 1. And a block and house joined by a hyphen,
// 199-31 CALLE 19, where both halves are numbers.
//
// The two hyphens mean different things, which is why they are separate
// alternatives rather than one optional hyphen. The one after a letter is
// punctuation inside a single number and comes out, A-17 to A17. The one
// between two numbers carries the block and the house and stays.
var primaryNumberForm = regexp.MustCompile(`^(?:[0-9]+(?:-[0-9]+)?|[A-Z]-?[0-9]+[A-Z]?)$`)

// letterPrefixed matches the shape whose hyphen is punctuation.
var letterPrefixed = regexp.MustCompile(`^[A-Z]-[0-9]+[A-Z]?$`)

// NormalizeStreetLine reads a Puerto Rico street line, returning the primary
// address number and the street name.
//
// The shape is the standard's, p. 26: "Spanish street names generally have the
// suffix element preceding the root street name, making it a prefix." So the
// line is a primary address number, then a street type, then the root name —
// 1234 CALLE AURORA, 585 AVE FD ROOSEVELT, 1025 PARQUE DEL REY — and the type
// is part of the name rather than a suffix that moved.
//
// Every type is spelled out, AVE included: an abbreviated type on input is
// expanded and never the other way around. p. 26 says "Developers MUST NOT
// abbreviate street names" and "MUST NOT translate CALLE to the suffix ST";
// p. 24 separately permits "the word AVENIDA or its abbreviation AVE" in this
// position. Both sentences describe the same text, and a MUST NOT over the
// same ground outranks a MAY, so AVE is not carved out as an exception.
// Aaron's ruling, on #117.
//
// The reference data settles it independently. zipcity's Pub28FeatureName
// spells the type out when it renders Puerto Rico TIGER records — its
// spanishPrefixOverrides maps AVE -> AVENIDA
// (internal/ustigerline/featnames/featnames.go:32), applied both to the base
// name (:145) and to the coded prefix type (:189). Observed there, under
// STATEFP 72: "AVE FD ROOSEVELT" -> "AVENIDA FD ROOSEVELT". A street this
// library emitted as AVE could never match that index entry.
//
// The cost is real and known, not a defect: p. 26's own example, 585 AVE FD
// ROOSEVELT, stops being a fixed point of this library — NormalizeStreetLine
// now returns "AVENIDA FD ROOSEVELT" for it, not "AVE FD ROOSEVELT".
//
// The root name is everything after the type. It is arbitrary text — the
// standard validates it against nothing — so the caller's line boundary is
// what says where it ends, the same way it is for an urbanization name.
//
// An error means these are not the tokens of a Puerto Rico street line, per
// CONTRIBUTING §1.6.
func NormalizeStreetLine(s string) (primaryNumber, streetName string, err error) {
	fields := strings.Fields(strings.ToUpper(strings.TrimSpace(s)))

	// A number, a type and at least one word of root name. Two fields are a
	// number and a bare type, which names no street.
	if len(fields) < 3 {
		return "", "", fmt.Errorf("Not a Puerto Rico street line")
	}

	number, ok := normalizePrimaryNumber(fields[0])
	if !ok {
		return "", "", fmt.Errorf("Not a Puerto Rico street line")
	}

	// The type is always spelled out, per the doc comment above, so this is a
	// direct call rather than a wrapper with a since-removed exception.
	streetType, err := NormalizeStreetType(fields[1])
	if err != nil {
		return "", "", err
	}

	return number, streetType + " " + strings.Join(fields[2:], " "), nil
}

// normalizePrimaryNumber returns the primary address number as the standard
// writes it, with the punctuating hyphen of a letter-prefixed number removed.
func normalizePrimaryNumber(field string) (string, bool) {
	if !primaryNumberForm.MatchString(field) {
		return "", false
	}

	if letterPrefixed.MatchString(field) {
		return strings.Replace(field, "-", "", 1), true
	}

	return field, true
}
