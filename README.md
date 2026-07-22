# Go Project US@

This module implements the
[Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
(which is an extension of the USPS publication 28 standard) for US Address
Normalization directly in Go.

## This implementation is currently incomplete and its API should be considered unstable.

This implementation is incomplete and has not reached release status. It is
being built in public to facilitate feedback from potential users and support
from potential backers.

**What works today:** structured `Address` normalization (`Normalize` /
`NormalizeWithOptions`), free-text `Parse` covering the main Project US@ / C# parity
workflows: multi-line and comma-separated input; overseas military; rural route and PO Box;
multi-token directionals; leading secondary / `#` reorder; double-suffix and state-as-street-name;
same-line business and narrative pre-street prefixes; grid-style double directionals;
fractional and hyphenated primaries; multi-secondary units; directional-as-name and
Avenue-letter edges; Puerto Rico Spanish street types and secondaries (including Apartado),
street/last-line formatting, and component packages under `pkg/`.

**Not yet:** exhaustive C# `StreetLineNormalizerTests` parity for every punctuation/
narrative edge, geocoder-backed disambiguation, and some rare dual-interpretation cases.
The API remains unstable.

**Prefer content form** (`Normalize`) when writing patient records; use
`NormalizeWithOptions` when preparing addresses for match/exchange (see Options
below). Use `Parse` when the input is a free-text multi-line or comma-separated
address string, then `Normalize` for content form.

### Parse (free-text)

```go
addr, err := goprojectusat.Parse("123 North Main Street Apt 4\nSpringfield IL 62701")
if err != nil {
	// handle
}
norm, err := goprojectusat.Normalize(addr)
// Format(norm) =>
// 123 N MAIN ST APT 4
// SPRINGFIELD IL 62701
```

Overseas military:

```go
addr, err := goprojectusat.Parse("PSC 3 BOX 4120\nAPO AE 09021-0002")
norm, err := goprojectusat.Normalize(addr)
// Format(norm) =>
// PSC 3 BOX 4120
// APO AE 09021-0002
```

Rural route and PO Box (stored wholly in `StreetName`, like military):

```go
addr, err := goprojectusat.Parse("Rural Route 91 Box A7\nSpringfield IL 62701")
// Format(Normalize(addr)) =>
// RR 91 BOX A7
// SPRINGFIELD IL 62701

addr, err = goprojectusat.Parse("PO Box 11890\nSpringfield IL 62701")
// Format(Normalize(addr)) =>
// PO BOX 11890
// SPRINGFIELD IL 62701
```

`Parse` is conservative: it splits on newlines and commas, peels last-line
city/region/postal, rewrites rural route / PO Box (and PR `Apartado`) street lines,
merges multi-token directionals (`SOUTH WEST` → `SW`), reorders leading secondary / `#`,
extracts same-line business/narrative prefixes before the house number, reverse-token
peels street components (including multi-secondary `BLDG 420 RM 120`), handles grid
directionals (`1016 E 1700 S`), fractions and hyphenated primaries, and uses a military
fast path for APO/FPO/DPO. When region is `PR`, Spanish street types (`CALLE`→`CLL`) and
secondaries (`URB`, `APARTAMENTO`) apply. It does not call `Normalize`; compose
`Normalize(Parse(raw))` for content form. Ambiguous input returns an error rather than
inventing structure.

Puerto Rico:

```go
addr, err := goprojectusat.Parse("Calle Luna 123\nSan Juan PR 00901")
// Format(Normalize(addr)) =>
// 123 LUNA CLL
// SAN JUAN PR 00901
```

Same-line business / narrative:

```go
addr, err := goprojectusat.Parse("Williamson Medical Center 3000 Edward Curd Lane\nSpringfield IL 62701")
// BusinessName WILLIAMSON MEDICAL CENTER; street 3000 EDWARD CURD LN
```

## Normalization of Input vs. Normalization for Comparison

The
[Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
specifically sets out to "enhance performance of patient matching algorithms
through improved address quality." Note that this purpose is focused on how
address data is standardized as it is collected and stored. This shows up in
certain details of the specification (for instance, "Numeric street names, for
example, 7TH ST or SEVENTH ST, MUST be conveyed exactly as it appears in the
patient’s official identification (government issued or insurance card)," a
task that is impossible during many patient matching operations.)

However, the specification also includes in its intended audience MPI and eMPI
vendors, government agencies, data scientists, and others whose application of
the standard would necessarily be focused on comparing address data without
access to a patient or source documents. The specification refers to these two
distinct scenarios as "content" and "exchange" and encourages systems "to
standardize patient address information according to the specification before
exchange and matching in such a way that limits information loss."

Where this implementation recognizes a difference in the needs of the content
and exchange use cases it will attempt to provide distinct interfaces that are
appropriate for each. In some cases, particularly noted in the street suffixes
specification, the specifications goals may come in conflict with each other.
Some plural and singular versions of a street suffix share the same abbreviation
but are considered distinct "primary" suffixes. Since the storage of a suffix by
its primary representation or abbreviation is left up to an implementation
(based on the space allocated for storage), this will necessarily lead to
instances where one implementation stores the address with the suffix
abbreviation while another does not - resulting in a form of infomration loss.
The exchange case is therefore required to treat both primary suffixes as
potentially matching. (Avoiding information loss would require distinct
abbreviations for each primary suffix, but, since they do not exist in the
specification, this implementation cannot create its own distinct standard
abbreviations for these cases.)

### Options (content vs exchange)

- **`Normalize`** — content form for storage: exact controlled vocabulary, preserves
  diacritics, keeps secondary designators as standard abbreviations (`APT`, `STE`, …).
- **`NormalizeWithOptions`** — same pipeline with exchange/matching knobs:
  - `Fuzzy` — mild typos on region and street suffix via package `Fuzzy*` helpers
  - `SecondaryAsHash` — rewrite secondary designators to `#` for comparison only
    (not correct for content storage)
  - `DiacriticMode` — `""` leave as-is; `"substitute"` / `"transliterate"` strip or
    map diacritics on free-text fields (then uppercased again)

Prefer content form when writing patient records; use options when preparing
addresses for match/exchange.

## Alternatives and related technologies

There are several related technologies that could be used instead of or in
conjunction with Go Project US@.

- [ProjectUsNormalizer](https://github.com/ica-carealign/project-us-normalizer):
  "C# library to facilitate address normalization in accordance with Project
  US@, the "Unified Specification for Addresses in Health Care." Portobello
  couldn't use this directly since we want a Go implementation.
- [golang-address](https://github.com/kminehart/golang-address) is 10 years old
  and recommends using [gopostal](https://github.com/openvenues/gopostal). It
  also only parses the street line of adresses.
- [gopostal](https://github.com/openvenues/gopostal) uses machine learning to
  select an appropriate parser for international addresses. It is built on
  [libpostal](https://github.com/openvenues/libpostal) and requires that library
  to be installed. libpostal is broadly deployed and used by several projects.
  Employing it through
  [libpostal-rest](https://github.com/johnlonganecker/libpostal-rest) on a
  [docker image](https://github.com/johnlonganecker/libpostal-rest-docker) might
  be a great way to tap in to that community.
- [Boostport address](https://github.com/Boostport/address) does address
  validation rather than normalization. It might pair well with a normalization
  library.
- [Pelias](https://pelias.io/) and their
  [docker](https://github.com/pelias/docker/) deployment could be interesting if
  you wanted to geocode two addresses and compare the location information or
  geocode an address and then reverse-geocode it to potentially get more
  information about it.

Ultimately, Portobello wanted a native Go implementation of the
[Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
available that didn't involve any external library installs or calls to other
services. It isn't intended to do anything more or less than that.
