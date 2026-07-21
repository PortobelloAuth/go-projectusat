# Free-Text Parse + Military Root Wiring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add conservative free-text `Parse` for multi-line / comma-separated patient addresses, and wire overseas military (APO/FPO/DPO) street + last lines into the root package so `Parse` and `Normalize`/`Format` work for military addresses.

**Architecture:** Keep component packages pure. Add `parse.go` in the root package with line-splitting, last-line extraction, reverse-token street-line parse, and a military fast path that calls `pkg/military`. `Normalize` remains field-based; when street/city/region look military, it normalizes via `military` helpers so content form is correct. Do not attempt full C# `StreetLineParser` parity (no business-prefix reordering, no RR/PO Box, no dual-secondary reordering) in this plan.

**Tech Stack:** Go 1.25+, module `github.com/PortobelloAuth/go-projectusat`, existing `pkg/*` packages. No new deps. Apache-2.0. Table-driven tests matching root/package style.

## Global Constraints

- Module path remains `github.com/PortobelloAuth/go-projectusat`.
- API is unstable; export small documented surfaces only.
- Alphabetical content is uppercase in normalized content output.
- Unknown fields: blank string; token `UNKNOWN` → blank.
- No CGO, libpostal, network, or external validation.
- Every task ends with `go test ./...` passing and `go vet ./...` clean.
- Prefer exact table lookup; fuzzy only via existing options on Normalize, not Parse.
- Parse is **conservative**: prefer error or leaving free-text in StreetName over inventing structure when ambiguous.
- Do not invent geocoder-backed disambiguation for numeric street names or dual directionals.
- Military: overseas only (APO/FPO/DPO + AE/AP/AA); domestic military street addresses use the normal path.
- PR body style for upstream: What was addressed / How to test / Configurations or dependencies.
- Commit after each task with a focused message.

## File Structure

| Path | Responsibility |
|------|----------------|
| `parse.go` | `Parse`, line split, last-line parse, street-line parse, military fast path |
| `parse_test.go` | Parse + military root integration tests |
| `goprojectusat.go` | Wire military into `Normalize` free-text street / city+region when applicable; import military |
| `goprojectusat_test.go` | Add Normalize military cases (or put those in parse_test if preferred — prefer `goprojectusat_test.go` for Normalize, `parse_test.go` for Parse) |
| `README.md` | Document Parse + military support; update “Not yet” |
| `pkg/military/*` | Unchanged API unless a tiny export is required (prefer root-only changes) |

## Scope (in)

1. `Parse(raw string) (Address, error)` — multi-line (`\n`) and single-line with commas.
2. Military fast path when street line matches military street pattern and last line matches military last line.
3. Standard last line: `CITY REGION POSTAL` (region via `region.NormalizeRegion`; postal via existing `normalizePostal`).
4. Standard street line reverse-token parse for common shapes:
   - primary number + optional predirectional + street name + suffix + optional postdirectional + secondary designator + optional secondary number
   - `#` secondary form (`# 3200` / `#3200`)
   - highway-style street names via `highways.NormalizeStreetName` on the street-name token span
5. `Normalize` military: if `StreetName` is a military street line, normalize with `military.NormalizeStreetLine` and leave PrimaryNumber/suffix/dir empty; if City is APO/FPO/DPO and Region is AE/AP/AA, leave as content-form (upper/postal already handled).
6. `Format` continues to work for military when StreetName holds full military street line and City/Region/Postal hold last-line parts.

## Scope (out)

- Rural route / PO Box free-text rewrite
- Secondary unit reordering when designator appears before primary number
- Business-name detection beyond “first extra line before street”
- Puerto Rico Spanish street types in Parse
- Fuzzy parse
- Full C# StreetLineNormalizerTests parity

## Public API (final)

```go
// Parse converts free-text multi-line or comma-separated address text into a
// structured Address (not yet Normalize'd). Call Normalize on the result for
// content form. Empty input returns an error.
func Parse(raw string) (Address, error)
```

Parse does **not** call Normalize; callers compose `Normalize(Parse(raw))` when they want content form. Integration tests may check both.

---

### Task 1: Military helpers on root Normalize + Format path

**Files:**
- Modify: `goprojectusat.go`
- Modify: `goprojectusat_test.go`

**Interfaces:**
- Consumes: `military.NormalizeStreetLine`, `military.NormalizeLastLine`, existing `Normalize`
- Produces: `Normalize` accepts Address where `StreetName` is e.g. `psc 3 box 4120` and City/Region/Postal are military last-line parts; output is content-form military address; `Format` yields two lines.

- [ ] **Step 1: Write failing tests** in `goprojectusat_test.go`

```go
func TestNormalizeMilitaryOverseas(t *testing.T) {
	in := Address{
		StreetName: "psc 3 box 4120",
		City:       "apo",
		Region:     "ae",
		Postal:     "09021-0002",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	// Military street line lives in StreetName; no primary/suffix/dir.
	if got.StreetName != "PSC 3 BOX 4120" {
		t.Errorf("StreetName = %q, want PSC 3 BOX 4120", got.StreetName)
	}
	if got.City != "APO" || got.Region != "AE" || got.Postal != "09021-0002" {
		t.Errorf("last line = %q %q %q", got.City, got.Region, got.Postal)
	}
	if got.PrimaryNumber != "" || got.StreetSuffix != "" {
		t.Errorf("expected empty primary/suffix for military street, got primary=%q suffix=%q",
			got.PrimaryNumber, got.StreetSuffix)
	}
	if Format(got) != "PSC 3 BOX 4120\nAPO AE 09021-0002" {
		t.Errorf("Format = %q", Format(got))
	}
}

func TestNormalizeMilitaryStreetOnlyWhenLooksMilitary(t *testing.T) {
	// Ordinary street must not be forced through military.
	in := Address{
		PrimaryNumber: "123",
		StreetName:    "Main",
		StreetSuffix:  "Street",
		City:          "Springfield",
		Region:        "IL",
		Postal:        "62701",
	}
	got, err := Normalize(in)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if got.StreetName != "MAIN" || got.StreetSuffix != "ST" {
		t.Fatalf("got StreetName=%q Suffix=%q", got.StreetName, got.StreetSuffix)
	}
}
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test . -run 'TestNormalizeMilitary' -v
```

Expected: FAIL (military path missing or StreetName stays free-text without military normalize).

- [ ] **Step 3: Implement in `goprojectusat.go`**

Add import for `pkg/military`.

In `NormalizeWithOptions`, after `freeTextField` for street name (or instead of plain highway path):

```go
// After base free-text for street name tokens:
// If entire street line (PrimaryNumber empty and StreetName set, or
// FormatStreetLine of input pre-normalize) matches military, normalize via military.

// Practical rule for this task:
// 1. Build a candidate street line:
//    candidate := joinNonEmpty(" ", a.PrimaryNumber, a.Predirectional, a.StreetName,
//        a.StreetSuffix, a.Postdirectional, a.SecondaryDesignator, a.SecondaryNumber)
//    If PrimaryNumber and StreetSuffix and directionals and secondary are all empty,
//    candidate is just StreetName after freeTextField.
// 2. If military.NormalizeStreetLine(candidate) succeeds, set:
//    out.PrimaryNumber, Predirectional, StreetSuffix, Postdirectional,
//    SecondaryDesignator, SecondaryNumber = ""
//    out.StreetName = normalized military line
//    Skip highway path for that case.
// 3. For City: if strings.EqualFold after upper is APO/FPO/DPO, keep upper form
//    (freeTextField already uppercases). Optionally validate with a small check:
//    if city is APO|FPO|DPO and region normalizes to AE|AP|AA, accept.
// 4. Region still goes through region.NormalizeRegion (AE/AP/AA already in map).
```

Concrete implementation sketch (adapt to existing control flow):

```go
// Inside NormalizeWithOptions, replace pure highways branch with:

streetCandidate := joinNonEmpty(" ",
	baseField(a.PrimaryNumber),
	baseField(a.Predirectional),
	baseField(a.StreetName),
	baseField(a.StreetSuffix),
	baseField(a.Postdirectional),
	baseField(a.SecondaryDesignator),
	baseField(a.SecondaryNumber),
)
if mil, err := military.NormalizeStreetLine(streetCandidate); err == nil {
	out.StreetName = mil
	// leave other street fields blank on out (already zero)
	// still process city/region/postal/business as usual
} else {
	// existing free-text + highway + directional + suffix + secondary path
}
```

Important: when military path hits, **do not** also try to normalize empty directionals/suffix as errors. Structure the function so military early-success returns after filling city/region/postal/business, or sets a flag `militaryStreet bool` that skips controlled-vocab street fields.

Also: city free-text for APO is fine via freeTextField → `APO`.

- [ ] **Step 4: Run tests**

```bash
go test ./...
go vet ./...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add goprojectusat.go goprojectusat_test.go
git commit -m "$(cat <<'EOF'
feat: wire military street lines through Normalize and Format

EOF
)"
```

---

### Task 2: Parse skeleton — empty/error, multi-line split, military full address

**Files:**
- Create: `parse.go`
- Create: `parse_test.go`

**Interfaces:**
- Consumes: `military.NormalizeStreetLine`, `military.NormalizeLastLine`, `textutil`
- Produces: `func Parse(raw string) (Address, error)`

- [ ] **Step 1: Write failing tests** in `parse_test.go`

```go
package goprojectusat

import "testing"

func TestParseEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\n"} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}

func TestParseMilitaryMultiline(t *testing.T) {
	raw := "PSC 3 BOX 4120\nAPO AE 09021-0002"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Parse returns structured fields; may not yet Normalize — content may be upper from parse clean.
	if got.StreetName != "PSC 3 BOX 4120" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.City != "APO" || got.Region != "AE" || got.Postal != "09021-0002" {
		t.Errorf("last = %q %q %q", got.City, got.Region, got.Postal)
	}
	norm, err := Normalize(got)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if Format(norm) != "PSC 3 BOX 4120\nAPO AE 09021-0002" {
		t.Errorf("Format = %q", Format(norm))
	}
}

func TestParseMilitaryCommaSeparated(t *testing.T) {
	raw := "UNIT 2050 BOX 4190, APO AP 96278-2050"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.StreetName != "UNIT 2050 BOX 4190" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.City != "APO" || got.Region != "AP" {
		t.Errorf("city/region = %q %q", got.City, got.Region)
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (Parse undefined)

```bash
go test . -run 'TestParse' -v
```

- [ ] **Step 3: Implement minimal `parse.go`**

```go
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
```

For Task 2 tests that only cover military + empty, stub is OK if those tests pass. Do not leave stub if any test hits civilian yet.

- [ ] **Step 4: Run**

```bash
go test ./...
go vet ./...
```

- [ ] **Step 5: Commit**

```bash
git add parse.go parse_test.go
git commit -m "$(cat <<'EOF'
feat: Parse military multi-line and comma-separated addresses

EOF
)"
```

---

### Task 3: Parse last line (city, region, postal) + simple one-line street fallback

**Files:**
- Modify: `parse.go`
- Modify: `parse_test.go`

**Interfaces:**
- Consumes: `region.NormalizeRegion`, `normalizePostal` (same package, unexported OK)
- Produces: `parseLastLine(s string) (city, region, postal string, err error)` used by `parseCivilian`

- [ ] **Step 1: Failing tests**

```go
func TestParseSimpleMultiline(t *testing.T) {
	raw := "123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Street parse may be partial in Task 3 — if Task 3 only does last line + entire
	// prior line as StreetName, document that; Task 4 splits street components.
	// This plan assigns last-line + whole street line as StreetName in Task 3,
	// component split in Task 4.
	if got.City != "SPRINGFIELD" && got.City != "Springfield" {
		// Prefer uppercase in parse for consistency with military path:
		// parseLastLine should uppercase city tokens.
	}
	// Exact expectations after implementation:
	// City SPRINGFIELD, Region IL, Postal 62701
	// StreetName may still be full "123 Main Street" until Task 4.
}

func TestParseLastLineCanadian(t *testing.T) {
	// "OTTAWA ON K1A 0B1" — postal is two tokens.
	raw := "10 Wellington Street\nOttawa ON K1A 0B1"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Region != "ON" {
		t.Errorf("Region = %q, want ON", got.Region)
	}
	// Postal after Normalize would be K1A 0B1; Parse may store "K1A 0B1"
	if !strings.Contains(strings.ToUpper(got.Postal), "K1A") {
		t.Errorf("Postal = %q", got.Postal)
	}
}
```

Use exact assertions in the real test file:

```go
func TestParseSimpleMultiline(t *testing.T) {
	raw := "123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.City != "SPRINGFIELD" {
		t.Errorf("City = %q, want SPRINGFIELD", got.City)
	}
	if got.Region != "IL" {
		t.Errorf("Region = %q, want IL", got.Region)
	}
	if got.Postal != "62701" {
		t.Errorf("Postal = %q, want 62701", got.Postal)
	}
	// Task 3: street line kept intact in StreetName (componentization in Task 4)
	if got.StreetName != "123 MAIN STREET" && got.PrimaryNumber == "" {
		// If implementer already uppercases street line into StreetName:
		if strings.ToUpper(got.StreetName) != "123 MAIN STREET" && got.PrimaryNumber != "123" {
			t.Errorf("street not captured: %+v", got)
		}
	}
}
```

Simpler exact Task 3 contract:

- Last line always parsed into City/Region/Postal (Region abbreviated via `region.NormalizeRegion`).
- All non-last lines joined: if one line, put in StreetName (uppercase collapsed); if two+, first lines → BusinessName, last-but-one → StreetName.

```go
func TestParseSimpleMultiline(t *testing.T) {
	raw := "123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.City != "SPRINGFIELD" || got.Region != "IL" || got.Postal != "62701" {
		t.Fatalf("last line = %+v", got)
	}
	if got.StreetName != "123 MAIN STREET" {
		t.Fatalf("StreetName = %q (Task 3 stores full street line)", got.StreetName)
	}
}

func TestParseWithBusinessLine(t *testing.T) {
	raw := "Acme Corp\n123 Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.BusinessName != "ACME CORP" {
		t.Errorf("BusinessName = %q", got.BusinessName)
	}
	if got.StreetName != "123 MAIN STREET" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
}
```

- [ ] **Step 2: Implement `parseLastLine` and `parseCivilian`**

Algorithm for last line (tokens uppercased, collapsed):

1. `tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))`
2. If len < 2, error `fmt.Errorf("invalid last line: %q", line)`.
3. Try region at various positions from the right:
   - US ZIP forms: last token matches `^\d{5}(-\d{4})?$` or last two tokens form ZIP+4; region is token before postal; city is join of tokens before region.
   - Canadian: last two tokens look like `A1A` + `1A1` (letter-digit-letter digit-letter-digit); region before that; city before region.
   - Fallback: last token is postal (normalizePostal), second-to-last region via `region.NormalizeRegion`, rest city. If region fails, try second-to-last as part of city and third-to-last as region (multi-word city).
4. Multi-word city: everything left of region token is city (joined with spaces).
5. Multi-word region names: try longest suffix of tokens (before postal) that `region.NormalizeRegion` accepts (1 or 2+ tokens joined). Prefer longest match that works.

Concrete helper sketch:

```go
func parseLastLine(line string) (city, reg, postal string, err error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	if len(tokens) < 2 {
		return "", "", "", fmt.Errorf("invalid last line: %q", line)
	}
	// Peel postal from the right.
	postal, rest, ok := peelPostal(tokens)
	if !ok {
		return "", "", "", fmt.Errorf("invalid last line postal: %q", line)
	}
	if len(rest) < 1 {
		return "", "", "", fmt.Errorf("invalid last line: missing region in %q", line)
	}
	// Longest region match from the right of rest.
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

func peelPostal(tokens []string) (postal string, rest []string, ok bool) {
	if len(tokens) == 0 {
		return "", nil, false
	}
	last := tokens[len(tokens)-1]
	// ##### or #####-####
	if matched := /* use same usZIPCompact or simple regex */; matched {
		return normalizePostal(last), tokens[:len(tokens)-1], true
	}
	// Canadian two-token: K1A 0B1
	if len(tokens) >= 2 {
		two := tokens[len(tokens)-2] + " " + last
		np := normalizePostal(two)
		// Heuristic: Canadian after normalize has letter.
		if len(np) >= 6 && containsLetter(np) {
			return np, tokens[:len(tokens)-2], true
		}
	}
	// Single alphanumeric postal fallback
	return normalizePostal(last), tokens[:len(tokens)-1], true
}
```

`parseCivilian`:

```go
func parseCivilian(lines []string) (Address, error) {
	if len(lines) == 1 {
		// Single segment: cannot reliably split street vs last line without a
		// last-line pattern. Try: if line has a recognizable trailing last line
		// of 3+ tokens with region+postal, split; else error.
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
```

For single-line civilian in Task 3: try to peel last line tokens from the right (postal, region, city as 1 token minimum) and remainder is street. If region match fails, return error.

```go
func parseSingleLineCivilian(line string) (Address, error) {
	tokens := strings.Fields(strings.ToUpper(textutil.CollapseSpace(line)))
	if len(tokens) < 4 { // need at least number? street city region postal — soft minimum 3 for city region postal + something
		// try military already handled; fail clearly
	}
	// Walk: peel postal, peel region (longest), city is one token (conservative) or all remaining after street?
	// Conservative single-line: city is single token immediately before region.
	// street is everything before city.
	postal, rest, ok := peelPostal(tokens)
	if !ok || len(rest) < 2 {
		return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
	}
	for n := min(3, len(rest)); n >= 1; n-- { // region up to 3 tokens e.g. DISTRICT OF COLUMBIA
		cand := strings.Join(rest[len(rest)-n:], " ")
		if abbr, e := region.NormalizeRegion(cand); e == nil {
			before := rest[:len(rest)-n]
			if len(before) < 2 {
				return Address{}, fmt.Errorf("cannot parse single-line address: %q", line)
			}
			// last token of before = city (conservative single-word city on single-line)
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
```

Note: multi-word cities on single-line are out of scope (error or only multi-line form). Multi-line supports multi-word cities via `parseLastLine`.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

- [ ] **Step 4: Commit**

```bash
git add parse.go parse_test.go
git commit -m "$(cat <<'EOF'
feat(parse): last-line city/region/postal and civilian multi-line skeleton

EOF
)"
```

---

### Task 4: Reverse-token street line componentization

**Files:**
- Modify: `parse.go`
- Modify: `parse_test.go`

**Interfaces:**
- Consumes: `directionals.AbbreviateDirectional`, `streetsuffixes.NormalizeStreetSuffixAbreviation`, `secondaryunit.Info` / `Normalize`, `highways.NormalizeStreetName`
- Produces: `parseStreetLine(line string) (Address, error)` filling street fields; residual name in StreetName

- [ ] **Step 1: Failing tests**

```go
func TestParseStreetComponents(t *testing.T) {
	raw := "123 North Main Street Apt 4\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.PrimaryNumber != "123" {
		t.Errorf("PrimaryNumber = %q", got.PrimaryNumber)
	}
	if got.Predirectional != "N" {
		t.Errorf("Predirectional = %q", got.Predirectional)
	}
	if got.StreetName != "MAIN" {
		t.Errorf("StreetName = %q", got.StreetName)
	}
	if got.StreetSuffix != "ST" {
		t.Errorf("StreetSuffix = %q", got.StreetSuffix)
	}
	if got.SecondaryDesignator != "APT" {
		t.Errorf("SecondaryDesignator = %q", got.SecondaryDesignator)
	}
	if got.SecondaryNumber != "4" {
		t.Errorf("SecondaryNumber = %q", got.SecondaryNumber)
	}
}

func TestParseStreetPostdirectionalAndHash(t *testing.T) {
	raw := "100 Main Street Southwest # 12\nMiami FL 33101"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Postdirectional != "SW" {
		t.Errorf("Postdirectional = %q", got.Postdirectional)
	}
	if got.SecondaryDesignator != "#" || got.SecondaryNumber != "12" {
		t.Errorf("secondary = %q %q", got.SecondaryDesignator, got.SecondaryNumber)
	}
}

func TestParseStreetHighway(t *testing.T) {
	raw := "3324 TN HIGHWAY 431\nSomewhere KY 40000"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// After component parse + highway normalize on name span:
	// PrimaryNumber 3324, StreetName TN HIGHWAY 431 (or KY depending on input)
	if got.PrimaryNumber != "3324" {
		t.Errorf("PrimaryNumber = %q", got.PrimaryNumber)
	}
	if got.StreetName != "TN HIGHWAY 431" {
		t.Errorf("StreetName = %q, want TN HIGHWAY 431", got.StreetName)
	}
}

func TestParseThenNormalizeRoundTrip(t *testing.T) {
	raw := "123 North Main Street Apt 4\nSpringfield IL 62701"
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	norm, err := Normalize(parsed)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	want := "123 N MAIN ST APT 4\nSPRINGFIELD IL 62701"
	if Format(norm) != want {
		t.Fatalf("Format = %q, want %q", Format(norm), want)
	}
}
```

- [ ] **Step 2: Implement `parseStreetLine`**

Clean line: `strings.ToUpper(textutil.CollapseSpace(textutil.StripPunctuation(line, textutil.StripOptions{KeepHyphen: true, KeepSlash: true})))` — keep `#`.

Note: `StripPunctuation` keeps `#` by default. Good.

Tokenize with `strings.Fields`. Also split glued `#12` → `#`, `12` if needed:

```go
func expandHashTokens(tokens []string) []string {
	var out []string
	for _, t := range tokens {
		if strings.HasPrefix(t, "#") && len(t) > 1 {
			out = append(out, "#", t[1:])
			continue
		}
		out = append(out, t)
	}
	return out
}
```

Reverse peel:

1. **Secondary:**  
   - If last token is `#` and no number — designator `#` only (rare).  
   - If second-to-last is secondary designator (`secondaryunit.Info` OK) and Info.Numbered, last is SecondaryNumber, designator normalized short.  
   - If last is secondary and !Numbered, designator only.  
   - If last is `#` wait — if second-to-last is `#`, last is number.  
   - If only last is secondary numbered without number, keep designator, empty number.

2. **Postdirectional:** if last remaining token is directional (`directionals.AbbreviateDirectional` OK), set Postdirectional to abbreviation.  
   - Optional: if last two tokens are single-letter directionals that combine (N+E→NE), merge (only N/S then E/W pairs).  
   - Do **not** merge NORTH SOUTH or EAST WEST consecutive (leave second as street name) — if both still in street name span, fine for v1.

3. **Suffix:** if last remaining is street suffix (`streetsuffixes.NormalizeStreetSuffixAbreviation`), set StreetSuffix.  
   - If the only tokens left would be empty street name and primary, still allow.  
   - Double suffix: if two suffixes in a row, outer is StreetSuffix, inner stays in street name as primary word (spell-out not required if we only store abbreviation for true suffix). Conservative: only peel one suffix from the right.

4. **Primary number:** first token if `looksLikePrimaryNumber(tok)` — has a digit, not purely ordinal street? Use: contains digit OR matches `^\d+[-/].*` OR alphanumeric with digit. Exclude tokens that are only directionals.  
   `looksLikePrimaryNumber`: any digit in token, or hyphenated number pattern.

5. **Predirectional:** if next token (after primary) is directional, set Predirectional abbreviation.

6. **Street name:** remaining tokens joined. Run `highways.NormalizeStreetName`; if success use that; else joined uppercase name. If empty street name after peels, error `fmt.Errorf("unrecognized street line: %q", line)`.

7. Assign fields onto Address partial and merge with last-line fields in `parseCivilian`.

```go
func parseStreetLine(line string) (Address, error) {
	// returns Address with only street-related fields set
}
```

Wire into `parseCivilian` / `parseSingleLineCivilian` so street line is componentized instead of dumped into StreetName only.

- [ ] **Step 3: Run**

```bash
go test ./...
go vet ./...
```

- [ ] **Step 4: Commit**

```bash
git add parse.go parse_test.go
git commit -m "$(cat <<'EOF'
feat(parse): reverse-token street line componentization

EOF
)"
```

---

### Task 5: README + edge-case hardening + full suite

**Files:**
- Modify: `README.md`
- Modify: `parse_test.go` (extra edges)
- Modify: `parse.go` if gaps found

- [ ] **Step 1: Edge tests**

```go
func TestParseUnknownTokenInRegionErrors(t *testing.T) {
	_, err := Parse("123 Main St\nSpringfield ZZ 62701")
	if err == nil {
		t.Fatal("expected error for bad region")
	}
}

func TestParseMilitaryRejectsCountryOnLastLine(t *testing.T) {
	// military package rejects extra tokens; Parse should not accept as military.
	// May fall through to civilian and fail region — either error is OK.
	_, err := Parse("PSC 3 BOX 4120\nAPO AE GERMANY 09021-0002")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseDirectionalsNE(t *testing.T) {
	raw := "101 Northeast Main Street\nSpringfield IL 62701"
	got, err := Parse(raw)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Predirectional != "NE" {
		t.Errorf("Predirectional = %q, want NE", got.Predirectional)
	}
}
```

- [ ] **Step 2: Update README**

Replace incomplete blurb:

**What works today:** structured `Address` normalization (`Normalize` / `NormalizeWithOptions`), free-text `Parse` (multi-line / comma-separated; military overseas fast path; reverse-token street parse), street/last-line formatting, and component packages under `pkg/`.

**Not yet:** rural route / PO Box free-text rewrite, secondary-before-primary reordering, Puerto Rico parse integration, full C#-parity street-line edge cases.

Document:

```go
addr, err := goprojectusat.Parse("123 North Main Street Apt 4\nSpringfield IL 62701")
norm, err := goprojectusat.Normalize(addr)
fmt.Println(goprojectusat.Format(norm))
// 123 N MAIN ST APT 4
// SPRINGFIELD IL 62701
```

Military:

```go
addr, err := goprojectusat.Parse("PSC 3 BOX 4120\nAPO AE 09021-0002")
```

- [ ] **Step 3: Full verification**

```bash
go test ./...
go vet ./...
gofmt -w parse.go parse_test.go goprojectusat.go goprojectusat_test.go README.md
```

- [ ] **Step 4: Commit**

```bash
git add README.md parse.go parse_test.go goprojectusat.go goprojectusat_test.go
git commit -m "$(cat <<'EOF'
docs: document Parse and military root support

EOF
)"
```

---

## Verification (every task)

```bash
go test ./...
go vet ./...
```

## Self-review (plan vs goal)

| Goal element | Task |
|--------------|------|
| Military in root Normalize/Format | 1 |
| Parse API + military multi-line/comma | 2 |
| Last line + business line + civilian skeleton | 3 |
| Street component parse + highway | 4 |
| Docs + edges + QA suite | 5 |
| No RR/PO/C# full parity (YAGNI) | Explicit out-of-scope |

Placeholder scan: no TBD steps; concrete tests and algorithms included.

Type consistency: `Parse(raw string) (Address, error)` throughout; military street stored in `StreetName`; Region is 2-letter after parse last-line.

## Plan test notes (controller)

Before dispatch:

1. Baseline `go test ./...` is green on branch.
2. Task 1 must not break existing Normalize tests (military early-path only when `military.NormalizeStreetLine` succeeds on candidate).
3. Task 2 stub `parseCivilian` must not ship alone — Tasks 3–4 complete civilian path before considering feature done.
4. `region.NormalizeRegion` already knows AE/AP/AA and Canadian provinces.
5. Do not call Normalize inside Parse.

---

## Upstream PR checklist (when opening)

```markdown
## What was addressed
- Free-text Parse for multi-line and comma-separated addresses
- Overseas military APO/FPO/DPO fast path in Parse and Normalize
- Reverse-token street line componentization for common civilian forms
- README updates for Parse and military support

## How to test / validate
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] Manual: Parse+Normalize+Format military example matches PSC/APO form
- [ ] Manual: Parse+Normalize+Format `123 North Main Street Apt 4\nSpringfield IL 62701`

## Configurations or dependencies
- [ ] None
```
