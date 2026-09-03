package generaldelivery

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
)

/*
General delivery is mail held at a post office for a patient to collect, rather
than delivered to a place. The standard says:

	GENERAL DELIVERY
	Developers MUST use the words GENERAL DELIVERY, all uppercase, spelled out
	(no abbreviation), as the patient street address line in the patient record
	if the patient has a general delivery address. Each general delivery record
	SHOULD carry the -9999 add-on code. The ZIP Code or ZIP+4 Code MUST be
	correctly applied for patient addresses with a general delivery. Note that
	General Delivery is not available at every post office.

The whole street address line is those two words. There is no number, no
suffix, no directional and no secondary unit, because there is no place on a
street being described — the post office named by the last line is the
destination, and general delivery is the instruction for what to do with the
mail when it gets there.
*/

/*
Deviations, per CONTRIBUTING §2.

The abbreviated spellings are accepted on input and never produced on output.
The MUST above governs what a record holds, and this library's job on the way
in is to recognize what an integrator was actually sent — an intake system that
already emitted GENERAL DELIVERY needs no normalizer. GEN DEL is not invented
for that purpose: it is the abbreviation the standard itself uses for this
concept, in the Puerto Rico table that pairs Entrega General with GEN DEL. The
two intermediate spellings are the combinations of the halves, which cost
nothing to accept and would otherwise fail for no reason a reader could state.

The Spanish spelling ENTREGA GENERAL is deliberately absent, for the reason
given in ruralroute.go: the Puerto Rico vocabulary is one table, and taking a
single row out of it would leave that dialect half recognized. pkg/addresstypes/
puertorico already holds that row today. It classifies it as a secondary
designator rather than as a street line, which reads like a mistake against the
text quoted above, but correcting it is a change to that package's vocabulary
and not this one's — see the pull request that added this file.

The -9999 add-on code is not required here, and neither is a postal code at
all. Both belong to the last line, which this address type takes as read; see
Candidates for why. Note also that the add-on is a SHOULD and the ZIP is a MUST
about a well formed record, and this package is a recognizer, not a validator:
an address whose ZIP is missing is a general delivery address that is missing
its ZIP, and refusing to read it would lose the very fact a caller would need
in order to say so.
*/

// standardForm is the one spelling this package produces. See the MUST above.
const standardForm = "GENERAL DELIVERY"

// recognizedSpellings is the authoritative table of spellings a general
// delivery line may arrive as. All of them normalize to standardForm.
var recognizedSpellings = []string{
	standardForm,
	"GENERAL DEL",
	"GEN DELIVERY",
	"GEN DEL",
}

// maxSpan is the longest spelling above, measured in tokens. It bounds how far
// Claims looks ahead.
var maxSpan = slices.Max(slices.Collect(func(yield func(int) bool) {
	for _, s := range recognizedSpellings {
		if !yield(len(strings.Fields(s))) {
			return
		}
	}
}))

// punctuation is everything that is neither a letter, a digit nor a space, and
// it becomes a space rather than nothing. Deleting it would glue the halves of
// GENERAL-DELIVERY into one word that no spelling matches. Digits are kept for
// the opposite reason: they are not punctuation to be tidied away, they are
// evidence that this line says something more than the phrase, and dropping
// them would make "GENERAL DELIVERY 5" normalize as though the 5 were never
// there.
var punctuation = regexp.MustCompile(`[^0-9A-Z ]+`)
var whitespace = regexp.MustCompile(`\s+`)

// Normalize returns the standardized general delivery street line, or an error
// when sn is not one.
//
// The match is against the whole of sn rather than a prefix of it. Unlike a
// post office box or a rural route there is nothing that may follow — the
// standard makes these two words the entire street address line — so
// GENERAL DELIVERY 5 is not a general delivery line with a stray token, it is
// something else that happens to start the same way. Reading it as this shape
// would silently drop the 5, and a dropped token is how two different addresses
// come to look identical.
//
// Per CONTRIBUTING §1.6, an error means "not mine" and nothing more. It carries
// no part of sn, because an error is the value most likely to reach a log.
func Normalize(sn string) (string, error) {
	capitalized := strings.ToUpper(sn)
	capitalized = punctuation.ReplaceAllString(capitalized, " ")
	capitalized = whitespace.ReplaceAllString(capitalized, " ")
	capitalized = strings.TrimSpace(capitalized)

	if slices.Contains(recognizedSpellings, capitalized) {
		return standardForm, nil
	}

	return "", fmt.Errorf("Not a recognized general delivery address")
}

// GeneralDeliveryAddress is the AddressType for a general delivery address.
//
// Claims reads the line as a street name, because that is the field the
// standard's own words name — "the patient street address line" — and because
// it is the field a caller comparing two addresses will look in.
type GeneralDeliveryAddress struct{}

// FormatStreetLine renders "GENERAL DELIVERY".
//
// Every other field is deliberately dropped rather than appended. There is no
// shape here for a number or a unit to take part in, so a value in one of them
// did not come from this address type, and rendering it would produce a line
// the standard does not have. Detail is dropped for the reason military.go
// gives: it holds a private mailbox number, a private mailbox is rented from a
// commercial mail receiving agency, and mail a post office is holding for
// collection was never at one.
//
// A street name that is not a general delivery line renders as nothing. This
// address type has one line and knows what it says, so it cannot pass the value
// through the way a type with a variable street name would, and emitting
// GENERAL DELIVERY over the top of some other street would state something the
// address never said. An empty line is visibly wrong; a fabricated one is not.
func (g *GeneralDeliveryAddress) FormatStreetLine(a *address.Address) string {
	if _, err := Normalize(a.StreetName); err != nil {
		return ""
	}

	return standardForm
}
