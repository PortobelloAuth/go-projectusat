# zipcity_checkout

A parser that uses embedded zipcity data

## Usage

As with other parsing functions, this plugs in through `parser.AddressParsingOptions`:

```go
type AddressParsingOptions struct {
  Verifier     AddressVerifier // func(*address.Address) (*address.Address, error)
  CustomParser ParsingFunc     // Parse(source string) (*address.Address, error)
}
```

## Packages

### `parse`

The parser. It runs go-projectusat's pipeline — tokenize, let every vocabulary
claim what it recognizes, read the last line from those claims, ask each address
type for its reading of the whole — and then chooses among the readings,
consulting `zipcity`'s reference data when the grammar alone cannot settle one.

**Reference data moves a candidate's score, not its confidence.** The score is
this package's own integer — the grammar's rung plus one for every question
reference data agrees with and minus one for every question it contradicts,
uncapped in either direction (see `rung` and `score` in `parse/parse.go`).
`zipcity`'s filters are built at a 0.005 false positive rate (set in zipcity's
`internal/bloomgenerator`), so a `false` is definitive while a `true` is
roughly 200:1 evidence rather than a confirmation. Both are worth acting on,
and neither settles a reading on its own. The data orders readings against
each other and never rates one: `CandidateAddress.Confidence` is left exactly
as the grammar wrote it, and `Parse` returns the `Address` alone, so no score
this package computes is ever visible to a caller.

A candidate that carries a street name asks about it too — `CheckZipAndStreet`
where the candidate has a ZIP, `CheckCityStateAndStreet` where it has a city
and state, both when it has all three. Asking both when both are available and
folding the answers before either moves the score squares the odds of a
spurious step instead of doubling them: both `true` folds to a single
`agrees`, both `false` folds to a single `contradicts`, and a split — one
shard finds the street, the other does not — is recorded as which side was
absent (`missingInZip` or `missingInCity`) rather than resolved, and moves
nothing. `zipcity` also offers directional-variant matching for a street
(`MatchZipAndStreet`), and that remains deliberately unused: querying several
spellings of the same street multiplies the odds of a spurious `true` for no
better reason than asking more than once. See `agreement`, `streetAgreement`,
and `foldStreetAnswers` in `parse/parse.go`.
