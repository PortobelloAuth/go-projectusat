// Package privatemailbox recognizes the private mailbox number a patient holds
// at a commercial mail receiving agency, the element address.Address.Detail
// carries.
package privatemailbox

import (
	"fmt"
	"regexp"
	"strings"
)

/*
A private mailbox is a box rented at a commercial mail receiving agency (CMRA),
a private company rather than a post office. The standard says, under Private
Mailbox Addresses:

	Patient addresses at a CMRA MUST include either the PMB identifier or the
	numerical identifier, followed by the appropriate private mailbox number.
	Developers MUST NOT use any other identifiers.
	...
	Developers MUST NOT combine the secondary address element of the address
	for the CMRA and the CMRA patient's private box number.

and gives four examples: PMB 234 on its own line over RR 1 BOX 12 and over
10 MAIN ST STE 11, and trailing, as 123 MAIN STREET PMB 4545 and
PO BOX 159753 PMB 3571. Publication 28 §285 is the source and names the
numerical identifier: "either the PMB identifier or the #, followed by the
appropriate private mailbox number. Use of any other identifier is prohibited."

The street line is the CMRA's own address and the mailbox is the patient's.
That is why the mailbox is a field of its own, address.Address.Detail, beside a
secondary unit and never a second one; see pkg/addresstypes for which address
types admit it and where.
*/

/*
Deviations, per CONTRIBUTING §2.

Only PMB is ever produced. The # form is accepted on input and normalized to
PMB, because one output form per address is what makes two addresses
comparable, and because the exception in §285 requires PMB whenever the CMRA's
own line carries a secondary element. Ruled on #76.

The # form is claimed below Exact. Under Publication 28 §213.2 the same tokens
are a secondary unit of unspecified type, and pkg/secondaryunit claims them so
at Exact. That order is the intended one: without evidence otherwise a # is a
secondary unit (#78), and the evidence — a secondary unit already placed on the
line — is something only an address type can see. This package offers the
reading; the address type decides.

The number is not restricted beyond carrying a digit, for the reason
secondaryunit gives: what a CMRA numbers its boxes is not this library's to
say. # WEST is not offered because it is not worth losing with.
*/

// identifier is the one form this package produces. See the MUST above.
const identifier = "PMB"

// form is a permitted identifier followed by a number. The space is optional
// because the tokenizer keeps #234 as one token, and the standard's own example
// writes it that way.
var form = regexp.MustCompile(`^(?:PMB|#) ?([0-9A-Z]*[0-9][0-9A-Z]*)$`)

// maxSpan is the most tokens a private mailbox occupies: the identifier and
// the number. It bounds how far Claims looks ahead.
const maxSpan = 2

// Normalize returns the standard form of a private mailbox, PMB and its number,
// or an error meaning the text is not one. See CONTRIBUTING §1.6.
func Normalize(text string) (string, error) {
	m := form.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(text)))
	if m == nil {
		return "", fmt.Errorf("Not a private mailbox")
	}

	return identifier + " " + m[1], nil
}
