// Package ordinarystreet reads the ordinary American street line, the one with
// no special format:
//
//	1600 PENNSYLVANIA AVE NW
//	123 MAIN ST APT 4
//	N6W23001 BLUEMOUND RD
//	43 E 200 N
//
// It is the default address type, and it is built differently from every other
// one in this tree. A post office box, a rural route and a military address are
// each a rigid pattern with its own recognizer, so each exports Claims saying
// "these tokens are my pattern". There is no pattern here. The street name is
// arbitrary text, and what bounds it is the vocabulary claims around it — the
// suffix, the directionals, the secondary unit — none of which this package
// owns.
//
// That is why this package exports no Claims. A Claims(tokens) signature would
// be a lie: this reading is a function of the other claims, not of the tokens,
// and there is nothing it could honestly say about tokens alone. It exports
// Candidates only, which already receives exactly what it needs. See #56 and
// #70 for the discussion that settled this.
//
// The street name is the residue: whatever lies between the primary number and
// the first accepted claim that ends the name. The primary number and the name
// are assigned together as one indivisible claim, because neither is
// recognizable without the other — 123 is a house number only because a name
// follows it on the same line, and the name is bounded only because the number
// opened the line.
package ordinarystreet

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
)

// OrdinaryStreetAddress is the AddressType for an ordinary street address.
//
// It carries no state, like every other AddressType: it names a shape, and one
// value of it serves every ordinary street address.
type OrdinaryStreetAddress struct{}

// FormatStreetLine renders "123 N MAIN ST SW APT 4".
//
// The ordering rule is not written here, because this package does not own it.
// address.Address.FormatStreetLine already owns it, as the branch it takes when
// an address has no special format — which is the definition of this type. So
// this delegates to that branch rather than restating the field order, and the
// copy without a Type is how the branch is reached. See CONTRIBUTING §1.2.
func (o *OrdinaryStreetAddress) FormatStreetLine(a *address.Address) string {
	ordinary := *a
	ordinary.Type = nil

	return ordinary.FormatStreetLine()
}
