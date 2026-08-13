# Go Project US@

This module implements the
[Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
(which is an extension of the USPS publication 28 standard) for US Address
Normalization directly in Go.

## License

This project is licensed under the Apache License 2.0 - see the
[LICENSE](LICENSE.md) file for details.

## Contributing and extending

This is a reference implementation, intended to be read, extended, and plugged
into other systems. [CONTRIBUTING.md](CONTRIBUTING.md) documents the development
guidelines that apply to everyone working on it — design principles, how the
package layout expresses the architecture, documentation and specification
traceability expectations, testing and patient-data rules, and guidance for
organizations extending the library downstream. Please read it before opening a
pull request.

## This implementation is currently incomplete and its API should be considered unstable

This implementation is incomplete and has not reached release status. It is
being built in public to facilitate feedback from potential users and support
from potential backers.

**What works today:** structured `Address` normalization (`Normalize` /
`NormalizeWithOptions`), street/last-line formatting, and component packages
under `pkg/` (regions, street suffixes, directionals, secondary units,
diacritics, highways, text helpers).

**Not yet:** free-text multi-line address parsing (`Parse`), and root-level
orchestration of Puerto Rico or military address flows. Use `pkg/puertorico` and
`pkg/military` directly for those vocabularies and line helpers until they are
wired into the root pipeline.

**Prefer content form** when writing patient records; `Normalize()` accepts
`USAtNormalizeOption`s - `WithContentNormalization()` is effectively the
default, but is useful for explicitly declaring your intent to use content
normalization. `WithMatchingNormalization()` when preparing addresses for
matching/exchange use cases (see Options below).

## Normalization of Input vs. Normalization for Comparison

The
[Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
specifically sets out to "enhance performance of patient matching algorithms
through improved address quality." Note that this purpose is focused on how
address data is standardized as it is collected and stored. This shows up in
certain details of the specification (for instance, "Numeric street names, for
example, 7TH ST or SEVENTH ST, MUST be conveyed exactly as it appears in the
patient’s official identification (government issued or insurance card)," a task
that is impossible during many patient matching operations.)

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
abbreviation while another does not - resulting in a form of information loss.
The exchange case is therefore required to treat both primary suffixes as
potentially matching. (Avoiding information loss would require distinct
abbreviations for each primary suffix, but, since they do not exist in the
specification, this implementation cannot create its own distinct standard
abbreviations for these cases.)

### Options (content vs exchange)

**`Normalize(string, ...USAtNormalizeOption)`** is the primary interface this
library exposes. It defaults to using "content form" normalization options that
are appropriate for storage: an exact controlled vocabulary, diacritic
preservation, and secondary designators as their standard abbreviations (`APT`,
`STE`, …). The `WithContentNormalization()` option allows a caller to explicitly
declare the intent to use these options and is strongly recommended for address
intake and storage scenarios.

The **`WithMatchingNormalization()`** option allows `Normalize()` to be used for
matching (aka exchange) scenarios. It allows:

- `Fuzzy` — mild typos on region and street suffix via package `Fuzzy*` helpers
- `SecondaryAsHash` — rewrite secondary designators to `#` for comparison only
  (not correct for content storage)
- `DiacriticMode` — an enumeration with 3 possible values:
  - `diacritics.KeepDiacritics` does not change diacritics
  - `diacritics.SubstituteDiacritics` substitutes diacritics acording to the
  ProjectUS@ specification
  - `diacritics.TransliterateDiacritics` transliterates diacritics, replacing
  them with ASCII letter combinations that commonly represent them.

The matching options provide a normalization that can match when the correct
secondary designator was not known by one of the parties or minor text variation
is detected. More sophisticated workflows allow these options and others to be
set individually.

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
