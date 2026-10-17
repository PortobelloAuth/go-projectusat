package parsertest

import (
	"embed"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
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
// not a decomposition into this package's Address fields on their own — so
// go-projectusat#223 built usAddressWantFields to translate the ones that do
// correspond field-for-field (see its doc comment for the table and for
// which tags don't). Want is still left empty here: it is this package's
// Pub 28 rendering of the string, which usaddress's tagging says nothing
// about one way or the other.
//
// Source: https://github.com/datamade/usaddress/tree/main/measure_performance/test_data
// at commit aa7699b5 (see usAddressCommit), MIT licensed, © 2014 The Atlanta
// Journal Constitution — see testdata/usaddress/NOTICE.
var UsAddressCases = buildUsAddressCases()

func buildUsAddressCases() []Case {
	var cases []Case
	for _, name := range usAddressFiles {
		records, err := readUsAddressFixture(name)
		if err != nil {
			panic(fmt.Sprintf("parsertest: reading usaddress fixture %s: %v", name, err))
		}
		for _, rec := range records {
			cases = append(cases, Case{
				Source: "usaddress",
				Note: fmt.Sprintf(
					"datamade/usaddress@%s measure_performance/test_data/%s (MIT, see testdata/usaddress/NOTICE); Want left unsettled — see UsAddressCases doc comment",
					usAddressCommit[:8], name,
				),
				Input:      rec.Input,
				WantFields: usAddressWantFields(rec.Tags),
			})
		}
	}
	return cases
}

// usAddressRecord is one <AddressString> case out of a usaddress fixture:
// the reconstructed plain string, and the tags that made it up in document
// order, text already trimmed. The order matters because a component with
// more than one word (StreetName "Martin"/"Luther"/"King", a PO box's
// USPSBoxType "P.O."/"Box") is repeated tag-per-word rather than given once,
// so usAddressWantFields has to append in the order they appeared to
// reassemble it.
type usAddressRecord struct {
	Input string
	Tags  []usAddressTag
}

type usAddressTag struct {
	Name string
	Text string
}

// readUsAddressFixture walks one usaddress tagged-XML file and returns one
// usAddressRecord per <AddressString> element. Its children tag individual
// components (AddressNumber, StreetName, ...) interleaved with plain
// character data (the whitespace and punctuation between them);
// concatenating all character data in document order, tagged or not,
// reproduces the original address string exactly, so this walks tokens
// rather than unmarshaling into a fixed struct shape. The fixtures tag only
// one level deep inside <AddressString>, so tracking a single current tag
// name is enough to attribute character data to it.
func readUsAddressFixture(name string) ([]usAddressRecord, error) {
	f, err := usAddressFixtures.Open("testdata/usaddress/" + name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	var records []usAddressRecord
	var inAddress bool
	var sb strings.Builder
	var tags []usAddressTag
	var current string
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
				tags = nil
			} else if inAddress {
				current = t.Name.Local
			}
		case xml.CharData:
			if inAddress {
				sb.Write(t)
				if current != "" {
					if text := strings.TrimSpace(string(t)); text != "" {
						tags = append(tags, usAddressTag{Name: current, Text: text})
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "AddressString" {
				inAddress = false
				if s := strings.Join(strings.Fields(sb.String()), " "); s != "" {
					records = append(records, usAddressRecord{Input: s, Tags: tags})
				}
			} else if inAddress && t.Name.Local == current {
				current = ""
			}
		}
	}
	return records, nil
}

// usAddressUnsettledTags are usaddress tags this package's Address struct
// has no field to decompose into. IntersectionSeparator, SecondStreetName
// and Recipient were the ones go-projectusat#223 found and agreed on — an
// intersection isn't a single street line, and an attention line isn't a
// Pub 28 field either. StreetNamePreType, USPSBoxGroupID, USPSBoxGroupType,
// CornerOf, LandmarkName and NotAddress turned up while building the
// mapping below: real tags in the vendored fixtures (rural-route box
// groups, highway-route prefixes like "US"/"FM"/"PR", landmark and
// corner-reference addresses, and explicitly-not-an-address rows) that
// #223's agreed table never mentioned, mapped or excluded. Rather than
// invent a reading for any of them, a case carrying one gets a nil
// WantFields — the same "unsettled, not failing" treatment Case.WantFields
// documents for every case without ground truth — instead of forcing one
// of our fields to hold data it was never meant to.
var usAddressUnsettledTags = map[string]bool{
	"IntersectionSeparator": true,
	"SecondStreetName":      true,
	"Recipient":             true,
	"StreetNamePreType":     true,
	"USPSBoxGroupID":        true,
	"USPSBoxGroupType":      true,
	"CornerOf":              true,
	"LandmarkName":          true,
	"NotAddress":            true,
}

// usAddressWantFields builds the Case.WantFields decomposition for one
// usaddress record, per the mapping go-projectusat#223 agreed on:
//
//	BusinessName:         BuildingName
//	PrimaryNumber:        AddressNumber (+ AddressNumberSuffix), USPSBoxID
//	Predirectional:       StreetNamePreDirectional
//	StreetName:           StreetName (+ StreetNamePreModifier, StreetNamePostModifier), USPSBoxType
//	StreetSuffix:         StreetNamePostType
//	Postdirectional:      StreetNamePostDirectional
//	SecondaryDesignator:  OccupancyType / SubaddressType
//	SecondaryNumber:      OccupancyIdentifier / SubaddressIdentifier
//	City:                 PlaceName
//	Region:               StateName
//	Postal:               ZipCode
//	Country:              CountryName
//	Type:                 &pobox.POBoxAddress{}, when USPSBoxID or USPSBoxType is present
//
// Area, PostalStation and Detail stay blank: usaddress has no tags for
// urbanizations, postal stations, or PMB-style detail numbers.
//
// Country is the one correction past what #223 settled on: the issue
// discussion assumed usaddress had no country tag at all, but
// labeled.xml and multi_word_state_addresses.xml both carry CountryName
// (5 cases each, "United"/"States"), so it is mapped the same way PlaceName
// and StateName are.
//
// A record whose tags include anything in usAddressUnsettledTags, or a tag
// this mapping has never seen at all, returns nil — fail closed rather than
// silently drop a usaddress tag this package doesn't know how to read yet.
// A record with no tagged components (todo once there's a fixture with
// character data outside every tag) also returns nil: nothing to assert.
func usAddressWantFields(tags []usAddressTag) *address.Address {
	if len(tags) == 0 {
		return nil
	}

	a := &address.Address{}
	isPOBox := false
	for _, t := range tags {
		if usAddressUnsettledTags[t.Name] {
			return nil
		}
		switch t.Name {
		case "BuildingName":
			a.BusinessName = textutil.JoinNonEmpty(" ", a.BusinessName, t.Text)
		case "AddressNumber", "AddressNumberSuffix":
			a.PrimaryNumber = textutil.JoinNonEmpty(" ", a.PrimaryNumber, t.Text)
		case "USPSBoxID":
			a.PrimaryNumber = textutil.JoinNonEmpty(" ", a.PrimaryNumber, t.Text)
			isPOBox = true
		case "StreetNamePreDirectional":
			a.Predirectional = textutil.JoinNonEmpty(" ", a.Predirectional, t.Text)
		case "StreetNamePreModifier", "StreetName", "StreetNamePostModifier":
			a.StreetName = textutil.JoinNonEmpty(" ", a.StreetName, t.Text)
		case "USPSBoxType":
			a.StreetName = textutil.JoinNonEmpty(" ", a.StreetName, t.Text)
			isPOBox = true
		case "StreetNamePostType":
			a.StreetSuffix = textutil.JoinNonEmpty(" ", a.StreetSuffix, t.Text)
		case "StreetNamePostDirectional":
			a.Postdirectional = textutil.JoinNonEmpty(" ", a.Postdirectional, t.Text)
		case "OccupancyType", "SubaddressType":
			a.SecondaryDesignator = textutil.JoinNonEmpty(" ", a.SecondaryDesignator, t.Text)
		case "OccupancyIdentifier", "SubaddressIdentifier":
			a.SecondaryNumber = textutil.JoinNonEmpty(" ", a.SecondaryNumber, t.Text)
		case "PlaceName":
			a.City = textutil.JoinNonEmpty(" ", a.City, t.Text)
		case "StateName":
			a.Region = textutil.JoinNonEmpty(" ", a.Region, t.Text)
		case "ZipCode":
			a.Postal = textutil.JoinNonEmpty(" ", a.Postal, t.Text)
		case "CountryName":
			a.Country = textutil.JoinNonEmpty(" ", a.Country, t.Text)
		default:
			return nil
		}
	}
	if isPOBox {
		a.Type = &pobox.POBoxAddress{}
	}
	return a
}
