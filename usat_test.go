package goprojectusat_test

import (
	"fmt"
	"sync"
	"testing"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// legacyNormalize is the body of the package-level Normalize as it was before
// USAt existed (go-projectusat#195), kept verbatim so the tests below compare
// USAt against the old behaviour rather than against itself.
func legacyNormalize(source string, opts ...goprojectusat.USAtNormalizeOption) (string, error) {
	popts := &parser.AddressParsingOptions{}
	nopts := &normalizer.AddressNormalizationOptions{}
	fopts := &address.FormatOptions{}
	for _, fn := range opts {
		err := fn(popts, nopts, fopts)
		if err != nil {
			return "", fmt.Errorf("Error setting normailzation options: %w", err)
		}
	}
	p, err := parser.New(*popts)
	if err != nil {
		return "", fmt.Errorf("Unable to instantiate parser: %w", err)
	}
	addr, err := p.Parse(source)
	if err != nil {
		return "", fmt.Errorf("Unable to parse address: %w", err)
	}
	n := normalizer.NewNomalizer(*nopts)
	addr, err = n.Normalize(addr)
	if err != nil {
		return "", fmt.Errorf("Unable to normalize address: %w", err)
	}
	return addr.Format(*fopts)
}

var optionSets = map[string][]goprojectusat.Option{
	"none":                  nil,
	"content":               {goprojectusat.WithContentNormalization()},
	"matching":              {goprojectusat.WithMatchingNormalization()},
	"single-line":           {goprojectusat.WithSingleLineFormatting()},
	"content+single-line":   {goprojectusat.WithContentNormalization(), goprojectusat.WithSingleLineFormatting()},
	"matching+single-line":  {goprojectusat.WithMatchingNormalization(), goprojectusat.WithSingleLineFormatting()},
	"fuzzy+hash":            {goprojectusat.WithFuzzyNormalization(), goprojectusat.WithSecondaryAsHash()},
	"substitute-diacritics": {goprojectusat.WithDiacriticNormalization(diacritics.SubstituteDiacritics), goprojectusat.WithDiacriticFormatting(diacritics.SubstituteDiacritics)},
	"matching-then-content": {goprojectusat.WithMatchingNormalization(), goprojectusat.WithContentNormalization()},
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TestUSAtMatchesLegacyNormalize runs every parsertest case through the
// pre-#195 pipeline, a USAt, and the package-level Normalize, under each
// option set, and requires identical output and identical errors.
func TestUSAtMatchesLegacyNormalize(t *testing.T) {
	for name, opts := range optionSets {
		t.Run(name, func(t *testing.T) {
			u, err := goprojectusat.New(opts...)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			for _, c := range parsertest.Cases {
				want, wantErr := legacyNormalize(c.Input, opts...)
				got, gotErr := u.Normalize(c.Input)
				if got != want || errString(gotErr) != errString(wantErr) {
					t.Errorf("USAt.Normalize(%q)\n got  %q, %v\n want %q, %v", c.Input, got, gotErr, want, wantErr)
				}
				pkg, pkgErr := goprojectusat.Normalize(c.Input, opts...)
				if pkg != want || errString(pkgErr) != errString(wantErr) {
					t.Errorf("Normalize(%q)\n got  %q, %v\n want %q, %v", c.Input, pkg, pkgErr, want, wantErr)
				}
			}
		})
	}
}

// TestUSAtDoesNotRebuildPerCall shows the per-call construction #195 is about
// is gone: a USAt allocates strictly less per Normalize than the package-level
// Normalize does when it is given options and must build one for the call.
// The stub parser keeps the measurement to construction, not parsing.
func TestUSAtDoesNotRebuildPerCall(t *testing.T) {
	stub := parser.ParsingFn(func(string) (*address.Address, error) {
		return &address.Address{PrimaryNumber: "1015", StreetName: "NORTH", StreetSuffix: "AVENUE"}, nil
	})
	opts := []goprojectusat.Option{goprojectusat.WithCustomAddressParser(stub), goprojectusat.WithContentNormalization()}
	u, err := goprojectusat.New(opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	reused := testing.AllocsPerRun(100, func() { _, _ = u.Normalize("1015 NORTH AVENUE") })
	perCall := testing.AllocsPerRun(100, func() { _, _ = goprojectusat.Normalize("1015 NORTH AVENUE", opts...) })
	t.Logf("allocs/op: USAt.Normalize %.0f, Normalize(opts...) %.0f", reused, perCall)
	if reused >= perCall {
		t.Errorf("expected USAt.Normalize (%.0f allocs/op) to allocate less than per-call construction (%.0f allocs/op)", reused, perCall)
	}
}

// TestUSAtParseMatchesPackageParse checks the struct and package-level Parse
// agree, with no options (the shared default) and with a verifier.
func TestUSAtParseMatchesPackageParse(t *testing.T) {
	upper := func(a *address.Address) (*address.Address, error) { return a, nil }
	u, err := goprojectusat.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	uv, err := goprojectusat.New(goprojectusat.WithParsedAddressVerifier(upper))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, c := range parsertest.Cases {
		want, wantErr := goprojectusat.Parse(c.Input)
		got, gotErr := u.Parse(c.Input)
		if fmt.Sprintf("%+v", got) != fmt.Sprintf("%+v", want) || errString(gotErr) != errString(wantErr) {
			t.Errorf("USAt.Parse(%q)\n got  %+v, %v\n want %+v, %v", c.Input, got, gotErr, want, wantErr)
		}
		wantV, wantVErr := goprojectusat.Parse(c.Input, parser.AddressParsingOptions{Verifier: upper})
		gotV, gotVErr := uv.Parse(c.Input)
		if fmt.Sprintf("%+v", gotV) != fmt.Sprintf("%+v", wantV) || errString(gotVErr) != errString(wantVErr) {
			t.Errorf("USAt.Parse(%q) with verifier\n got  %+v, %v\n want %+v, %v", c.Input, gotV, gotVErr, wantV, wantVErr)
		}
	}
}

// TestUSAtNormalizeAddressMatchesNormalize checks that NormalizeAddress
// (go-projectusat#206), applied after Parse, agrees with the combined
// Normalize on the same input and options — it is the same normalizer, only
// reachable a part at a time.
func TestUSAtNormalizeAddressMatchesNormalize(t *testing.T) {
	u, err := goprojectusat.New(goprojectusat.WithFuzzyNormalization())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, c := range parsertest.Cases {
		want, wantErr := u.Normalize(c.Input)

		parsed, parseErr := u.Parse(c.Input)
		if errString(parseErr) != errString(wantErr) && parseErr != nil {
			continue // Normalize's error came from normalization, not parsing; not this test's case.
		}
		if parseErr != nil {
			if errString(parseErr) != errString(wantErr) {
				t.Errorf("Parse(%q) err = %v, want %v", c.Input, parseErr, wantErr)
			}
			continue
		}

		normalized, normErr := u.NormalizeAddress(parsed)
		got := ""
		gotErr := normErr
		if normErr == nil {
			got, gotErr = normalized.Format()
		}
		if got != want || errString(gotErr) != errString(wantErr) {
			t.Errorf("Parse+NormalizeAddress(%q)\n got  %q, %v\n want %q, %v", c.Input, got, gotErr, want, wantErr)
		}
	}
}

// TestUSAtNormalizeGluedPunctuationDesignator is Aaron's go-projectusat#207
// repro: "Apt." keeps its period glued to the word (the tokenizer only drops
// a token that is punctuation alone), and that period must not stop the
// designator from being recognized. The Puerto Rico diacritics case from his
// comment is excluded here — that one is #205's open architectural question,
// not this bug.
func TestUSAtNormalizeGluedPunctuationDesignator(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{
			input:  "123 Main Street Apt 23\nNorfolk, Virginia 23510",
			expect: "123 MAIN ST # 23\nNORFOLK VA 23510",
		},
		{
			input:  "123 Main Street Apt. 23\nNorfolk, Virginia 23510",
			expect: "123 MAIN ST # 23\nNORFOLK VA 23510",
		},
	}

	usat, err := goprojectusat.New(
		goprojectusat.WithDiacriticNormalization(diacritics.SubstituteDiacritics),
		goprojectusat.WithSecondaryAsHash(),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, c := range cases {
		got, err := usat.Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) error: %v", c.input, err)
			continue
		}
		if got != c.expect {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.expect)
		}
	}
}

// TestUSAtConcurrentUse shares one USAt across goroutines; run with -race.
func TestUSAtConcurrentUse(t *testing.T) {
	u, err := goprojectusat.New(goprojectusat.WithMatchingNormalization())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	want := make([]string, len(parsertest.Cases))
	for i, c := range parsertest.Cases {
		want[i], _ = u.Normalize(c.Input)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for i, c := range parsertest.Cases {
				if got, _ := u.Normalize(c.Input); got != want[i] {
					t.Errorf("concurrent Normalize(%q) = %q, want %q", c.Input, got, want[i])
				}
			}
		})
	}
	wg.Wait()
}

func ExampleNew() {
	usat, err := goprojectusat.New(goprojectusat.WithMatchingNormalization(), goprojectusat.WithSingleLineFormatting())
	if err != nil {
		panic(err)
	}
	for _, in := range []string{
		"152 South Tech Dr Apartment 3200, Salt Lake City, UT 84101",
		"PO Box 11890, Phoenix, Arizona 85001",
	} {
		out, err := usat.Normalize(in)
		if err != nil {
			panic(err)
		}
		fmt.Println(out)
	}
	// Output:
	// 152 S TECH DR # 3200 SALT LAKE CITY UT 84101
	// PO BOX 11890 PHOENIX AZ 85001
}

func ExampleUSAt_Parse() {
	usat, err := goprojectusat.New()
	if err != nil {
		panic(err)
	}
	addr, err := usat.Parse("Acme Widgets, 152 South Tech Drive Suite 3200, Salt Lake City, Utah 84101")
	if err != nil {
		panic(err)
	}
	fmt.Println("BusinessName:", addr.BusinessName)
	fmt.Println("PrimaryNumber:", addr.PrimaryNumber)
	fmt.Println("Predirectional:", addr.Predirectional)
	fmt.Println("StreetName:", addr.StreetName)
	fmt.Println("StreetSuffix:", addr.StreetSuffix)
	fmt.Println("SecondaryDesignator:", addr.SecondaryDesignator)
	fmt.Println("SecondaryNumber:", addr.SecondaryNumber)
	fmt.Println("City:", addr.City)
	fmt.Println("Region:", addr.Region)
	fmt.Println("Postal:", addr.Postal)

	// Output:
	// BusinessName: ACME WIDGETS
	// PrimaryNumber: 152
	// Predirectional: S
	// StreetName: TECH
	// StreetSuffix: DR
	// SecondaryDesignator: STE
	// SecondaryNumber: 3200
	// City: SALT LAKE CITY
	// Region: UT
	// Postal: 84101
}

// Parse and NormalizeAddress together are Normalize without the formatting,
// leaving the address in fields. Here WithMatchingNormalization is applied by
// NormalizeAddress, not by Parse, and replaces the designator with #.
func ExampleUSAt_NormalizeAddress() {
	usat, err := goprojectusat.New(goprojectusat.WithMatchingNormalization())
	if err != nil {
		panic(err)
	}
	parsed, err := usat.Parse("152 South Tech Drive Suite 3200, Salt Lake City, UT 84101")
	if err != nil {
		panic(err)
	}
	normalized, err := usat.NormalizeAddress(parsed)
	if err != nil {
		panic(err)
	}
	fmt.Println("parsed:", parsed.SecondaryDesignator, parsed.SecondaryNumber)
	fmt.Println("normalized:", normalized.SecondaryDesignator, normalized.SecondaryNumber)

	// Output:
	// parsed: STE 3200
	// normalized: # 3200
}

// An address need not come from Parse. One that holds only some of its
// fields normalizes just those.
func ExampleUSAt_NormalizeAddress_fields() {
	usat, err := goprojectusat.New()
	if err != nil {
		panic(err)
	}
	normalized, err := usat.NormalizeAddress(&address.Address{City: "salt lake city", Region: "utah"})
	if err != nil {
		panic(err)
	}
	fmt.Println(normalized.City)
	fmt.Println(normalized.Region)

	// Output:
	// SALT LAKE CITY
	// UT
}

func benchmarkCases(b *testing.B, normalize func(string) (string, error)) {
	b.ReportAllocs()
	for b.Loop() {
		for _, c := range parsertest.Cases {
			_, _ = normalize(c.Input)
		}
	}
}

// BenchmarkNormalize compares the package-level Normalize with options, which
// builds a USAt per call, against one USAt reused across calls.
func BenchmarkNormalize(b *testing.B) {
	opts := []goprojectusat.Option{goprojectusat.WithMatchingNormalization()}
	b.Run("package-with-options", func(b *testing.B) {
		benchmarkCases(b, func(s string) (string, error) { return goprojectusat.Normalize(s, opts...) })
	})
	b.Run("package-default", func(b *testing.B) {
		benchmarkCases(b, func(s string) (string, error) { return goprojectusat.Normalize(s) })
	})
	b.Run("usat-reused", func(b *testing.B) {
		u, err := goprojectusat.New(opts...)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkCases(b, u.Normalize)
	})
}
