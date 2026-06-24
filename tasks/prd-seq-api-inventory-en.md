# PRD: seq — A Chainable Generic Library Built on iter.Seq (API Inventory)

> Language: [简体中文](prd-seq-api-inventory.md) | English


## Introduction

`seq` is a Go generic library built around the standard library's lazy iterator `iter.Seq` / `iter.Seq2` (Go 1.23+), offering Scala-style, chainable collection operations enabled by **Go 1.27's generic methods feature** ([golang/go#77273](https://github.com/golang/go/issues/77273)).

The core problem it solves: today, libraries like `samber/lo` can only use top-level functions (`lo.Map(lo.Filter(...))`, read inside-out), because before Go 1.27, **methods could not declare their own type parameters** and could not express the chainable `.Map()` that takes `[]T` and returns `[]U`. Go 1.27 lifts that restriction, making a discoverable, chainable, lazy pipeline like `From(xs).Filter(...).Map(...).Collect()` possible for the first time.

This document **produces an API inventory only**: a complete listing of the methods on `Seq[T]` and `Seq2[K,V]`, the accompanying package-level (free) functions, and a clear statement of whether each API is a "method" or a "free function" and why. No implementation. The reader may be a junior developer or an AI agent, so each API gets a one-line semantic note.

### Key design constraint (determines the method vs. free-function split)

This is the central rule of the whole library; every split derives from it:

1. **A method's type parameters are fresh and independent; they cannot add a constraint onto the receiver's `T`.** `Seq[T any]` declares `T` as `any` at the type level, so any operation that requires **the element itself** to satisfy `comparable` / `cmp.Ordered` / a numeric constraint **cannot be a method**, only a free function. Examples: `Distinct` (needs `comparable`), `Max` (needs `Ordered`), `Sum` (needs `Numeric`).
2. **Escape hatch: a method's own type parameters may carry constraints.** So "by key" variants can return to method form. Examples: `DistinctBy[K comparable](key func(T) K)`, `GroupBy[K comparable]`, because `K` is the method's own parameter, not the receiver's `T`.
3. **Operations involving multiple generator types or a specific instantiated receiver can only be free functions.** Examples: `Zip[A, B](a Seq[A], b Seq[B])` (two independent types), `Flatten[T](s Seq[Seq[T]])` (the receiver must be the specific instantiation `Seq[Seq[T]]`, which a method cannot generically attach to).
4. **Generic methods do not satisfy interfaces** (a hard line of the proposal) — this library does not attempt to provide a polymorphic interface abstraction over collections.

### Known capability boundaries (relative to Scala)

- No for-comprehension syntax sugar: everything is a bare method chain.
- No HKT / type classes: cannot abstract Functor/Monad; every method is a concrete method on a concrete type.
- No universal equality: `==` is a constraint, not a universal capability, so `Distinct`/`Max` etc. must be free functions.
- No tuples: `Zip` and friends return a custom `Pair[A,B]` struct.

## Goals

- Produce a **complete, unambiguous** API inventory, each entry tagged with its classification (method / free function) plus a one-line semantic.
- Clarify the responsibility boundary and the conversion entry points between the two core types `Seq[T]` and `Seq2[K,V]`.
- Make the split rule mechanically verifiable: anything that constrains `T` itself must be a free function; anything that uses only the method's own constrained parameters may be a method.
- Cover `lo`'s common operations + the portable parts of Scala collections + a reasonable long tail.
- Provide a directly actionable API contract for the subsequent `/prd-to-spec` (technical design) and `/to-issues` (issue decomposition).

## User Stories

### US-001: Define core types and the Pair helper type
**Description:** As a library author, I need to define the `Seq[T]`, `Seq2[K,V]` defined types and `Pair[A,B]`, as the carrier of all APIs.

**Acceptance Criteria:**
- [ ] `type Seq[T any] iter.Seq[T]` defined, zero-cost convertible with `iter.Seq[T]`
- [ ] `type Seq2[K, V any] iter.Seq2[K, V]` defined, zero-cost convertible with `iter.Seq2[K,V]`
- [ ] `type Pair[A, B any] struct { Left A; Right B }` defined
- [ ] `type Numeric` constraint (integers + floats) defined, for use by numeric free functions
- [ ] `go build ./...` passes, `go vet` reports no warnings

### US-002: Constructor entry free functions (Seq[T])
**Description:** As a user, I need to create `Seq[T]` from sources like slices, variadic arguments, ranges, and channels to start a method chain.

**Acceptance Criteria:**
- [ ] Implement all constructor functions listed in FR-2, with signatures matching the docs
- [ ] `From([]int{1,2,3}).Collect()` returns `[1 2 3]`
- [ ] `Range(0, 3).Collect()` returns `[0 1 2]`
- [ ] Infinite sources (`Generate`/`Iterate`/infinite `Repeat`) with `Take` do not loop forever
- [ ] Unit tests cover each constructor, `go test ./...` passes

### US-003: Intermediate transformation methods (lazy, return Seq)
**Description:** As a user, I need lazy intermediate operations like Map/Filter/FlatMap that chain without materializing immediately.

**Acceptance Criteria:**
- [ ] Implement all intermediate methods listed in FR-3
- [ ] `Map[U]` can change the element type: `From([]int{1,2}).Map(strconv.Itoa).Collect()` returns `["1","2"]`
- [ ] Intermediate operations are lazy: the source is not traversed before a terminal operation (verify with a counter side effect)
- [ ] `TakeRight`/`DropRight`/`Slice`/`Init` and other materializing intermediates behave correctly
- [ ] Chains compose: `Filter(...).Map(...).Take(n)` compiles and is semantically correct
- [ ] Unit tests cover, `go test ./...` passes

### US-004: Terminal methods (drive evaluation, return non-Seq)
**Description:** As a user, I need terminal operations like Collect/Fold/Reduce/Find to drive the pipeline and extract results.

**Acceptance Criteria:**
- [ ] Implement all terminal methods listed in FR-4
- [ ] `Reduce` returns `(zero, false)` on an empty sequence
- [ ] `Find`/`First`/`Last`/`Nth` return `(v, true)` on hit, `(zero, false)` on miss
- [ ] `FindLast`/`FindLastIndex` return the last satisfying element and its index
- [ ] `CountBy` returns the count satisfying the predicate; `GroupCount` returns a `map[K]int` grouped by key
- [ ] `MaxByKey`/`MinByKey`/`SumBy`/`MeanBy` aggregate by a projection function
- [ ] `Any`/`All`/`None` short-circuit (stop on hit, verify with a counter)
- [ ] Unit tests cover empty-sequence boundaries, `go test ./...` passes

### US-005: Constrained free functions (comparable / Ordered / Numeric)
**Description:** As a user, I need operations like Distinct/Max/Sum/Sort that depend on element constraints; by the language rule they can only be free functions.

**Acceptance Criteria:**
- [ ] Implement all constrained free functions listed in FR-5
- [ ] `Distinct[T comparable]` deduplicates while preserving first-occurrence order
- [ ] `Compact` drops zero values; `Without` excludes specified values
- [ ] `SymmetricDifference` returns the symmetric difference (lodash xor semantics)
- [ ] `Max`/`Min`/`Sum`/`Product`/`Mean` return the agreed value on an empty sequence (see FR-5 notes)
- [ ] Each function signature carries the correct constraint (`comparable`/`cmp.Ordered`/`Numeric`), compiler-enforced
- [ ] Unit tests cover, `go test ./...` passes

### US-006: Multi-sequence / nested free functions (Zip / Flatten / Concat)
**Description:** As a user, I need to combine multiple sequences or flatten nested sequences; due to multiple type parameters or a specific instantiated receiver, these can only be free functions.

**Acceptance Criteria:**
- [ ] Implement all multi-sequence functions listed in FR-6
- [ ] `Zip[A,B]` stops when the shorter sequence is exhausted
- [ ] `ZipWith` produces `Seq[C]` via a combining function; `ZipMap` forms a `map[K]V`
- [ ] `Flatten[T](Seq[Seq[T]])` flattens one level correctly
- [ ] `Concat` joins multiple sequences in order
- [ ] Unit tests cover, `go test ./...` passes

### US-007: Seq2[K,V] methods and free functions
**Description:** As a user, I need the transformation and terminal operations of `Seq2[K,V]` in key-value / indexed / map scenarios.

**Acceptance Criteria:**
- [ ] Implement all `Seq2` methods and accompanying free functions listed in FR-7
- [ ] `Keys()`/`Values()` return `Seq[K]`/`Seq[V]`
- [ ] `ToMap[K comparable,V]` is a free function (because a map key needs `comparable`)
- [ ] `Associate[T,K comparable,V]` projects a `Seq[T]` into a `Seq2[K,V]`
- [ ] `FromMap`/`Enumerate` entry points are available
- [ ] Unit tests cover, `go test ./...` passes

### US-008: API inventory document and split-rationale table
**Description:** As a maintainer and downstream consumer, I need an authoritative API inventory document, annotating each entry with its classification and reason, as a contract.

**Acceptance Criteria:**
- [ ] Generate `API.md` in the project, listing every API signature + one-line semantic, partitioned into "method / free function"
- [ ] Each "free function" entry notes why it cannot be a method (constrains T / multiple types / nested instantiation)
- [ ] The document matches the code signatures (manually verifiable)
- [ ] Markdown renders without formatting errors

## Functional Requirements

### FR-1: Core types
- FR-1: The system shall define `Seq[T any]` (a defined type of `iter.Seq[T]`), `Seq2[K,V any]`, the `Pair[A,B any]` struct, the `Tuple3[A,B,C any]` and `Tuple4[A,B,C,D any]` structs, and the `Numeric` constraint.

### FR-2: Constructor entries (free functions — no receiver, so all functions)
- FR-2: The system shall provide the following constructor free functions:

| Function | Semantic |
|---|---|
| `From[T any](s []T) Seq[T]` | Create a sequence from a slice |
| `Of[T any](items ...T) Seq[T]` | Create a sequence from variadic arguments |
| `Empty[T any]() Seq[T]` | Empty sequence |
| `Range(start, end int) Seq[int]` | Integer sequence over `[start, end)` |
| `RangeStep(start, end, step int) Seq[int]` | Stepped integer sequence |
| `Repeat[T any](n int, v T) Seq[T]` | Repeat `v` a total of `n` times |
| `RepeatInf[T any](v T) Seq[T]` | Repeat `v` infinitely |
| `Generate[T any](f func() T) Seq[T]` | Infinite sequence, calling `f` for each element |
| `Iterate[T any](init T, f func(T) T) Seq[T]` | Infinite sequence `init, f(init), f(f(init))...` |
| `FromChannel[T any](ch <-chan T) Seq[T]` | Create from a channel (single-use source) |
| `FromMap[K comparable, V any](m map[K]V) Seq2[K,V]` | Create a Seq2 from a map |

### FR-3: Intermediate transformation methods (lazy methods on `Seq[T]`, return a new Seq)
- FR-3: The system shall provide the following methods on `Seq[T]`:

| Method | Semantic | Note |
|---|---|---|
| `Map[U any](f func(T) U) Seq[U]` | Element `T→U` transformation | core of 1.27 generic methods |
| `FlatMap[U any](f func(T) Seq[U]) Seq[U]` | Transform to subsequences then flatten | |
| `FilterMap[U any](f func(T) (U, bool)) Seq[U]` | Filter + transform combined | |
| `Filter(pred func(T) bool) Seq[T]` | Keep elements satisfying `pred` | |
| `Reject(pred func(T) bool) Seq[T]` | Drop elements satisfying `pred` | |
| `Take(n int) Seq[T]` | Take the first `n` | |
| `Drop(n int) Seq[T]` | Skip the first `n` | |
| `TakeRight(n int) Seq[T]` | Take the last `n` (materializes internally, lodash takeRight) | |
| `DropRight(n int) Seq[T]` | Drop the last `n` (materializes internally, lodash dropRight) | |
| `Slice(start, end int) Seq[T]` | Subrange `[start, end)` (lodash slice) | |
| `Init() Seq[T]` | Drop the last element (Scala init, materializes internally) | |
| `Tail() Seq[T]` | Drop the first element (Scala tail, equiv `Drop(1)`) | |
| `TakeWhile(pred func(T) bool) Seq[T]` | Take until the first that doesn't satisfy | |
| `DropWhile(pred func(T) bool) Seq[T]` | Skip leading elements that satisfy | |
| `Scan[U any](init U, f func(U, T) U) Seq[U]` | Emit each accumulated value (incl. init) | |
| `Peek(f func(T)) Seq[T]` | Side-channel side effect (debug/log), passes elements through | |
| `Chunk(size int) Seq[Seq[T]]` | Fixed-length non-overlapping chunks (Scala grouped) | |
| `Window(size, step int) Seq[Seq[T]]` | Sliding window (Scala sliding) | |
| `Intersperse(sep T) Seq[T]` | Insert a separator element between elements | |
| `DistinctBy[K comparable](key func(T) K) Seq[T]` | Deduplicate by key | `K` is the method's own constrained param |
| `SortBy(less func(T, T) bool) Seq[T]` | Sort by comparator (materializes internally) | |
| `Reverse() Seq[T]` | Reverse (materializes internally) | |
| `Concat(other Seq[T]) Seq[T]` | Append another sequence at the tail | |
| `Enumerate() Seq2[int, T]` | Pair with index, convert to `Seq2[int,T]` (Scala zipWithIndex) | |

### FR-4: Terminal methods (on `Seq[T]`, drive evaluation)
- FR-4: The system shall provide the following terminal methods on `Seq[T]`:

| Method | Semantic |
|---|---|
| `Collect() []T` | Materialize to a slice |
| `ForEach(f func(T))` | Run a side effect on each element |
| `ForEachIndexed(f func(int, T))` | Indexed traversal |
| `Fold[U any](init U, f func(U, T) U) U` | Left fold to an accumulator (Scala foldLeft) |
| `Reduce(f func(T, T) T) (T, bool)` | Reduce with no initial value, `(zero,false)` on empty |
| `Count() int` | Element count |
| `CountBy(pred func(T) bool) int` | Count satisfying `pred` |
| `GroupCount[K comparable](key func(T) K) map[K]int` | Count grouped by key (lodash countBy semantics, `K` is the method's own constrained param) |
| `Find(pred func(T) bool) (T, bool)` | First satisfying element |
| `FindIndex(pred func(T) bool) (int, bool)` | Index of the first satisfying element |
| `FindLast(pred func(T) bool) (T, bool)` | Last satisfying element (lodash findLast, materializes internally) |
| `FindLastIndex(pred func(T) bool) (int, bool)` | Index of the last satisfying element (materializes internally) |
| `Any(pred func(T) bool) bool` | Whether any satisfies (short-circuit, Scala exists) |
| `All(pred func(T) bool) bool` | Whether all satisfy (short-circuit, Scala forall) |
| `None(pred func(T) bool) bool` | Whether none satisfy |
| `First() (T, bool)` | First element |
| `Last() (T, bool)` | Last element |
| `Nth(n int) (T, bool)` | The `n`-th element |
| `IsEmpty() bool` | Whether empty |
| `GroupBy[K comparable](key func(T) K) map[K][]T` | Group by key |
| `KeyBy[K comparable](key func(T) K) map[K]T` | Index by key (later wins on same key) |
| `Partition(pred func(T) bool) ([]T, []T)` | Binary split by predicate (satisfy / not) |
| `Span(pred func(T) bool) ([]T, []T)` | Split at the first non-satisfying (Scala span) |
| `MaxBy(less func(T, T) bool) (T, bool)` | Max by comparator |
| `MinBy(less func(T, T) bool) (T, bool)` | Min by comparator |
| `MaxByKey[K cmp.Ordered](key func(T) K) (T, bool)` | Max by projected key (lodash maxBy, `K` is the method's own constrained param) |
| `MinByKey[K cmp.Ordered](key func(T) K) (T, bool)` | Min by projected key (lodash minBy, `K` is the method's own constrained param) |
| `SumBy[U Numeric](f func(T) U) U` | Sum after projection (lodash sumBy, `U` is the method's own constrained param) |
| `MeanBy[U Numeric](f func(T) U) float64` | Mean after projection (lodash meanBy) |
| `Join(sep string, str func(T) string) string` | Stringify each element then join with `sep` (Scala mkString) |

### FR-5: Constrained free functions (constrain `T` itself — so cannot be methods)
- FR-5: The system shall provide the following free functions, each unable to be a method because it constrains `T` itself:

| Function | Semantic | Empty-sequence convention |
|---|---|---|
| `Distinct[T comparable](s Seq[T]) Seq[T]` | Deduplicate, preserve first-occurrence order | — |
| `Contains[T comparable](s Seq[T], v T) bool` | Whether `v` is present (short-circuit) | `false` |
| `IndexOf[T comparable](s Seq[T], v T) (int, bool)` | First index of `v` | `(0,false)` |
| `LastIndexOf[T comparable](s Seq[T], v T) (int, bool)` | Last index of `v` (lodash lastIndexOf, materializes internally) | `(0,false)` |
| `CountValues[T comparable](s Seq[T]) map[T]int` | Occurrence count per value | empty map |
| `Max[T cmp.Ordered](s Seq[T]) (T, bool)` | Maximum | `(zero,false)` |
| `Min[T cmp.Ordered](s Seq[T]) (T, bool)` | Minimum | `(zero,false)` |
| `Mean[T Numeric](s Seq[T]) float64` | Mean (lodash mean) | `0` |
| `Sum[T Numeric](s Seq[T]) T` | Sum | `zero` |
| `Product[T Numeric](s Seq[T]) T` | Product | `1` |
| `Sort[T cmp.Ordered](s Seq[T]) Seq[T]` | Ascending sort (materializes internally) | empty sequence |
| `Equal[T comparable](a, b Seq[T]) bool` | Element-wise equality | both empty is `true` |
| `Compact[T comparable](s Seq[T]) Seq[T]` | Drop zero-value elements (lodash compact) | empty sequence |
| `Without[T comparable](s Seq[T], vals ...T) Seq[T]` | Exclude specified values (lodash without) | empty sequence |
| `Union[T comparable](seqs ...Seq[T]) Seq[T]` | Union (deduplicated) | empty sequence |
| `Intersect[T comparable](a, b Seq[T]) Seq[T]` | Intersection | empty sequence |
| `Difference[T comparable](a, b Seq[T]) Seq[T]` | Difference `a−b` | — |
| `SymmetricDifference[T comparable](a, b Seq[T]) Seq[T]` | Symmetric difference (lodash xor) | — |
| `ToSet[T comparable](s Seq[T]) map[T]struct{}` | Convert to a set | empty map |
| `JoinStrings(s Seq[string], sep string) string` | Join a string sequence (free-function Join) | `""` |

### FR-6: Multi-sequence / nested free functions (multiple type params or nested instantiation — so cannot be methods)
- FR-6: The system shall provide the following free functions:

| Function | Semantic | Why not a method |
|---|---|---|
| `Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]` | Pair two sequences, stop at the shorter | two independent type params |
| `ZipWith[A, B, C any](a Seq[A], b Seq[B], f func(A, B) C) Seq[C]` | Pair and combine via `f` (lodash zipWith) | multiple independent type params |
| `ZipMap[K comparable, V any](keys Seq[K], vals Seq[V]) map[K]V` | Pair two sequences into a map (lodash zipObject) | multiple type params + constrains K |
| `Zip3[A, B, C any](a Seq[A], b Seq[B], c Seq[C]) Seq[Tuple3[A,B,C]]` | Three-way pairing | multiple type params |
| `Zip4[A, B, C, D any](a Seq[A], b Seq[B], c Seq[C], d Seq[D]) Seq[Tuple4[A,B,C,D]]` | Four-way pairing | multiple type params |
| `Unzip[A, B any](s Seq2[A, B]) (Seq[A], Seq[B])` | Split a Seq2 into two sequences | multiple return types |
| `Flatten[T any](s Seq[Seq[T]]) Seq[T]` | Flatten one nested level | receiver must be the `Seq[Seq[T]]` instantiation |
| `Concat[T any](seqs ...Seq[T]) Seq[T]` | Join multiple sequences in order (variadic) | variadic aggregation fits a function better |
| `Interleave[T any](seqs ...Seq[T]) Seq[T]` | Take elements in round-robin | variadic aggregation |

> Note: `Tuple3[A,B,C]` and `Tuple4[A,B,C,D]` are defined alongside FR-1 (multi-element extensions of `Pair`, capped at four; use a named struct for more elements).

### FR-7: Seq2[K,V] methods and free functions
- FR-7a: The system shall provide the following methods on `Seq2[K,V]`:

| Method | Semantic |
|---|---|
| `MapValues[U any](f func(V) U) Seq2[K, U]` | Transform value only |
| `MapKeys[J any](f func(K) J) Seq2[J, V]` | Transform key only |
| `Map[J, U any](f func(K, V) (J, U)) Seq2[J, U]` | Transform key/value together |
| `Filter(pred func(K, V) bool) Seq2[K, V]` | Filter by (k,v) |
| `Keys() Seq[K]` | Project out the key sequence |
| `Values() Seq[V]` | Project out the value sequence |
| `ForEach(f func(K, V))` | Side-effect traversal |
| `Fold[U any](init U, f func(U, K, V) U) U` | Left fold |
| `Count() int` | Pair count |
| `Find(pred func(K, V) bool) (K, V, bool)` | First satisfying (k,v) |
| `Any(pred func(K, V) bool) bool` | Existence (short-circuit) |
| `All(pred func(K, V) bool) bool` | Universality (short-circuit) |

- FR-7b: The system shall provide the following `Seq2` accompanying free functions (constrain K itself, so functions):

| Function | Semantic |
|---|---|
| `ToMap[K comparable, V any](s Seq2[K, V]) map[K]V` | Materialize to a map |
| `CollectPairs[K, V any](s Seq2[K, V]) []Pair[K, V]` | Materialize to a Pair slice |
| `Entries[K, V any](pairs []Pair[K, V]) Seq2[K, V]` | Create from a Pair slice |
| `Associate[T any, K comparable, V any](s Seq[T], f func(T) (K, V)) Seq2[K, V]` | Project `Seq[T]` via `f` into `Seq2[K,V]` (closes the Open Question; constrains K, so a free function) |

### FR-8: Documentation
- FR-8: The system shall generate `API.md` at the project root, listing every API's signature and one-line semantic partitioned into "method / free function", with free functions annotated by the reason category they cannot be methods.

## Non-Goals (Out of Scope)

- **Error handling / `Seq2[T, error]` short-circuit chains**: explicitly not in this version. Scala's `Try`/`Either`/for-comprehension error flow has no elegant counterpart on Go's lazy sequences; deferred to a separate PRD.
- **HKT / type class abstraction** (Functor/Monad/an abstractable `Collection` interface): not supported by the language, not attempted.
- **for-comprehension syntax sugar**: Go has no equivalent syntax, not simulated.
- **Parallel execution** (vs. `lo/parallel`): this version does sequential lazy pipelines only; parallelism deferred.
- **In-place mutation** (vs. `lo/mutable`): conflicts with the lazy immutable positioning, not done.
- **Concrete implementation and performance tuning**: this PRD defines the API contract only, no implementation code or benchmarks.
- **Generic methods satisfying interfaces**: a hard line of the proposal, no polymorphic collection interface.
- **Arbitrary-depth flatten** (lodash `flattenDeep`/`flattenDepth`): Go's type system cannot express the type "nested sequence of arbitrary depth", so only a fixed one-level `Flatten` is provided; deeper nesting requires explicit repeated calls or a custom structure.
- **Scala-style high-arity tuples (Tuple5–Tuple22)**: tuples are capped at `Tuple4`. Go has no tuple literal or pattern destructuring, the fields of a high-arity `TupleN` can only be names like `Field1..FieldN` with no semantics, readability is poor, and `ZipN`/`Unzip`/tests must be maintained per arity; aggregation over more fields should use a named `struct` (with meaningful field names), so no five-or-more-element tuples and no corresponding `Zip5+`.

## Technical Considerations

- **Target Go version: 1.27** (generic methods). `go.mod` declares `go 1.27`, and a local 1.27 toolchain is required (currently `go1.27rc1`, to switch to the stable release once available). Compilation/testing runs under the 1.27 toolchain.
- **Defined type, not a struct wrapper**: `type Seq[T any] iter.Seq[T]` converts with `iter.Seq` at zero cost and feeds seamlessly into the standard library's `slices.Collect`, `maps.Keys`, etc.
- **Lazy semantics**: intermediate methods return an unevaluated `Seq`; only terminal methods drive traversal. `API.md` must annotate each method as "intermediate" or "terminal".
- **Re-iteration semantics**: a `Seq` from a slice can be re-iterated; one from a channel is single-use. Must be clarified in the docs.
- **`cmp.Ordered`** comes from the standard library `cmp` package; `Numeric` needs a custom constraint.
- **Performance note**: a deep chain = nested `yield` closure calls, and a hot inner loop may not match a hand-written loop; the docs should give applicability-boundary hints.

## Success Metrics

- The API inventory covers 100% of the entries listed in FR-2 through FR-7, with no omissions.
- Every free function maps to one of the three classification reasons (constrains T / multiple type params / nested instantiation).
- The split rule is mechanically auditable: the inventory contains no entry that "constrains T itself yet is listed as a method".
- Downstream can run `/to-issues` directly from this inventory to decompose into implementable Issues, with no second round of clarification on API shape.

## Open Questions

> The following decisions were settled in review (kept for the record):
> - **MapIndexed**: not provided separately; indexing always goes through `Enumerate().Map(...)`, keeping `Map(func(T) U)` as the single form.
> - **Tuple arity**: provided up to `Tuple4`, capped; more elements use a custom struct.
> - **Join naming**: unified into the `Join` family — method `Join(sep, fn)`, free function `JoinStrings(s, sep)`.
> - **Count naming**: `CountBy(pred)` = predicate count (keeps original meaning); count grouped by key is named `GroupCount[K]`.
> - **By vs ByKey convention**: `SumBy`/`MeanBy` = aggregate by projected value; `MaxByKey`/`MinByKey` = take extreme by projected key; `MaxBy`/`MinBy` = comparator versions. The docs must state this convention clearly.

Remaining items to verify (not design decisions, confirmed at implementation/test time):

- Whether the inner `Seq[Seq[T]]` returned by `Chunk`/`Window` also supports the full method chain (it should — it is itself a `Seq`); to be verified in tests that the nested chain works.
