//go:build ignore

// usaddressgenerator.go - Run via 'go generate ./pkg/address/parser/parsertest/'
// (or 'go run pkg/address/parser/parsertest/internal/usaddressgenerator/usaddressgenerator.go'
// from the repository root) to regenerate usaddress_generated.go in parsertest.
//
// It reads the vendored datamade/usaddress tagged-XML fixtures, maps each
// record's usaddress tags onto an address.Address (see ../../README.md for the
// mapping and why), runs that through the content normalizer, and writes the
// result out as Go source. The output is committed so that every change to a
// WantFields value is a diff a reviewer reads in a pull request, rather than
// something that moves silently whenever the mapping or a normalizer does
// (go-projectusat#223).
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"go/format"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// fixtureFiles are the datamade/usaddress tagged-address corpora parsertest
// vendors, in the order their cases appear in UsAddressCases. Two further
// files at the same upstream path, synthetic_osm_data.xml and
// synthetic_clean_osm_data.xml, are deliberately left out: they are
// machine-generated from OpenStreetMap rather than real addresses, and together
// they are larger than all four files below combined.
var fixtureFiles = []string{
	"labeled.xml",
	"multi_word_state_addresses.xml",
	"simple_address_patterns.xml",
	"us50_test_tagged.xml",
}

const outputName = "usaddress_generated.go"

// record is one <AddressString> case out of a usaddress fixture: the
// reconstructed plain string, and the tags that made it up in document order,
// text already trimmed. The order matters because a component with more than
// one word (StreetName "Martin"/"Luther"/"King", a PO box's USPSBoxType
// "P.O."/"Box") is repeated tag-per-word rather than given once, so
// rawAddress has to append in the order they appeared to reassemble it.
type record struct {
	Input string
	Tags  []tag
}

type tag struct {
	Name string
	Text string
}

// unsettledTags are usaddress tags this package's Address struct has no
// field to decompose into. A record carrying one is emitted with a nil
// WantFields and the tag named as the reason, rather than forcing one of our
// fields to hold data it was never meant to. See ../../README.md for why each
// is here; USPSBoxGroupID and USPSBoxGroupType (rural route and highway
// contract boxes) are expected to move out in a follow-on change.
var unsettledTags = map[string]bool{
	"IntersectionSeparator": true,
	"SecondStreetName":      true,
	"USPSBoxGroupID":        true,
	"USPSBoxGroupType":      true,
	"CornerOf":              true,
	"LandmarkName":          true,
	"NotAddress":            true,
}

// overridesName is the checked-in file of per-case overrides, next to this
// source. See loadOverrides.
const overridesName = "overrides.json"

// override replaces what the generator would emit for one input, and says why.
// With no Want the case is left unsettled with Reason as the report text: the
// missing WantFields is a decision, not a parse failure. With a Want those
// fields are the expected answer instead of the normalizer's.
//
// Prefer fixing the normalizer, which fixes every input and not one; an
// override is for a fixture the normalizer cannot be right about, such as a
// negative case a fixture author built to be rejected, or a Recipient run that
// mixes a business name with a person (Project US@ v1.0 pp.34-35 removes ATTN
// lines but keeps a business name, and nothing in the tags says where one ends,
// so guessing BusinessName would be invented data). Each entry needs a Reason
// for the reviewer of the diff that adds it (go-projectusat#241).
type override struct {
	Input  string      `json:"input"`
	Reason string      `json:"reason"`
	Want   *wantFields `json:"want,omitempty"`
}

// wantFields is the JSON form of an expected address.Address, named after its
// fields. POBox stands in for Type, the only type the mapping produces.
type wantFields struct {
	POBox               bool   `json:"poBox,omitempty"`
	BusinessName        string `json:"businessName,omitempty"`
	PrimaryNumber       string `json:"primaryNumber,omitempty"`
	Predirectional      string `json:"predirectional,omitempty"`
	StreetName          string `json:"streetName,omitempty"`
	StreetSuffix        string `json:"streetSuffix,omitempty"`
	Postdirectional     string `json:"postdirectional,omitempty"`
	SecondaryDesignator string `json:"secondaryDesignator,omitempty"`
	SecondaryNumber     string `json:"secondaryNumber,omitempty"`
	City                string `json:"city,omitempty"`
	Region              string `json:"region,omitempty"`
	Postal              string `json:"postal,omitempty"`
	Country             string `json:"country,omitempty"`
}

func (w wantFields) address() *address.Address {
	a := &address.Address{
		BusinessName:        w.BusinessName,
		PrimaryNumber:       w.PrimaryNumber,
		Predirectional:      w.Predirectional,
		StreetName:          w.StreetName,
		StreetSuffix:        w.StreetSuffix,
		Postdirectional:     w.Postdirectional,
		SecondaryDesignator: w.SecondaryDesignator,
		SecondaryNumber:     w.SecondaryNumber,
		City:                w.City,
		Region:              w.Region,
		Postal:              w.Postal,
		Country:             w.Country,
	}
	if w.POBox {
		a.Type = &pobox.POBoxAddress{}
	}
	return a
}

// loadOverrides reads overridesName, keyed by input. It fails on an unknown
// field, a missing reason, or a repeated input, so a typo cannot silently
// leave a case un-overridden.
func loadOverrides(path string) (map[string]override, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var list []override
	if err := dec.Decode(&list); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	overrides := map[string]override{}
	for _, o := range list {
		if o.Reason == "" {
			return nil, fmt.Errorf("%s: override for %q has no reason", filepath.Base(path), o.Input)
		}
		if _, dup := overrides[o.Input]; dup {
			return nil, fmt.Errorf("%s: %q is overridden twice", filepath.Base(path), o.Input)
		}
		overrides[o.Input] = o
	}
	return overrides, nil
}

// generatedCase is one emitted entry: the fixture file, the input string, and
// either a normalized WantFields or the reason there is none.
type generatedCase struct {
	File       string
	Input      string
	WantFields *address.Address
	Unsettled  string
}

func main() {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("usaddressgenerator: cannot locate own source file")
	}
	pkgDir := filepath.Join(filepath.Dir(self), "..", "..")
	fixtureDir := filepath.Join(pkgDir, "testdata", "usaddress")

	overrides, err := loadOverrides(filepath.Join(filepath.Dir(self), overridesName))
	if err != nil {
		log.Fatalf("usaddressgenerator: %v", err)
	}
	used := map[string]bool{}

	n := normalizer.NewContentNomalizer()
	var cases []generatedCase
	for _, name := range fixtureFiles {
		records, err := readFixture(filepath.Join(fixtureDir, name))
		if err != nil {
			log.Fatalf("usaddressgenerator: reading %s: %v", name, err)
		}
		for _, rec := range records {
			c := generatedCase{File: name, Input: rec.Input}
			raw, reason := rawAddress(rec.Tags)
			if raw == nil {
				c.Unsettled = reason
			} else if o, ok := overrides[rec.Input]; ok {
				used[rec.Input] = true
				if o.Want == nil {
					c.Unsettled = o.Reason
				} else {
					c.WantFields = o.Want.address()
				}
			} else if normalized, err := n.Normalize(raw); err != nil {
				c.Unsettled = "normalizer: " + err.Error()
			} else if again, err := n.Normalize(normalized); err != nil || !again.Equals(normalized) {
				// A WantFields value the normalizer would itself change is
				// not a settled answer: which pass is right is a normalizer
				// question this record cannot decide. Fail closed until the
				// normalizer is idempotent on the input.
				c.Unsettled = notIdempotent
			} else {
				c.WantFields = normalized
			}
			cases = append(cases, c)
		}
	}

	for _, input := range slices.Sorted(maps.Keys(overrides)) {
		if !used[input] {
			log.Fatalf("usaddressgenerator: %s: %q matches no settleable fixture case; remove it", overridesName, input)
		}
	}

	src, err := render(cases)
	if err != nil {
		log.Fatalf("usaddressgenerator: %v", err)
	}
	out := filepath.Join(pkgDir, outputName)
	if err := os.WriteFile(out, src, 0o644); err != nil {
		log.Fatalf("usaddressgenerator: %v", err)
	}

	// Counts only: the generated file is where the values are reviewed.
	settled := 0
	for _, c := range cases {
		if c.WantFields != nil {
			settled++
		}
	}
	fmt.Printf("usaddressgenerator: wrote %s: %d cases, %d with WantFields\n", outputName, len(cases), settled)
}

// readFixture walks one usaddress tagged-XML file and returns one record per
// <AddressString> element. Its children tag individual components
// (AddressNumber, StreetName, ...) interleaved with plain character data (the
// whitespace and punctuation between them); concatenating all character data
// in document order, tagged or not, reproduces the original address string
// exactly, so this walks tokens rather than unmarshaling into a fixed struct
// shape. The fixtures tag only one level deep inside <AddressString>, so
// tracking a single current tag name is enough to attribute character data to
// it.
func readFixture(path string) ([]record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	var records []record
	var inAddress bool
	var sb strings.Builder
	var tags []tag
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
						tags = append(tags, tag{Name: current, Text: text})
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "AddressString" {
				inAddress = false
				if s := strings.Join(strings.Fields(sb.String()), " "); s != "" {
					records = append(records, record{Input: s, Tags: tags})
				}
			} else if inAddress && t.Name.Local == current {
				current = ""
			}
		}
	}
	return records, nil
}

// rawAddress assembles the literal, unnormalized Address for one usaddress
// record per the mapping in ../../README.md. It returns nil and a reason for a
// record it will not decompose: one carrying a tag in unsettledTags, one
// carrying a tag the mapping has never seen (fail closed rather than silently
// drop it), or one with no tagged components at all.
func rawAddress(tags []tag) (*address.Address, string) {
	if len(tags) == 0 {
		return nil, "no tagged components"
	}

	a := &address.Address{}
	isPOBox := false
	for _, t := range tags {
		if unsettledTags[t.Name] {
			return nil, "usaddress tag " + t.Name + " has no Address field"
		}
		// usaddress keeps the separator that followed a token inside its
		// tag ("Chicago,", "IL,"). It delimits the component; it is not
		// part of it.
		text := strings.TrimRight(t.Text, ",;")
		switch t.Name {
		case "Recipient":
			// Project US@ v1.0 pp.34-35, "Remove Certain Words":
			// "Developers SHOULD replace or remove certain words ...
			// Remove ATTENTION, ATTN:". An attention or recipient line is
			// not part of the address, so it contributes to no field and
			// the rest of the record is decomposed as usual.
		case "BuildingName":
			a.BusinessName = textutil.JoinNonEmpty(" ", a.BusinessName, text)
		case "AddressNumber", "AddressNumberSuffix":
			a.PrimaryNumber = textutil.JoinNonEmpty(" ", a.PrimaryNumber, text)
		case "USPSBoxID":
			a.PrimaryNumber = textutil.JoinNonEmpty(" ", a.PrimaryNumber, text)
			isPOBox = true
		case "StreetNamePreDirectional":
			a.Predirectional = textutil.JoinNonEmpty(" ", a.Predirectional, text)
		case "StreetNamePreModifier", "StreetNamePreType", "StreetName", "StreetNamePostModifier":
			// StreetNamePreType is the route designator ahead of a
			// highway number ("US Hwy 50", "FM 544", "County Road 3668").
			// It is part of the route's name, the same way pkg/highways
			// claims a whole highway designation as a street name.
			a.StreetName = textutil.JoinNonEmpty(" ", a.StreetName, text)
		case "USPSBoxType":
			a.StreetName = textutil.JoinNonEmpty(" ", a.StreetName, text)
			isPOBox = true
		case "StreetNamePostType":
			a.StreetSuffix = textutil.JoinNonEmpty(" ", a.StreetSuffix, text)
		case "StreetNamePostDirectional":
			a.Postdirectional = textutil.JoinNonEmpty(" ", a.Postdirectional, text)
		case "OccupancyType", "SubaddressType":
			a.SecondaryDesignator = textutil.JoinNonEmpty(" ", a.SecondaryDesignator, text)
		case "OccupancyIdentifier", "SubaddressIdentifier":
			a.SecondaryNumber = textutil.JoinNonEmpty(" ", a.SecondaryNumber, text)
		case "PlaceName":
			a.City = textutil.JoinNonEmpty(" ", a.City, text)
		case "StateName":
			a.Region = textutil.JoinNonEmpty(" ", a.Region, text)
		case "ZipCode":
			a.Postal = textutil.JoinNonEmpty(" ", a.Postal, text)
		case "CountryName":
			a.Country = textutil.JoinNonEmpty(" ", a.Country, text)
		default:
			return nil, "usaddress tag " + t.Name + " is not in the mapping"
		}
	}
	// A bare "# 234" is a numbered unit with the designator unknown, which
	// Project US@ writes as designator "#". When a designator is already
	// present ("# APT 3"), the "#" is redundant and is dropped.
	if number, ok := strings.CutPrefix(a.SecondaryNumber, "#"); ok {
		a.SecondaryNumber = strings.TrimSpace(number)
		if a.SecondaryDesignator == "" {
			a.SecondaryDesignator = "#"
		}
	}
	if *a == (address.Address{}) {
		return nil, "no address components outside Recipient"
	}
	if isPOBox {
		a.Type = &pobox.POBoxAddress{}
	}
	return a, ""
}

// render writes cases out as the gofmt'd source of usaddress_generated.go,
// headed by a generation report a reviewer can check the diff against.
func render(cases []generatedCase) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("// Code generated by internal/usaddressgenerator/usaddressgenerator.go. DO NOT EDIT.\n\n")
	b.WriteString("package parsertest\n\n")
	writeReport(&b, cases)

	usesPOBox := slices.ContainsFunc(cases, func(c generatedCase) bool {
		return c.WantFields != nil && c.WantFields.Type != nil
	})
	b.WriteString("import (\n\t\"github.com/PortobelloAuth/go-projectusat/pkg/address\"\n")
	if usesPOBox {
		b.WriteString("\t\"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/pobox\"\n")
	}
	b.WriteString(")\n\n")

	b.WriteString("var usAddressGenerated = []usAddressCase{\n")
	for _, c := range cases {
		fmt.Fprintf(&b, "{\nFile: %q,\nInput: %q,\n", c.File, c.Input)
		if c.WantFields == nil {
			fmt.Fprintf(&b, "Unsettled: %q,\n", c.Unsettled)
		} else {
			lit, err := addressLiteral(c.WantFields)
			if err != nil {
				return nil, fmt.Errorf("%s %q: %w", c.File, c.Input, err)
			}
			fmt.Fprintf(&b, "WantFields: %s,\n", lit)
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")

	return format.Source(b.Bytes())
}

// writeReport writes the per-file and per-reason counts as a comment, so a
// change in how many cases settle, or why the rest do not, is visible at the
// top of the diff rather than only case by case.
func writeReport(b *bytes.Buffer, cases []generatedCase) {
	b.WriteString("// Generation report\n//\n")
	b.WriteString("// Cases with WantFields, per fixture file:\n//\n")
	for _, name := range fixtureFiles {
		total, settled := 0, 0
		for _, c := range cases {
			if c.File == name {
				total++
				if c.WantFields != nil {
					settled++
				}
			}
		}
		fmt.Fprintf(b, "//\t%-32s %4d / %4d\n", name, settled, total)
	}

	reasons := map[string]int{}
	settled := 0
	for _, c := range cases {
		if c.WantFields != nil {
			settled++
			continue
		}
		reasons[reasonCategory(c.Unsettled)]++
	}
	fmt.Fprintf(b, "//\t%-32s %4d / %4d\n//\n", "total", settled, len(cases))
	b.WriteString("// Cases without WantFields, by reason:\n//\n")
	for _, r := range slices.Sorted(maps.Keys(reasons)) {
		fmt.Fprintf(b, "//\t%4d  %s\n", reasons[r], r)
	}
	b.WriteString("\n")
}

// notIdempotent is the reason given when normalizing the normalized address a
// second time changes it again.
const notIdempotent = "normalizer: not idempotent, a second pass changes the result"

// reasonCategory groups normalizer errors by the field that failed rather
// than by the value, so the report stays a short list.
func reasonCategory(reason string) string {
	if rest, ok := strings.CutPrefix(reason, "normalizer: "); ok {
		field, _, _ := strings.Cut(rest, ":")
		return "normalizer: " + field
	}
	return reason
}

// typeLiteral returns the Go expression for a's Type. Only the types the
// mapping produces are supported; anything else is a generator bug.
func typeLiteral(a *address.Address) (string, bool) {
	if a == nil {
		return "", false
	}
	switch a.Type.(type) {
	case nil:
		return "", true
	case *pobox.POBoxAddress:
		return "&pobox.POBoxAddress{}", true
	}
	return "", false
}

// addressLiteral renders a as an &address.Address{...} literal carrying only
// its non-empty fields.
func addressLiteral(a *address.Address) (string, error) {
	typ, ok := typeLiteral(a)
	if !ok {
		return "", fmt.Errorf("unsupported address type %T", a.Type)
	}

	var b strings.Builder
	b.WriteString("&address.Address{\n")
	if typ != "" {
		fmt.Fprintf(&b, "Type: %s,\n", typ)
	}
	for _, f := range []struct{ name, value string }{
		{"BusinessName", a.BusinessName},
		{"Area", a.Area},
		{"PostalStation", a.PostalStation},
		{"PrimaryNumber", a.PrimaryNumber},
		{"Predirectional", a.Predirectional},
		{"StreetName", a.StreetName},
		{"StreetSuffix", a.StreetSuffix},
		{"Postdirectional", a.Postdirectional},
		{"SecondaryDesignator", a.SecondaryDesignator},
		{"SecondaryNumber", a.SecondaryNumber},
		{"Detail", a.Detail},
		{"City", a.City},
		{"Region", a.Region},
		{"Postal", a.Postal},
		{"Country", a.Country},
	} {
		if f.value != "" {
			fmt.Fprintf(&b, "%s: %q,\n", f.name, f.value)
		}
	}
	b.WriteString("}")
	return b.String(), nil
}
