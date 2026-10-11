package textutil

import (
	"strings"
	"unicode/utf8"
)

// ASCIISet is a set of ASCII characters, for the character-class rewrites the
// vocabularies and address types run on every candidate span the parser
// considers. Done with package regexp, those rewrites were a fifth of parsing
// CPU under go-projectusat#194; the loops here do the same work in one pass and
// allocate nothing when there is nothing to change.
type ASCIISet [utf8.RuneSelf]bool

// NewASCIISet returns the set of the bytes in chars, which must all be ASCII.
func NewASCIISet(chars string) *ASCIISet {
	var set ASCIISet
	for i := 0; i < len(chars); i++ {
		if chars[i] >= utf8.RuneSelf {
			panic("textutil.NewASCIISet: non-ASCII byte in " + chars)
		}
		set[chars[i]] = true
	}
	return &set
}

const (
	digits = "0123456789"
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower  = "abcdefghijklmnopqrstuvwxyz"
)

// UpperAlnumSpace is [0-9A-Z ].
var UpperAlnumSpace = NewASCIISet(digits + upper + " ")

// AlnumSpace is [a-zA-Z0-9 ].
var AlnumSpace = NewASCIISet(digits + upper + lower + " ")

func (set *ASCIISet) has(c byte) bool {
	return c < utf8.RuneSelf && set[c]
}

// ReplaceRunsOutside replaces each maximal run of characters not in set with
// repl, exactly as regexp `[^set]+` with ReplaceAllString(s, repl) would.
//
// It works on bytes where the regexp works on runes, and gets the same runs:
// the set holds only ASCII, and every byte of a multi-byte or invalid UTF-8
// sequence is at or above utf8.RuneSelf, so a rune is outside the set exactly
// when each of its bytes is. FuzzReplaceRunsOutside holds it to the regexp.
func (set *ASCIISet) ReplaceRunsOutside(s, repl string) string {
	i := 0
	for i < len(s) && set.has(s[i]) {
		i++
	}
	if i == len(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	b.WriteString(s[:i])
	for i < len(s) {
		if set.has(s[i]) {
			b.WriteByte(s[i])
			i++
			continue
		}
		for i < len(s) && !set.has(s[i]) {
			i++
		}
		b.WriteString(repl)
	}
	return b.String()
}

// isRE2Space reports whether c is in RE2's \s class, [\t\n\f\r ]. That is
// deliberately not unicode.IsSpace: it leaves out \v and every non-ASCII space.
func isRE2Space(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\f' || c == '\r'
}

// CollapseRE2Space replaces each run of [\t\n\f\r ] with a single space,
// exactly as regexp `\s+` with ReplaceAllString(s, " ") would. Unlike
// CollapseSpace it does not trim, and it leaves other Unicode spaces alone.
// FuzzCollapseRE2Space holds it to the regexp.
func CollapseRE2Space(s string) string {
	// Find the first place the output differs: a run longer than one byte, or
	// a single whitespace byte that is not already a space.
	i := 0
	for ; i < len(s); i++ {
		if !isRE2Space(s[i]) {
			continue
		}
		if s[i] != ' ' || (i+1 < len(s) && isRE2Space(s[i+1])) {
			break
		}
	}
	if i == len(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	b.WriteString(s[:i])
	for i < len(s) {
		if !isRE2Space(s[i]) {
			b.WriteByte(s[i])
			i++
			continue
		}
		for i < len(s) && isRE2Space(s[i]) {
			i++
		}
		b.WriteByte(' ')
	}
	return b.String()
}
