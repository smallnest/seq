package seq

import (
	"cmp"
	"iter"
	"slices"
)

// Numeric / ordered free functions. These constrain T itself (Numeric or
// cmp.Ordered), so by the library's 划分铁律 they cannot be methods on Seq[T]
// (whose T is any). They are the free-function counterpart to the SeqNumeric /
// SeqOrdered subtype methods (issue #10) and keep identical semantics.

// Max returns the maximum element of s by the natural ordering, or None if s
// is empty.
func Max[T cmp.Ordered](s Seq[T]) Optional[T] {
	var best T
	first := true
	for v := range iter.Seq[T](s) {
		if first {
			best = v
			first = false
			continue
		}
		if cmp.Less(best, v) {
			best = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(best)
}

// Min returns the minimum element of s by the natural ordering, or None if s
// is empty.
func Min[T cmp.Ordered](s Seq[T]) Optional[T] {
	var best T
	first := true
	for v := range iter.Seq[T](s) {
		if first {
			best = v
			first = false
			continue
		}
		if cmp.Less(v, best) {
			best = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(best)
}

// Sum returns the sum of all elements. An empty sequence yields the zero
// value of T.
func Sum[T Numeric](s Seq[T]) T {
	var sum T
	for v := range iter.Seq[T](s) {
		sum += v
	}
	return sum
}

// Product returns the product of all elements. An empty sequence yields 1
// (the multiplicative identity).
func Product[T Numeric](s Seq[T]) T {
	var prod T = 1
	for v := range iter.Seq[T](s) {
		prod *= v
	}
	return prod
}

// Mean returns the arithmetic mean of the elements as a float64, or 0 for an
// empty sequence.
func Mean[T Numeric](s Seq[T]) float64 {
	var sum T
	n := 0
	for v := range iter.Seq[T](s) {
		sum += v
		n++
	}
	if n == 0 {
		return 0
	}
	return float64(sum) / float64(n)
}

// Sort returns a Seq that yields s's elements in ascending natural order. The
// input is materialized internally, so the result may be iterated repeatedly.
// An empty sequence sorts to empty.
func Sort[T cmp.Ordered](s Seq[T]) Seq[T] {
	collected := slices.Collect(iter.Seq[T](s))
	slices.Sort(collected)
	return From(collected)
}
