package ruralroute

import (
	"strings"
	"testing"
)

// guardInputs are strings normalize accepts, and strings built to probe each
// step mayBeRoute's argument leans on: leading punctuation alphanumspace
// deletes, leading whitespace it keeps, non-ASCII runes that upper-case into
// ASCII (ı -> I, ſ -> S), number markers rewritten ahead of the designator,
// and designators glued into longer words.
var guardInputs = []string{
	"RURAL ROUTE 91 BOX A7", "Rural Route 91 Box A7", "RFD 82 BOX 12",
	"RD 51 # 25", "RFD Route 4 #87a", "RR 2 BOX 18 Bryan Dairy Rd",
	"RR03 BOX 98D", "RR 3 BX 98D", "HC 4 BOX 12", "HCR 4 BOX 12",
	"HIGHWAY CONTRACT ROUTE 4 BOX 12", "STAR ROUTE 4 BOX 12",
	"STAR ROUTE No. 4 # 12", "rt 4 no 12", "RT #4 NUMBER 12",
	"-RR 4 BOX 12", "#RR 4 BOX 12", "...RR 4 BOX 12", " RR 4 BOX 12",
	"\tRR 4 BOX 12", "\nRR 4 BOX 12", "éRR 4 BOX 12", "ſTAR ROUTE 4 BOX 12",
	"ıRR 4 BOX 12", "4 RR BOX 12", "BOX 12 RR 4", "NO RR 4 BOX 12",
	"# 4 RR 4 BOX 12", "BIRD 4 BOX 12", "SHORT RD 4 BOX 12", "S RR 4 BOX 12",
	"HRR 4 BOX 12", "RRD 4 BOX 12", "RDRR 4 # 12", "Main St", "", " ", "#",
	"PO BOX 12", "RR", "RR 4", "RR 4 BOX", "R.R. 4 BOX 12", "H.C. 4 BOX 12",
	"R-R 4 B-O-X 12", "RR 4 BOX 12", " RR 4 BOX 12", "�RR 4 BOX 12",
	"\xffRR 4 BOX 12",
}

func TestMayBeRouteIsNecessary(t *testing.T) {
	for _, in := range guardInputs {
		for _, s := range []string{in, strings.ToLower(in)} {
			checkGuard(t, s)
		}
	}
}

func FuzzMayBeRouteIsNecessary(f *testing.F) {
	for _, in := range guardInputs {
		f.Add(in)
	}
	f.Fuzz(checkGuard)
}

// checkGuard fails when mayBeRoute rejects a string normalize accepts, or
// when Normalize and normalize disagree.
func checkGuard(t *testing.T, s string) {
	want, wantErr := normalize(s)
	got, gotErr := Normalize(s)
	if wantErr == nil && !mayBeRoute(s) {
		t.Errorf("mayBeRoute(%q) = false, but normalize accepts it as %q", s, want)
	}
	if got != want || (gotErr == nil) != (wantErr == nil) {
		t.Errorf("Normalize(%q) = %q, %v; normalize = %q, %v", s, got, gotErr, want, wantErr)
	}
}
