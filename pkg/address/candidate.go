package address

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
)

// CandidateAddress is one address type's complete reading of an address.
//
// A Claim is a statement about a run of tokens and a LineClaim is a statement
// about a line. This is the statement about the whole thing: given these
// tokens, this is the address, read as a PO box or a rural route or a military
// address, and this is how much of it that reading accounts for.
//
// Every address type may produce candidates for the same tokens, and one type
// producing a candidate does not preclude another from producing its own — the
// same rule that governs Claims one level down. A type that has no reading to
// offer returns none, which is different from offering a weak one: PSC standing
// alone is not a poor military address, it is not a military address.
//
// Choosing between candidates is the parser's job, and it is most of what the
// parser has left to do.
type CandidateAddress struct {
	// Address is the reading. Its Type names the address type that produced it.
	Address *Address

	// Confidence is how strongly the whole reading is held, on the shared scale
	// in pkg/address/parser/claim.
	//
	// A candidate is only as good as the weakest reading in it, so this is the
	// minimum over the accepted claims and then one step down for each run of
	// leftover tokens.
	//
	// That is the opposite of how a last line is rated, and the difference is
	// the point. A line pattern is evidence in its own right: a five digit
	// number is a weak ZIP on its own and a strong one after a region at the
	// end of a line, because the pattern corroborates its members. There is no
	// pattern above the address doing that job. An address is the conjunction
	// of its lines and nothing vouches for the conjunction, so it inherits its
	// weakest part rather than rising above it.
	Confidence claim.Confidence

	// Claims are the claims this reading accepted, which is what lets a caller
	// audit the address rather than take it on trust. Claims the reading
	// rejected are not carried here; that record belongs to the line that
	// rejected them.
	Claims []claim.Claim

	// Leftover are the runs of tokens the reading does not account for.
	//
	// They are why a candidate can be complete and still be wrong. Tokens with
	// no sensible place to live are the clearest signal that an address type
	// has been applied to an address it does not fit, and each run costs the
	// candidate a step of confidence. They are not an assignment queue: nothing
	// downstream should try to place them, because a reading that could place
	// them would have.
	Leftover []claim.Span
}
