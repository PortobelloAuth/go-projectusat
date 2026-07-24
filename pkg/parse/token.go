package parse

import (
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/highways"
	"github.com/PortobelloAuth/go-projectusat/pkg/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/puertorico"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// Token is an address token with relative position and punctuation context.
// Tokens are the unit of scoring and component assignment.
type Token struct {
	// Text is the uppercase alphanumeric (plus retained # / - / .) form.
	Text string
	// Index is the 0-based position among tokens in the full address.
	Index int
	// CommaAfter is true when a comma immediately followed this token in the source.
	CommaAfter bool
	// Line is the 0-based logical line index after newline splitting.
	Line int
}

// Scores holds per-package likelihood that a token belongs to that vocabulary.
// 0 means "not this part"; higher is more confident (typically 80–100).
type Scores struct {
	Region       int
	Directional  int
	StreetSuffix int
	Secondary    int
	Highway      int
	Military     int
	PuertoRico   int
	Postal       int
}

// ScoreToken scores a single token against each address-part package.
func ScoreToken(text string) Scores {
	var s Scores
	if v, _ := region.Score(text); v > 0 {
		s.Region = v
	}
	if v, _ := directionals.Score(text); v > 0 {
		s.Directional = v
	}
	if v, _ := streetsuffixes.Score(text); v > 0 {
		s.StreetSuffix = v
	}
	if v, _ := secondaryunit.Score(text); v > 0 {
		s.Secondary = v
	}
	if v, _ := highways.Score(text); v > 0 {
		s.Highway = v
	}
	if v, _ := military.Score(text); v > 0 {
		s.Military = v
	}
	if v, _ := puertorico.Score(text); v > 0 {
		s.PuertoRico = v
	}
	if postalScore(text) > 0 {
		s.Postal = postalScore(text)
	}
	return s
}

func postalScore(tok string) int {
	u := strings.ToUpper(strings.TrimSpace(tok))
	if u == "" {
		return 0
	}
	if usZIPCompact.MatchString(u) {
		return 100
	}
	if caPostalCompact.MatchString(u) {
		return 100
	}
	if caPostalFSA.MatchString(u) || caPostalLDU.MatchString(u) {
		return 70
	}
	// Loose international alnum postal
	hasDigit := false
	for _, r := range u {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if hasDigit && len(u) >= 3 && len(u) <= 12 {
		return 40
	}
	return 0
}

// Tokenize splits raw address text into Tokens, tracking newlines as line breaks
// and commas as CommaAfter on the preceding token. Whitespace separates tokens.
// Hash, hyphen, and digit-adjacent periods are preserved in token text (after
// uppercasing / space collapse of the source segments).
func Tokenize(raw string) []Token {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	var tokens []Token
	idx := 0
	for lineNo, line := range strings.Split(raw, "\n") {
		line = textutil.CollapseSpace(line)
		if line == "" {
			continue
		}
		// Walk the line, splitting on whitespace and recording commas.
		var cur strings.Builder
		flush := func(commaAfter bool) {
			if cur.Len() == 0 {
				return
			}
			text := strings.ToUpper(cur.String())
			tokens = append(tokens, Token{
				Text:       text,
				Index:      idx,
				CommaAfter: commaAfter,
				Line:       lineNo,
			})
			idx++
			cur.Reset()
		}
		for _, r := range line {
			switch {
			case r == ',':
				flush(true)
			case unicode.IsSpace(r):
				flush(false)
			default:
				cur.WriteRune(r)
			}
		}
		flush(false)
	}
	return tokens
}

// texts returns the Text fields of tokens.
func texts(tokens []Token) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t.Text
	}
	return out
}
