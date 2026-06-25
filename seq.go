// Package seq provides chainable, lazy collection pipelines built on top of
// the standard library's [iter.Seq] / [iter.Seq2] iterators, using Go 1.27's
// generic methods feature (golang/go#77273).
//
// The headline capability — reading data flow left to right — comes from
// generic methods such as [Seq.Map], which can change the element type:
//
//	seq.From([]int{1, 2}).Map(strconv.Itoa).Collect() // ["1" "2"]
//
// Seq[T] is a defined type over [iter.Seq[T]], not a struct wrapper, so it
// interconverts with [iter.Seq[T]] at zero cost and can be fed directly to
// [slices.Collect], [maps.Keys] and the rest of the standard library iterator
// ecosystem. Because Seq declares T as any at the type level, operations that
// require T itself to satisfy comparable / [cmp.Ordered] / [Numeric] cannot be
// methods on Seq[T]; they are provided as free functions (see [Distinct],
// [Max], [Sum]) or recovered as methods on the constrained subtypes
// [SeqComparable], [SeqOrdered] and [SeqNumeric].
//
// # Function arguments
//
// Every higher-order operation (Map, Filter, FlatMap, Reduce, ForEach, and the
// rest) requires a non-nil function argument. Because operations are lazy, a
// nil function passed to an intermediate operation panics not at the call site
// but when a terminal operation later drives iteration; a nil function passed
// to a terminal operation panics immediately. Callers must not pass nil.
package seq

import (
	"iter"
)

// Seq is a defined type over [iter.Seq[T]]. It interconverts with iter.Seq[T]
// at zero cost: Seq[T](it) and iter.Seq[T](s) are both compile-time
// conversions with no runtime overhead.
//
// T is declared as any at the type level. Any operation that needs T to
// satisfy a constraint (comparable, [cmp.Ordered], [Numeric]) is therefore a
// free function rather than a method — see the design doc's "划分铁律".
type Seq[T any] iter.Seq[T]

// Seq2 is a defined type over [iter.Seq2[K, V]], the key-value iterator. It
// interconverts with iter.Seq2[K, V] at zero cost, the same way [Seq] does
// with iter.Seq[T].
type Seq2[K, V any] iter.Seq2[K, V]

// Pair holds two values of (possibly) different types. It is the element type
// produced by [Zip] and consumed by [Unzip] / Entries.
type Pair[A, B any] struct {
	Left  A
	Right B
}

// Tuple3 holds three values. Tuples cap at four elements; aggregate more
// fields with a named struct instead.
type Tuple3[A, B, C any] struct {
	First  A
	Second B
	Third  C
}

// Tuple4 holds four values.
type Tuple4[A, B, C, D any] struct {
	First  A
	Second B
	Third  C
	Fourth D
}

// Numeric is the constraint satisfied by all integer and floating-point types.
// It backs the [Sum], [Product] and [Mean] free functions, the SumBy/MeanBy
// methods, and the [SeqNumeric] subtype.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}
