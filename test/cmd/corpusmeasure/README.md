# Measuring a parser against the reference corpora

`parsertest` (`pkg/address/parser/parsertest`) carries three corpora of
address test cases — `SpecCases`, `HistoricalCases`, and `OursCases` — plus a
runner, `Run`, that scores any `parser.ParsingFunc` against them. This
directory is a small cli command that runs all three suites and prints
pass/settled/total counts, plus every failing case, for one parser.

## Running

```sh
go run ./test/cmd/corpusmeasure
```

This defaults to `zipcityembedded`, the parser go-projectusat uses by
default.

## Testing your own parser implementation

`parsertest.Run` only needs a `parser.ParsingFunc` — anything with a
`Parse(source string) (*address.Address, error)` method, or a plain func
wrapped in `parser.ParsingFn`. To measure a different implementation, change
`newParserUnderTest` in `corpusmeasure.go` to return it instead, e.g.:

```go
func newParserUnderTest() (parser.ParsingFunc, error) {
	return parser.ParsingFn(func(source string) (*address.Address, error) {
		// your implementation
	}), nil
}
```

This is the same corpus and runner `addressparsers` scores itself against
(`internal/probe/refcorpusmeasure.go` there), so a number reported here and a
number reported there are directly comparable.
