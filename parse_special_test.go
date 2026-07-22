package goprojectusat

import "testing"

func TestRewriteSpecialStreetLineRuralRoute(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Rural Route 91 Box A7", "RR 91 BOX A7"},
		{"RURAL ROUTE 91 BOX A7", "RR 91 BOX A7"},
		{"RFD 61 #87b", "RR 61 BOX 87B"},
		{"RD 61 # 87b", "RR 61 BOX 87B"},
		{"RR0061 #87b", "RR 61 BOX 87B"},
		{"RR 0061 BOX 87B", "RR 61 BOX 87B"},
		{"rr 12 box 3", "RR 12 BOX 3"},
		{"RUTA 5 BOX 12", "RR 5 BOX 12"},
		{"RUTA RURAL 8 BOX 9A", "RR 8 BOX 9A"},
		// Phrase variants: RFD Route, Rural Route NO., Spanish box, glued hash
		{"RFD Route 61 Box 87b", "RR 61 BOX 87B"},
		{"Rural Route NO. 91 Box A7", "RR 91 BOX A7"},
		{"Rural Route Number 91 Box A7", "RR 91 BOX A7"},
		{"RUTA 5 BUZON 12", "RR 5 BOX 12"},
		{"RUTA 5 BZN 12", "RR 5 BOX 12"},
		{"RR0061#87b", "RR 61 BOX 87B"},
		// Trailing junk after box is dropped
		{"RR 91 BOX A7 SPRINGFIELD", "RR 91 BOX A7"},
		{"Rural Route 2 Box 10 Main Street", "RR 2 BOX 10"},
	}
	for _, tc := range cases {
		got, ok := rewriteSpecialStreetLine(tc.in)
		if !ok {
			t.Errorf("rewriteSpecialStreetLine(%q) = false, want true", tc.in)
			continue
		}
		if got != tc.want {
			t.Errorf("rewriteSpecialStreetLine(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRewriteSpecialStreetLinePOBox(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Post office Box G", "PO BOX G"},
		{"PO Box 11890", "PO BOX 11890"},
		{"P.O. Box 11890", "PO BOX 11890"},
		{"POST OFFICE BOX 42", "PO BOX 42"},
		{"po box abc", "PO BOX ABC"},
		{"POBOX 99", "PO BOX 99"},
	}
	for _, tc := range cases {
		got, ok := rewriteSpecialStreetLine(tc.in)
		if !ok {
			t.Errorf("rewriteSpecialStreetLine(%q) = false, want true", tc.in)
			continue
		}
		if got != tc.want {
			t.Errorf("rewriteSpecialStreetLine(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRewriteSpecialStreetLineNoMatch(t *testing.T) {
	// Bare RD/road highway forms must NOT become rural route without BOX/#.
	// Ordinary streets and incomplete RR/PO must fall through.
	noMatch := []string{
		"RD 5A",
		"RD 61",
		"ROAD 5A",
		"123 Main Street",
		"RR 91",          // route without box
		"RURAL ROUTE 91", // without box
		"RFD 61",         // without box
		"BOX 12",         // no RR prefix
		"PO BOX",         // missing box id
		"POST OFFICE",    // incomplete
		"Main Street Box 4",
	}
	for _, in := range noMatch {
		if got, ok := rewriteSpecialStreetLine(in); ok {
			t.Errorf("rewriteSpecialStreetLine(%q) = %q, true; want no match", in, got)
		}
	}
}
