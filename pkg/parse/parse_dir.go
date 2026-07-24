package parse

import "strings"

// expandHyphenatedDirectionals rewrites compound directionals joined by a
// hyphen into their standard compound abbreviation:
//
//	NORTH-EAST / N-E / NORTHEAST → NE (token already compound passes through
//	via AbbreviateDirectional later; hyphenated forms become NE here)
//
// Only N/S + E/W compounds are rewritten. Other hyphenated tokens are left
// for expandHyphenatedStreetTokens.
func expandHyphenatedDirectionals(tokens []string) []string {
	if len(tokens) == 0 {
		return tokens
	}
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if merged, ok := hyphenatedDirectional(t); ok {
			out = append(out, merged)
			continue
		}
		out = append(out, t)
	}
	return out
}

// hyphenatedDirectional returns the compound abbreviation for tokens like
// "NORTH-EAST", "N-W", "SOUTH-EAST". Returns ("", false) when not a compound.
func hyphenatedDirectional(tok string) (string, bool) {
	i := strings.IndexByte(tok, '-')
	if i <= 0 || i >= len(tok)-1 {
		return "", false
	}
	// Only a single hyphen joining two parts.
	if strings.Count(tok, "-") != 1 {
		return "", false
	}
	a, b := tok[:i], tok[i+1:]
	return mergeDirectionalPair(a, b)
}

// mergeDirectionTokens scans tokens left-to-right and merges adjacent cardinal
// pairs that form a compound directional:
//
//	(N|NORTH)+(E|EAST) → NE, (N|NORTH)+(W|WEST) → NW,
//	(S|SOUTH)+(E|EAST) → SE, (S|SOUTH)+(W|WEST) → SW
//
// Opposite pairs (NORTH SOUTH, EAST WEST, etc.) are left unmerged so they can
// remain street-name material. Already-compound tokens (NE, NORTHEAST, …) are
// single tokens and pass through unchanged.
func mergeDirectionTokens(tokens []string) []string {
	if len(tokens) < 2 {
		return tokens
	}
	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if i+1 < len(tokens) {
			if merged, ok := mergeDirectionalPair(tokens[i], tokens[i+1]); ok {
				out = append(out, merged)
				i++ // consume second token of the pair
				continue
			}
		}
		out = append(out, tokens[i])
	}
	return out
}

// mergeDirectionalPair returns the compound abbreviation when a and b form a
// valid N/S then E/W pair; otherwise ("", false).
func mergeDirectionalPair(a, b string) (string, bool) {
	ns, okNS := cardinalNS(a)
	if !okNS {
		return "", false
	}
	ew, okEW := cardinalEW(b)
	if !okEW {
		return "", false
	}
	return ns + ew, true
}

func cardinalNS(tok string) (string, bool) {
	switch strings.ToUpper(tok) {
	case "N", "NORTH":
		return "N", true
	case "S", "SOUTH":
		return "S", true
	default:
		return "", false
	}
}

func cardinalEW(tok string) (string, bool) {
	switch strings.ToUpper(tok) {
	case "E", "EAST":
		return "E", true
	case "W", "WEST":
		return "W", true
	default:
		return "", false
	}
}
