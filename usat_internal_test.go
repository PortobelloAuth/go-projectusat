package goprojectusat

import "testing"

// TestDefaultUSAtIsBuiltOnce mirrors zipcityembedded's
// TestOnlyASingleParserIsInstantiated one level up: the package-level
// functions reuse one USAt, and with it one parser and normalizer.
func TestDefaultUSAtIsBuiltOnce(t *testing.T) {
	u1, err := defaultUSAt()
	if err != nil {
		t.Fatalf("defaultUSAt: %v", err)
	}
	const addr = "1015 NORTH AVENUE, SALT LAKE CITY, UT 84101"
	if _, err := Normalize(addr); err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if _, err := Parse(addr); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	u2, err := defaultUSAt()
	if err != nil {
		t.Fatalf("defaultUSAt: %v", err)
	}
	if u1 != u2 || u1.parser != u2.parser || u1.normalizer != u2.normalizer {
		t.Errorf("expected the default USAt to be built once:\n %p\n %p", u1, u2)
	}
}
