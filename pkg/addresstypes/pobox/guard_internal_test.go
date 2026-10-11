package pobox

import "testing"

var hashGuardInputs = []string{
	"", "#", " # ", "PO BOX 12", "PO BOX #12", "Main St # 4", "\t#\n",
	"＃4", "PO # BOX", "\xff#", "BOX 12",
}

func TestHashGuard(t *testing.T) {
	for _, s := range hashGuardInputs {
		checkHashGuard(t, s)
	}
}

func FuzzHashGuard(f *testing.F) {
	for _, s := range hashGuardInputs {
		f.Add(s)
	}
	f.Fuzz(checkHashGuard)
}

// checkHashGuard fails when Normalize would skip hashPattern on a string it
// matches.
func checkHashGuard(t *testing.T, s string) {
	hasHash := false
	for i := 0; i < len(s); i++ {
		hasHash = hasHash || s[i] == '#'
	}
	if !hasHash && hashPattern.MatchString(s) {
		t.Errorf("no # in %q, but hashPattern matches it", s)
	}
}
