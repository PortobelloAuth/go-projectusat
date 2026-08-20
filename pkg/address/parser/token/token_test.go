package token_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

func TestJoin(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"single token", "WYOMING", "WYOMING"},
		{"multi token name", "SOUTH CAROLINA", "SOUTH CAROLINA"},
		{
			// Tokenize drops commas and collapses runs of whitespace, so Join
			// cannot reproduce its input; it reproduces the tokens.
			name: "punctuation and spacing are not preserved",
			in:   "WEST JORDAN,  UT",
			want: "WEST JORDAN UT",
		},
		{"no tokens", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := token.Join(token.Tokenize(tc.in)); got != tc.want {
				t.Errorf("Join(Tokenize(%q)) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// Claims are made over sub-runs of a token slice, so Join has to be correct on
// a slice that does not start at the beginning.
func TestJoinSubSlice(t *testing.T) {
	tokens := token.Tokenize("8011 SOUTH CAROLINA AVE")

	if got := token.Join(tokens[1:3]); got != "SOUTH CAROLINA" {
		t.Errorf("Join(tokens[1:3]) = %q, want %q", got, "SOUTH CAROLINA")
	}
}

func TestLineEnd(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		start int
		want  int
	}{
		{"single line", "PO BOX 11890", 0, 3},
		{"from the middle of a line", "PO BOX 11890", 1, 3},
		{"first line of two", "PO BOX 11890\nDENVER CO 80201", 0, 3},
		{"last token of the first line", "PO BOX 11890\nDENVER CO 80201", 2, 3},
		{"second line of two", "PO BOX 11890\nDENVER CO 80201", 3, 6},
		{"third line", "ACME INC\nPO BOX 11890\nDENVER CO 80201", 2, 5},
		{"blank line between", "PO BOX 11890\n\nDENVER CO 80201", 0, 3},
		{"no tokens", "", 0, 0},
		{"start past the end", "PO BOX 11890", 9, 3},
		{"negative start", "PO BOX 11890", -1, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := token.Tokenize(tc.in)
			if got := token.LineEnd(tokens, tc.start); got != tc.want {
				t.Errorf("LineEnd(%q, %d) = %d, want %d", tc.in, tc.start, got, tc.want)
			}
		})
	}
}

// A span bounded by LineEnd never contains a line break, which is the property
// the recognizers depend on: Join flattens the span with spaces, so a break
// inside it would be invisible to them.
func TestLineEndBoundsASpanToOneLine(t *testing.T) {
	tokens := token.Tokenize("PO BOX 11890\nDENVER CO 80201")

	for start := range tokens {
		for end := start + 1; end <= token.LineEnd(tokens, start); end++ {
			for i := start; i < end; i++ {
				if tokens[i].Line != tokens[start].Line {
					t.Errorf("span [%d,%d) crosses from line %d to line %d",
						start, end, tokens[start].Line, tokens[i].Line)
				}
			}
		}
	}
}
