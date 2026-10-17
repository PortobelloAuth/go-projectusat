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

## v0.1.0 Alpha Status

This library is currently in Alpha release status before version v0.1.0. We're
being deliberately conservative about versioning to give ourselves room to adapt
the API to necessary feedback.

However, `goprojectusat` appears to be _fully functional_ according to all of
our internal testing. We need testing and feedback from potential users to
determine if there are additional issues to address. This project is open source
and built in public to facilitate feedback from potential users and support from
potential backers.

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

## API

Package documentation for the latest version is available at
[https://pkg.go.dev/github.com/PortobelloAuth/go-projectusat](https://pkg.go.dev/github.com/PortobelloAuth/go-projectusat).
A less comprehensive introduction is also included here.

### New()

```golang
func New(opts ...goprojectusat.Option) (*goprojectusat.USAt, error)
```

`goprojectusat.New(opts ...goprojectusat.Option)` is the primary interface this
library exposes. It allow the caller to instantiate a `*goprojectusat.USAt`
instance that can be used for parsing and normalization, applying the rules
defined by the provided `*goprojectusat.Option`s.

`*goprojectusat.Option`s are applied in order with the last `Option` applied
taking precedence over earlier options.

Creating a new `goprojectusat.USAt` instance usually looks something like this:

```golang
usat, err := goprojectusat.New(
  goprojectusat.WithDiacriticNormalization(diacritics.SubstituteDiacritics),
  goprojectusat.WithSecondaryAsHash(),
)
if err != nil {
  return nil, err
}
```

### Options (content vs exchange)

**`USAt.Normalize(string)`** is the most comprehensive method on a `USAt`
instance. It takes a string, parses it as an address, normalizes it, and
reformats it all according to the
[Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
Specification. It excercises the every aspect of `goprojectusat` and every
`projectusat.Option` impacts how it works in some way.

`USAt.Normalize(string)` defaults to using "content form" normalization options
that are appropriate for storage: an exact controlled vocabulary, diacritic
preservation, and secondary designators as their standard abbreviations (`APT`,
`STE`, …). The **`WithContentNormalization()`** `projectusat.Option` allows a
caller to explicitly declare the intent to use these options. It is strongly
recommended for address intake and storage scenarios.

The **`WithMatchingNormalization()`** `projectusat.Option` allows
`USAt.Normalize()` to be used for matching (aka exchange) scenarios. It enables:

- `Fuzzy` — mild typos on region and street suffix via package `Fuzzy*` helpers
- `SecondaryAsHash` — rewrite secondary designators to `#` for comparison only
  (not correct for content storage)
- `DiacriticMode` set to `diacritics.SubstituteDiacritics` to substitute
  diacritics according to the ProjectUS@ specification

The matching options provide a normalization that can match when the correct
secondary designator was not known by one of the parties or minor text variation
is detected. More sophisticated workflows allow these options and others to be
set individually.

These 2 high-level `goprojectusat.Options` actualy set 3 values each, rather
than just 1 as most other `goprojectusat.Options` do.

#### WithContentNormalization

`WithContentNormalization()` explicitly sets the options recommended for
"content form" normalization. These options that are appropriate for address
intake and storage: an exact controlled vocabulary, diacritic preservation, and
secondary designators as their standard abbreviations (`APT`, `STE`, …). These
same options are used by default. However, using this Option clearly indicates
the caller's intent to use content normalization.

#### WithMatchingNormalization

`WithMatchingNormalization()` sets the options recommended for "exchange form",
also known as matching normalization. These options that are appropriate for
address matching: fuzzy matching of certain vocabularies, diacritic
substitution, and numbered secondary designators converted to `"#"`.

#### `WithFuzzyNormalization()`

Some of `goprojectusat`'s normalization functions - particularly
`streetsuffixes` and `region` - have support for fuzzy matching using
Damerau-Levenshtein edit distance. `WithFuzzyNormalization()` allows this
setting to be enabled.

#### `WithSecondaryAsHash()`

[Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
allows the secondary designator of an address to be `"#"` when it is not known.
As a result, any address with a numbered secondary designator may also be
recorded with a `"#"` (unless the address also contains a Private Mailbox.) For
a matching / exchange use case, normalization must resolve to a single value in
order to match. The `WithSecondaryAsHash()` option normalizes a numbered
secondary designator to `"#"` to facilitate such matching.

Please note `WithSecondaryAsHash()` currently only effects the first secondary
designator parsed. Addresses with multiple secondary designators - such as a
building number and a street - may not be as easily matched.

#### `WithDiacriticNormalization()`

[Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)
allows diacritics to be retained or substituted according to a table it
supplies. `WithDiacriticNormalization()` allows the caller to specify how they
would like the parser and normalizer to handle diacritic substitution.

`DiacriticMode` is an enumeration with 3 possible values:

- `diacritics.KeepDiacritics` does not change diacritics
- `diacritics.SubstituteDiacritics` substitutes diacritics according to the
  ProjectUS@ specification
- `diacritics.TransliterateDiacritics` transliterates diacritics, replacing them
  with ASCII letter combinations that commonly represent them.

#### `WithParsedAddressVerifier()`

`WithParsedAddressVerifier()` provides a `parser.AddessVerifier` function to a
USAt instance's internal parser. This function is called after an address has
been parsed by the internal parser (which may be a custom address parser or
`parser.ParsingFunc`) and before the internal address normalizer.

```golang
import (
  "github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
)

usat, err := goprojectusat.New(
  // parser.IdentityVerifier simply returns the address it recieves as
  // a parameter. It is the default implementation. It is used here
  // as a simple illustration of the API.
  goprojectusat.WithParsedAddressVerifier(parser.IdentityVerifier),
)
if err != nil {
  return nil, err
}

return usat, nil
```

#### `WithCustomAddressParser()`

`WithCustomAddressParser()` allows passing in a `parser.ParsingFunc` to replace
the core parsing function of a USAt instance. By default `goprojectusat.USAt`
relies on a parser that uses
[zipcity](https://github.com/poetic-systems/zipcity) to disambiguate between
streets, cites, and regions. `zipcity`'s bloom filters are an efficient method
for doing this without calling an API or running a database. However, they may
not serve every use case. WithCustomAddressParser lets a caller use their own
data store, API, or other custom parsing logic.

```golang
import (
  "github.com/PortobelloAuth/go-projectusat/pkg/address/parser/zipcityembedded"
)

zc, err := zipcityembedded.New()
if err != nil {
  return nil, err
}

usat, err := goprojectusat.New(
  // zipcityembedded.New() returns a parser wrapping a single instance of
  // a parser backed by zipcity's embedded bloom filters. It is the default
  // implementation. It is used here as a simple illustration of the API.
  goprojectusat.WithCustomAddressParser(zc),
)
if err != nil {
  return nil, err
}

return usat, nil
```

A custom parsing function should be tested using a copy of
[`test/cmd/corpusmeasure/corpusmeasure.go`](./test/cmd/corpusmeasure/corpusmeasure.go).
This will give you an idea how well your parser is working against our corpus of
test addresses. See ["Testing a parser"](./test/cmd/README.md#testing-a-parser)
for more information on this process.

#### `WithSingleLineFormatting()`

`WithSingleLineFormatting()` specifies that addresses should be formatted
without newlines.

#### `WithDiacriticFormatting()`

`WithDiacriticFormatting()` allows formatting an address that has been parsed
keeping diacritics in a format with diacritics substituted or transliterated,
without changing the source address data.

## Where the standards are insufficient

Project US@ and Pub 28 are silent, ambiguous or self-contradictory in places.
[docs/spec-gaps.md](docs/spec-gaps.md) records each one and the decision this
implementation made.

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
  be a great way to tap in to that community. A parser that does so lives in its
  own repository and plugs in through `ParsingFunc`.
- [Boostport address](https://github.com/Boostport/address) does address
  validation based on Google's published rules rather than normalization. It
  might pair well with a normalization library.
- [Pelias](https://pelias.io/) and their
  [docker](https://github.com/pelias/docker/) deployment could be interesting if
  you wanted to geocode two addresses and compare the location information or
  geocode an address and then reverse-geocode it to potentially get more
  information about it.

### Address Verification Services

There are many different address verification services that support US
addresses. Examples include Smarty, Melissa, or USPS' own service. These
generally license USPS' deliverability data; only the USPS can tell you if it
will deliver to a given address. The licenses on this data generally do not
allow it to be made available in an open source product and services built with
it must meet strict requirements to ensure that the data is not misappropriated.

One primary use case for `go-projectusat` is to serve as the address
normalization component for `go-opm-core` (a privacy preserving entity matching
library) and similar person or patient matching technologies. While per query
network latency and per query monetary costs might be reasonable, calling an
external service for address verification would create an additional entity with
access logs containing exactly the personally identifiable information and
personal health information `go-opm-core` is intended to keep from being
broadcast to parties that do not already have it.

As documented in the sections on `WithParsedAddressVerifier()` and
`WithCustomAddressParser()`, users who wish to do so may incorporate an external
address verification service or data they have licensed from USPS in a custom
verifier - run after parsing - or parser. Portobello Health Systems would be
happy to provide service and support for such an implementation.
