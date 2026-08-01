package token_test

import (
	"slices"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
)

func TestTokenize(t *testing.T) {
	addr := "43 E 200 N, NORTH SALT LAKE, UT"
	want := []token.Token{
		{"43", 0, 1, 0, -1},
		{"E", 0, 1, 1, -1},
		{"200", 0, 1, 2, -1},
		{"N", 0, 1, 3, -1},
		{"NORTH", 0, 1, 4, 0},
		{"SALT", 0, 1, 5, -1},
		{"LAKE", 0, 1, 6, -1},
		{"UT", 0, 1, 7, 1},
	}

	out := token.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	// no commas
	addr = "43 E 200 N NORTH SALT LAKE UT"
	want = []token.Token{
		{"43", 0, 1, 0, -1},
		{"E", 0, 1, 1, -1},
		{"200", 0, 1, 2, -1},
		{"N", 0, 1, 3, -1},
		{"NORTH", 0, 1, 4, -1},
		{"SALT", 0, 1, 5, -1},
		{"LAKE", 0, 1, 6, -1},
		{"UT", 0, 1, 7, -1},
	}

	out = token.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	addr = "3253 W 9200 S, West Jordan, UT  84088"
	want = []token.Token{
		{"3253", 0, 1, 0, -1},
		{"W", 0, 1, 1, -1},
		{"9200", 0, 1, 2, -1},
		{"S", 0, 1, 3, -1},
		{"West", 0, 1, 4, 0},
		{"Jordan", 0, 1, 5, -1},
		{"UT", 0, 1, 6, 1},
		{"84088", 0, 1, 7, -1},
	}

	out = token.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	addr = "3253 W 9200 S\nWest Jordan, UT  84088\nUSA"
	want = []token.Token{
		{"3253", 0, 3, 0, -1},
		{"W", 0, 3, 1, -1},
		{"9200", 0, 3, 2, -1},
		{"S", 0, 3, 3, -1},
		{"West", 1, 3, 0, -1},
		{"Jordan", 1, 3, 1, -1},
		{"UT", 1, 3, 2, 0},
		{"84088", 1, 3, 3, -1},
		{"USA", 2, 3, 0, -1},
	}

	out = token.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}

	// 1445 VALLEYHIGH DR NW ROCHESTER MN 55901-0776 UNITED STATES
	addr = "1445 VALLEYHIGH DR NW ROCHESTER MN 55901-0776 UNITED STATES"
	want = []token.Token{
		{"1445", 0, 1, 0, -1},
		{"VALLEYHIGH", 0, 1, 1, -1},
		{"DR", 0, 1, 2, -1},
		{"NW", 0, 1, 3, -1},
		{"ROCHESTER", 0, 1, 4, -1},
		{"MN", 0, 1, 5, -1},
		{"55901-0776", 0, 1, 6, -1},
		{"UNITED", 0, 1, 7, -1},
		{"STATES", 0, 1, 8, -1},
	}

	out = token.Tokenize(addr)
	if !slices.Equal(out, want) {
		t.Fatalf("Tokenize returned unexpected tokens: %v want: %v", out, want)
	}
}
