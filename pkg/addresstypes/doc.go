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
//	pobox            whole line: PO BOX n                a Detail (PMB or #), trailing or above *
//	ruralroute       whole line: RR/HC n BOX n           a Detail (PMB or #), trailing or above
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
// box line has no secondary unit at all — Pub 28 §281 standardizes the PO Box
// delivery line as PO BOX and the box number, and §241 does the same for the
// rural route line — so on pobox and ruralroute the # identifier has nowhere
// else to be read and is admitted as the mailbox the same as PMB is; only
// ordinarystreet, which does have a secondary unit position, takes # as the
// mailbox solely beside a unit already placed (Aaron on #78).
//
// Both a secondary unit and a Detail may also be read from the line
// immediately above a type's own street line, not only from the position
// trailing it: Pub 28 §213.3 puts a secondary unit there when it does not fit
// on the street line, and §285's four-line CMRA form puts PMB or # there
// instead of trailing the street line. A type reads that line only when it is
// covered exactly by one claim it would otherwise admit at the end of its own
// line, under the same rule for which identifiers count as the mailbox that
// governs the trailing position (#98).
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
// no street vocabulary of its own (#60, #71).
package addresstypes
