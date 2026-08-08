package pobox_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/pobox"
)

var cases = []struct {
	In   string
	Want string
}{
	{"POST OFFICE BOX 11890", "PO BOX 11890"},
	{"PO Box 11890", "PO BOX 11890"},
	{"POB 11890", "PO BOX 11890"},
	{"PO Box #11890", "PO BOX 11890"},
	{"POB#11890", "PO BOX 11890"},
	{"Caller 11890", "PO BOX 11890"},
	{"FIRM CALLER 11890", "PO BOX 11890"},
	{"BIN 11890", "PO BOX 11890"},
	{"Lockbox 11890", "PO BOX 11890"},
	{"DRAWER 11890", "PO BOX 11890"},
	{"POST OFFICE BOX G", "PO BOX G"},
	{"PO Box K", "PO BOX K"},
	{"POB M", "PO BOX M"},
	{"Caller G", "PO BOX G"},
	{"FIRM CALLER G", "PO BOX G"},
	{"BIN J", "PO BOX J"},
	{"Lockbox N", "PO BOX N"},
	{"DRAWER S", "PO BOX S"},
}

func TestNormalize(t *testing.T) {
	for _, tc := range cases {
		out, err := pobox.Normalize(tc.In)
		if err != nil {
			t.Errorf("%s", err)
		}
		if out != tc.Want {
			t.Errorf("Unexpected normalized PO Box text %s for %s. Expected: %s", out, tc.In, tc.Want)
		}
	}
}

func TestNotPOBox(t *testing.T) {
	notpobox := "Main St"
	out, err := pobox.Normalize(notpobox)
	if err == nil {
		t.Errorf("Expected error for non-po Box: %s got: %s", notpobox, out)
	}
}
