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
// what Normalize should return, so it fits this corpus's model. amadsen
// reviewed each of these cases on go-projectusat#112 and settled the ones
// that were open; a case whose Want is still empty is honestly unsettled,
// not an oversight — see Case's doc comment.
var OursCases = []Case{
	// Whether WEST belongs to the street (postdirectional) or the city.
	// Both readings name a real Florida city, so getting this wrong doesn't
	// error, it silently sends the address to the other one. Unlike the
	// NORTH PARK / PAUL pair below, the data does not discriminate here —
	// PALM BEACH and WEST PALM BEACH are both real cities — so the comma
	// still decides.
	{
		Source: "ours",
		Note:   "addressparsers#17: WEST before the comma is the street's postdirectional, not the city",
		Input:  "123 MAIN ST WEST, PALM BEACH, FL",
		Want:   "123 MAIN ST W\nPALM BEACH FL",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: WEST after the comma is the start of the city WEST PALM BEACH",
		Input:  "123 MAIN ST, WEST PALM BEACH, FL",
		Want:   "123 MAIN ST\nWEST PALM BEACH FL",
	},
	// Whether SW glued to JORDAN is a postdirectional plus a name (S, WEST
	// JORDAN) or a two-letter postdirectional on its own (SW) before a city
	// named just JORDAN. Utah has no city named plain JORDAN, so the first
	// reading is correct and both forms should reach it. amadsen on
	// go-projectusat#112: a data dependent parser can handle this case, a
	// grammatical one cannot, and the glued SW is a strange enough shape
	// that it is unlikely a human ever typed it that way — it most plausibly
	// comes from an earlier normalizer having mangled the data. It is kept
	// anyway, with this note, because it is precisely the case that
	// separates a data-dependent parser from a grammatical one.
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
	// amadsen on go-projectusat#112: what this test set should not replace
	// is these two delimiter-free variants, which have no comma anywhere to
	// mark the street/city boundary. Measured on addressparsers main at
	// 0957bbb (go-projectusat 9427c72, zipcity 9d0ac0f) on 2026-09-22,
	// through goprojectusat.Normalize with
	// parse.New(parse.Options{UseReferenceData: true}) and
	// WithContentNormalization(): "3253 W 9200 S WEST JORDAN UT" returns
	// "3253 WEST 9200 SOUTH\nWEST JORDAN UT", while "3253 WEST 9200 SOUTH
	// WEST JORDAN UT" returns "3253 W 9200 S\nWEST JORDAN UT". With no
	// delimiter the two legal spellings of one address cross over and
	// produce two different strings for the street line — one address, two
	// hashes.
	{
		Source: "ours",
		Note:   "addressparsers#17: W 9200 S spelling with no delimiter before WEST JORDAN",
		Input:  "3253 W 9200 S WEST JORDAN UT",
		Want:   "3253 W 9200 S\nWEST JORDAN UT",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: WEST 9200 SOUTH spelling with no delimiter before WEST JORDAN",
		Input:  "3253 WEST 9200 SOUTH WEST JORDAN UT",
		Want:   "3253 W 9200 S\nWEST JORDAN UT",
	},
	// Whether E ST is the alphabet street E with a suffix, or EAST ST is a
	// directional street name. Both are real and common. amadsen on
	// go-projectusat#112 settles the rule: a valid directional street name
	// stays written out, a valid alphabetic street name stays the letter,
	// and only data saying the supplied reading is not real while the
	// alternative is may switch one to the other. go-projectusat#110 is
	// where that data-driven promotion gets built, not the reason the
	// expectation is unknown. A last line is supplied because the parser
	// admits no address type without one, and E ST NW is real in DC.
	// Measured today, EAST ST returns "123 E STREET\nWASHINGTON DC
	// 20001", which is neither reading: EAST taken as a predirectional
	// and abbreviated, ST taken as the name and spelled out. It is
	// expected to fail until the street window lands.
	{
		Source: "ours",
		Note:   "addressparsers#17: the alphabet street E keeps its own form",
		Input:  "123 E ST\nWASHINGTON DC 20001",
		Want:   "123 E ST\nWASHINGTON DC 20001",
	},
	{
		Source: "ours",
		Note:   "addressparsers#17: EAST ST is a directional street name and stays written out",
		Input:  "123 EAST ST\nWASHINGTON DC 20001",
		Want:   "123 EAST ST\nWASHINGTON DC 20001",
	},
	// Whether ST belongs to the street (NORTH PARK ST, city PAUL) or the
	// city (NORTH PARK, city ST PAUL). amadsen on go-projectusat#112,
	// quoting addressparsers#17's rule back at us: a last-line reading whose
	// city zipcity has seen beats one it has not. Minnesota has never had a
	// place called PAUL while ST PAUL is a real Minnesota city, so the data
	// discriminates, and the comma after ST in the marked form reads as a
	// likely typo or OCR error that should be overridden — both forms
	// resolve to the same city. Per
	// https://pe.usps.com/text/pub28/28c2_008.htm, city names are spelled
	// in their entirety, so the city renders SAINT PAUL, not ST PAUL. The
	// street half still rests on addressparsers#24 (whether NORTH PARK
	// decomposes into a directional plus a suffix) and on the rule settled
	// above for EAST ST — a valid directional street name stays written
	// out — which is why NORTH PARK is written out here rather than
	// abbreviated to N PARK. Measured today: the comma'd form returns
	// "123 N PARK ST\nPAUL MN" and the unmarked form returns "123 NORTH
	// PARK\nST PAUL MN"; both are expected to fail until step 2
	// (addressparsers#17) and the street window land.
	{
		Source: "ours",
		Note:   "addressparsers#19: the comma in NORTH PARK ST, PAUL reads as a typo or OCR error once the data shows MN has no city PAUL, only SAINT PAUL",
		Input:  "123 NORTH PARK ST, PAUL, MN",
		Want:   "123 NORTH PARK\nSAINT PAUL MN",
	},
	{
		Source: "ours",
		Note:   "addressparsers#19: ST PAUL renders SAINT PAUL per USPS Pub 28 (spell city names in their entirety)",
		Input:  "123 NORTH PARK, ST PAUL, MN",
		Want:   "123 NORTH PARK\nSAINT PAUL MN",
	},
}
