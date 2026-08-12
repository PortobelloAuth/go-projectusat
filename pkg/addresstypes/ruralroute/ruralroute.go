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
var alphanumspace = regexp.MustCompile("[^0-9A-Z ]+")
var ruralroutePattern = regexp.MustCompile(`^RR [1-9A-Z][0-9A-Z]* BOX [0-9A-Z]+`)
var routeHashPattern = regexp.MustCompile(`R(D|R|FD|FD ROUTE|URAL ROUTE)\s*(#|NUMBER|NUM|NO)\s*`)
var boxHashPattern = regexp.MustCompile(`([0-9A-Z]+)\s*(BOX\s+)?(#|NUMBER|NUM|NO)\s*`)
var whitespace = regexp.MustCompile(`\s+`)
var leadingzero = regexp.MustCompile(`(RR|BOX)\s*0+`)

var recognizedDesignators = []string{
	"RURAL ROUTE",
	// NOTE: longer values get replaced first!
	"RFD ROUTE",
	"RFD",
	"RR",
	"RD",
}
var ruralrouteReplacements = slices.Collect(func(yield func(string) bool) {
	for _, v := range recognizedDesignators {
		if !yield(v) {
			return
		}
		if !yield("RR") {
			return
		}
	}
})

var ruralrouteReplacer = strings.NewReplacer(ruralrouteReplacements...)

func Normalize(sn string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(sn)
	capitalized = routeHashPattern.ReplaceAllString(capitalized, "R$1 ")
	capitalized = boxHashPattern.ReplaceAllString(capitalized, "$1 BOX ")
	capitalized = alphanumspace.ReplaceAllString(capitalized, "")
	capitalized = whitespace.ReplaceAllString(capitalized, " ")

	replaced := ruralrouteReplacer.Replace(capitalized)
	replaced = leadingzero.ReplaceAllString(replaced, "$1 ")

	suffix := ruralroutePattern.ReplaceAllString(replaced, "")
	replaced, _ = strings.CutSuffix(replaced, suffix)

	// See if we replaced anything
	if ruralroutePattern.MatchString(replaced) {
		return replaced, nil
	}

	return "", fmt.Errorf("Not a recognized rural route")
}

func NewRuralRoute(a *address.Address) (*address.Address, error) {
	// TODO: check a to see if the type fits
	a.Type = &RuralRouteAddress{}
	return a, nil
}

type RuralRouteAddress struct{}

func (rr *RuralRouteAddress) FormatStreetLine(a *address.Address) string {
	return textutil.JoinNonEmpty(" ",
		a.Predirectional,
		a.StreetName,
		a.PrimaryNumber,
		a.Postdirectional,
		a.SecondaryDesignator,
		a.SecondaryNumber,
	)
}
