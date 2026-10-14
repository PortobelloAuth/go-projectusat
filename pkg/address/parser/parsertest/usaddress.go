package parsertest

import (
	"embed"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// usAddressCommit is the datamade/usaddress commit the fixtures in
// testdata/usaddress were vendored from; see testdata/usaddress/NOTICE for
// the license and the full provenance of each file.
const usAddressCommit = "aa7699b53a0843fc443f9e87285b88cbd9eaf50a"

//go:embed testdata/usaddress/*.xml
var usAddressFixtures embed.FS

// usAddressFiles are the datamade/usaddress tagged-address corpora this suite
// vendors. Two further files at the same path, synthetic_osm_data.xml and
// synthetic_clean_osm_data.xml, are deliberately left out of this pass: they
// are machine-generated from OpenStreetMap rather than real addresses, and
// together they are larger than all four files below combined. Pulling them
// in is a scope decision for a later issue, not this one, if the volume this
// suite already gives #194's load testing is not enough.
var usAddressFiles = []string{
	"labeled.xml",
	"multi_word_state_addresses.xml",
	"simple_address_patterns.xml",
	"us50_test_tagged.xml",
}

// UsAddressCases are the tagged-address test corpora datamade/usaddress
// vendors for its own measure_performance suite, pulled in for go-projectusat#218:
// a larger body of real-world address strings to exercise the parser against
// (and, per Aaron's comment there, give #194's load/perf testing more volume
// than our own corpora provide) at a scale too large to transcribe by hand.
//
// Each file tags every address component (AddressNumber, StreetName,
// PlaceName, ...) individually. Those tags are usaddress's own token labels,
// not a decomposition into this package's Address fields, and they carry no
// claim about what this project's Pub 28 rendering of the address should be
// — so Want and WantFields are left empty here rather than guessed at. See
// Case's doc comment: an empty Want means ground truth is not yet known, and
// inventing one from a label scheme that isn't ours would be worse than
// admitting we don't have it. What this suite does assert, via corpusmeasure
// and go test, is narrower but still real: that every one of these addresses
// parses without error.
//
// Source: https://github.com/datamade/usaddress/tree/main/measure_performance/test_data
// at commit aa7699b5 (see usAddressCommit), MIT licensed, © 2014 The Atlanta
// Journal Constitution — see testdata/usaddress/NOTICE.
var UsAddressCases = buildUsAddressCases()

func buildUsAddressCases() []Case {
	var cases []Case
	for _, name := range usAddressFiles {
		addrs, err := readUsAddressFixture(name)
		if err != nil {
			panic(fmt.Sprintf("parsertest: reading usaddress fixture %s: %v", name, err))
		}
		for _, in := range addrs {
			cases = append(cases, Case{
				Source: "usaddress",
				Note: fmt.Sprintf(
					"datamade/usaddress@%s measure_performance/test_data/%s (MIT, see testdata/usaddress/NOTICE); Want left unsettled — see UsAddressCases doc comment",
					usAddressCommit[:8], name,
				),
				Input: in,
			})
		}
	}
	return cases
}

// readUsAddressFixture reconstructs the plain address strings out of one
// usaddress tagged-XML file. Each case is an <AddressString> element whose
// children tag individual components (AddressNumber, StreetName, ...)
// interleaved with plain character data (the whitespace and punctuation
// between them); concatenating all character data in document order, tagged
// or not, reproduces the original address string exactly, so this walks
// tokens rather than unmarshaling into a fixed struct shape.
func readUsAddressFixture(name string) ([]string, error) {
	f, err := usAddressFixtures.Open("testdata/usaddress/" + name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	var addrs []string
	var inAddress bool
	var sb strings.Builder
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "AddressString" {
				inAddress = true
				sb.Reset()
			}
		case xml.CharData:
			if inAddress {
				sb.Write(t)
			}
		case xml.EndElement:
			if t.Name.Local == "AddressString" {
				inAddress = false
				if s := strings.Join(strings.Fields(sb.String()), " "); s != "" {
					addrs = append(addrs, s)
				}
			}
		}
	}
	return addrs, nil
}
