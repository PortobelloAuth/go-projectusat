package goprojectusat

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/PortobelloAuth/go-projectusat/pkg/military"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/textutil"
)

// Parse converts free-text multi-line or comma-separated address text into a
// structured Address. It does not call Normalize; compose Normalize(Parse(raw))
// for content form. Empty input returns an error.
func Parse(raw string) (Address, error) {
	lines := splitAddressLines(raw)
	if len(lines) == 0 {
		return Address{}, fmt.Errorf("empty address")
	}

	// Military fast path: last line military last-line AND some earlier line military street.
	if len(lines) >= 2 {
		if city, reg, zip, err := military.NormalizeLastLine(lines[len(lines)-1]); err == nil {
			// Prefer the line immediately before last as street; else scan.
			for i := len(lines) - 2; i >= 0; i-- {
				if street, err := military.NormalizeStreetLine(lines[i]); err == nil {
					var business string
					if i > 0 {
						business = strings.Join(lines[:i], " ")
					}
					return Address{
						BusinessName: business,
						StreetName:   street,
						City:         city,
						Region:       reg,
						Postal:       zip,
					}, nil
				}
			}
		}
	}

	// Single-line military "STREET, LAST" already split by splitAddressLines into 2 lines via comma.
	// Fall through to Task 3/4 for non-military.
	return parseCivilian(lines)
}

// splitAddressLines splits on newlines, then on commas within lines, collapses
// space per segment, drops empties. Uppercases lightly via CollapseSpace only
// (leave casing to Normalize) — actually for military helpers we upper inside
// military package; keep segments trimmed collapsed.
func splitAddressLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		// Comma-separated segments within a line become separate logical lines.
		for _, part := range strings.Split(line, ",") {
			part = textutil.CollapseSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func parseCivilian(lines []string) (Address, error) {
	if len(lines) == 1 {
		// Single segment: peel trailing last-line tokens; remainder is street.
		return parseSingleLineCivilian(lines[0])
	}
	city, reg, zip, err := parseLastLine(lines[len(lines)-1])
	if err != nil {
		return Address{}, err
	}
	streetLine := lines[len(lines)-2]
	business := ""
	if len(lines) > 2 {
		business = strings.ToUpper(textutil.CollapseSpace(strings.Join(lines[:len(lines)-2], " ")))
	}
	return Address{
		BusinessName: business,
		StreetName:   strings.ToUpper(textutil.CollapseSpace(streetLine)),
		City:         city,
		Region:       reg,
		Postal:       zip,
	}, nil
}

// parseLastLine extracts city, region (abbreviated), and postal from a last line
// like "SPRINGFIELD IL 62701" or "OTTAWA ON K1A 0B1". City is multi-word capable.
func parseLastLine(line string) (city, reg, postal string, err error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	if len(tokens) < 2 {
		return "", "", "", fmt.Errorf("invalid last line: %q", line)
	}
	postal, rest, ok := peelPostal(tokens)
	if !ok {
		return "", "", "", fmt.Errorf("invalid last line postal: %q", line)
	}
	if len(rest) < 1 {
		return "", "", "", fmt.Errorf("invalid last line: missing region in %q", line)
	}
	// Longest region match from the right of rest (e.g. DISTRICT OF COLUMBIA).
	for n := len(rest); n >= 1; n-- {
		cand := strings.Join(rest[len(rest)-n:], " ")
		if abbr, e := region.NormalizeRegion(cand); e == nil {
			city = strings.Join(rest[:len(rest)-n], " ")
			if city == "" {
				return "", "", "", fmt.Errorf("invalid last line: missing city in %q", line)
			}
			return city, abbr, postal, nil
		}
	}
	return "", "", "", fmt.Errorf("unrecognized region in last line: %q", line)
}

// peelPostal removes US ZIP / ZIP+4 or Canadian postal from the right of tokens.
func peelPostal(tokens []string) (postal string, rest []string, ok bool) {
	if len(tokens) == 0 {
		return "", nil, false
	}
	last := tokens[len(tokens)-1]

	// ##### or #####-#### (and compact #########)
	if usZIPCompact.MatchString(last) {
		return normalizePostal(last), tokens[:len(tokens)-1], true
	}

	if len(tokens) >= 2 {
		prev := tokens[len(tokens)-2]
		// Two-token US ZIP+4: 62701 1234
		if usZIPCompact.MatchString(prev + last) || usZIPCompact.MatchString(prev+"-"+last) {
			return normalizePostal(prev + "-" + last), tokens[:len(tokens)-2], true
		}
		// Canadian two-token: K1A 0B1
		two := prev + " " + last
		np := normalizePostal(two)
		if len(np) >= 6 && containsLetter(np) {
			return np, tokens[:len(tokens)-2], true
		}
	}

	// Single alphanumeric postal fallback
	return normalizePostal(last), tokens[:len(tokens)-1], true
}

func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// parseSingleLineCivilian peels postal + region + single-token city from the
// right; remainder becomes StreetName. Multi-word cities require multi-line form.
func parseSingleLineCivilian(line string) (Address, error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	postal, rest, ok := peelPostal(tokens)
	if !ok || len(rest) < 2 {
		return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
	}
	// Region up to 3 tokens e.g. DISTRICT OF COLUMBIA
	for n := min(3, len(rest)); n >= 1; n-- {
		cand := strings.Join(rest[len(rest)-n:], " ")
		if abbr, e := region.NormalizeRegion(cand); e == nil {
			before := rest[:len(rest)-n]
			if len(before) < 2 {
				return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
			}
			// Conservative: last token of before = city; rest = street.
			city := before[len(before)-1]
			streetToks := before[:len(before)-1]
			return Address{
				StreetName: strings.Join(streetToks, " "),
				City:       city,
				Region:     abbr,
				Postal:     postal,
			}, nil
		}
	}
	return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
}
