package highways

import (
	"strings"
	"testing"
)

// guardInputs are every highway example the standard gives, each leading
// word normalizeTokens tests for, state names and codes in front of a route,
// glued interstates, and strings that probe how the guard finds the first
// word: leading and embedded non-ASCII space, runes that upper-case into
// ASCII (ı -> I, ſ -> S), and invalid UTF-8.
var guardInputs = append([]string{
	"FM", "FARM TO", "I", "IH", "I 10", "IH 280", "I10A", "IH10", "I1", "IX10",
	"INTERSTATE 10", "US 1", "CR A", "TSR 4", "TOWNSHIP ROAD 4", "RANCH ROAD 4",
	"STATE ROUTE A", "ST RT 1", "RD 1", "ROAD 1", "RT 1", "RTE 1", "ROUTE 1",
	"EXPRESSWAY 1", "TX 1", "TEXAS 1", "NEW MEXICO 66", "NEW HAMPSHIRE HWY 1",
	"PUERTO RICO 2", "AMERICAN SAMOA 1", "NEW", "NORTH", "WEST VIRGINIA",
	"FEDERATED STATES OF MICRONESIA HWY 1", "MICRONESIA HIGHWAY 1", "FM FM 1",
	"MAIN", "1 HWY 64", " HWY 64", " HWY 64", "\u0085HWY 64", "ıH 280",
	"ſR 220", "ſT RD 1", "\xffHWY 64", "HWY\xff 64", "\xc3", " ",
}, claimsInputs...)

func TestMayLeadHighwayIsNecessary(t *testing.T) {
	for _, in := range guardInputs {
		for _, s := range []string{in, strings.ToLower(in)} {
			checkGuard(t, s)
		}
	}
}

func FuzzMayLeadHighwayIsNecessary(f *testing.F) {
	for _, in := range guardInputs {
		f.Add(in)
	}
	f.Fuzz(checkGuard)
}

// checkGuard fails when the leading-word guard rejects a name the unguarded
// recognizer accepts, when NormalizeStreetName and the unguarded path
// disagree, or when leadingWord disagrees with strings.Fields.
func checkGuard(t *testing.T, name string) {
	s := collapseSpace(strings.ToUpper(strings.TrimSpace(name)))
	if s == "" {
		return
	}

	want, wantErr := normalizeUpper(s)
	got, gotErr := NormalizeStreetName(name)

	lead, _, _ := strings.Cut(s, " ")
	if wantErr == nil && !mayLeadHighway(lead) {
		t.Errorf("mayLeadHighway(%q) = false, but %q normalizes to %q", lead, name, want)
	}
	if got != want || (gotErr == nil) != (wantErr == nil) ||
		(gotErr != nil && gotErr.Error() != wantErr.Error()) {
		t.Errorf("NormalizeStreetName(%q) = %q, %v; unguarded = %q, %v", name, got, gotErr, want, wantErr)
	}

	fields := strings.Fields(strings.ToUpper(name))
	word, ok := leadingWord(name)
	if ok != (len(fields) > 0) || (ok && word != fields[0]) {
		t.Errorf("leadingWord(%q) = %q, %v; strings.Fields = %q", name, word, ok, fields)
	}
}
