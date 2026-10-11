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
	return addr.Format(*fopts), nil
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
