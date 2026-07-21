package goprojectusat

import (
	"fmt"
	"strings"

	"github.com/PortobelloAuth/go-projectusat/pkg/military"
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

// parseCivilian is implemented in Task 3–4; for Task 2 stub:
func parseCivilian(lines []string) (Address, error) {
	return Address{}, fmt.Errorf("non-military parse not implemented")
}
