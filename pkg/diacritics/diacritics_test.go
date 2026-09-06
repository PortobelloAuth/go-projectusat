package diacritics_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
	appendixa "github.com/poetic-systems/addresstables/diacritics"
)

// transliterations names the rows where Transliterate deliberately parts
// company with Appendix A. The standard substitutes a single ASCII letter for
// every character it lists — ae becomes a, thorn becomes p — and
// transliteration spells the digraphs out instead. Every other row agrees
// with the standard, so only the disagreements are written down.
var transliterations = map[rune]string{
	'Æ': "ae",
	'æ': "ae",
	'Ð': "d",
	'ð': "d",
	'Þ': "th",
	'þ': "th",
	'ß': "ss",
	'Œ': "oe",
	'œ': "oe",
}

// TestSubstituteMatchesAppendixA checks the rune ranges in this package
// against Project US@ Appendix A, row for row, as addresstables holds it.
// The implementation is not a table, so this is where the two are made to
// agree: a range that stops covering a character the appendix lists, or that
// substitutes a different letter for it, fails here.
func TestSubstituteMatchesAppendixA(t *testing.T) {
	rows := 0

	for d := range appendixa.All() {
		rows++

		t.Run(d.Description, func(t *testing.T) {
			got, err := diacritics.Substitute(string(d.Rune))
			if err != nil {
				t.Fatalf("Substitute(%q): %v", string(d.Rune), err)
			}
			if got != d.Substitute {
				t.Errorf("Substitute(%q) = %q, want %q", string(d.Rune), got, d.Substitute)
			}
		})
	}

	if rows != appendixa.Len() {
		t.Fatalf("ranged over %d rows, want %d", rows, appendixa.Len())
	}
}

// TestTransliterateMatchesAppendixAExceptWhereItSpellsDigraphs covers the
// same rows through Transliterate, which follows the appendix everywhere
// except the nine characters transliterations names.
func TestTransliterateMatchesAppendixAExceptWhereItSpellsDigraphs(t *testing.T) {
	seen := map[rune]bool{}

	for d := range appendixa.All() {
		want := d.Substitute
		if spelled, ok := transliterations[d.Rune]; ok {
			want = spelled
			seen[d.Rune] = true
		}

		t.Run(d.Description, func(t *testing.T) {
			got, err := diacritics.Transliterate(string(d.Rune))
			if err != nil {
				t.Fatalf("Transliterate(%q): %v", string(d.Rune), err)
			}
			if got != want {
				t.Errorf("Transliterate(%q) = %q, want %q", string(d.Rune), got, want)
			}
		})
	}

	for r := range transliterations {
		if !seen[r] {
			t.Errorf("transliterations names %q, which Appendix A does not list", string(r))
		}
	}
}
