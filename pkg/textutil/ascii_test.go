package textutil_test

import (
	"regexp"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// asciiInputs probe run boundaries, every RE2 space, spaces RE2 does not
// count (\v, NBSP, U+3000), multi-byte runes, and invalid UTF-8.
var asciiInputs = []string{
	"", " ", "  ", "RR 4 BOX 12", "R.R. 4 #12", "--", "a-b--c", "-a-",
	"Main St.", "PO Box #4", "\t\n\f\r ", " \t ", "a\vb", "a b",
	"a　b", "Peñuelas", "CAFÉ", "\xff", "a\xffb\xfe", "é-é", "\x00\x7f",
	"GENERAL  DELIVERY", " GENERAL-DELIVERY. ", "x \n y", "end ",
}

var (
	upperAlnumSpaceRE = regexp.MustCompile(`[^0-9A-Z ]+`)
	alnumSpaceRE      = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	re2SpaceRE        = regexp.MustCompile(`\s+`)
)

func checkReplaceRunsOutside(t *testing.T, s string) {
	for _, c := range []struct {
		set *textutil.ASCIISet
		re  *regexp.Regexp
	}{
		{textutil.UpperAlnumSpace, upperAlnumSpaceRE},
		{textutil.AlnumSpace, alnumSpaceRE},
	} {
		for _, repl := range []string{"", " "} {
			want := c.re.ReplaceAllString(s, repl)
			if got := c.set.ReplaceRunsOutside(s, repl); got != want {
				t.Errorf("ReplaceRunsOutside(%q, %q) = %q; %s gives %q", s, repl, got, c.re, want)
			}
		}
	}
}

func checkCollapseRE2Space(t *testing.T, s string) {
	want := re2SpaceRE.ReplaceAllString(s, " ")
	if got := textutil.CollapseRE2Space(s); got != want {
		t.Errorf("CollapseRE2Space(%q) = %q; regexp gives %q", s, got, want)
	}
}

func TestReplaceRunsOutside(t *testing.T) {
	for _, s := range asciiInputs {
		checkReplaceRunsOutside(t, s)
	}
}

func FuzzReplaceRunsOutside(f *testing.F) {
	for _, s := range asciiInputs {
		f.Add(s)
	}
	f.Fuzz(checkReplaceRunsOutside)
}

func TestCollapseRE2Space(t *testing.T) {
	for _, s := range asciiInputs {
		checkCollapseRE2Space(t, s)
	}
}

func FuzzCollapseRE2Space(f *testing.F) {
	for _, s := range asciiInputs {
		f.Add(s)
	}
	f.Fuzz(checkCollapseRE2Space)
}
