# Project US@ Normalizer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver working, tested Go code for Project US@ address normalization that can be PR'd incrementally to [PortobelloAuth/go-projectusat](https://github.com/PortobelloAuth/go-projectusat).

**Architecture:** Keep the existing `pkg/*` component packages as pure normalizers (lookup tables + `Normalize*` functions). Add a root-package structured `Address` model and field-level / whole-address normalizers that compose those packages. Defer free-text multi-line parsing until structured normalization is solid. Prefer small, reviewable PRs that each leave `go test ./...` green.

**Tech Stack:** Go 1.25+, module `github.com/PortobelloAuth/go-projectusat`, deps `github.com/anyascii/go`, `github.com/hbollon/go-edlib`, `golang.org/x/text`. Apache-2.0. Table-driven tests matching `pkg/directionals` and `pkg/diacritics` style.

## Global Constraints

- Module path remains `github.com/PortobelloAuth/go-projectusat` (do not rename for local branding).
- API is unstable; still export small, documented surfaces only.
- Alphabetical address text is uppercase in normalized **content** output (Project US@ letter-case rule), except where a package already documents a deliberate deviation (e.g. `diacritics.Substitute` returns lowercase per Appendix A table).
- Unknown fields: blank string; the token `UNKNOWN` normalizes to blank.
- Do not add CGO, libpostal, network calls, or external validation services.
- Every task ends with `go test ./...` passing.
- Match existing error style: `fmt.Errorf("Unrecognized ...")` for unknown tokens.
- Prefer exact table lookup; fuzzy matching is opt-in via `Fuzzy*` functions only.
- PR body for upstream uses: What was addressed / How to test / Configurations or dependencies.

## File Structure

| Path | Responsibility |
|------|----------------|
| `pkg/region/region.go` | State / province / military region → 2-letter code |
| `pkg/region/region_test.go` | Region tests |
| `pkg/streetsuffixes/streetsuffixes.go` | Street suffix primary + abbreviation maps |
| `pkg/streetsuffixes/streetsuffixes_test.go` | Street suffix tests |
| `pkg/secondaryunit/secondaryunit.go` | Secondary unit designators |
| `pkg/secondaryunit/secondaryunit_test.go` | Secondary unit tests |
| `pkg/directionals/` | Predirectional / postdirectional (done) |
| `pkg/diacritics/` | Appendix A diacritics (done) |
| `pkg/highways/highways.go` | Highway / route street-name normalization |
| `pkg/highways/highways_test.go` | Highway examples from the standard |
| `pkg/military/military.go` | APO/FPO/DPO street-line helpers |
| `pkg/military/military_test.go` | Military examples |
| `pkg/puertorico/puertorico.go` | Spanish street / unit vocabulary for PR |
| `pkg/puertorico/puertorico_test.go` | Puerto Rico vocabulary tests |
| `goprojectusat.go` | Root types + `Normalize` for structured addresses |
| `goprojectusat_test.go` | Root integration tests (replace comment-only file content gradually; keep useful standard quotes only where they document behavior) |
| `textutil.go` (new, root or `pkg/textutil`) | Shared punctuation / whitespace / UNKNOWN helpers |

PR grouping (upstream):

1. **Foundations** — Tasks 1–3 (tests + region data fixes)
2. **Components** — Tasks 4–6 (highways, military, puertorico)
3. **Structured normalize** — Tasks 7–9 (text helpers, Address model, Normalize)
4. **Options** — Task 10 (matching mode: secondary unit as `#`, fuzzy flags)

---

### Task 1: Region tests + data fixes

**Files:**
- Create: `pkg/region/region_test.go`
- Modify: `pkg/region/region.go` (`regionMap` entries for Delaware and District of Columbia)

**Interfaces:**
- Consumes: `NormalizeRegion(string) (string, error)`, `FuzzyNormalizeRegion(string) (string, error)`
- Produces: Correct map data; full test coverage of representative regions

- [x] **Step 1: Write the failing tests** (file created in this plan deliverable)

Run:

```bash
go test ./pkg/region/ -v
```

Expected before fix: FAIL on `District of Columbia` (returns `DE`) and `Delaware` (unrecognized).

- [x] **Step 2: Fix region map data**

In `pkg/region/region.go`:

```go
"DELAWARE":                       "DE",
"DELEWARE":                       "DE", // common misspelling retained
"DE":                             "DE",
"DISTRICT OF COLUMBIA":           "DC",
"DC":                             "DC",
```

- [x] **Step 3: Run tests**

```bash
go test ./pkg/region/ -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/region/region.go pkg/region/region_test.go
git commit -m "test(region): cover NormalizeRegion and fix DC/Delaware map entries"
```

---

### Task 2: Street suffix tests

**Files:**
- Create: `pkg/streetsuffixes/streetsuffixes_test.go`

**Interfaces:**
- Consumes: `NormalizeStreetSuffix`, `NormalizeStreetSuffixAbreviation`, `FuzzyNormalizeStreetSuffix`, `FuzzyNormalizeStreetSuffixAbreviation`
- Produces: Confidence that common suffixes and alts resolve correctly

- [x] **Step 1: Add table-driven tests** (file created in this plan deliverable)

- [x] **Step 2: Run tests**

```bash
go test ./pkg/streetsuffixes/ -v
```

Expected: PASS (no code changes required if tables already match USPS alts)

- [ ] **Step 3: Commit**

```bash
git add pkg/streetsuffixes/streetsuffixes_test.go
git commit -m "test(streetsuffixes): add table-driven primary and abbreviation tests"
```

---

### Task 3: Secondary unit tests

**Files:**
- Create: `pkg/secondaryunit/secondaryunit_test.go`

**Interfaces:**
- Consumes: `Normalize(string) (string, error)`, `Info(string) (*SecondaryUnit, error)`
- Produces: Coverage of numbered vs non-numbered units and unknown token errors

- [x] **Step 1: Add tests** (file created in this plan deliverable)

- [x] **Step 2: Run tests**

```bash
go test ./pkg/secondaryunit/ -v
```

Expected: PASS

- [ ] **Step 3: Commit** (or fold Tasks 1–3 into one foundations PR)

```bash
git add pkg/secondaryunit/secondaryunit_test.go
git commit -m "test(secondaryunit): cover Normalize and Info for unit designators"
```

**Upstream PR 1 (Foundations):** Tasks 1–3.

---

### Task 4: Highways normalizer

**Files:**
- Modify: `pkg/highways/highways.go`
- Create/replace: `pkg/highways/highways_test.go`

**Interfaces:**
- Consumes: none from other packages initially (state abbreviation may call `region.NormalizeRegion` when a full state name prefixes a highway)
- Produces:

```go
// NormalizeStreetName normalizes highway-style primary street names per Project US@.
// Input is the street name portion only (not full address). Returns uppercase.
func NormalizeStreetName(name string) (string, error)
```

Behavior (from comments already in `highways.go`):

| Input | Output |
|-------|--------|
| `COUNTY HWY 60E` | `COUNTY HIGHWAY 60E` |
| `CNTY HWY 20` | `COUNTY HIGHWAY 20` |
| `COUNTY RD 441` | `COUNTY ROAD 441` |
| `CR 1185` | `COUNTY ROAD 1185` |
| `FARM TO MARKET 1200` | `FM 1200` |
| `HWY FM 1320` | `FM 1320` |
| `HWY 64` | `HIGHWAY 64` |
| `I10` | `INTERSTATE 10` |
| `IH280` | `INTERSTATE 280` |
| `INTERSTATE HWY 680` | `INTERSTATE 680` |
| `RT 88` | `ROUTE 88` |
| `SR 220` | `STATE ROAD 220` |
| `US HWY 44` | `US HIGHWAY 44` |
| `KENTUCKY 440` | `KY HIGHWAY 440` |
| `KY 1207` | `KY HIGHWAY 1207` |

- [ ] **Step 1: Write failing table tests** covering the rows above

- [ ] **Step 2: Run tests — expect FAIL** (function missing)

```bash
go test ./pkg/highways/ -v
```

- [ ] **Step 3: Implement `NormalizeStreetName`**

Approach:

1. Uppercase and collapse whitespace.
2. Apply ordered rewrite rules (regex or token-scan) for interstate, FM, county/state/township/ranch/US highway patterns.
3. Expand `HWY`→`HIGHWAY`, `RD`→`ROAD`, `RT`/`RTE`→`ROUTE` only when they are highway vocabulary, not when they would be ordinary street suffixes outside these patterns (this function only handles highway street-name forms).
4. When a full US state name is a leading token of a highway form, replace with two-letter code via `region.NormalizeRegion`.

- [ ] **Step 4: Run tests — expect PASS**

- [ ] **Step 5: Commit**

```bash
git add pkg/highways/
git commit -m "feat(highways): normalize highway and route street names"
```

---

### Task 5: Military address helpers

**Files:**
- Modify: `pkg/military/military.go`
- Replace: `pkg/military/military_test.go` (today is comments only)

**Interfaces:**

```go
type AddressType string // "CMR" | "OMC" | "PSC" | "UMR" | "UNIT"

// NormalizeStreetLine normalizes the military street line, e.g.
// "PSC 3 BOX 4120" stays "PSC 3 BOX 4120" with uppercase and single spaces.
func NormalizeStreetLine(line string) (string, error)

// NormalizeLastLine normalizes "APO AE 09021-0002" style last lines.
// City must be APO, FPO, or DPO; region AE/AP/AA; ZIP or ZIP+4.
func NormalizeLastLine(line string) (city, region, postal string, err error)
```

Examples from standard:

| Street | Last line city/region/ZIP |
|--------|---------------------------|
| `PSC 3 BOX 4120` | `APO` / `AE` / `09021-0002` |
| `UNIT 2050 BOX 4190` | `APO` / `AP` / `96278-2050` |
| `UNIT 100100 BOX 4120` | `FPO` / `AP` / `96691-0104` |
| `UNIT 8400 BOX 0000` | `DPO` / `AE` / `09498-0048` |

- [ ] **Step 1: Write failing tests** for street + last line

- [ ] **Step 2: Implement minimal parsers** (token-based; reject city/country names on overseas military last lines)

- [ ] **Step 3: `go test ./pkg/military/ -v` PASS**

- [ ] **Step 4: Commit**

```bash
git add pkg/military/
git commit -m "feat(military): normalize APO/FPO/DPO street and last lines"
```

---

### Task 6: Puerto Rico vocabulary

**Files:**
- Modify: `pkg/puertorico/puertorico.go`
- Replace: `pkg/puertorico/puertorico_test.go`

**Interfaces:**

```go
// NormalizeStreetSuffix maps Spanish PR street types to Spanish primary form
// (standard: keep Spanish, do not force English).
// Example: "AVE" or "AVENIDA" -> "AVENIDA" (or short "AVE" via Abbreviate).
func NormalizeStreetType(s string) (string, error)
func AbbreviateStreetType(s string) (string, error)
func NormalizeSecondary(s string) (string, error) // Apartamento->APT, etc.
```

Table from existing comments: AVENIDA/AVE, CALLE/CLL, CAMINO/CAM, … plus secondary URB, COND, etc.

- [ ] **Step 1: Failing table tests**

- [ ] **Step 2: Implement maps** (same pattern as `directionals` / `secondaryunit`)

- [ ] **Step 3: Tests pass; commit**

```bash
git add pkg/puertorico/
git commit -m "feat(puertorico): Spanish street and secondary unit vocabulary"
```

**Upstream PR 2 (Components):** Tasks 4–6.

---

### Task 7: Shared text normalization helpers

**Files:**
- Create: `pkg/textutil/textutil.go`
- Create: `pkg/textutil/textutil_test.go`

**Interfaces:**

```go
package textutil

// CollapseSpace turns all whitespace runs into a single ASCII space and trims ends.
func CollapseSpace(s string) string

// StripPunctuation removes Project US@ special characters, preserving:
// - hyphen in primary number and ZIP+4 contexts (caller-controlled via options)
// - pound sign #
// Default: remove * , . ( ) " : ; ` @ & and most hyphens.
func StripPunctuation(s string, opts StripOptions) string

type StripOptions struct {
	KeepHyphen bool
	KeepSlash  bool // fractional addresses 1/2
}

// NormalizeUnknown returns "" if s is empty or equals "UNKNOWN" (any case).
func NormalizeUnknown(s string) string

// Upper is strings.ToUpper with UNKNOWN handling applied first.
func Upper(s string) string
```

- [ ] **Step 1: Tests for whitespace, UNKNOWN, punctuation, keep-hyphen ZIP and primary number samples**

- [ ] **Step 2: Implement**

- [ ] **Step 3: Commit**

```bash
git add pkg/textutil/
git commit -m "feat(textutil): shared whitespace, punctuation, and UNKNOWN helpers"
```

---

### Task 8: Root Address model

**Files:**
- Modify: `goprojectusat.go`
- Modify: `goprojectusat_test.go` (add real tests; keep short package doc)

**Interfaces:**

```go
package goprojectusat

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Address struct {
	BusinessName string // firm / business line (optional)

	// Street line elements
	PrimaryNumber      string
	Predirectional     string
	StreetName         string
	StreetSuffix       string
	Postdirectional    string
	SecondaryDesignator string // APT, STE, ...
	SecondaryNumber    string

	// Last line
	City   string
	Region string // state / province / military "state"
	Postal string // ZIP, ZIP+4, or Canadian postal code

	Country string // optional; often blank for domestic
}

// Normalize returns a content-normalized copy (uppercase, standard abbreviations).
func Normalize(a Address) (Address, error)
```

Normalization rules for `Normalize`:

1. Apply `textutil.NormalizeUnknown` + uppercase to all string fields.
2. `Region` → `region.NormalizeRegion`.
3. `Predirectional` / `Postdirectional` → `directionals.AbbreviateDirectional` when non-empty.
4. `StreetSuffix` → `streetsuffixes.NormalizeStreetSuffixAbreviation` when non-empty.
5. `SecondaryDesignator` → `secondaryunit.Normalize` when non-empty.
6. `StreetName` → try `highways.NormalizeStreetName`; if it errors, keep collapsed uppercase street name (ordinary names are free text).
7. `Postal`: US ZIP `#####` or `#####-####`; strip invalid punctuation; leave Canadian patterns as uppercase alphanumeric with space collapsed to single space if present.
8. Diacritics: do **not** strip by default in content mode (optional later). Document that callers may pre-run `diacritics.Substitute`.

- [ ] **Step 1: Write tests** with structured inputs:

```go
// 123 Main Street, Apt 4, Springfield IL 62701
in := Address{
  PrimaryNumber: "123", StreetName: "Main", StreetSuffix: "Street",
  SecondaryDesignator: "Apartment", SecondaryNumber: "4",
  City: "Springfield", Region: "Illinois", Postal: "62701",
}
// expect StreetSuffix "ST", SecondaryDesignator "APT", Region "IL", all upper
```

- [ ] **Step 2: Implement `Normalize`**

- [ ] **Step 3: `go test .` PASS**

- [ ] **Step 4: Commit**

```bash
git add goprojectusat.go goprojectusat_test.go
git commit -m "feat: structured Address and Normalize for Project US@ content form"
```

---

### Task 9: Last-line and street-line formatters

**Files:**
- Modify: `goprojectusat.go`
- Modify: `goprojectusat_test.go`

**Interfaces:**

```go
// FormatStreetLine joins street elements with single spaces; omits blanks.
func FormatStreetLine(a Address) string

// FormatLastLine joins CITY REGION POSTAL with single spaces (1+ spaces allowed by standard; we use one).
func FormatLastLine(a Address) string

// Format returns business line (if any), street line, last line separated by \n.
func Format(a Address) string
```

- [ ] **Step 1: Tests** for element order:  
  `PRIMARY PREDIR STREET SUFFIX POSTDIR SEC SECNUM`

- [ ] **Step 2: Implement join helpers**

- [ ] **Step 3: Commit**

```bash
git add goprojectusat.go goprojectusat_test.go
git commit -m "feat: format street and last lines from structured Address"
```

**Upstream PR 3 (Structured normalize):** Tasks 7–9.

---

### Task 10: Matching / exchange options

**Files:**
- Modify: `goprojectusat.go`
- Modify: `goprojectusat_test.go`

**Interfaces:**

```go
type Options struct {
	// Fuzzy enables FuzzyNormalize* for region and street suffix.
	Fuzzy bool
	// SecondaryAsHash rewrites secondary designators to "#" for matching
	// (not correct for content storage; for exchange/matching only).
	SecondaryAsHash bool
	// DiacriticMode: "" = leave as-is, "substitute" = diacritics.Substitute then upper, "transliterate" = anyascii path
	DiacriticMode string
}

func NormalizeWithOptions(a Address, opts Options) (Address, error)
```

- [ ] **Step 1: Tests** proving `Apartment`→`#` when option set; fuzzy recovers mild typos

- [ ] **Step 2: Implement**

- [ ] **Step 3: README note** on content vs exchange

- [ ] **Step 4: Commit + Upstream PR 4**

```bash
git add goprojectusat.go goprojectusat_test.go README.md
git commit -m "feat: NormalizeWithOptions for fuzzy and matching secondary units"
```

---

### Task 11 (later / optional): Free-text parse

**Out of scope for first four PRs.** When started:

- Parse multiline / comma-separated candidate into `Address` using notes in `goprojectusat.go`.
- Do not invent geocoder-backed disambiguation for numeric street names or dual directionals.
- New API: `Parse(raw string) (Address, error)` with conservative heuristics and errors on ambiguity.

---

## Verification (every PR)

```bash
go test ./...
go vet ./...
```

Optional before open PR:

```bash
gofmt -w $(find . -name '*.go' -not -path './.git/*')
```

## Upstream PR checklist

1. Branch from latest `origin/main`.
2. One concern per PR (Foundations / Components / Structured / Options).
3. PR body:

```markdown
## What was addressed
- ...

## How to test / validate
- [ ] `go test ./...` passes
- [ ] ...

## Configurations or dependencies
- [ ] None (or list)
```

4. Do not force-push shared branches; do not rewrite published history.

## Self-review (plan vs goal)

| Goal element | Task |
|--------------|------|
| Working tested components | 1–6 |
| Composable structured API | 7–9 |
| Content vs exchange | 10 |
| PR-back-able increments | PR grouping above |
| Free-text parse | 11 deferred |

Placeholder scan: no TBD steps; concrete APIs and commands included.

---

## Current deliverable (this session)

Alongside this plan, foundation tests for Tasks 1–3 are added under:

- `pkg/region/region_test.go`
- `pkg/streetsuffixes/streetsuffixes_test.go`
- `pkg/secondaryunit/secondaryunit_test.go`

Task 1 data fixes (`DELAWARE`, `DISTRICT OF COLUMBIA` → `DC`) should be applied so those tests pass.
