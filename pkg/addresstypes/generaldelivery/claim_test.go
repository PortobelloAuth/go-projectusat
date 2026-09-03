package generaldelivery_test

import (
	"testing"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/generaldelivery"
)

func claims(source string) []claim.Claim {
	return generaldelivery.Claims(token.Tokenize(source))
}

func TestTheSpelledOutPhraseIsOneExactClaim(t *testing.T) {
	got := claims("GENERAL DELIVERY\nSPRINGFIELD IL 62701-9999")

	if len(got) != 1 {
		t.Fatalf("claims = %d, want 1", len(got))
	}

	c := got[0]
	if c.Confidence != claim.ConfidenceExact {
		t.Errorf("Confidence = %d, want %d", c.Confidence, claim.ConfidenceExact)
	}

	if len(c.Parts) != 1 {
		t.Fatalf("Parts = %d, want 1", len(c.Parts))
	}

	p := c.Parts[0]
	if p.Part != claim.PartStreetName || p.Value != "GENERAL DELIVERY" {
		t.Errorf("part = %v %q, want street name %q", p.Part, p.Value, "GENERAL DELIVERY")
	}

	if p.Start != 0 || p.Length != 2 {
		t.Errorf("span = [%d,%d), want [0,2)", p.Start, p.Start+p.Length)
	}
}

func TestAnAbbreviationIsClaimedALittleLessConfidently(t *testing.T) {
	// The lookup is exact either way. GEN DEL is rated below the spelled out
	// form because DEL is an ordinary word inside a longer name, so the parser
	// is given something to prefer with.
	got := claims("GEN DEL\nSPRINGFIELD IL 62701")

	if len(got) != 1 {
		t.Fatalf("claims = %d, want 1", len(got))
	}

	if got[0].Confidence != claim.ConfidenceStrong {
		t.Errorf("Confidence = %d, want %d", got[0].Confidence, claim.ConfidenceStrong)
	}

	if got[0].Parts[0].Value != "GENERAL DELIVERY" {
		t.Errorf("value = %q, want %q", got[0].Parts[0].Value, "GENERAL DELIVERY")
	}
}

func TestNeitherWordIsClaimedOnItsOwn(t *testing.T) {
	// A street named for a general is a street, and DELIVERY alone says nothing.
	// It is the pair that carries the meaning.
	for _, source := range []string{
		"400 GENERAL BLVD\nSPRINGFIELD IL 62701",
		"DELIVERY RD\nSPRINGFIELD IL 62701",
		"GENERAL\nSPRINGFIELD IL 62701",
	} {
		t.Run(source, func(t *testing.T) {
			if got := claims(source); len(got) != 0 {
				t.Errorf("claims = %d, want none", len(got))
			}
		})
	}
}

func TestThePhraseIsNotClaimedAcrossALineBreak(t *testing.T) {
	// A delivery address line is a line, and token.Join flattens the break, so
	// without the bound "GENERAL\nDELIVERY RD" reads as a general delivery line
	// with a stray suffix after it.
	if got := claims("GENERAL\nDELIVERY\nSPRINGFIELD IL 62701"); len(got) != 0 {
		t.Errorf("claims = %d, want none", len(got))
	}
}

func TestALongerNameLeavesItsExtraTokensUnclaimed(t *testing.T) {
	// The claim covers the phrase and no more. What is left over is what tells
	// a candidate it did not account for the whole address, which is the
	// reading the parser should weigh rather than one this package suppresses.
	got := claims("GENERAL DELIVERY LN\nSPRINGFIELD IL 62701")

	if len(got) != 1 {
		t.Fatalf("claims = %d, want 1", len(got))
	}

	if p := got[0].Parts[0]; p.Start != 0 || p.Length != 2 {
		t.Errorf("span = [%d,%d), want [0,2)", p.Start, p.Start+p.Length)
	}
}
