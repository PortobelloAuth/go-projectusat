package puertorico

import (
	"fmt"
	"maps"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/claim"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/token"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
	"github.com/poetic-systems/addresstables/directionals"
)

// spanishDirectionMap and spanishDirectionShortMap hold only the eight
// Spanish directionals (go-projectusat#154): NORTE, SUR, ESTE, OESTE,
// NORESTE, SUDESTE, NOROESTE, SUDOESTE. Each is keyed on its own Short
// abbreviation, never on the English row addresstables/directionals points
// it at — p.25 is explicit that "developers MUST NOT translate
// directionals", so NOROESTE must abbreviate to NO, not to NW.
//
// pkg/directionals is the shared English-only vocabulary, and its own doc
// comment defers exactly this decision to Puerto Rico addresses rather than
// making it: accepting a Spanish spelling belongs to this package, per
// Aaron's ruling on #71 that a Puerto-Rico-specific vocabulary does not
// become a change to the general one.
var spanishDirectionMap = maps.Collect(func(yield func(string, string) bool) {
	for d := range directionals.Spanish() {
		if !yield(d.Full, d.Short) {
			return
		}
	}
})

var spanishDirectionShortMap = maps.Collect(func(yield func(string, string) bool) {
	for k, v := range spanishDirectionMap {
		if !yield(v, k) {
			return
		}
	}
})

// abbreviateSpanishDirectional reduces a Spanish directional word or
// abbreviation to its Pub 28 abbreviation. Unlike
// pkg/directionals.AbbreviateDirectional, the value returned is always the
// Spanish abbreviation (NOROESTE -> NO), never the English one the Spanish
// row's English field names.
func abbreviateSpanishDirectional(d string) (string, error) {
	capitalized := strings.ToUpper(d)

	if abbrev, ok := spanishDirectionMap[capitalized]; ok {
		return abbrev, nil
	}

	if _, ok := spanishDirectionShortMap[capitalized]; ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("unrecognized Spanish directional")
}

// directionalMaxSpan is the longest Spanish directional in the vocabulary,
// measured in tokens: a compound spelled as two words, e.g. NORTE ESTE.
const directionalMaxSpan = 2

// directionalClaims returns every Spanish-directional reading of tokens.
//
// This mirrors pkg/directionals.Claims's structural shape exactly — both
// PartPredirectional and PartPostdirectional are claimed for every match,
// a single token and a two-token compound (recognized the same way, by
// concatenating each token's abbreviation and checking the vocabulary
// recognizes the result) are both offered, and confidence is graded the
// same way: exact for a bare abbreviation, strong for one spelled-out word
// or an abbreviation compound, likely for a two-word spelled-out compound.
// What differs is only the data: this reads solely from
// addresstables/directionals.Spanish(), never English() or All(), so this
// package's Spanish recognition cannot regress into a general vocabulary
// change (#71's precedent about Puerto Rico addresses).
func directionalClaims(tokens []token.Token) []claim.Claim {
	var claims []claim.Claim

	for i := range tokens {
		// A compound hard wrapped across two lines is not one compound. See
		// token.LineEnd.
		span := min(directionalMaxSpan, token.LineEnd(tokens, i)-i)
		for length := span; length >= 1; length-- {
			abbreviation, ok := abbreviateSpanishSpan(tokens[i : i+length])
			if !ok {
				continue
			}

			confidence := spanishSpanConfidence(tokens[i : i+length])
			for _, part := range []claim.Part{claim.PartPredirectional, claim.PartPostdirectional} {
				claims = append(claims, claim.Claim{
					Confidence: confidence,
					Parts: []claim.ClaimPart{{
						Start:  i,
						Length: length,
						Part:   part,
						Value:  abbreviation,
					}},
				})
			}
		}
	}

	return claims
}

// abbreviateSpanishSpan reduces a run of tokens to a single Spanish
// directional abbreviation, reporting whether the vocabulary recognizes it.
// See pkg/directionals.abbreviateSpan, which this mirrors against Spanish
// data only.
func abbreviateSpanishSpan(tokens []token.Token) (string, bool) {
	var combined strings.Builder
	for _, t := range tokens {
		// There should not be any punctuation in directionals.
		clean := textutil.StripPunctuation(t.Text, textutil.StripOptions{
			KeepHyphen: false,
			KeepSlash:  false,
		})

		abbreviation, err := abbreviateSpanishDirectional(clean)
		if err != nil {
			return "", false
		}

		combined.WriteString(abbreviation)
	}

	if len(tokens) == 1 {
		return combined.String(), true
	}

	if _, ok := spanishDirectionShortMap[combined.String()]; !ok {
		return "", false
	}

	return combined.String(), true
}

// substituteSpanishDirectionals returns tokens[start:end] with any
// Spanish-directional claim entirely inside that range replaced by a single
// token holding the claimed abbreviation, so a recognizer downstream (e.g.
// NormalizeStreetLine) sees "NO" wherever the line spelled out NOROESTE.
//
// Only PartPredirectional/PartPostdirectional claims are considered — the
// same two parts directionalClaims always claims together — and only those
// whose span falls entirely within [start, end); one reaching outside it is
// not a reading of this line. Where more than one length matches at the same
// starting token (a single word, and, where it also opens a compound, two),
// the longest wins, the same preference pkg/directionals.Claims documents
// for a compound over the two directionals it is made of.
func substituteSpanishDirectionals(tokens []token.Token, claims []claim.Claim, start, end int) []token.Token {
	best := map[int]claim.ClaimPart{}
	for _, c := range claims {
		for _, p := range c.Parts {
			if p.Part != claim.PartPredirectional && p.Part != claim.PartPostdirectional {
				continue
			}
			if p.Start < start || p.End() > end {
				continue
			}
			if cur, ok := best[p.Start]; !ok || p.Length > cur.Length {
				best[p.Start] = p
			}
		}
	}

	out := make([]token.Token, 0, end-start)
	for i := start; i < end; {
		if p, ok := best[i]; ok {
			substituted := tokens[i]
			substituted.Text = p.Value
			out = append(out, substituted)
			i += p.Length
			continue
		}

		out = append(out, tokens[i])
		i++
	}

	return out
}

// spanishSpanConfidence rates a matched run of tokens. See
// pkg/directionals.spanConfidence, which this mirrors against Spanish data
// only.
func spanishSpanConfidence(tokens []token.Token) claim.Confidence {
	spelledOut := false
	for _, t := range tokens {
		clean := textutil.StripPunctuation(t.Text, textutil.StripOptions{
			KeepHyphen: false,
			KeepSlash:  false,
		})
		if _, isFullWord := spanishDirectionMap[strings.ToUpper(clean)]; isFullWord {
			spelledOut = true
		}
	}

	switch {
	case len(tokens) > 1 && spelledOut:
		return claim.ConfidenceLikely
	case len(tokens) > 1:
		return claim.ConfidenceStrong
	case spelledOut:
		return claim.ConfidenceStrong
	}

	return claim.ConfidenceExact
}
