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

/*
À 192 a U+00C0 Capital letter A with grave accent
Á 193 a U+00C1 Capital letter A with acute accent
Â 194 a U+00C2 Capital letter A with circumflex accent
Ã 195 a U+00C3 Capital letter A with tilde
Ä 196 a U+00C4 Capital letter A with dieresis or umlaut mark
Å 197 a U+00C5 Capital letter A with ring above
Æ 198 a U+00C6 Capital letter AE diphthong
Ç 199 c U+00C7 Capital letter C with cedilla
È 200 e U+00C8 Capital letter E with grave accent
É 201 e U+00C9 Capital letter E with acute accent
Ê 202 e U+00CA Capital letter E with circumflex accent
Ë 203 e U+00CB Capital letter E with dieresis or umlaut mark
Ì 204 i U+00CC Capital letter I with grave accent
Í 205 i U+00CD Capital letter I with acute accent
Î 206 i U+00CE Capital letter I with circumflex
Ï 207 i U+00CF Capital letter I with dieresis or umlaut mark
Ð 208 e U+00D0 Capital letter ETH (Icelandic)
Ñ 209 n U+00D1 Capital letter N with tilde
Ò 210 o U+00D2 Capital letter O with grave accent
Ó 211 o U+00D3 Capital letter O with acute accent
Ô 212 o U+00D4 Capital letter O with circumflex
Õ 213 o U+00D5 Capital letter O with tilde
Ö 214 o U+00D6 Capital letter O with dieresis or umlaut mark
Ø 216 o U+00D8 Capital letter O with slash
Ù 217 u U+00D9 Capital letter U with grave accent
Ú 218 u U+00DA Capital letter U with acute accent
Û 219 u U+00DB Capital letter U with circumflex
Ü 220 u U+00DC Capital letter U with dieresis or umlaut mark
Ý 221 y U+00DD Capital letter Y with acute accent
Þ 222 p U+00DE Capital letter THORN
ß 223 s U+00DF Small letter sharp s - ess-zed
à 224 a U+00E0 Small letter a with grave accent
á 225 a U+00E1 Small letter a with acute accent
â 226 a U+00E2 Small letter a with circumflex
ã 227 a U+00E3 Small letter a with tilde
ä 228 a U+00E4 Small letter a with dieresis or umlaut mark
å 229 a U+00E5 Small letter a with ring above
æ 230 a U+00E6 Small letter ae
ç 231 c U+00E7 Small letter c with cedilla
è 232 e U+00E8 Small letter e with grave accent
é 233 e U+00E9 Small letter e with acute accent
ê 234 e U+00EA Small letter e with circumflex
ë 235 e U+00EB Small letter e with dieresis
ì 236 i U+00EC Small letter i with grave accent
í 237 i U+00ED Small letter i with acute accent
î 238 i U+00EE Small letter i with circumflex
ï 239 i U+00EF Small letter i with diaresis
ð 240 e U+00FO Small letter eth
ñ 241 n U+00F1 Small letter n with tilde
ò 242 o U+00F2 Small letter o with grave accent
ó 243 o U+00F3 Small letter o with acute accent
ô 244 o U+00F4 Small letter o with circumflex
õ 245 o U+00F5 Small letter o with tilde
ö 246 o U+00F6 Small letter o with dieresis
ø 248 o U+00F8 Small letter o with slash
ù 249 u U+00F9 Small letter u with grave accent
ú 250 u U+00FA Small letter u with acute accent
û 251 u U+00FB Small letter u with circumflex
ü 252 u U+00FC Small letter u with dieresis
ý 253 y U+00FD Small letter y with acute accent
þ 254 p U+00FE Small letter thorn
ÿ 255 y U+00FF Small letter y with dieresis
Œ 338 o U+0152 Capital letter OE
œ 339 o U+0153 Small letter oe
Š 352 s U+0160 Capital letter S with caron
š 353 s U+0161 Small letter s with caron
Ÿ 376 y U+0178 Capital letter Y with dieresis
*/

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
