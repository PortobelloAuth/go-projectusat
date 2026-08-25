package ruralroute

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

/*
The rural route number in a patient record MUST be standardized as "RR ___ BOX ___"

	RURAL ROUTE 91 BOX A7       -> RR 91 BOX A7
	RFD 82 BOX 12               -> RR 82 BOX 12
	RD 51 # 25                  -> RR 51 BOX 25
	RFD Route 4 #87a          -> RR 4 BOX 87A
	RR 2 BOX 18 Bryan Dairy Rd  -> RR 2 BOX 18
	RR03 BOX 98D                -> RR 3 BOX 98D

Developers:
SHOULD NOT use the words RURAL, NUMBER, NO., or the pound sign (#).
MUST NOT add a leading zero before the rural route number.
SHOULD include hyphens as part of the box number only when they are part of the address.
SHOULD change the designations RFD and RD (as a meaning for rural or rural free delivery) to RR.
SHOULD NOT allow additional designations, such as town or street names, on the patient Street
Address Line of rural route addresses.
*/
/*
Highway contract routes share this package with rural routes.

The standard describes them in two places and defines them in neither. The
Business Addressing Standards section names a "rural/highway contract route
address (with route and box numbers)" in one breath, and the Puerto Rico
Addressing section gives the Spanish for one — Ruta Estrella <-> Highway
Contract — beside Ruta Rural <-> Rural Route. Both readings say the same thing
about its shape: a route designator, a route number, and a box.

That is the rural route pattern with a different designator, so it is
normalized by the rural route code rather than beside it. Making it its own
package would have meant a second copy of the number-marker handling, the
leading-zero rule and the box split, differing only in two letters — CONTRIBUTING
§1.2, where DRY is about knowledge and this is the same knowledge. Making it its
own address type would have meant a second FormatStreetLine that formats
identically, and a second look-alike for #52's shape-is-not-provenance hazard to
catch out.

What it is NOT is a highway name. pkg/highways normalizes COUNTY HIGHWAY 140 as
a street name, which is a road that an address sits on. A highway contract route
is a mail delivery route, and HC 4 is no more a highway than RR 4 is rural.

Deviation, per CONTRIBUTING §2: the accepted spellings below are the structural
parallels of the rural route ones. The standard's own list of spellings that
"SHOULD" be changed to HC is not quoted anywhere in this repository, so it has
not been checked against one. STAR ROUTE is a likely member — Ruta Estrella is
literally that — and is deliberately absent until the wording is confirmed.

The Spanish spellings are absent for the same reason they are absent from the
rural route half: Ruta Estrella, Ruta Rural, Buzon and Apartado are one Puerto
Rico vocabulary, and taking a single row out of that table would leave
RUTA ESTRELLA 4 BOX 12 recognized and RUTA ESTRELLA 4 BUZON 12 not.
*/

var alphanumspace = regexp.MustCompile("[^0-9A-Z ]+")
var boxHashPattern = regexp.MustCompile(`([0-9A-Z]+)\s*(BOX\s+)?(#|NUMBER|NUM|NO)\s*`)
var whitespace = regexp.MustCompile(`\s+`)

// designator is a spelling a route may be written with, and the standardized
// form it becomes.
type designator struct {
	Spelling string
	Standard string
}

// recognizedDesignators is the authoritative table of route spellings.
//
// NOTE: longer values must come first within a family! strings.Replacer and Go
// alternation both prefer the earliest listed match at a position, not the
// longest, so RFD before RR would leave RFD ROUTE half replaced.
var recognizedDesignators = []designator{
	{"RURAL ROUTE", "RR"},
	{"RFD ROUTE", "RR"},
	{"RFD", "RR"},
	{"RR", "RR"},
	{"RD", "RR"},
	{"HIGHWAY CONTRACT ROUTE", "HC"},
	{"HIGHWAY CONTRACT", "HC"},
	{"HCR", "HC"},
	{"HC", "HC"},
}

// standardDesignators are the distinct standardized forms, in the order they
// first appear above. Deriving them keeps the table the only place a designator
// is written down.
var standardDesignators = slices.Collect(func(yield func(string) bool) {
	var seen []string
	for _, d := range recognizedDesignators {
		if slices.Contains(seen, d.Standard) {
			continue
		}
		seen = append(seen, d.Standard)
		if !yield(d.Standard) {
			return
		}
	}
})

var spellings = slices.Collect(func(yield func(string) bool) {
	for _, d := range recognizedDesignators {
		if !yield(d.Spelling) {
			return
		}
	}
})

var routePattern = regexp.MustCompile(
	`^(` + strings.Join(standardDesignators, "|") + `) [1-9A-Z][0-9A-Z]* BOX [0-9A-Z]+`)

var routeHashPattern = regexp.MustCompile(
	`(` + strings.Join(spellings, "|") + `)\s*(#|NUMBER|NUM|NO)\s*`)

var leadingzero = regexp.MustCompile(
	`(` + strings.Join(append(slices.Clone(standardDesignators), "BOX"), "|") + `)\s*0+`)

var routeReplacements = slices.Collect(func(yield func(string) bool) {
	for _, d := range recognizedDesignators {
		if !yield(d.Spelling) {
			return
		}
		if !yield(d.Standard) {
			return
		}
	}
})

var routeReplacer = strings.NewReplacer(routeReplacements...)

func Normalize(sn string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(sn)
	capitalized = routeHashPattern.ReplaceAllString(capitalized, "$1 ")
	capitalized = boxHashPattern.ReplaceAllString(capitalized, "$1 BOX ")
	capitalized = alphanumspace.ReplaceAllString(capitalized, "")
	capitalized = whitespace.ReplaceAllString(capitalized, " ")

	replaced := routeReplacer.Replace(capitalized)
	replaced = leadingzero.ReplaceAllString(replaced, "$1 ")

	suffix := routePattern.ReplaceAllString(replaced, "")
	replaced, _ = strings.CutSuffix(replaced, suffix)

	// See if we replaced anything
	if routePattern.MatchString(replaced) {
		return replaced, nil
	}

	return "", fmt.Errorf("Not a recognized rural route or highway contract route")
}

func NewRuralRoute(a *address.Address) (*address.Address, error) {
	// TODO: check a to see if the type fits
	a.Type = &RuralRouteAddress{}
	return a, nil
}

type RuralRouteAddress struct{}

// FormatStreetLine renders "RR 4 BOX 125", with a secondary unit and a private
// mailbox number where the address carries them.
//
// Detail qualifies the secondary number and so follows it, exactly as in the
// ordinary street line. The standard's CMRA section shows a private mailbox
// over a rural route — "PMB 234" above "RR 1 BOX 12" — which is why this shape
// renders it at all. That example puts it on its own line; this library emits
// only the trailing form, because one output form per address is what makes
// two addresses comparable.
func (rr *RuralRouteAddress) FormatStreetLine(a *address.Address) string {
	// Rural Route Order: STREET PRIMARY SEC SECNUM DETAIL.
	// Other address parts are silently dropped.
	return textutil.JoinNonEmpty(" ",
		a.StreetName,
		a.PrimaryNumber,
		a.SecondaryDesignator,
		a.SecondaryNumber,
		a.Detail,
	)
}
