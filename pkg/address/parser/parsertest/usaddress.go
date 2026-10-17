package parsertest

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
)

//go:generate go run ./internal/usaddressgenerator/usaddressgenerator.go

// usAddressCommit is the datamade/usaddress commit the fixtures in
// testdata/usaddress were vendored from; see testdata/usaddress/NOTICE for
// the license and the full provenance of each file.
const usAddressCommit = "aa7699b53a0843fc443f9e87285b88cbd9eaf50a"

// UsAddressCases are the tagged-address test corpora datamade/usaddress
// vendors for its own measure_performance suite, pulled in for go-projectusat#218:
// a larger body of real-world address strings to exercise the parser against
// (and, per Aaron's comment there, give #194's load/perf testing more volume
// than our own corpora provide) at a scale too large to transcribe by hand.
//
// Each file tags every address component (AddressNumber, StreetName,
// PlaceName, ...) individually. WantFields is those tags mapped onto this
// package's Address fields and then put through the content normalizer, so
// it holds what Project US@ says the decomposition should be rather than
// usaddress's literal text. That is done ahead of time by a generator and
// committed as usaddress_generated.go, so that every WantFields value is
// reviewed in a pull request and cannot drift with the implementation it is
// checking (go-projectusat#223). README.md in this directory has the tag
// mapping, the tags left unsettled, and how to regenerate.
//
// Want is still left empty: it is this package's Pub 28 rendering of the
// string, which usaddress's tagging says nothing about one way or the other.
//
// Source: https://github.com/datamade/usaddress/tree/main/measure_performance/test_data
// at commit aa7699b5 (see usAddressCommit), MIT licensed, © 2014 The Atlanta
// Journal Constitution — see testdata/usaddress/NOTICE.
var UsAddressCases = buildUsAddressCases()

// usAddressCase is one generated entry of usaddress_generated.go. Unsettled
// says why WantFields is nil, and is empty when it is not.
type usAddressCase struct {
	File       string
	Input      string
	WantFields *address.Address
	Unsettled  string
}

func buildUsAddressCases() []Case {
	cases := make([]Case, 0, len(usAddressGenerated))
	for _, g := range usAddressGenerated {
		note := fmt.Sprintf(
			"datamade/usaddress@%s measure_performance/test_data/%s (MIT, see testdata/usaddress/NOTICE); Want left unsettled — see UsAddressCases doc comment",
			usAddressCommit[:8], g.File,
		)
		if g.Unsettled != "" {
			note += "; WantFields unsettled: " + g.Unsettled
		}
		cases = append(cases, Case{
			Source:     "usaddress",
			Note:       note,
			Input:      g.Input,
			WantFields: g.WantFields,
		})
	}
	return cases
}
