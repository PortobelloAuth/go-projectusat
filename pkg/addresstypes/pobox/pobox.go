package pobox

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

/*
Post Office Box addresses in a patient record MUST be standardized as PO Box ___

	POST OFFICE BOX 11890 -> PO BOX 11890
	POST OFFICE BOX G     -> PO BOX G

Developers MUST NOT add a leading zero before the post office box number.
PO Box addresses often appear with the words CALLER, FIRM CALLER, BIN, LOCKBOX, or DRAWER,
or other synonyms. When this occurs, developers MUST change these words to PO BOX in the patient
record.
PO Box services in some locations allow for an option to use the Post Office street address for the
address, along with the PO Box number preceded by a “#” sign or “UNIT” designation.
*/
var poboxPattern = regexp.MustCompile(`^PO BOX [0-9A-Z]+$`)
var hashPattern = regexp.MustCompile(`\s*#\s*`)

var recognizedDesignators = []string{
	"POST OFFICE BOX",
	"PO BOX",
	"POB",
	// NOTE: longer values get replaced first!
	"FIRM CALLER",
	"CALLER",
	"BIN",
	"LOCKBOX",
	"DRAWER",
}
var poboxReplacements = slices.Collect(func(yield func(string) bool) {
	for _, v := range recognizedDesignators {
		if !yield(v) {
			return
		}
		if !yield("PO BOX") {
			return
		}
	}
})

var poboxReplacer = strings.NewReplacer(poboxReplacements...)

func Normalize(sn string) (string, error) {
	// NOTE / FIXME?: Normalize is currently used by Claims to recognize a whole pobox
	// street line, not just the different ways we can write PO BOX (the street name.)
	// This means it cannot be used to normalize as street name, which does not have a
	// primary number incorporated (as CheckStreetLine() and the regex it now wraps
	// expect.) As a result, this code has some duplication in NormalizeStreetName()
	// while we figure out what can or should change.

	// capitalize
	capitalized := strings.ToUpper(sn)
	capitalized = hashPattern.ReplaceAllString(capitalized, " ")

	replaced := poboxReplacer.Replace(capitalized)

	// See if we recognized a PO Box
	if CheckStreetLine(replaced) {
		return replaced, nil
	}

	return "", fmt.Errorf("Not a recognized PO Box")
}

func NormalizeStreetName(streetname string) (string, error) {
	capitalized := strings.ToUpper(streetname)
	capitalized = hashPattern.ReplaceAllString(capitalized, " ")

	replaced := poboxReplacer.Replace(capitalized)

	if replaced != "PO BOX" {
		return "", fmt.Errorf("Not a recognized PO Box street name")
	}

	return replaced, nil
}

func CheckStreetLine(streetline string) bool {
	return poboxPattern.MatchString(streetline)
}
