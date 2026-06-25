# seq

[![Go Reference](https://pkg.go.dev/badge/github.com/smallnest/seq.svg)](https://pkg.go.dev/github.com/smallnest/seq)
[![CI](https://github.com/smallnest/seq/actions/workflows/ci.yml/badge.svg)](https://github.com/smallnest/seq/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.27-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

English | [简体中文](README_CN.md)

**Chainable, lazy collection pipelines for Go, built on `iter.Seq`.**

`seq` is a generic Go library that wraps the standard library's lazy iterators `iter.Seq` / `iter.Seq2` (Go 1.23+) and gives them Scala-style, left-to-right, chainable operations:

```go
sum := seq.From([]int{1, 2, 3, 4, 5, 6}).
    Filter(func(x int) bool { return x%2 == 0 }).
    SumBy(func(x int) int { return x * x })
```

> ⚠️ **This library requires Go 1.27.** The chainable methods depend on the generic methods proposal ([golang/go#77273](https://github.com/golang/go/issues/77273)), which **has been accepted** and is implemented in Go 1.27. A 1.27 toolchain (currently `go1.27rc1`) is required to build. See [Status](#status) below.

## Why it exists

Today, libraries like [`samber/lo`](https://github.com/samber/lo) can only offer top-level functions, so a pipeline reads inside-out:

```go
// Reading order is the reverse of data flow: you start at the innermost Filter.
sum := lo.Sum(
    lo.Map(
        lo.Filter([]int{1, 2, 3, 4, 5, 6}, func(x int, _ int) bool { return x%2 == 0 }),
        func(x int, _ int) int { return x * x },
    ),
)
```

This isn't a style preference — it's a language limitation. Before Go 1.27, **a method cannot declare its own type parameters**, so a `.Map()` that turns `Seq[int]` into `Seq[string]` is impossible to express:

```go
// Pre-1.27: this does not compile. Methods may not have their own [U any].
func (s Seq[T]) Map[U any](f func(T) U) Seq[U]
```

[golang/go#77273](https://github.com/golang/go/issues/77273) lifts that restriction. Chainable, lazy, discoverable pipelines become possible — and that is the entire reason this library exists.

## Design

### Core types are *defined types*, not struct wrappers

```go
type Seq[T any]      iter.Seq[T]       // = func(yield func(T) bool)
type Seq2[K, V any]  iter.Seq2[K, V]
```

This gives zero-cost conversion: any `iter.Seq[T]` can be used directly as a `Seq[T]` and vice versa, and the result feeds straight into `slices.Collect`, `maps.Keys`, etc. The cost is that `Seq[T]` declares `T` as `any` at the type level — and that single fact decides half the API.

### The dividing rule: method vs. free function

Because `Seq[T any]` pins `T` to `any`, a method's type parameters are fresh and independent — **they cannot add a constraint back onto the receiver's `T`**. So:

- Operations that constrain `T` itself **must be free functions**:

  ```go
  func Distinct[T comparable](s Seq[T]) Seq[T]
  func Max[T cmp.Ordered](s Seq[T]) (T, bool)
  func Sum[T Numeric](s Seq[T]) T
  ```

- Operations that only use a method's own constrained parameter **can be methods** (the "escape hatch"):

  ```go
  func (s Seq[T]) DistinctBy[K comparable](key func(T) K) Seq[T]
  func (s Seq[T]) GroupBy[K comparable](key func(T) K) map[K][]T
  ```

- Operations over multiple generator types or a specific instantiation **must be free functions**:

  ```go
  func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]   // two independent types
  func Flatten[T any](s Seq[Seq[T]]) Seq[T]           // receiver must be Seq[Seq[T]]
  ```

The rule is mechanically checkable: ask "does it constrain `T` itself?" If yes, it's a function; if no, it can be a method.

## API at a glance

| Group | Form | Examples |
|-------|------|----------|
| Constructors | free functions | `From`, `Of`, `Range`, `RangeStep`, `Repeat`, `Generate`, `Iterate`, `Times`, `FromChannel`, `FromMap` |
| Intermediate (lazy) | methods | `Map`, `Filter`, `FlatMap`, `FilterMap`, `Reject`, `Take`/`Drop`, `TakeWhile`/`DropWhile`, `Scan`, `Chunk`, `Window`, `DistinctBy`, `Enumerate` |
| Terminal | methods | `Collect`, `Fold`, `Reduce`, `Find`, `Any`/`All`/`None`, `GroupBy`, `KeyBy`, `Partition`, `SumBy`, `MaxByKey`, `Join` |
| Grouping | free functions | `PartitionBy` (ordered groups as `Seq2[K,[]T]`) |
| Constrained | free functions | `Distinct`, `Contains`, `Max`/`Min`, `Sum`/`Product`/`Mean`, `Sort`, `Union`/`Intersect`/`Difference`, `Compact`, `Without` |
| Constrained subtypes | functions (entry) + methods | `Comparable`/`Ordered`/`Numbers` enter; then chain `.Distinct()`, `.Max()`, `.Sum()`, `.Replace()`/`.ReplaceAll()`; downgrade with `.Ordered()`/`.Comparable()` |
| Multi-sequence | free functions | `Zip`/`Zip3`/`Zip4`, `ZipWith`, `ZipMap`, `Unzip`, `Flatten`, `Concat`, `Interleave` |
| `Seq2[K,V]` | methods + functions | `MapValues`, `MapKeys`, `Keys`, `Values`; `ToMap`, `CollectPairs`, `Associate` |
| `Optional[T]` | type + methods + function | `Some`/`None`/`ToOptional`; `Get`, `Unwrap`/`UnwrapOr`/`OrElse`, `Map`, `Filter`, `FlatMap`; `MapOptional` (type-changing) |

The full, authoritative list with one-line semantics for each entry will live in `API.md`. The design rationale is in [`tasks/design-seq.md`](tasks/design-seq.md); the complete API inventory is in [`tasks/prd-seq-api-inventory.md`](tasks/prd-seq-api-inventory.md).

### Optional[T]

`Optional[T]` is a zero-dependency, opt-in wrapper over the `(T, bool)` pair returned by `Find`, `First`, `Last`, `Nth`, `Reduce` and the other partial-result methods. It exists only as caller-side sugar — **no `Seq`/`Seq2` method takes or returns it**, so the package keeps its zero-cost interop with `iter.Seq` and the standard library. Bridge from any `(T, bool)` method with `ToOptional`:

```go
out := seq.ToOptional(s.Find(func(x int) bool { return x > 2 })).
    Map(func(x int) int { return x * 10 }).
    OrElse(-1)
```

Because Go methods cannot introduce new type parameters, `Optional.Map` is same-type only; use the package-level `MapOptional[T, U]` to change the element type (e.g. `Optional[int]` → `Optional[string]`).

## Scope

**Not included** (by deliberate decision, see the design doc for the reasoning):

- **Error-handling chains** (`Seq2[T, error]` short-circuiting) — no clean fit on Go's lazy iterators; deferred to a separate proposal.
- **Parallel execution** (`lo/parallel`) — this version is sequential and lazy only.
- **In-place mutation** (`lo/mutable`) — conflicts with the lazy, immutable model.
- **Arbitrary-depth flatten** (`flattenDeep`) — Go's type system can't express it; only one-level `Flatten`.
- **High-arity tuples** (`Tuple5`–`Tuple22`) — capped at `Tuple4`; use named structs for more fields.
- **HKT / type classes** and **generic methods satisfying interfaces** — the language doesn't support these.

## Status

**Draft.** The library requires **Go 1.27** (`go.mod` declares `go 1.27`); a 1.27 toolchain is needed to build, currently `go1.27rc1`. Work is split into two batches:

- **Batch ① — core types and free functions** (`From`, `Distinct`, `Max`, `Zip`, …). Independently useful — effectively "a `lo` with the correct constraint split" — and compiles on Go 1.23+ if split out on its own.
- **Batch ② — the chainable methods** (`Map`, `Filter`, `Fold`, …) that depend on Go 1.27 generic methods.

The generic methods proposal ([golang/go#77273](https://github.com/golang/go/issues/77273)) has been accepted and is implemented in Go 1.27. The remaining gap is only that 1.27 is currently an RC, not yet a stable release; pinning the minimum version at 1.27 is the cost we accept for the method chain.

## License

See [LICENSE](LICENSE).
