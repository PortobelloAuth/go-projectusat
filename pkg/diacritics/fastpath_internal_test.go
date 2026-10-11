package diacritics

import "testing"

// fastPathInputs are ASCII strings Substitute short-circuits, and non-ASCII
// ones it must not, so the fuzzer starts on both sides of isASCII.
var fastPathInputs = []string{
	"", " ", "Main St", "MAIN ST", "123 Calle Luna", "RR 4 BOX 12",
	"\t\n\x00\x7f", "#-/.,'", "Peñuelas", "CAFÉ", "ſ", "é", "\xff",
	"Ærøskøbing", "STRAßE", "a\x80b",
}

func TestSubstituteASCIIFastPath(t *testing.T) {
	for _, s := range fastPathInputs {
		checkFastPath(t, s)
	}
}

func FuzzSubstituteASCIIFastPath(f *testing.F) {
	for _, s := range fastPathInputs {
		f.Add(s)
	}
	f.Fuzz(checkFastPath)
}

// checkFastPath fails when Substitute and the full transform chain disagree.
func checkFastPath(t *testing.T, s string) {
	want, wantErr := substitute(s)
	got, gotErr := Substitute(s)
	if got != want || (gotErr == nil) != (wantErr == nil) {
		t.Errorf("Substitute(%q) = %q, %v; substitute = %q, %v", s, got, gotErr, want, wantErr)
	}
}
