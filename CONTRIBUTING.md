# Contributing to Go Project US@

This project is a **reference implementation**. That word carries obligations
beyond "working code": the source is meant to be read by people deciding how
Project US@ should behave, extended by organizations with vocabularies and
dialects we have not anticipated, and trusted by systems handling patient data.

The guidelines below exist so that a stranger can read this codebase, form a
correct mental model of it in an afternoon, and extend it without asking us
first.

This document has two parts:

- [Part 1 — Contributing upstream](#part-1--contributing-upstream), for changes
  you intend to land in this repository.
- [Part 2 — Extending downstream](#part-2--extending-downstream), for
  organizations plugging this library into their own systems.

Several sections below are marked **(evolving)**. Those reflect areas where the
project has a working position but expects to tighten it after further internal
discussion. Follow the stated intent; expect the specifics to sharpen.

---

## Part 1 — Contributing upstream

### 1. Design principles

These are ordered. When two of them pull in opposite directions, the earlier one
usually wins.

#### 1.1 Keep it simple (KISS)

Prefer the simplest model that is actually correct for the specification. A
simple rule that a reader can hold in their head is worth more than a clever one
that handles a case the standard never asks about.

The temptation in address parsing is to reach for an algorithm when a
description will do. If the specification says a value is "everything before the
address at the end," implement that — do not build a merge algorithm to
rediscover it.

Complexity that the specification genuinely requires is welcome. Complexity that
anticipates a requirement nobody has stated is not.

#### 1.2 Don't repeat yourself (DRY)

Duplicated knowledge is the failure mode this project is most prone to, because
so much of it is vocabulary data: suffixes, directionals, regions, secondary
designators. A second copy of a lookup table is not a second convenience, it is
a second source of truth that will drift from the first.

Derive rather than duplicate. The established pattern here is a single
authoritative slice of `[]XInfo` records, with lookup maps built from it at
package level:

```go
// One authoritative table.
var regions = []RegionInfo{ /* ... */ }

// Maps derived from it, not maintained alongside it.
var byAbbreviation = maps.Collect(/* iterate regions */)
```

Note that these derived values are package-level variables, **not** built in an
`init()` function. This codebase does not use `init()`; derived state should be
visibly derived at its declaration, where a reader will find it.

DRY applies to knowledge, not to characters. Two functions that look alike but
encode different rules from different sections of the standard are not
duplication, and merging them makes the code harder to check against the spec.

#### 1.3 Encapsulation: organize by concern, gather responsibility

The package layout is the architecture. `pkg/` is divided by address-part
concern — `region`, `streetsuffixes`, `directionals`, `secondaryunit`,
`highways`, `pobox`, `ruralroute`, `military`, `puertorico`, `postalcode`,
`country`, `diacritics`, `textutil` — and each of those packages owns everything
about its concern.

The practical test: **code that parses a thing belongs with the code that
normalizes that thing.** Region parsing lives with region normalization. If you
find yourself adding region knowledge to the parser, or street-suffix knowledge
to the normalizer, the abstraction has leaked and the fix is to move the
knowledge, not to duplicate it.

Two structural rules follow from this:

- **The repository root is the access surface.** `goprojectusat.go` is the file
  through which other functionality is generally accessed. It orchestrates; it
  does not implement. Do not add new files at the root — new capability belongs
  in a package under `pkg/`, exposed at the root only if it is part of the
  public entry point.
- **Cross-cutting behavior gets an interface, not a conditional.** A dialect or
  variant (Puerto Rico addressing, military addressing) is expressed as a
  high-level interface with its own implementation, not as `if`/`else` branches
  scattered through each section of the pipeline. The moment a variant needs its
  second branch in a second place, it needs an interface instead.

#### 1.4 Names carry the design

A name should say what the code does, in the vocabulary of the problem domain.
`extractStreet` and `remove` say what happens; `peel` requires the reader to
reverse-engineer an intention.

Where the specification supplies a term, use the specification's term. A reader
holding the standard in one hand and this code in the other should not have to
maintain a translation table.

#### 1.5 Context over position

Address components are not reliably identified by where they sit in a string.
Where a token appears *in relation to other token types* often carries more
information about its meaning than its raw position does. Prefer designs that
score and compare interpretations over designs that assume a fixed layout, and
where the input is genuinely ambiguous, prefer surfacing possible
interpretations to silently choosing one.

### 2. Documentation and intent

Documentation is how intent survives the person who had it. Working code
communicates *what* happens; it very rarely communicates *why this and not the
obvious alternative* — and in a spec implementation, the why is the valuable
part.

The standard to meet: **every new piece of code should communicate its concept
easily, through the code itself or through a comment.** Prefer the code. Reach
for a comment when the code cannot carry the meaning on its own.

Concretely:

- Every exported identifier gets a doc comment, in standard Go form, starting
  with the identifier's name.
- Every package gets a package comment explaining what concern it owns.
- **Citing the specification section a rule implements is encouraged**, and
  strongly encouraged for any rule whose motivation is not self-evident. It is
  not a hard requirement, but a block comment quoting the relevant text is the
  single highest-value comment you can write in this repository — it lets the
  next reader verify correctness without hunting through a PDF.
- **Deviations from the specification must be documented.** This is a hard
  requirement. Anywhere the implementation knowingly departs from the standard,
  resolves an ambiguity in it, or must choose between two of its goals that
  conflict, write down what you chose and why. Put it in a comment at the site,
  and if it changes what a user of the library should expect, in the README as
  well. An undocumented deviation in a reference implementation is a defect even
  when the behavior is right.
- Comments explain intent, tradeoffs, and specification linkage. They should not
  restate what the code plainly says.

### 3. Testing

- **Every behavior change ships with tests.** Whether you write the test first
  is your call; the project does not mandate an order.
- **Cover the specification's own examples where you reasonably can.** The
  examples in the standard are the closest thing this project has to an
  acceptance suite, and they are free — when you implement a rule, the examples
  the standard gives for that rule are the natural test cases. **(evolving —**
  the project may formalize how completely specification examples must be
  covered; today this is a strong encouragement, not a gate.**)**
- Follow the testing conventions already present in the tree rather than
  inventing new ones: tests live in the `_test` package, cross-package tests
  live under `test/multipkg_test/`, and a test's purpose is marked in its name
  and failure message rather than in a special-purpose filename.
- **Test data must be synthetic** — invented, or drawn from the public
  specification. Never commit real patient data, and never commit
  de-identified or scrubbed real records. See [§5](#5-handling-patient-data).

### 4. Changes should be small and readable

Optimize for the reviewer, and for the person doing archaeology on this commit
two years from now.

- **One concern per change.** A pull request should be describable in a single
  sentence without the word "and." If you cannot, it is two pull requests.
- Refactoring and behavior change belong in separate commits, ideally separate
  pull requests. A diff that moves code *and* changes what it does is very hard
  to review and very easy to approve incorrectly.
- **Commit subjects are plain and imperative** — "Add pobox normalization",
  "Fix decimal style street names". This project does not use
  conventional-commit prefixes.
- Keep the tree clean and reasonable: formatted with `gofmt`, free of `go vet`
  complaints, free of leftover debug output, and building against the toolchain
  in `go.mod`. **(evolving —** the project expects to formalize CI requirements
  and merge criteria; until then, apply the obvious professional standard.**)**
- **Open an issue or discussion before large changes** — a new package, a change
  to the public API surface, or a new approach to parsing. This is not
  gatekeeping; it is so outside contributors do not invest a weekend in a design
  the maintainers have already ruled out. Small, self-evident fixes need no
  preamble.

### 5. Handling patient data

Addresses processed by this library are protected health information in most of
the contexts this project is built for. Treat every value flowing through it as
sensitive by default, and prefer designs that make it hard for a downstream
integrator to leak that data by accident.

The one firm rule today: **test fixtures and examples must use synthetic data
only** — invented addresses, or addresses published in the specification itself.
Real records are not acceptable, including de-identified ones.

**(evolving)** Further requirements — around logging of input values, network
egress and telemetry in the core packages, and dependency review — are under
internal discussion and will be added here. In the meantime, assume the strict
reading: do not log or print address values, and treat anything that leaves the
process as a decision that needs justification in the pull request.

---

## Part 2 — Extending downstream

This library is meant to go into organizations and be extended. If you are
adding a dialect, a vocabulary, a parser, or a matching strategy for your own
deployment, this part is for you.

### 1. Extension philosophy

The project's position is that extension should happen through **composition and
data**, not through forking or through growing the API surface one function at a
time. Two habits follow from that, and they are worth internalizing before you
reach for a new exported function:

**Extend behavior through options, not through new entry points.** The public
surface uses the functional-options pattern — `Normalize(source string, opts
...USAtNormalizeOption)` — precisely so that new behavior does not require a new
`NormalizeSomethingElse()`. If your change would add a function that differs
from an existing one by a flag, it probably wants to be an option.

**Extend vocabulary through data, not through lookups.** When a component needs
to know something new about a region, a suffix, or a unit designator, the
answer is usually a new field on that concern's `Info` record — with the lookup
maps derived from it as described in
[Part 1 §1.2](#12-dont-repeat-yourself-dry) —
rather than a new standalone lookup function.

**Plug points are interfaces with function adapters.** Where the library invites
you to substitute behavior, it does so with an interface plus a function-typed
adapter, following the `http.Handler` / `http.HandlerFunc` idiom. `ParsingFunc`
and `ParsingFn` are the existing example: implement the interface when you have
state, pass a function when you do not. `pkg/address/parser/libpostalhttp` is a
worked example of an external parser plugged in this way — and of the project's
preference for keeping anything that reaches outside the process in its own
clearly-fenced package.

More specific extension-design guidance, including a fuller account of which
seams are intended to be extended, is expected to follow. **(evolving)** In the
meantime, the direction to reason from: prefer options over new functions,
prefer data over lookups, prefer interfaces over branches — and if the extension
you need does not fit those, open a discussion. That is a signal the library is
missing a seam, which is a better bug for us to hear about than to have you work
around.

### 2. API stability

This implementation has not reached release status and its API should be
considered unstable; see the README for what works today and what does not. If
you are building on it now, pin a commit and expect to read the diff on upgrade.

---

## Questions

Open an issue. This project is being built in public specifically to get
feedback from the people who intend to use it, and a question that exposes an
unclear design is a contribution.
