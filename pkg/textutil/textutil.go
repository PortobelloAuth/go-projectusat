package textutil

import (
	"strings"
	"unicode"
)

// StripOptions controls which punctuation characters StripPunctuation retains.
type StripOptions struct {
	KeepHyphen bool
	KeepSlash  bool // fractional addresses 1/2
}

// CollapseSpace turns all whitespace runs into a single ASCII space and trims ends.
func CollapseSpace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inSpace := false
	started := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if started {
				inSpace = true
			}
			continue
		}
		if inSpace {
			b.WriteByte(' ')
			inSpace = false
		}
		b.WriteRune(r)
		started = true
	}
	return b.String()
}

// StripPunctuation removes Project US@ special characters, preserving:
//   - hyphen in primary number and ZIP+4 contexts (caller-controlled via options)
//   - pound sign #
//   - slash when KeepSlash is set (fractional addresses)
//
// Default: remove * , . ( ) " : ; ` @ & / ' and hyphens.
func StripPunctuation(s string, opts StripOptions) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if stripRune(r, opts) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func stripRune(r rune, opts StripOptions) bool {
	switch r {
	case '*', ',', '.', '(', ')', '"', ':', ';', '`', '@', '&', '\'', '\u2019':
		return true
	case '-':
		return !opts.KeepHyphen
	case '/':
		return !opts.KeepSlash
	default:
		return false
	}
}

// NormalizeUnknown returns "" if s is empty or equals "UNKNOWN" (any case).
func NormalizeUnknown(s string) string {
	if s == "" || strings.EqualFold(s, "UNKNOWN") {
		return ""
	}
	return s
}

// Upper is strings.ToUpper with UNKNOWN handling applied first.
func Upper(s string) string {
	return strings.ToUpper(NormalizeUnknown(s))
}
