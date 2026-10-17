package token

import (
	"regexp"
	"strings"
	"unicode"
)

type Token struct {
	Text         string
	Line         int
	OfLines      int
	Position     int
	FollowsComma int // -1 means no, 0+ is the ordinal of the comma it follows on the line
}

var bycommaspace = regexp.MustCompile(`([^,\s#]+|#[^,\s#]*|[,\s]+)`)
var whitespace = regexp.MustCompile(`\s`)

// embeddedpunctuation is punctuation no address part contains, removed from
// inside a token as well: "ATTN:" and "(IRVING" read as ATTN and IRVING.
// Apostrophes, hyphens, slashes, ampersands and "#" stay: names, fractions and
// secondary numbers use them.
var embeddedpunctuation = regexp.MustCompile(`[():;"|*]`)
var punctuationalone = regexp.MustCompile(`^(\||\*|\.|\(|\)|"|:|;|@|&|'|\x{2019})$`)

// dropPeriods removes the periods from a token unless a digit sits beside one:
// "U.S." and "N." read as US and N, while "39.4" is a number and "NO.2" is
// claimed as written. Nothing downstream can use a period, and the tables hold
// none, so "N. MIAMI" must reach them as "N MIAMI".
func dropPeriods(s string) string {
	r := []rune(s)
	out := make([]rune, 0, len(r))
	for i, c := range r {
		if c == '.' && !(i > 0 && unicode.IsDigit(r[i-1])) && !(i+1 < len(r) && unicode.IsDigit(r[i+1])) {
			continue
		}
		out = append(out, c)
	}

	return string(out)
}

func Tokenize(source string) []Token {
	tokens := make([]Token, 0)
	lines := strings.Split(source, "\n")
	for lnum, ln := range lines {
		texts := bycommaspace.FindAllString(ln, -1)
		follows := -1
		pcomma := -1
		pos := 0
		for _, txt := range texts {
			// if the token is only punctuation, remove it
			clean := punctuationalone.ReplaceAllString(txt, "")

			// clean up excess whitespace
			txt = dropPeriods(embeddedpunctuation.ReplaceAllString(whitespace.ReplaceAllString(clean, ""), ""))
			if len(txt) > 0 {
				if txt[0] == ',' {
					pcomma += len(txt) // there might be more than one comma
					follows = pcomma
				} else {
					tokens = append(tokens, Token{
						Text:         txt,
						Line:         lnum,
						OfLines:      len(lines),
						Position:     pos,
						FollowsComma: follows,
					})
					follows = -1
					pos += 1
				}
			}
		}
	}

	return tokens
}

// Join reconstructs the text of a run of tokens, separated by single spaces.
//
// Tokenize discards the whitespace and commas between tokens, so this is the
// canonical way to ask "what did these tokens say" — every package that matches
// a multi-token span against a vocabulary has to agree on the answer, or claims
// made over the same span will not compare.
func Join(tokens []Token) string {
	texts := make([]string, len(tokens))
	for i, t := range tokens {
		texts[i] = t.Text
	}

	return strings.Join(texts, " ")
}

// LineEnd reports the index one past the last token on start's line.
//
// A delivery address line is a line. The patterns that occupy one whole — a
// post office box, a rural route, a military street line — are recognized by
// handing a span to a Normalize function, and Join flattens that span with
// spaces, so a line break inside it is invisible to the recognizer. Without
// this bound "PO BOX\nDENVER CO 80201" matches the post office box pattern and
// takes the city for a box number.
//
// The same bound applies to a vocabulary phrase, for a different reason. There
// the pattern is not a line, so the question is whether a phrase hard-wrapped
// across two — "NORTH\nCAROLINA" — is input this library should still read as
// one. It is not: an address whose lines were wrapped by a fixed-width form
// rather than by address structure is expected to be cleaned up before it
// arrives here, so a break is a structural signal and not whitespace. The rule
// is that a claim never spans a line break, whatever kind of claim it is.
//
// A start outside the tokens has no line, and the answer is the end of them.
func LineEnd(tokens []Token, start int) int {
	if start < 0 || start >= len(tokens) {
		return len(tokens)
	}

	end := start
	for end < len(tokens) && tokens[end].Line == tokens[start].Line {
		end++
	}

	return end
}
