// Package ruralroute reads a Puerto Rico rural route or highway contract
// route address.
//
// It sits under puertorico rather than beside pkg/addresstypes/ruralroute
// because the two encode different rules from different sections of the
// standard, not one rule in two languages. Aaron's call on #117: "Please do
// not mix this with the English ruralroute package. They are analogous, but I
// don't think they share any actual logic."
//
// The three differences, p. 22 against p. 30:
//
//   - RFD, RD and RT become RR. On the mainland that is a SHOULD; in Puerto
//     Rico it is a MUST.
//   - The Spanish spellings — RURAL, RUTA RURAL, RUTA ESTRELLA, BUZON, BZN —
//     exist only here, and p. 30 states them as words a developer MUST NOT
//     leave in place.
//   - A sector name written with the route MUST be eliminated. The mainland
//     has no rule like it; the nearest thing, p. 22, only SHOULD NOT allow
//     additional designations. See Claims for what that costs.
//
// p. 30, verbatim:
//
//	A rural route address in the patient record MUST be standardized as
//	follows: RR___ BOX___
//
//	Developers MUST NOT use the words RURAL, RUTA RURAL, BUZON, or BZN. The
//	designations RFD, RD, and RT (meaning rural route) MUST be changed to RR
//	and developers MUST have a space between RR and the route number and BOX
//	and the box number.
//
//	Developers MUST NOT add a leading zero before the rural route number.
//
// and p. 30 again, for the highway contract route:
//
//	Highway contract route addresses MUST be standardized as HC____BOX____.
//	It is basically the same format utilized for rural routes. Likewise,
//	Health IT developers MUST NOT include leading zeros before the route
//	number.
//
// p. 31 adds one sentence to that, and nothing else: the rule against
// additional designations applies to highway contract routes too.
//
// The vocabulary itself is data and is held once, in
// github.com/poetic-systems/addresstables/puertorico, beside the street types
// and the secondary designators this package's parent already reads from
// there.
package ruralroute

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/poetic-systems/addresstables/puertorico"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// routeWords is the authoritative vocabulary, in the longest-first order
// addresstables guarantees. Order is load-bearing twice over: strings.Replacer
// prefers the earliest listed match at a position rather than the longest, so
// RURAL ahead of RUTA RURAL would leave RUTA RR behind, and the same is true
// of the scan in designatorLength.
var routeWords = slices.Collect(puertorico.RouteWords())

// standardDesignators are the distinct route designators a normalized line may
// open with. Deriving them keeps addresstables the only place a designator is
// written down.
var standardDesignators = distinct(func(w puertorico.RouteWord) (string, bool) {
	return w.Standard, w.Standard != "BOX"
})

// replacements pairs every spelling with the form the standard requires, for
// strings.Replacer. A standard form is its own spelling in the table, so this
// is one pass rather than two.
var replacements = slices.Collect(func(yield func(string) bool) {
	for _, w := range routeWords {
		if !yield(w.Spelling) || !yield(w.Standard) {
			return
		}
	}
})

var routeReplacer = strings.NewReplacer(replacements...)

var alphanumspace = regexp.MustCompile("[^0-9A-Z ]+")
var whitespace = regexp.MustCompile(`\s+`)

var designatorGroup = "(" + strings.Join(standardDesignators, "|") + ")"
var numberedGroup = "(" + strings.Join(append(slices.Clone(standardDesignators), "BOX"), "|") + ")"

// routePattern is the standardized form, "RR 3 BOX 9800". It anchors at the
// head and not at the tail: the standard's own sector examples put text after
// the pattern that a developer MUST eliminate, so a line that continues past
// it is still this address, with the remainder to be dropped.
var routePattern = regexp.MustCompile(`^` + designatorGroup + ` [1-9A-Z][0-9A-Z]* BOX [0-9A-Z]+`)

// gluednumber matches a designator written straight onto the number it
// introduces, which p. 30 forbids — "developers MUST have a space between RR
// and the route number and BOX and the box number". The standard's own RR03
// example would otherwise survive only because the leading zero happens to be
// rewritten with a space, and a route number starting with any other digit has
// no such luck.
var gluednumber = regexp.MustCompile(numberedGroup + `(\d)`)

var leadingzero = regexp.MustCompile(numberedGroup + `\s*0+`)

// Normalize returns the standardized form of a Puerto Rico route line, and an
// error when the text is not one.
//
// It is the recognizer as well as the formatter, per CONTRIBUTING §1.6: the
// rule for what counts as this address lives in one place, and a caller asks
// by handing over a span and reading the error.
//
// Diacritics are folded first, because the vocabulary is Spanish and the
// standard writes its own example as "Buzón". Then punctuation goes, which is
// what turns the standard's "17-A" into "17A" — a hyphen inside a box number
// is not part of the address here, where p. 22 lets a mainland one be.
func Normalize(line string) (string, error) {
	folded, err := diacritics.Substitute(line)
	if err != nil {
		return "", err
	}

	capitalized := strings.ToUpper(folded)
	capitalized = alphanumspace.ReplaceAllString(capitalized, "")
	capitalized = whitespace.ReplaceAllString(capitalized, " ")

	replaced := routeReplacer.Replace(strings.TrimSpace(capitalized))
	replaced = gluednumber.ReplaceAllString(replaced, "$1 $2")
	replaced = leadingzero.ReplaceAllString(replaced, "$1 ")

	standardized := routePattern.FindString(replaced)
	if standardized == "" {
		return "", fmt.Errorf("Not a Puerto Rico rural route or highway contract route")
	}

	return standardized, nil
}

// PuertoRicoRouteAddress is the AddressType for a Puerto Rico route address.
type PuertoRicoRouteAddress struct{}

// FormatStreetLine renders "RR 3 BOX 9800".
//
// The field order is p. 30's "RR___ BOX___", which the mainland rural route
// line happens to share — the standard says as much, "It is basically the same
// format utilized for rural routes". Writing it out again rather than reaching
// across to pkg/addresstypes/ruralroute is CONTRIBUTING §1.2: what is repeated
// is four field names, and what is not repeated is the knowledge, which is a
// different page of the standard with a different rule about what may follow
// the box number.
func (r *PuertoRicoRouteAddress) FormatStreetLine(a *address.Address) string {
	return textutil.JoinNonEmpty(" ", a.StreetName, a.PrimaryNumber)
}

// distinct collects the distinct values keep returns for the vocabulary, in
// the order they first appear in it.
func distinct(keep func(puertorico.RouteWord) (string, bool)) []string {
	var values []string
	for _, w := range routeWords {
		value, ok := keep(w)
		if !ok || slices.Contains(values, value) {
			continue
		}
		values = append(values, value)
	}

	return values
}
