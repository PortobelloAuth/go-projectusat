package privatemailbox_test

import (
	"reflect"
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/privatemailbox"
)

// Every address here is published in the specification or invented. See
// CONTRIBUTING §5.

// reading is a claim flattened to the token text it covers, so cases read as
// "these words, this strongly, normalized to this".
type reading struct {
	text       string
	confidence claim.Confidence
	value      string
}

func readings(source string) []reading {
	tokens := token.Tokenize(source)
	var out []reading
	for _, c := range privatemailbox.Claims(tokens) {
		if len(c.Parts) != 1 || c.Parts[0].Part != claim.PartDetail {
			panic("a private mailbox is one Detail part")
		}
		p := c.Parts[0]
		out = append(out, reading{token.Join(tokens[p.Start:p.End()]), c.Confidence, p.Value})
	}

	return out
}

func TestClaims(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []reading
	}{
		{
			name: "PMB is exact",
			in:   "123 MAIN STREET PMB 4545\nHERNDON VA 22071",
			want: []reading{{"PMB 4545", claim.ConfidenceExact, "PMB 4545"}},
		},
		{
			name: "the numerical identifier is contested and normalized to PMB",
			in:   "# 234\nRR 1 BOX 12\nHERNDON VA 22071",
			want: []reading{{"# 234", claim.ConfidenceLikely, "PMB 234"}},
		},
		{
			name: "the spec writes the numerical identifier against the number",
			in:   "#234\n10 MAIN ST STE 11\nHERNDON VA 22071",
			want: []reading{{"#234", claim.ConfidenceLikely, "PMB 234"}},
		},
		{
			name: "lower case and a trailing box are read",
			in:   "po box 159753 pmb 3571\nherndon va 22071",
			want: []reading{{"pmb 3571", claim.ConfidenceExact, "PMB 3571"}},
		},
		{
			name: "an identifier alone is a fragment, not a weak claim",
			in:   "PMB\nHERNDON VA 22071",
			want: nil,
		},
		{
			name: "a number without a digit is not offered",
			in:   "# WEST\nHERNDON VA 22071",
			want: nil,
		},
		{
			name: "the pair does not cross a line break",
			in:   "PMB\n234 MAIN ST\nHERNDON VA 22071",
			want: nil,
		},
		{
			name: "no other identifier is permitted",
			in:   "PRIVATE MAILBOX 234\nHERNDON VA 22071",
			want: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := readings(c.in); !reflect.DeepEqual(got, c.want) {
				t.Errorf("Claims(%q)\n got %v\nwant %v", c.in, got, c.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"PMB 234": "PMB 234",
		"# 234":   "PMB 234",
		"#234":    "PMB 234",
		"pmb 12a": "PMB 12A",
	} {
		got, err := privatemailbox.Normalize(in)
		if err != nil || got != want {
			t.Errorf("Normalize(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	for _, in := range []string{"PMB", "#", "# WEST", "PMB 234 5", "STE 11"} {
		if got, err := privatemailbox.Normalize(in); err == nil {
			t.Errorf("Normalize(%q) = %q, want an error", in, got)
		}
	}
}
