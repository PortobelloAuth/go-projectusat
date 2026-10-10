package zipcityembedded

import "testing"

func TestOnlyASingleParserIsInstantiated(t *testing.T) {
	p1, err := New()
	if err != nil {
		t.Fatalf("Failed to instatiate parser 1: %s", err)
	}

	p2, err := New()
	if err != nil {
		t.Fatalf("Failed to instatiate parser 2: %s", err)
	}

	if p1.parser != p2.parser {
		t.Errorf("Expected both parsers to wrap the same parser:\n %v\n %v", p1.parser, p2.parser)
	}
}
