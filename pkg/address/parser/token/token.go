package token

import (
	"regexp"
	"strings"
)

type Token struct {
	Text         string
	Line         int
	OfLines      int
	Position     int
	FollowsComma int // -1 means no, 0+ is the ordinal of the comma it follows on the line
}

var bycommaspace = regexp.MustCompile(`([^,\s]+|[,\s]+)`)
var whitespace = regexp.MustCompile(`\s`)

func Tokenize(source string) []Token {
	tokens := make([]Token, 0)
	lines := strings.Split(source, "\n")
	for lnum, ln := range lines {
		texts := bycommaspace.FindAllString(ln, -1)
		follows := -1
		pcomma := -1
		pos := 0
		for _, txt := range texts {
			txt = whitespace.ReplaceAllString(txt, "")
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
