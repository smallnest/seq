package seq

import (
	"cmp"
	"iter"
)

// Folding, grouping, and aggregation terminal methods. These use Go 1.27
// method-level type parameters where the aggregation introduces a new type
// (Fold[U], SumBy/MeanBy[U]) or a constrained key (GroupCount[K], GroupBy[K],
// KeyBy[K], MaxByKey/MinByKey[K]).

// Fold left-folds the sequence into an accumulator starting from init (Scala
// foldLeft). U is the method's own type parameter, so the result type may
// differ from T.
func (s Seq[T]) Fold[U any](init U, f func(U, T) U) U {
	acc := init
	for v := range iter.Seq[T](s) {
		acc = f(acc, v)
	}
	return acc
}

// GroupCount counts elements per key extracted by `key`, returning map[K]int
// (lodash countBy). K is the method's own constrained type parameter.
func (s Seq[T]) GroupCount[K comparable](key func(T) K) map[K]int {
	out := make(map[K]int)
	for v := range iter.Seq[T](s) {
		out[key(v)]++
	}
	return out
}

// GroupBy groups elements by key extracted by `key`, returning map[K][]T in
// first-occurrence order within each group. K is the method's own constrained
// type parameter.
func (s Seq[T]) GroupBy[K comparable](key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for v := range iter.Seq[T](s) {
		k := key(v)
		out[k] = append(out[k], v)
	}
	return out
}

// PartitionBy splits s into groups keyed by `key`, yielding (key, group) pairs
// in the order each key first appears — unlike [Seq.GroupBy], whose map is
// unordered. Within a group, elements keep their relative order from s.
//
// Although the result is a [Seq2] (a lazy type), PartitionBy is internally
// materializing: it must consume all of s before yielding any group, because a
// group is not known to be complete until the source is exhausted. Do not call
// it on an unbounded sequence. K is comparable, so by the library's 划分铁律
// this is a free function with a method-level-style type parameter rather than
// a method on Seq[T].
func PartitionBy[T any, K comparable](s Seq[T], key func(T) K) Seq2[K, []T] {
	return Seq2[K, []T](func(yield func(K, []T) bool) {
		index := make(map[K]int)
		var order []K
		var groups [][]T
		for v := range iter.Seq[T](s) {
			k := key(v)
			i, ok := index[k]
			if !ok {
				i = len(groups)
				index[k] = i
				order = append(order, k)
				groups = append(groups, nil)
			}
			groups[i] = append(groups[i], v)
		}
		for i, k := range order {
			if !yield(k, groups[i]) {
				return
			}
		}
	})
}

// KeyBy indexes elements by key extracted by `key`, returning map[K]T. On
// duplicate keys the later element overwrites (lodash keyBy). K is the
// method's own constrained type parameter.
func (s Seq[T]) KeyBy[K comparable](key func(T) K) map[K]T {
	out := make(map[K]T)
	for v := range iter.Seq[T](s) {
		out[key(v)] = v
	}
	return out
}

// MaxBy returns the element for which less reports it as greatest, or None if
// empty.
func (s Seq[T]) MaxBy(less func(a, b T) bool) Optional[T] {
	var best T
	first := true
	for v := range iter.Seq[T](s) {
		if first {
			best = v
			first = false
			continue
		}
		if less(best, v) {
			best = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(best)
}

// MinBy returns the element for which less reports it as least, or None if
// empty.
func (s Seq[T]) MinBy(less func(a, b T) bool) Optional[T] {
	var best T
	first := true
	for v := range iter.Seq[T](s) {
		if first {
			best = v
			first = false
			continue
		}
		if less(v, best) {
			best = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(best)
}

// MaxByKey returns the element whose projected key is greatest (lodash maxBy).
// K is the method's own constrained type parameter (cmp.Ordered). Returns None
// if empty.
func (s Seq[T]) MaxByKey[K cmp.Ordered](key func(T) K) Optional[T] {
	var best T
	var bestKey K
	first := true
	for v := range iter.Seq[T](s) {
		k := key(v)
		if first {
			best, bestKey = v, k
			first = false
			continue
		}
		if cmp.Less(bestKey, k) {
			best, bestKey = v, k
		}
	}
	if first {
		return None[T]()
	}
	return Some(best)
}

// MinByKey returns the element whose projected key is least (lodash minBy).
// K is the method's own constrained type parameter (cmp.Ordered). Returns None
// if empty.
func (s Seq[T]) MinByKey[K cmp.Ordered](key func(T) K) Optional[T] {
	var best T
	var bestKey K
	first := true
	for v := range iter.Seq[T](s) {
		k := key(v)
		if first {
			best, bestKey = v, k
			first = false
			continue
		}
		if cmp.Less(k, bestKey) {
			best, bestKey = v, k
		}
	}
	if first {
		return None[T]()
	}
	return Some(best)
}

// SumBy sums the projected numeric value of each element (lodash sumBy). U is
// the method's own constrained type parameter (Numeric).
func (s Seq[T]) SumBy[U Numeric](f func(T) U) U {
	var sum U
	for v := range iter.Seq[T](s) {
		sum += f(v)
	}
	return sum
}

// MeanBy returns the arithmetic mean of the projected numeric values as a
// float64 (lodash meanBy), or 0 for an empty sequence.
func (s Seq[T]) MeanBy[U Numeric](f func(T) U) float64 {
	var sum U
	n := 0
	for v := range iter.Seq[T](s) {
		sum += f(v)
		n++
	}
	if n == 0 {
		return 0
	}
	return float64(sum) / float64(n)
}
