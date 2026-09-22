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
	// Whether E ST is the alphabet street E with a suffix, or EAST ST is a
	// directional street name. Both are real and common; amadsen on
	// go-projectusat#108 and #110 is that the string cannot say which is
	// meant and the input's form is kept until the data decides, so these
	// two must not collapse into one. A last line is supplied because the
	// parser admits no address type without one, and E ST NW is real in DC.
	{
		Source: "ours",
		Note:   "addressparsers#17: the alphabet street E keeps its own form",
		Input:  "123 E ST\nWASHINGTON DC 20001",
		Want:   "123 E ST\nWASHINGTON DC 20001",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: whether EAST ST abbreviates to E ST is go-projectusat#110's question and is not settled — the two are different streets if it does not",
		Input:  "123 EAST ST\nWASHINGTON DC 20001",
		Want:   "",
	},
	// Whether ST belongs to the street (NORTH PARK ST, city PAUL) or the
	// city (NORTH PARK, city ST PAUL). Both readings are grammatical and
	// Minnesota has never had a place called PAUL, so addressparsers#19
	// settles the unmarked form to ST PAUL from the data — but whether a
	// comma after ST should hold the older reading is the judgement call
	// raised for amadsen on addressparsers#19 and not yet answered, and
	// whether the city renders ST PAUL or SAINT PAUL is a second open
	// question. Both are carried unsettled rather than guessed.
	{
		Source: "ours",
		Note:   "addressparsers#19: whether a comma after ST keeps the city PAUL is unanswered",
		Input:  "123 NORTH PARK ST, PAUL, MN",
		Want:   "",
	},
	{
		Source: "ours",
		Note:   "addressparsers#19: the reading is settled, the ST PAUL / SAINT PAUL rendering is not",
		Input:  "123 NORTH PARK, ST PAUL, MN",
		Want:   "",
	},
}
