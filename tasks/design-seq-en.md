Title: Chainable, lazy collection pipelines for iter.Seq, using Go 1.27 generic methods
Author(s): chaoyuepan
Last updated: 2026-06-24
Discussion at: (none yet — under internal review; to be filled once a project issue exists)
Status: Draft

> Language: [简体中文](design-seq.md) | English


## Abstract

We propose `seq`: a generic collection library wrapping the standard library's lazy iterators `iter.Seq` / `iter.Seq2` (Go 1.23+), making `From(xs).Filter(...).Map(...).Collect()` — a left-to-right, discoverable, lazy pipeline — possible in Go for the first time. Its foundation is the **generic methods feature in Go 1.27** ([golang/go#77273](https://github.com/golang/go/issues/77273)). Before that, a method could not declare its own type parameters, so `.Map()` could not express "input `Seq[T]`, output `Seq[U]`", which is why libraries like `samber/lo` can only offer inside-out nested functions `lo.Map(lo.Filter(...))`.

The most important promise, and the biggest constraint: **this is a design aimed at Go 1.27.** The generic methods proposal has been accepted and is implemented in Go 1.27; Go 1.27 RC (`go1.27rc1`) is installed locally, `go.mod` declares `go 1.27`, and the full API compiles and tests. This document does not provide a "works on older Go too" compatibility implementation — that is deliberate, and the reasoning is in Rationale. The only thing still pending is the official stable 1.27 release (currently an RC).

This library **defines the API contract only**, with no implementation. It runs on one dividing rule throughout: any operation that constrains the element type `T` itself (`Distinct` needs `comparable`, `Max` needs `Ordered`, `Sum` needs `Numeric`) can only be a free function; only operations that use a method's own constrained parameter (`DistinctBy[K comparable]`, `GroupBy[K comparable]`) can return to method form. This rule is forced by the language, not a style preference, and the reasoning is below.

## Background

Here is what "take evens, square them, sum them" looks like with `lo` today:

```go
// Read inside-out: look at the innermost Filter first, then Map, then Sum.
// Reading order is the reverse of data flow.
sum := lo.Sum(
    lo.Map(
        lo.Filter([]int{1, 2, 3, 4, 5, 6}, func(x int, _ int) bool {
            return x%2 == 0
        }),
        func(x int, _ int) int { return x * x },
    ),
)
```

This code has three concrete pain points:

1. **The reading direction is twisted.** The data flow is filter → map → sum, but your eyes have to jump to the innermost `Filter` first, then read outward backwards. The deeper the nesting, the more you bounce between parentheses.

2. **Each step materializes a new slice.** `lo.Filter` returns a new `[]int`, `lo.Map` returns another; nothing can short-circuit lazily. To "find the first even number's square" you'd have to rewrite the whole thing by hand, losing the pipeline's early exit.

3. **Type transformation is especially awkward.** When `[]int` becomes `[]string` via `Map`, the type parameters on `lo.Map[int, string]` have to be spelled out explicitly or semi-explicitly; it won't chain.

Why doesn't `lo` do `xs.Filter(...).Map(...)`? Because before Go 1.27, **a method cannot declare its own type parameters**. For `Map` to turn `Seq[int]` into `Seq[string]`, it needs a method-level type parameter `U`:

```go
// Pre-1.27: this does not compile. Methods may not have their own [U any].
func (s Seq[T]) Map[U any](f func(T) U) Seq[U]
```

This is a hard wall at the language level. It is not that `lo` doesn't want to chain — it can't. [golang/go#77273](https://github.com/golang/go/issues/77273) lifts that restriction, and chainable pipelines become possible — that is the entire reason this library exists.

## Design

### Core types are "defined types", not struct wrappers

We define `Seq[T]` as a defined type of `iter.Seq[T]`, rather than wrapping a struct:

```go
type Seq[T any] iter.Seq[T]      // = func(yield func(T) bool)
type Seq2[K, V any] iter.Seq2[K, V]
```

The benefit is zero-cost conversion: any `iter.Seq[T]` can be used directly as a `Seq[T]` and vice versa, and it feeds straight into the standard library's `slices.Collect`, `maps.Keys`. The cost is that `Seq[T]` declares `T` as `any` at the type level — this fact will immediately decide half the API's classification.

Plus two helper types and one constraint:

```go
type Pair[A, B any] struct { Left A; Right B }
type Tuple3[A, B, C any] struct { ... }   // capped at four
type Tuple4[A, B, C, D any] struct { ... }
type Numeric interface { ~int | ~int8 | ... | ~float64 }  // integers + floats
```

### Start from the smallest chain

The simplest pipeline: from a slice, filter, materialize back to a slice.

```go
seq.From([]int{1, 2, 3, 4}).
    Filter(func(x int) bool { return x%2 == 0 }).
    Collect()                                    // [2 4]
```

It reads in the same direction as the data flows: left to right. The middle `Filter` is a **lazy intermediate operation** — it returns a new `Seq[int]`, and at this moment no element has been traversed. Only the terminal `Collect` at the end drives the whole chain.

### Type transformation: the home turf of 1.27 generic methods

`Map` can change the element type. This is the library's signature capability, and the very `.Map()` that was impossible before 1.27:

```go
seq.From([]int{1, 2}).
    Map(strconv.Itoa).        // Seq[int] → Seq[string]
    Collect()                  // ["1" "2"]
```

The `U` in `Map[U any](f func(T) U) Seq[U]` is the type parameter the method declares for itself. Rewriting the twisted `lo` code from the start:

```go
// After: left to right, read in one pass, and lazy.
sum := seq.SumBy(
    seq.From([]int{1, 2, 3, 4, 5, 6}).
        Filter(func(x int) bool { return x%2 == 0 }).
        Map(func(x int) int { return x * x }),
)
// Want "the first even number's square"? Just swap the end for .First() — the chain short-circuits and won't run the whole sequence.
```

### The dividing rule: whichever type the constraint lands on decides method vs. function

This is the center of the whole library; every API classification follows from it. Because `Seq[T]` pins `T` to `any`, **a method's type parameters are fresh and independent — they cannot add a constraint back onto the receiver's `T`**. Therefore:

- `Distinct` (needs `T` to be `comparable`), `Max` (needs `Ordered`), `Sum` (needs `Numeric`) — **none can be methods**, only free functions:

```go
func Distinct[T comparable](s Seq[T]) Seq[T]
func Max[T cmp.Ordered](s Seq[T]) (T, bool)
func Sum[T Numeric](s Seq[T]) T
```

- But the "by key" variants can return to method form, because the constraint lands on the method's own parameter `K`, not the receiver's `T`:

```go
func (s Seq[T]) DistinctBy[K comparable](key func(T) K) Seq[T]
func (s Seq[T]) GroupBy[K comparable](key func(T) K) map[K][]T
```

- Operations over multiple independent generator types, or where the receiver must be a specific instantiation, can also only be free functions:

```go
func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]   // two independent types
func Flatten[T any](s Seq[Seq[T]]) Seq[T]            // receiver must be Seq[Seq[T]]
```

The benefit of this rule is that it is **mechanically checkable**: take any API and ask "does it constrain `T` itself?" — if yes it must be a function, if no it can be a method. The inventory must not contain any entry that "constrains `T` yet is listed as a method".

### Constrained subtypes: giving the chain back to Distinct/Max/Sum

The rule has a side effect: once operations that constrain `T` become free functions, the pipeline goes back to reading inside-out. `seq.Sum(seq.Distinct(seq.From(xs)))` is the same nesting as the `lo` we set out to kill.

The fix is to introduce three **constrained subtypes** that pin "the constraint on `T`" onto the type ahead of time, so their `Distinct`/`Max`/`Sum` can legitimately be methods:

```go
type SeqComparable[T comparable]  iter.Seq[T]   // .Distinct() .Contains() .CountValues() .ToSet() .Union() .Equal() …
type SeqOrdered[T cmp.Ordered]    iter.Seq[T]   // .Max() .Min() .Sort()  + everything from Comparable
type SeqNumeric[T Numeric]        iter.Seq[T]   // .Sum() .Product() .Mean() + everything from Ordered
```

The **entry into these types is still a free function** — that's a hard requirement of the rule, unavoidable. Because `Seq[T any]`'s `T` is `any`, converting it to `SeqComparable[T comparable]` amounts to requiring `any` to satisfy `comparable`, which only a constrained free function can do:

```go
func Comparable[T comparable](s Seq[T]) SeqComparable[T]
func Ordered[T cmp.Ordered](s Seq[T]) SeqOrdered[T]
func Numbers[T Numeric](s Seq[T]) SeqNumeric[T]   // name avoids the Numeric constraint itself
```

**But "downgrading" between constraints can be a method** — this is the key insight. Constraints form a hierarchy: `Numeric ⊂ Ordered ⊂ comparable` (numbers can be ordered, and can be compared for equality). A stronger constraint already satisfies a weaker one, so going from a stronger type to a weaker one doesn't re-constrain `T` and can hang as a method:

```go
func (s SeqNumeric[T]) Ordered() SeqOrdered[T]      // legal: Numeric satisfies Ordered
func (s SeqOrdered[T]) Comparable() SeqComparable[T] // legal: Ordered satisfies comparable
```

Net effect: **a pipeline pays the free-function cost only once, at the entry, and is a method chain the whole way after — including the type conversions.**

```go
// Before: inside-out, two layers of wrapping
sum := seq.Sum(seq.Distinct(seq.From(xs)))
// After: one entry, then a full chain
sum := seq.Numbers(seq.From(xs)).Distinct().Sum()
```

The cost, said in the open: the lazy operations that preserve `T` (`Filter`/`Take`/`Drop`/`TakeWhile`…) must be **re-exposed on each subtype** (returning the same type, so the chain doesn't break), which is this approach's main API-surface cost. `Map` changes `T` and returns a plain `Seq[U]`; call `Comparable()` again when needed.

### Boundaries: what shapes this library does not do

- **No for-comprehension syntax sugar**; everything is a bare method chain.
- **No HKT / type classes**; Functor/Monad cannot be abstracted, every method is a concrete method on a concrete type.
- **Generic methods do not satisfy interfaces** (a hard line of the 1.27 proposal), so this library offers no polymorphic interface abstraction over collections.
- **No universal equality**; `==` is a constraint, not a universal capability, which is exactly why `Distinct`/`Max` must be functions.

The full API inventory (constructors, intermediate methods, terminal methods, constrained functions, multi-sequence functions, the `Seq2` family) is in the PRD `tasks/prd-seq-api-inventory.md`; this document does not repeat it.

## Rationale

### Why stake method chaining on the newest language version

This is the decision most open to challenge, so answer it head-on. We could build an "older-Go compatible" version — all free functions, no method chaining. We did not take that road, because what comes out of it is just another `lo`, and `lo` is already good enough. This library's one irreplaceable value is the method chain; remove it and the project has no reason to exist. The generic methods proposal (#77273) has been accepted and is implemented in Go 1.27, and the locally installed 1.27 RC already compiles the full API — so this is not a bet on whether the feature passes, it is accepting a clear cost: pinning the minimum supported Go version at 1.27 and leaving older-version users behind. We think the trade is worth it.

### Why a "defined type" instead of a struct wrapper

The rejected option: `type Seq[T any] struct { it iter.Seq[T] }`. The temptation of a struct wrapper is the ability to hang arbitrary methods and extend fields. We didn't choose it, because it severs the zero-cost conversion with `iter.Seq` — every time a user wants to feed `slices.Collect` they'd have to unwrap `.it` first, and the seamless fit with the standard library ecosystem is gone. A defined type lets `Seq[T]` and `iter.Seq[T]` be used implicitly as each other, and that benefit outweighs the struct's flexibility.

### Why Distinct/Max/Sum are functions, not methods — this is not our choice

Someone will ask "why not make `Distinct` into `s.Distinct()`, it's handier". The answer is: on `Seq[T any]` it can't be done, not that we don't want to. `Seq[T any]` declares `T` as `any` at the type level, and a method cannot add a `comparable` constraint onto `T` (a method's type parameters are fresh new parameters, they cannot reach back to constrain the receiver). This is a language rule, not API taste. We provide two escape hatches: "by key" method variants (`DistinctBy`), and the constrained subtypes above (`From(xs).Comparable().Distinct()`) — pin the constraint onto the type ahead of time and the method chain comes back. The free function `Distinct` on a bare `Seq[T]` is kept as a fallback for when you don't want to convert types.

### Why the constrained-subtype entry is a function while downgrading is a method

This is the asymmetry most open to challenge in the `SeqComparable`/`SeqOrdered`/`SeqNumeric` design. The rejected option: make `Comparable()` a method on `Seq[T]` too, so the whole path is free-function-free. It can't be done — `Seq[T any]`'s `T` is `any`, and a method returning `SeqComparable[T comparable]` amounts to requiring `any` to satisfy `comparable`, which won't compile. So the first hop from the `any` world into the constrained world must be a free function. Downgrading inside the constrained world (`Numeric → Ordered → comparable`) can be a method, because a stronger constraint already satisfies a weaker one and no new constraint on `T` is added. In one line: **cross a constraint boundary with a function, move within a constraint boundary with a method.** This is fully consistent with the rule, not an exception.

The other rejected option: build only `SeqNumeric`. We didn't choose it, because then `Distinct` (needs only `comparable`) and `Max` (needs only `Ordered`) would either be forced into `SeqNumeric` (too strong a constraint — strings couldn't chain `Distinct`) or stay free functions (the chain breaks again). Three types aligned with three constraints let each operation land on the weakest constraint it actually needs.

### Why tuples cap at Tuple4 and don't follow Scala to Tuple22

The rejected option: Scala-style `Tuple1`–`Tuple22`. We didn't choose it, because Go has no tuple literal and no pattern destructuring; the fields of a high-arity `TupleN` can only be named `Field1..FieldN`, which reads poorly, and `ZipN`/`Unzip`/tests would have to be maintained per arity. `Zip`/`Zip3`/`Zip4` already cover nearly all real scenarios, and aggregation over more fields should use a named struct (with meaningful field names). Capping at four is the balance point between benefit and noise.

### Why no error handling / parallelism / in-place mutation

- **Error flow (`Seq2[T, error]` short-circuit chains)**: Scala's `Try`/`Either`/for-comprehension have no elegant counterpart on Go's lazy sequences, and forcing it would pollute every signature. Deferred to a separate PRD.
- **Parallelism (vs. `lo/parallel`)**: this version does sequential lazy pipelines only. Parallelism involves entirely different semantics and correctness guarantees, and should not be crammed into v1.
- **In-place mutation (vs. `lo/mutable`)**: directly conflicts with the "lazy + immutable pipeline" positioning. Not done.

## Compatibility

**This is a purely additive new library with zero breakage to any existing code** — it has no "existing callers". The real cost is not backward compatibility, but that it pins the minimum supported Go version at 1.27.

The costs, listed honestly:

- **The minimum Go version is 1.27.** `go.mod` declares `go 1.27`, and a local 1.27 toolchain is required (currently `go1.27rc1`). Any project still on Go 1.26 or earlier cannot adopt this library — that shuts out potential users, and it must be said in the open.
- **We currently depend on an RC, not a stable release.** Before the official 1.27 release, the syntax details of generic methods may still be adjusted (e.g. where method type parameters are declared); if so, the affected method signatures must be aligned accordingly. We align with the accepted proposal, not with an assumption that the details stay unchanged.

Migration path: the library is itself opt-in — nobody is forced to depend on it. For consumers, we recommend adopting it in production only after Go 1.27 is officially released (the current RC is fine for early adoption and development), with all unit tests passing on the 1.27 toolchain as a gate before release.

## Implementation

Landing in three steps, each with a verifiable output (the local 1.27 RC is ready, so all three can be compiled and tested immediately):

1. **Types and contract first.** Land the type definitions `Seq`/`Seq2`/`Pair`/`Tuple3`/`Tuple4`/`Numeric` and all **free functions** (`From`, `Distinct`, `Max`, `Zip`, and the constrained-subtype entries `Comparable`/`Ordered`/`Numbers`…) first, verifying the dividing rule holds. This batch does not depend on generic methods and, even split out on its own, compiles on Go 1.23+ as an independently publishable subset.

2. **The method-chain part.** Methods with method-level type parameters like `Map`/`Filter`/`FlatMap`, plus the methods and downgrade methods on the constrained subtypes (`SeqComparable`/`SeqOrdered`/`SeqNumeric`), depend on #77273 (Go 1.27), with "all unit tests pass on the 1.27 toolchain" as the done gate.

3. **Generate the authoritative API.md.** Partitioned by "method / free function", listing every signature plus a one-line semantic, with each free function annotated by the reason it cannot be a method (constrains T / multiple type params / nested instantiation). Documentation and code signatures cross-checked by hand.

The free-function set from step 1 is itself a subset that compiles independently on Go 1.23+ — equivalent to "a `lo` with the correct constraint split", still valuable to users who can't yet upgrade to 1.27.

## Appendix

### FAQ

**Q: Why not just use `samber/lo`?**
A: `lo` solves "do these operations exist"; this library solves "can you chain them left-to-right, lazily". Different positioning; `lo`'s full capability set cannot chain before 1.27, which is precisely this library's entry point.

**Q: Why do `SumBy` (method) and `Sum` (function) coexist?**
A: `Sum[T Numeric]` constrains `T` itself and on a bare `Seq[T]` can only be a function; `SumBy[U Numeric](f func(T) U)` lands its constraint on the method's own `U` and can be a method. Naming convention: `By` = aggregate by projected value (method), `ByKey` = take the extreme by projected key (method), bare `Max`/`Min`/`Sum` = functions that constrain T. To call `Sum` in a chain, convert the sequence to `SeqNumeric`: `From(xs).Numbers().Sum()`.

**Q: Constrained subtypes (SeqComparable/SeqOrdered/SeqNumeric) vs. bare Seq — which do I use?**
A: Default to `Seq[T]`. When you want to chain `Sum`/`Max`/`Distinct` over numeric/comparable/ordered elements, convert once with an entry function (`Numbers`/`Ordered`/`Comparable`), then it's a full method chain — and you can switch among the three with downgrade methods (`Numbers().Ordered().Comparable()`). When you don't want to convert types, the free-function versions on a bare `Seq[T]` (`Sum(s)`/`Distinct(s)`) remain available.

**Q: What about performance?**
A: A deep chain is nested `yield` closure calls, and a hot inner loop may not match a hand-written loop. This library's selling point is readability and composability, not peak performance. API.md will note the applicability boundaries.

**Q: Can the source sequence be re-iterated?**
A: It depends on the source. A `Seq` from a slice can be re-iterated; one from a channel is a single-use source. The docs will annotate each one.
