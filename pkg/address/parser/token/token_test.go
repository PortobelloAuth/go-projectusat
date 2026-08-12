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
