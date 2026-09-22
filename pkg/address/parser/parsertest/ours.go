package parsertest

// OursCases are hard cases the address stack found on its own, rather than
// transcribed from the standard: addresses whose comma placement or
// abbreviation puts two real readings within one edit of each other, so a
// parser that gets the wrong one silently misroutes a patient instead of
// erroring. They come from addressparsers#17, and from the probes that
// measured it — internal/probe/pairs_probe.go and
// internal/probe/commapairs_probe.go on addressparsers' ap17-slice2 branch
// (addressparsers#22) — which inspected parsed fields or comma-agreement
// rather than asserting a Normalize return value; Want below states each as
// what Normalize should return, so it fits this corpus's model. That
// judgment call is the author's, not a transcription: check it before
// trusting it as ground truth.
var OursCases = []Case{
	// Whether WEST belongs to the street (postdirectional) or the city.
	// Both readings name a real Florida city, so getting this wrong doesn't
	// error, it silently sends the address to the other one.
	{
		Source: "ours",
		Note:   "addressparsers#17: WEST before the comma is the street's postdirectional, not the city",
		Input:  "123 MAIN ST WEST, PALM BEACH, FL",
		Want:   "123 MAIN ST W\nPALM BEACH FL",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: WEST after the comma is the start of the city PALM BEACH",
		Input:  "123 MAIN ST, WEST PALM BEACH, FL",
		Want:   "123 MAIN ST\nWEST PALM BEACH FL",
	},
	// Whether SW glued to JORDAN is a postdirectional plus a name (S, WEST
	// JORDAN) or a two-letter postdirectional on its own (SW) before a city
	// named just JORDAN. Utah has no city named plain JORDAN, so the first
	// reading is correct and both forms should reach it.
	{
		Source: "ours",
		Note:   "addressparsers#17: glued SW JORDAN should still reach 9200 S / WEST JORDAN",
		Input:  "3253 W 9200 SW JORDAN, UT",
		Want:   "3253 W 9200 S\nWEST JORDAN UT",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: 9200 S, WEST JORDAN written unambiguously is a fixed point",
		Input:  "3253 W 9200 S, WEST JORDAN, UT",
		Want:   "3253 W 9200 S\nWEST JORDAN UT",
	},
	// Whether EAST is a directional word that abbreviates, or already a
	// street name. With no suffix or city to disambiguate, the standard's
	// own rule (p.16-17: a directional after the name abbreviates) says
	// EAST ST reaches E ST.
	{
		Source: "ours",
		Note:   "addressparsers#17: E ST is already a fixed point",
		Input:  "123 E ST",
		Want:   "123 E ST",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: EAST ST should reach E ST",
		Input:  "123 EAST ST",
		Want:   "123 E ST",
	},
	// Whether ST belongs to the street (NORTH PARK ST, city PAUL) or the
	// city (NORTH PARK, city ST PAUL). Both are real Minnesota places; ST
	// PAUL is the far more common one, and SAINT PAUL is how the content
	// normalizer spells it out (cf. the ST CLOUD parity cases).
	{
		Source: "ours",
		Note:   "addressparsers#17: ST before the comma is the street's suffix, city is PAUL",
		Input:  "123 NORTH PARK ST, PAUL, MN",
		Want:   "123 NORTH PARK ST\nPAUL MN",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: ST after the comma starts the city SAINT PAUL",
		Input:  "123 NORTH PARK, ST PAUL, MN",
		Want:   "123 NORTH PARK\nSAINT PAUL MN",
	},
}
