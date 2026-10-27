# parsertest

`parsertest` ships the corpus of test cases (`Cases`) any parser implementation
can be scored against, and the runner (`Run`, `CountPass`, `CountPassFields`)
that scores it. This file documents the `UsAddressCases` subset and how its
`WantFields` values are produced. Those values are the cases most likely to need
review when they change.

## usaddress cases

`UsAddressCases` comes from the tagged-address corpora that
[datamade/usaddress](https://github.com/datamade/usaddress) vendors for its own
`measure_performance` suite. The corpora are pinned at commit `aa7699b5` and are
MIT licensed; see `testdata/usaddress/NOTICE`. Four files are used:

| File | Cases |
| --- | --- |
| `labeled.xml` | 146 |
| `multi_word_state_addresses.xml` | 5 |
| `simple_address_patterns.xml` | 8 |
| `us50_test_tagged.xml` | 687 |

Two other files live at the same upstream path and are left out:
`synthetic_osm_data.xml` and `synthetic_clean_osm_data.xml`. They are generated
by machine from OpenStreetMap rather than taken from real addresses, and together
they are larger than the four files above combined.

Each fixture record tags every token of an address string with a usaddress
component label. The labels are defined in usaddress's documentation:
<https://github.com/datamade/usaddress/blob/aa7699b53a0843fc443f9e87285b88cbd9eaf50a/docs/index.rst>.

### From usaddress tags to `address.Address`

Three steps turn a record into `WantFields`:

1. **Map.** Each tag's text goes into an `address.Address` field according to the
   table below. A component that spans several words is tagged once per word,
   so the words are appended in document order. usaddress keeps a trailing
   separator inside the tag ("Chicago,"), and that trailing `,` or `;` is removed.
   This is a tokenization fix only. No other normalization happens in this
   step.
2. **Normalize.** The raw address is passed through
   `normalizer.NewContentNomalizer().Normalize`, so `WantFields` holds the
   decomposition Project US@ says the address should have, not usaddress's
   literal text. If the result changes when it is normalized a second time,
   the case is left unsettled (see below).
3. **Refine.** The normalized address is passed through the full normalizer
   (`goprojectusat.New().NormalizeAddress`), which re-reads it with the parser and
   the ZIP's data, and the result is adopted when it is a refinement: usaddress
   often leaves the suffix or a directional inside `StreetName` ("MENDENHALL MALL
   ROAD"), and the full normalizer splits it out. The result is refused, and the
   content-normalized address kept, when it changes the primary number, region,
   postal code, secondary unit or business name, drops a directional or the city,
   or is not itself a fixed point. A refused case carries the reason and what
   the normalizer gave in `Case.Note` and in the generated file, and the header
   report counts them by reason. Those are the places the parser and the fixture
   disagree and someone should look. Note that the street-line split in an
   adopted result is the parser's own reading, so agreement there is by
   construction; the independent evidence is the fixture's tags for the number,
   city, region, postal code and directionals.

| usaddress tag | `address.Address` field |
| --- | --- |
| `BuildingName` | `BusinessName` |
| `AddressNumber`, `AddressNumberSuffix` | `PrimaryNumber` |
| `USPSBoxID` | `PrimaryNumber`, and `Type` becomes `*pobox.POBoxAddress` |
| `StreetNamePreDirectional` | `Predirectional` |
| `StreetNamePreModifier`, `StreetNamePreType`, `StreetName`, `StreetNamePostModifier` | `StreetName`, in that document order |
| `USPSBoxType` | `StreetName`, and `Type` becomes `*pobox.POBoxAddress` |
| `StreetNamePostType` | `StreetSuffix` |
| `StreetNamePostDirectional` | `Postdirectional`; `StreetName` when a `StreetNamePostType` follows it (Project US@ p.18, #277) |
| `OccupancyType`, `SubaddressType` | `SecondaryDesignator` |
| `OccupancyIdentifier`, `SubaddressIdentifier` | `SecondaryNumber` |
| `PlaceName` | `City` |
| `StateName` | `Region` |
| `ZipCode` | `Postal` |
| `CountryName` | `Country` |
| `Recipient` | dropped (see below) |

**`Recipient` is dropped.** Project US@ Technical Specification v1.0,
pp. 34–35, "Remove Certain Words", says: "Developers SHOULD replace or remove
certain words ... Remove ATTENTION, ATTN:". An attention or recipient line is not
part of the address, so it fills no field and the rest of the record is
decomposed as usual. A record made of nothing but a `Recipient` is left
unsettled.

**`StreetNamePreType` is part of `StreetName`.** usaddress uses this tag for the
route designator in front of a highway number, as in "US Hwy 50", "FM 544" and
"County Road 3668". `pkg/highways` claims the whole highway designation as a
street name, and the mapping agrees with it. The normalizer then expands the
designator ("US HIGHWAY 50").

### Cases left without `WantFields`

A case with nothing that can be asserted gets a nil `WantFields`, and the reason
is kept in the generated record and in `Case.Note`. A nil `WantFields` means
"not settled" and never counts as a parser failure (see `Case.WantFields`). A
case is left unsettled for any of these reasons:

- **A tag with no `Address` field.** The tags are `IntersectionSeparator` and
  `SecondStreetName` (intersections), `CornerOf`, `LandmarkName`, and
  `NotAddress`.
- **`USPSBoxGroupID` and `USPSBoxGroupType`** are rural route and highway
  contract boxes. Mapping them is **deferred to a follow-on change**. Until
  that change lands, these cases stay unsettled.
- **A tag the mapping does not know.** The generator fails closed rather than
  silently dropping the tag.
- **The normalizer rejects the raw address.** This happens when a controlled
  vocabulary (region, directionals, street suffix, secondary designator) does
  not recognize the value, or when the PO Box normalizer does not recognize the
  box street line.
- **A reviewed override.** `internal/usaddressgenerator/overrides.json` holds
  per-input overrides, each with a mandatory `reason`. With no `want` the case
  is left unsettled and the reason is reported; with a `want` (fields named as in
  `address.Address`, `poBox` for the PO Box type; every other record is an
  `OrdinaryStreetAddress`) that address replaces the
  normalizer's. Fix the normalizer first, since that fixes every input; an
  override is for a fixture no normalizer could be right about, such as a
  negative case or a recipient run that mixes a business and a person. The
  generator fails on an unknown field, a missing reason, a repeated input, or an
  input that matches no case.
- **The normalizer is not idempotent on the value.** If normalizing the
  normalized address again changes it, the record cannot tell which pass is
  right, so it asserts neither.

The header of `usaddress_generated.go` carries a report with per-file counts
and the number of cases left unsettled for each reason.
`TestUsAddressWantFields` enforces the settled total as a ratchet, and it checks
that every committed `WantFields` is already in normalized form.

## Regenerating

`usaddress_generated.go` is generated and committed, so any change to a
`WantFields` value shows up as a diff a reviewer reads in a pull request. It
cannot drift along with the code it checks. The generator is
`internal/usaddressgenerator/usaddressgenerator.go`. It is a `//go:build ignore`
`main` package, so it is never part of the build graph. Regenerate it after
changing the mapping, the fixtures, or a normalizer the mapping depends on. Run
either of these from the repository root:

```sh
go generate ./pkg/address/parser/parsertest/
# or
go run pkg/address/parser/parsertest/internal/usaddressgenerator/usaddressgenerator.go
```

Then run `go test ./pkg/address/parser/parsertest/`. If the settled count
changed, update `wantSettled` in `TestUsAddressWantFields`, and review the diff
in `usaddress_generated.go` before committing.
