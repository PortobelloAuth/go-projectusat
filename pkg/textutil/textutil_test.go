package textutil_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

func TestCollapseSpace(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"  ", ""},
		{"hello", "hello"},
		{"hello world", "hello world"},
		{"  hello   world  ", "hello world"},
		{"hello\t\tworld", "hello world"},
		{"hello\nworld", "hello world"},
		{"hello\r\nworld", "hello world"},
		{"a \t\n\r  b", "a b"},
		{"  MAIN  ST  ", "MAIN ST"},
		{"\u00A0nbsp\u00A0around\u00A0", "nbsp around"}, // non-breaking space
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := textutil.CollapseSpace(tc.in)
			if got != tc.want {
				t.Fatalf("CollapseSpace(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestStripPunctuation(t *testing.T) {
	cases := []struct {
		name string
		in   string
		opts textutil.StripOptions
		want string
	}{
		{
			name: "default strips specials",
			in:   `*Hello*, "World". (Test): foo; bar` + "`" + `baz@qux&`,
			opts: textutil.StripOptions{},
			want: `Hello World Test foo bar` + `bazqux`,
		},
		{
			name: "default removes hyphen and slash",
			in:   "12345-6789 1/2 MAIN-ST",
			opts: textutil.StripOptions{},
			want: "123456789 12 MAINST",
		},
		{
			name: "keep hyphen for ZIP+4",
			in:   "12345-6789",
			opts: textutil.StripOptions{KeepHyphen: true},
			want: "12345-6789",
		},
		{
			name: "keep hyphen for primary number",
			in:   "112-10 BRONX RD",
			opts: textutil.StripOptions{KeepHyphen: true},
			want: "112-10 BRONX RD",
		},
		{
			name: "keep slash for fraction",
			in:   "123 1/2 MAIN ST",
			opts: textutil.StripOptions{KeepSlash: true},
			want: "123 1/2 MAIN ST",
		},
		{
			name: "keep hyphen and slash",
			in:   "112-10 1/2 MAIN ST",
			opts: textutil.StripOptions{KeepHyphen: true, KeepSlash: true},
			want: "112-10 1/2 MAIN ST",
		},
		{
			name: "pound sign preserved",
			in:   "123 MAIN ST #4",
			opts: textutil.StripOptions{},
			want: "123 MAIN ST #4",
		},
		{
			name: "mixed punctuation with pound",
			in:   "APT. #4, BLDG*2",
			opts: textutil.StripOptions{},
			want: "APT #4 BLDG2",
		},
		{
			name: "period in street name",
			in:   "N. MAIN ST.",
			opts: textutil.StripOptions{},
			want: "N MAIN ST",
		},
		{
			name: "empty",
			in:   "",
			opts: textutil.StripOptions{},
			want: "",
		},
		{
			name: "letters and digits untouched",
			in:   "ABC123",
			opts: textutil.StripOptions{},
			want: "ABC123",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := textutil.StripPunctuation(tc.in, tc.opts)
			if got != tc.want {
				t.Fatalf("StripPunctuation(%q, %+v) = %q, want %q", tc.in, tc.opts, got, tc.want)
			}
		})
	}
}

func TestNormalizeUnknown(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"UNKNOWN", ""},
		{"unknown", ""},
		{"Unknown", ""},
		{"UnKnOwN", ""},
		{"hello", "hello"},
		{"UNKNOWN CITY", "UNKNOWN CITY"}, // whole field token only
		{"KNOWN", "KNOWN"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := textutil.NormalizeUnknown(tc.in)
			if got != tc.want {
				t.Fatalf("NormalizeUnknown(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestUpper(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"UNKNOWN", ""},
		{"unknown", ""},
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"main st", "MAIN ST"},
		{"café", "CAFÉ"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := textutil.Upper(tc.in)
			if got != tc.want {
				t.Fatalf("Upper(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
