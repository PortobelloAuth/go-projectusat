package goprojectusat

import (
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// rewriteSpecialStreetLine rewrites rural-route and PO Box free-text street
// lines to Project US@ / USPS content form.
//
// Rural route: RURAL ROUTE / RFD / RD (with box) / RR / RUTA → "RR N BOX X"
// (leading zeros stripped from the route; trailing tokens after the box dropped).
// Bare "RD 5A" (highway-style, no BOX/#) does not match.
//
// PO Box: POST OFFICE BOX / PO BOX → "PO BOX X".
//
// On match returns the rewritten uppercase collapsed line and true.
func rewriteSpecialStreetLine(line string) (string, bool) {
	cleaned := strings.ToUpper(textutil.CollapseSpace(
		textutil.StripPunctuation(line, textutil.StripOptions{KeepHyphen: true}),
	))
	if cleaned == "" {
		return "", false
	}

	if s, ok := rewritePOBox(cleaned); ok {
		return s, true
	}
	if s, ok := rewriteRuralRoute(cleaned); ok {
		return s, true
	}
	return "", false
}

// rewritePOBox matches POST OFFICE BOX / PO BOX / POBOX + box id.
func rewritePOBox(cleaned string) (string, bool) {
	tokens := strings.Fields(cleaned)
	if len(tokens) == 0 {
		return "", false
	}

	// POBOX123 glued form
	if strings.HasPrefix(tokens[0], "POBOX") && tokens[0] != "POBOX" {
		boxID := tokens[0][len("POBOX"):]
		return "PO BOX " + boxID, true
	}

	i := 0
	switch {
	case len(tokens) >= 3 && tokens[0] == "POST" && tokens[1] == "OFFICE" && tokens[2] == "BOX":
		i = 3
	case len(tokens) >= 2 && tokens[0] == "PO" && tokens[1] == "BOX":
		i = 2
	case len(tokens) >= 1 && tokens[0] == "POBOX":
		i = 1
	default:
		return "", false
	}

	if i >= len(tokens) {
		return "", false
	}
	// Box identifier is the next token; drop any trailing junk.
	return "PO BOX " + tokens[i], true
}

// rewriteRuralRoute matches rural-route prefixes with a numeric route and a
// box marker (BOX or #) plus box id. Output: "RR {route} BOX {id}".
func rewriteRuralRoute(cleaned string) (string, bool) {
	tokens := expandHashTokens(strings.Fields(cleaned))
	tokens = expandGluedRRTokens(tokens)
	if len(tokens) == 0 {
		return "", false
	}

	i, ok := matchRuralPrefix(tokens)
	if !ok {
		return "", false
	}

	if i >= len(tokens) || !isAllDigits(tokens[i]) {
		return "", false
	}
	route := stripLeadingZeros(tokens[i])
	if route == "" {
		route = "0"
	}
	i++

	// Require BOX or # — this keeps bare "RD 5A" / "RD 61" out of RR.
	if i >= len(tokens) {
		return "", false
	}
	marker := tokens[i]
	if marker != "BOX" && marker != "#" {
		return "", false
	}
	i++

	if i >= len(tokens) {
		return "", false
	}
	boxID := tokens[i]
	// Trailing town/street after the box is intentionally dropped.
	return "RR " + route + " BOX " + boxID, true
}

// matchRuralPrefix consumes a rural-route designator and returns the next index.
// Accepts: RURAL ROUTE, RUTA RURAL, RFD, RR, RD, RUTA.
func matchRuralPrefix(tokens []string) (int, bool) {
	if len(tokens) >= 2 && tokens[0] == "RURAL" && tokens[1] == "ROUTE" {
		return 2, true
	}
	if len(tokens) >= 2 && tokens[0] == "RUTA" && tokens[1] == "RURAL" {
		return 2, true
	}
	if len(tokens) >= 1 {
		switch tokens[0] {
		case "RFD", "RR", "RD", "RUTA":
			return 1, true
		}
	}
	return 0, false
}

// expandGluedRRTokens splits tokens like "RR0061" / "RFD61" / "RD12" into
// designator + digits so "RR0061 #87b" parses as RR / 0061 / # / 87B.
func expandGluedRRTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens)+1)
	for _, t := range tokens {
		if prefix, num, ok := splitGluedRR(t); ok {
			out = append(out, prefix, num)
			continue
		}
		out = append(out, t)
	}
	return out
}

func splitGluedRR(t string) (prefix, num string, ok bool) {
	for _, p := range []string{"RURALROUTE", "RUTARURAL", "RFD", "RR", "RD", "RUTA"} {
		if strings.HasPrefix(t, p) && len(t) > len(p) {
			rest := t[len(p):]
			if isAllDigits(rest) {
				// Map multi-word glued forms back to spaced designators via
				// matchRuralPrefix-compatible single tokens; RURALROUTE is
				// handled specially.
				switch p {
				case "RURALROUTE":
					// Expand to RURAL ROUTE <num> as three logical pieces:
					// caller only gets two strings; re-inject as RURAL, ROUTE handled below.
					// Use RR as equivalent single designator for glued form.
					return "RR", rest, true
				case "RUTARURAL":
					return "RR", rest, true
				default:
					return p, rest, true
				}
			}
		}
	}
	return "", "", false
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func stripLeadingZeros(s string) string {
	i := 0
	for i < len(s) && s[i] == '0' {
		i++
	}
	return s[i:]
}
