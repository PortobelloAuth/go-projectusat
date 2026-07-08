package diacritics_test

import (
	"fmt"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

var testDiacritics = []struct {
	Desc string
	In   string
	Sub  string
	T    string
}{
	{In: "À", Sub: "a", T: "a", Desc: "Capital letter A with grave accent"},
	{In: "Á", Sub: "a", T: "a", Desc: "Capital letter A with acute accent"},
	{In: "Â", Sub: "a", T: "a", Desc: "Capital letter A with circumflex accent"},
	{In: "Ã", Sub: "a", T: "a", Desc: "Capital letter A with tilde"},
	{In: "Ä", Sub: "a", T: "a", Desc: "Capital letter A with dieresis or umlaut mark"},
	{In: "Å", Sub: "a", T: "a", Desc: "Capital letter A with ring above"},
	{In: "Æ", Sub: "a", T: "ae", Desc: "Capital letter AE diphthong"},
	{In: "Ç", Sub: "c", T: "c", Desc: "Capital letter C with cedilla"},
	{In: "È", Sub: "e", T: "e", Desc: "Capital letter E with grave accent"},
	{In: "É", Sub: "e", T: "e", Desc: "Capital letter E with acute accent"},
	{In: "Ê", Sub: "e", T: "e", Desc: "Capital letter E with circumflex accent"},
	{In: "Ë", Sub: "e", T: "e", Desc: "Capital letter E with dieresis or umlaut mark"},
	{In: "Ì", Sub: "i", T: "i", Desc: "Capital letter I with grave accent"},
	{In: "Í", Sub: "i", T: "i", Desc: "Capital letter I with acute accent"},
	{In: "Î", Sub: "i", T: "i", Desc: "Capital letter I with circumflex"},
	{In: "Ï", Sub: "i", T: "i", Desc: "Capital letter I with dieresis or umlaut mark"},
	{In: "Ð", Sub: "e", T: "d", Desc: "Capital letter ETH (Icelandic)"},
	{In: "Ñ", Sub: "n", T: "n", Desc: "Capital letter N with tilde"},
	{In: "Ò", Sub: "o", T: "o", Desc: "Capital letter O with grave accent"},
	{In: "Ó", Sub: "o", T: "o", Desc: "Capital letter O with acute accent"},
	{In: "Ô", Sub: "o", T: "o", Desc: "Capital letter O with circumflex"},
	{In: "Õ", Sub: "o", T: "o", Desc: "Capital letter O with tilde"},
	{In: "Ö", Sub: "o", T: "o", Desc: "Capital letter O with dieresis or umlaut mark"},
	{In: "Ø", Sub: "o", T: "o", Desc: "Capital letter O with slash"},
	{In: "Ù", Sub: "u", T: "u", Desc: "Capital letter U with grave accent"},
	{In: "Ú", Sub: "u", T: "u", Desc: "Capital letter U with acute accent"},
	{In: "Û", Sub: "u", T: "u", Desc: "Capital letter U with circumflex"},
	{In: "Ü", Sub: "u", T: "u", Desc: "Capital letter U with dieresis or umlaut mark"},
	{In: "Ý", Sub: "y", T: "y", Desc: "Capital letter Y with acute accent"},
	{In: "Þ", Sub: "p", T: "th", Desc: "Capital letter THORN"},
	{In: "ß", Sub: "s", T: "ss", Desc: "Small letter sharp s - ess-zed"},
	{In: "à", Sub: "a", T: "a", Desc: "Small letter a with grave accent"},
	{In: "á", Sub: "a", T: "a", Desc: "Small letter a with acute accent"},
	{In: "â", Sub: "a", T: "a", Desc: "Small letter a with circumflex"},
	{In: "ã", Sub: "a", T: "a", Desc: "Small letter a with tilde"},
	{In: "ä", Sub: "a", T: "a", Desc: "Small letter a with dieresis or umlaut mark"},
	{In: "å", Sub: "a", T: "a", Desc: "Small letter a with ring above"},
	{In: "æ", Sub: "a", T: "ae", Desc: "Small letter ae"},
	{In: "ç", Sub: "c", T: "c", Desc: "Small letter c with cedilla"},
	{In: "è", Sub: "e", T: "e", Desc: "Small letter e with grave accent"},
	{In: "é", Sub: "e", T: "e", Desc: "Small letter e with acute accent"},
	{In: "ê", Sub: "e", T: "e", Desc: "Small letter e with circumflex"},
	{In: "ë", Sub: "e", T: "e", Desc: "Small letter e with dieresis"},
	{In: "ì", Sub: "i", T: "i", Desc: "Small letter i with grave accent"},
	{In: "í", Sub: "i", T: "i", Desc: "Small letter i with acute accent"},
	{In: "î", Sub: "i", T: "i", Desc: "Small letter i with circumflex"},
	{In: "ï", Sub: "i", T: "i", Desc: "Small letter i with diaresis"},
	{In: "ð", Sub: "e", T: "d", Desc: "Small letter eth"},
	{In: "ñ", Sub: "n", T: "n", Desc: "Small letter n with tilde"},
	{In: "ò", Sub: "o", T: "o", Desc: "Small letter o with grave accent"},
	{In: "ó", Sub: "o", T: "o", Desc: "Small letter o with acute accent"},
	{In: "ô", Sub: "o", T: "o", Desc: "Small letter o with circumflex"},
	{In: "õ", Sub: "o", T: "o", Desc: "Small letter o with tilde"},
	{In: "ö", Sub: "o", T: "o", Desc: "Small letter o with dieresis"},
	{In: "ø", Sub: "o", T: "o", Desc: "Small letter o with slash"},
	{In: "ù", Sub: "u", T: "u", Desc: "Small letter u with grave accent"},
	{In: "ú", Sub: "u", T: "u", Desc: "Small letter u with acute accent"},
	{In: "û", Sub: "u", T: "u", Desc: "Small letter u with circumflex"},
	{In: "ü", Sub: "u", T: "u", Desc: "Small letter u with dieresis"},
	{In: "ý", Sub: "y", T: "y", Desc: "Small letter y with acute accent"},
	{In: "þ", Sub: "p", T: "th", Desc: "Small letter thorn"},
	{In: "ÿ", Sub: "y", T: "y", Desc: "Small letter y with dieresis"},
	{In: "Œ", Sub: "o", T: "oe", Desc: "Capital letter OE"},
	{In: "œ", Sub: "o", T: "oe", Desc: "Small letter oe"},
	{In: "Š", Sub: "s", T: "s", Desc: "Capital letter S with caron"},
	{In: "š", Sub: "s", T: "s", Desc: "Small letter s with caron"},
	{In: "Ÿ", Sub: "y", T: "y", Desc: "Capital letter Y with dieresis"},
}

func TestSubstitute(t *testing.T) {
	for _, td := range testDiacritics {
		t.Run(fmt.Sprintf("Substitute %s", td.Desc), func(t *testing.T) {
			actual, err := diacritics.Substitute(td.In)
			if err != nil {
				t.Errorf("%s", err)
			}
			if actual != td.Sub {
				t.Errorf("incorrect substitution from '%s' to '%s'. expected '%s'", td.In, actual, td.Sub)
			}
		})
	}
}

func TestTransliterate(t *testing.T) {
	for _, td := range testDiacritics {
		t.Run(fmt.Sprintf("Substitute %s", td.Desc), func(t *testing.T) {
			actual, err := diacritics.Transliterate(td.In)
			if err != nil {
				t.Errorf("%s", err)
			}
			if actual != td.T {
				t.Errorf("incorrect substitution from '%s' to '%s'. expected '%s'", td.In, actual, td.Sub)
			}
		})
	}
}
