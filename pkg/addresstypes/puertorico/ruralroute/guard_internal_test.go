package ruralroute

import (
	"strings"
	"testing"
)

var rewriteGuardInputs = []string{
	"", "RR 4 BOX 12", "RR4 BOX 12", "RR 04 BOX 012", "HC0", "BOX 0",
	"BOX0", "RR A BOX B", "RUTA RURAL 4 BUZON 12", "RR\t0", "RR ٣ BOX 4",
	"RR 4BOX12", "\xffRR0",
}

func TestRewriteGuards(t *testing.T) {
	for _, s := range rewriteGuardInputs {
		checkRewriteGuards(t, s)
	}
}

func FuzzRewriteGuards(f *testing.F) {
	for _, s := range rewriteGuardInputs {
		f.Add(s)
	}
	f.Fuzz(checkRewriteGuards)
}

// checkRewriteGuards fails when Normalize would skip gluednumber or
// leadingzero on a string it matches.
func checkRewriteGuards(t *testing.T, s string) {
	if !strings.ContainsAny(s, "0123456789") && gluednumber.MatchString(s) {
		t.Errorf("no digit in %q, but gluednumber matches it", s)
	}
	if strings.IndexByte(s, '0') < 0 && leadingzero.MatchString(s) {
		t.Errorf("no 0 in %q, but leadingzero matches it", s)
	}
}
