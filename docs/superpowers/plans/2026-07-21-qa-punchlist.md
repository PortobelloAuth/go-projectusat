# QA Punchlist — Parse / Normalize / Format

> Generated 2026-07-21 by a three-agent QA team (E2E, C# parity, adversarial).
> Full agent reports: `.superpowers/sdd/qa-agent{1,2,3}-*.md`

**Baseline:** `go test ./...` green on `feat/project-usat-normalizer` @ `d9d434c` (+ local punchlist commits).  
**Happy path:** 59/59 E2E cases from Agent 1 pass.  
**C# street sample:** ~72% OK on 98-case battery (Agent 2).

## How to use

1. Each item has a **falsifiable test** in `punchlist_test.go` (build tag `punchlist`).
2. Run acceptance (expected RED until fixed):

```bash
go test -tags punchlist . -run 'TestPunchlist_' -v -count=1
```

3. Fix one PL id at a time; when its test goes GREEN, mark the item done and move the test into the main suite (drop the build tag / rename into `parse_test.go`).
4. Do not “fix” by weakening the test to match wrong behavior.

## Severity legend

| Level | Meaning |
|-------|---------|
| **blocker** | Silent corruption of common patient addresses; fix before release |
| **major** | Clear wrong answer on realistic input; high user impact |
| **minor** | Edge / rare / partial C# parity; fix when touching area |

## Priority order (recommended)

1. PL-001 space-separated military (silent wrong)
2. PL-002 single-line multi-word city
3. PL-003 peelPostal Canadian / trailing country greed
4. PL-004 comma last-line / comma secondary
5. PL-005 postdir peels sole street name (`123 South`)
6. PL-006 splitPreStreet digit-leading firm (`3M`)
7. PL-007 grid decimal period stripped
8. PL-008 WAY/WY and state-as-name collisions
9. Remaining majors, then minors

---

## Blocker / High

### PL-001: Space-separated military misparsed as business + junk street

- **Severity:** blocker  
- **Sources:** A3-52  
- **Repro:** `Parse("PSC 3 BOX 4120 APO AE 09021-0002")`  
- **Actual:** `BusinessName=PSC`, `StreetName=BOX 4120`, `City=APO` (or similar wrong split)  
- **Expected:** Same as multi-line military: `StreetName=PSC 3 BOX 4120`, `City=APO`, `Region=AE`, `Postal=09021-0002`  
- **Root cause:** Military fast path requires ≥2 logical lines; space-only form never matches; `splitPreStreet` steals `PSC`.  
- **Falsifiable test:** `TestPunchlist_PL001_SpaceSeparatedMilitary`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-002: Single-line multi-word city invents wrong city

- **Severity:** blocker  
- **Sources:** A1-005, A3-54  
- **Repro:** `Parse("123 Main Street New York NY 10005")`  
- **Actual:** `City=YORK`, street polluted with `NEW`  
- **Expected:** `City=NEW YORK`, `Region=NY`, `Postal=10005`, street `123 MAIN ST`  
- **Root cause:** `parseSingleLineCivilian` takes only one city token.  
- **Falsifiable test:** `TestPunchlist_PL002_SingleLineMultiWordCity`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-003: `peelPostal` over-accepts letter-bearing pairs (CA compact + trailing USA)

- **Severity:** blocker (two symptoms, one root)  
- **Sources:** A1-004, A1-010, A3-33/53/59  
- **Repro A:** `Parse("10 Wellington Street\nOttawa ON K1A0B1")` → error or wrong region  
- **Repro B:** `Parse("123 Main Street\nSpringfield IL 62701 USA")` → `Postal="62701 USA"` after Normalize  
- **Expected A:** region `ON`, postal `K1A 0B1` (normalized spacing)  
- **Expected B:** postal `62701`, country `USA` or country stripped from last line per Project US@  
- **Root cause:** Canadian two-token branch accepts any `len≥6` letter-bearing join; trailing tokens stick to postal.  
- **Falsifiable tests:** `TestPunchlist_PL003a_CompactCanadianPostal`, `TestPunchlist_PL003b_TrailingCountryOnLastLine`  
- **Status:** **done** (minion wave 2026-07-21)

---

## Major

### PL-004: Comma splits destroy last lines and multi-unit streets

- **Severity:** major  
- **Sources:** A1-001, A1-002, A2-007  
- **Repro A:** `Parse("123 Main Street\nSpringfield, IL 62701")` → error  
- **Repro B:** `Parse("123 Main Street, Apt 4\nSpringfield IL 62701")` → error or wrong  
- **Repro C:** `Parse("4004 PINE Circle, Court\nSpringfield IL 62701")` (C# → `PINE CIRCLE CT`)  
- **Expected:** commas as soft separators within a line, not always hard line breaks for last-line/unit contexts  
- **Falsifiable tests:** `TestPunchlist_PL004a_CommaInLastLine`, `TestPunchlist_PL004b_CommaBeforeApt`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-005: Postdirectional peels the only street-name token

- **Severity:** major  
- **Sources:** A3-51  
- **Repro:** `Parse("123 South\nSpringfield IL 62701")`  
- **Actual:** `unrecognized street line`  
- **Expected:** primary `123`, street name `SOUTH` (directional-as-name), matching predir empty-name guard already used for predirectionals  
- **Falsifiable test:** `TestPunchlist_PL005_DirectionalOnlyStreetName`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-006: Digit-leading firm name steals primary number

- **Severity:** major  
- **Sources:** A3-40  
- **Repro:** `Parse("3M Corporation 100 Main Street\nSpringfield IL 62701")`  
- **Actual:** `PrimaryNumber=3M` (or business empty / primary wrong)  
- **Expected:** business contains `3M CORPORATION`, primary `100`, street `MAIN ST`  
- **Falsifiable test:** `TestPunchlist_PL006_DigitLeadingBusiness`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-007: Grid-style decimal period stripped

- **Severity:** major  
- **Sources:** A2-011  
- **Repro:** `Parse("123 Road 39.4\nSpringfield IL 62701")`  
- **Actual:** street involves `394` (period removed)  
- **Expected:** C# / grid rule keeps decimal: `ROAD 39.4` (or `123 ROAD 39.4`)  
- **Falsifiable test:** `TestPunchlist_PL007_GridDecimalPeriod`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-008: WAY/WY and state-as-name collisions

- **Severity:** major  
- **Sources:** A2-002  
- **Repro:** `Parse("8011 WY WY\nSpringfield IL 62701")`  
- **Actual:** `8011 WY WY` (not Wyoming + Way)  
- **Expected (C#):** `8011 WYOMING WAY`  
- **Also:** `Way` suffix should abbreviate to `WAY` not `WY` when it is a street suffix  
- **Falsifiable test:** `TestPunchlist_PL008_WyomingWay`  
- **Status:** open

### PL-009: Directional-as-name under/over expansion

- **Severity:** major  
- **Sources:** A2-001  
- **Repro A:** `1014 BAY W DRIVE` → actual `1014 BAY W DR` (W not expanded to WEST in name)  
- **Expected (C#):** `1014 BAY WEST DR`  
- **Repro B:** over-merge cases where `N E` mid-name should not always become NE (context-sensitive)  
- **Falsifiable test:** `TestPunchlist_PL009_BayWestDrive`  
- **Status:** open

### PL-010: State as *portion* of street name not abbreviated

- **Severity:** major  
- **Sources:** A2-004, A2-005  
- **Repro:** `8100 Montana Treasure Avenue` → should be `8100 MT TREASURE AVE`  
- **Repro:** `South Carolina county road 22` mishandles leading SOUTH as predir  
- **Falsifiable tests:** `TestPunchlist_PL010a_MontanaTreasure`, `TestPunchlist_PL010b_SouthCarolinaCountyRoad`  
- **Status:** open

### PL-011: Leading UNIT + trailing RM order inverted

- **Severity:** major  
- **Sources:** A2-008  
- **Repro:** C# case `Unit 3200 152 Tech Dr Room 12`  
- **Expected:** `152 TECH DR UNIT 3200 RM 12` (order of secondaries)  
- **Falsifiable test:** `TestPunchlist_PL011_UnitThenRoomOrder`  
- **Status:** open

### PL-012: Business + Suite before primary misparsed

- **Severity:** major  
- **Sources:** A2-009  
- **Repro:** `UCENT Building Suite 480 411 N Central Ave`  
- **Expected (C#):** business retained, primary `411`, suite as secondary  
- **Falsifiable test:** `TestPunchlist_PL012_BusinessSuiteThenPrimary`  
- **Status:** open

### PL-013: Rural route phrase variants incomplete

- **Severity:** major  
- **Sources:** A2-006, A2-010, A1-009  
- **Repro:** `RFD Route 61 Box 87b`, `Rural Route NO. 91 Box A7`, Spanish `BZN`/`BUZON`, glued `RR0061#87b`  
- **Expected:** normalize to `RR n BOX id`  
- **Falsifiable tests:** `TestPunchlist_PL013a_RFDRoutePhrase`, `TestPunchlist_PL013b_GluedRRHash`  
- **Status:** open

### PL-014: Trailing junk blocks suffix peel

- **Severity:** major  
- **Sources:** A1-003  
- **Repro:** street like `Oak Boulevard Box 9` (non-RR) loses Boulevard peel  
- **Falsifiable test:** `TestPunchlist_PL014_TrailingJunkAfterSuffix`  
- **Status:** open

### PL-015: Parse vs structured Normalize diverge on highway BYP/FRONTAGE

- **Severity:** major  
- **Sources:** A1-006  
- **Repro:** same highway string via Parse path vs field-level Address differs on BYPASS/FRONTAGE  
- **Expected:** identical content form  
- **Falsifiable test:** `TestPunchlist_PL015_HighwayParseNormalizeParity`  
- **Status:** **done** (already GREEN under `-tags punchlist` as of punchlist commit — no code change required; keep as regression)

### PL-016: Spanish URBANIZACION rejected

- **Severity:** major  
- **Sources:** A1-007  
- **Repro:** PR secondary `Urbanizacion` / `URBANIZACION` (correct Spanish spelling)  
- **Actual:** error or unrecognized; only English `URBANIZATION` in map  
- **Expected:** map to `URB`  
- **Falsifiable test:** `TestPunchlist_PL016_UrbanizacionSpelling`  
- **Status:** open

### PL-017: UNIT BOX street + civilian last line inconsistent

- **Severity:** major  
- **Sources:** A3-57  
- **Repro:** `Parse("UNIT 2050 BOX 4190\nSpringfield IL 62701")` errors; structured Normalize of same street OK  
- **Expected:** consistent accept or clear error; prefer accept street as military form even with civilian city  
- **Falsifiable test:** `TestPunchlist_PL017_UnitBoxCivilianCity`  
- **Status:** **done** (minion wave 2026-07-21)

### PL-018: Mid-line hash secondary not reordered

- **Severity:** major  
- **Sources:** A3-55  
- **Repro:** `100 #12 Main Street`  
- **Actual:** `# 12` absorbed into street name  
- **Expected:** primary `100`, secondary `#`/`12`, name `MAIN ST`  
- **Falsifiable test:** `TestPunchlist_PL018_MidLineHashSecondary`  
- **Status:** open

---

## Minor / GAP

### PL-019: Hyphenated compound directional not merged

- **Severity:** minor · **Class:** GAP  
- **Sources:** A2-012  
- **Repro:** `NORTH-EAST MAIN STREET`  
- **Expected (C#):** `NE MAIN ST`  
- **Falsifiable test:** `TestPunchlist_PL019_HyphenatedDirectional`  
- **Status:** open

### PL-020: Apostrophe not stripped

- **Severity:** minor · **Class:** GAP  
- **Sources:** A2-013  
- **Repro:** `West Main' rd`  
- **Expected (C#):** `W MAIN RD`  
- **Falsifiable test:** `TestPunchlist_PL020_ApostropheStrip`  
- **Status:** open

### PL-021: Hyphenated street token blocks suffix (`Main-Street`)

- **Severity:** minor  
- **Sources:** A3-58  
- **Repro:** `100 Main-Street`  
- **Actual:** suffix not peeled  
- **Falsifiable test:** `TestPunchlist_PL021_HyphenatedStreetToken`  
- **Status:** open

### PL-022: KEY peeled as secondary; PRAIRIE typo in tables

- **Severity:** minor  
- **Sources:** A2-003  
- **Repro:** state/street cases involving KEY / PRAIRIE  
- **Falsifiable test:** `TestPunchlist_PL022_KeyAndPrairie`  
- **Status:** open

### PL-023: Apt/Suite as street name fail

- **Severity:** minor  
- **Sources:** A1-008  
- **Repro:** streets literally named like suite patterns without numbers  
- **Falsifiable test:** `TestPunchlist_PL023_AptAsStreetName`  
- **Status:** **done** (already GREEN — "Suite Dreams Lane" style case passes; keep as regression)

### PL-024: POB alias for PO Box

- **Severity:** minor  
- **Sources:** A1-009  
- **Repro:** `POB 11890`  
- **Expected:** `PO BOX 11890`  
- **Falsifiable test:** `TestPunchlist_PL024_POBAlias`  
- **Status:** open

---

## Explicit non-goals (do not file as bugs)

- Exhaustive C# `StreetLineNormalizerTests` 100% match without a dedicated parity project  
- Geocoder-backed numeric street name resolution  
- Inventing structure when Parse correctly returns an error on empty/ambiguous input  

## QA team artifacts

| Agent | Focus | Report |
|-------|-------|--------|
| 1 E2E | Happy path + realistic failures | `.superpowers/sdd/qa-agent1-e2e-report.md` |
| 2 C# parity | 98-case street battery | `.superpowers/sdd/qa-agent2-csharp-parity-report.md` |
| 3 Adversarial | Edges, panics, silent wrong | `.superpowers/sdd/qa-agent3-adversarial-report.md` |

## Definition of done for this punchlist

- [ ] All **blocker** items GREEN under `go test -tags punchlist`
- [ ] All **major** items GREEN or explicitly deferred with issue link
- [ ] Main suite `go test ./...` still green
- [ ] README claims match implemented behavior
