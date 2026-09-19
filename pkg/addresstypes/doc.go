// Package addresstypes holds one package per address format the standard
// distinguishes, and no code of its own. This comment is the one place that
// says what an address type is and what each may take from the vocabularies.
//
// Publication 28 and Project US@ describe two kinds of rule. Some live inside
// an address type: a military street line is a facility designator, its
// number, BOX and a box number, and nothing outside package military can say
// so. Others are shared: a directional, a street suffix, a secondary unit, a
// region, a postal code mean the same wherever they appear, and every type may
// lean on them. An address type is the statement of which pattern it owns and
// which shared claims it admits beside that pattern.
//
// The parser gathers every vocabulary's claims into one pool and hands the
// pool, with the tokens and a last line reading, to every type's Candidates.
// This table is what each type may take from the pool. Everything in the
// internal column is recognized from the tokens by the package itself; nothing
// in it is ever read off another package's claim, which is how a rural route
// and a military street line — the same shape, a street name over a primary
// number — are kept apart. See #52 and #70.
//
//	type             internal                            admits from the pool
//	military         whole line: CMR/OMC/PSC/UMR/UNIT    nothing
//	                 n BOX n; APO/FPO/DPO as the city;
//	                 AA/AE/AP as the region
//	generaldelivery  whole line: GENERAL DELIVERY        nothing
//	pobox            whole line: PO BOX n                a trailing Detail (PMB or #) *
//	ruralroute       whole line: RR/HC n BOX n           a trailing Detail (PMB or #)
//	puertorico       URB name; its own street types,     region and postal code
//	                 numbered streets and secondaries
//	ordinarystreet   none                                everything
//
// Detail is a private mailbox number and it is not a secondary designator.
// The standard's CMRA section is explicit that developers MUST NOT combine the
// secondary address element of the CMRA's own address with the patient's
// private box number, which is why Detail is a field of its own on
// address.Address rather than a second secondary. A type that admits a
// secondary unit therefore admits Detail beside it, never instead of it. A
// box line has no secondary unit at all — Pub 28 §281 makes the PO Box line
// PO BOX and its number, §241 the rural route line RR n BOX n — so on pobox
// and ruralroute a # identifier can only be the mailbox and is admitted the
// same as PMB. ordinarystreet has a secondary unit position, and takes # as
// the mailbox only beside a unit already placed (#78).
//
// * The same section says the words PO BOX and the private mailbox number MUST
// NOT be used on the street address line, and two lines later gives
// "PO BOX 159753 PMB 3571" as a correct form. The reading consistent with the
// example is that PO BOX must not stand in for PMB — they name different
// things — so a post office box admits a trailing Detail the way a rural route
// does. The decision is recorded on pobox.FormatStreetLine per CONTRIBUTING §2.
//
// ordinarystreet is the catchall because it is the one type with no pattern of
// its own: it is assembled from shared claims only, so it is the only type that
// can read everything. Catchall means lowest precedence when a closed form also
// reads the tokens, not default when unsure. A line that pobox reads as a box is
// a box, whatever ordinarystreet could make of the same tokens.
//
// Where the code does not yet match its row, the row is the intent and the gap
// is tracked: puertorico has Claims for the urbanization but no Candidates and
// no street vocabulary of its own (#60, #71); a unit or mailbox on the line
// above the street line is read by no type (#98).
package addresstypes
