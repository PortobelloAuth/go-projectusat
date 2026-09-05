package diacritics

import (
	"fmt"
	"strings"
	"unicode"

	anyascii "github.com/anyascii/go"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/unicode/rangetable"
)

type DiacriticMode int

const (
	KeepDiacritics          DiacriticMode = iota // 0 (Default/Zero-value)
	SubstituteDiacritics                         // 1
	TransliterateDiacritics                      // 2
)

// diacriticMap is the Appendix A substitution, held as rune ranges rather
// than as rows: every character in a range substitutes to the letter it is
// keyed by. The table itself is Project US@ Appendix A, held once in
// github.com/poetic-systems/addresstables/diacritics, and
// TestSubstituteMatchesAppendixA checks this implementation against it row
// for row.
var diacriticMap = map[rune]*unicode.RangeTable{
	'a': {
		R16: []unicode.Range16{
			{Lo: 0x00C0, Hi: 0x00C6, Stride: 1}, // Uppercase "A" with various diacritics
			{Lo: 0x00E0, Hi: 0x00E6, Stride: 1}, // Lowercase "a" with various diacritics
		},
	},

	'c': {
		R16: []unicode.Range16{
			{Lo: 0x00C7, Hi: 0x00C7, Stride: 1}, // Capital letter C with cedilla
			{Lo: 0x00E7, Hi: 0x00E7, Stride: 1}, // Small letter c with cedilla
		},
	},

	'e': {
		R16: []unicode.Range16{
			{Lo: 0x00C8, Hi: 0x00CB, Stride: 1}, // Uppercase "E" with various diacritics
			{Lo: 0x00D0, Hi: 0x00D0, Stride: 1}, // Capital letter ETH
			{Lo: 0x00E8, Hi: 0x00EB, Stride: 1}, // Lowercase "a" with various diacritics
			{Lo: 0x00F0, Hi: 0x00F0, Stride: 1}, // Small letter eth
		},
	},

	'i': {
		R16: []unicode.Range16{
			{Lo: 0x00CC, Hi: 0x00CF, Stride: 1}, // Uppercase "I" with various diacritics
			{Lo: 0x00EC, Hi: 0x00EF, Stride: 1}, // Lowercase "i" with various diacritics
		},
	},

	'n': {
		R16: []unicode.Range16{
			{Lo: 0x00D1, Hi: 0x00D1, Stride: 1}, // Capital letter "N" with tilde
			{Lo: 0x00F1, Hi: 0x00F1, Stride: 1}, // Small letter "n" with tilde
		},
	},

	'o': {
		R16: []unicode.Range16{
			{Lo: 0x00D2, Hi: 0x00D8, Stride: 1}, // Capital letter "O" with various diacritics
			{Lo: 0x00F2, Hi: 0x00F8, Stride: 1}, // Small letter "o" with various diacritics
			{Lo: 0x0152, Hi: 0x0153, Stride: 1}, // Capital and small letter "oe"
		},
	},

	'u': {
		R16: []unicode.Range16{
			{Lo: 0x00D9, Hi: 0x00DC, Stride: 1}, // Capital letter "U" with various diacritics
			{Lo: 0x00F9, Hi: 0x00FC, Stride: 1}, // Small letter "u" with various diacritics
		},
	},

	'y': {
		R16: []unicode.Range16{
			{Lo: 0x00DD, Hi: 0x00DD, Stride: 1}, // Capital letter "Y" with acute accent
			{Lo: 0x00FD, Hi: 0x00FD, Stride: 1}, // Small letter "y" with acute accent
			{Lo: 0x00FF, Hi: 0x00FF, Stride: 1}, // Small letter "y" with dieresis
			{Lo: 0x0178, Hi: 0x0178, Stride: 1}, // Capital letter "Y" with dieresis
		},
	},

	'p': {
		R16: []unicode.Range16{
			{Lo: 0x00DE, Hi: 0x00DE, Stride: 1}, // Capital letter THORN
			{Lo: 0x00FE, Hi: 0x00FE, Stride: 1}, // Small letter thorn
		},
	},

	's': {
		R16: []unicode.Range16{
			{Lo: 0x00DF, Hi: 0x00DF, Stride: 1}, // Small letter sharp s - ess-zed
			{Lo: 0x0160, Hi: 0x0161, Stride: 1}, // Capital and small letter S with caron
		},
	},
}

var allDiacriticRanges = rangetable.Merge(
	diacriticMap['a'],
	diacriticMap['c'],
	diacriticMap['e'],
	diacriticMap['i'],
	diacriticMap['n'],
	diacriticMap['o'],
	diacriticMap['u'],
	diacriticMap['y'],
	diacriticMap['p'],
	diacriticMap['s'],
)

func Substitute(source string) (string, error) {
	// Chain the transformation: NFD Decomposition -> Remove Diacritics -> NFC Recomposition
	mapping := func(r rune) rune {
		for k, v := range diacriticMap {
			if runes.In(v).Contains(r) {
				return k
			}
		}
		return r
	}
	tdiacritics := runes.If(runes.In(allDiacriticRanges), runes.Map(mapping), nil)
	t := transform.Chain(norm.NFD, tdiacritics, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	clean, _, err := transform.String(t, source)
	if err != nil {
		return "", fmt.Errorf("error while substituting diacritics: %w", err)
	}

	// NOTE: the table in the standard document shows lowercase values - which seems silly
	// for a standard that wants everything uppercase, but that's how it is.
	return strings.ToLower(clean), nil
}

func Transliterate(source string) (string, error) {
	clean := anyascii.Transliterate(source)

	return strings.ToLower(clean), nil
}

func Normalize(source string, mode DiacriticMode) (string, error) {
	switch mode {
	case SubstituteDiacritics:
		return Substitute(source)
	case TransliterateDiacritics:
		return Transliterate(source)
	default:
	}
	return source, nil
}
