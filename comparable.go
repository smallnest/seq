package seq

import (
	"iter"
	"slices"
	"strings"
)

// Comparable free functions. These constrain T to comparable (or cmp.Ordered
// for Max/Min/Sort, Numeric for Sum/Product/Mean — see numeric.go). By the
// 划分铁律 they cannot be methods on Seq[T any], so they are free functions.
// Their semantics match the SeqComparable subtype methods (issue #10).

// Distinct returns a Seq that yields each distinct element once, preserving
// first-occurrence order.
func Distinct[T comparable](s Seq[T]) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range iter.Seq[T](s) {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	})
}

// Contains reports whether v occurs in s. It short-circuits on the first match.
func Contains[T comparable](s Seq[T], v T) bool {
	for x := range iter.Seq[T](s) {
		if x == v {
			return true
		}
	}
	return false
}

// IndexOf returns the first index of v in s, or None if v is absent.
func IndexOf[T comparable](s Seq[T], v T) Optional[int] {
	i := 0
	for x := range iter.Seq[T](s) {
		if x == v {
			return Some(i)
		}
		i++
	}
	return None[int]()
}

// LastIndexOf returns the last index of v in s, or None if v is absent. It
// materializes the input internally to scan from the end.
func LastIndexOf[T comparable](s Seq[T], v T) Optional[int] {
	collected := slices.Collect(iter.Seq[T](s))
	for i := len(collected) - 1; i >= 0; i-- {
		if collected[i] == v {
			return Some(i)
		}
	}
	return None[int]()
}

// CountValues returns a map from each element to the number of times it
// occurs. An empty sequence yields an empty (non-nil) map.
func CountValues[T comparable](s Seq[T]) map[T]int {
	counts := make(map[T]int)
	for v := range iter.Seq[T](s) {
		counts[v]++
	}
	return counts
}

// Equal reports whether a and b yield the same elements in the same order. Two
// empty sequences are equal.
func Equal[T comparable](a, b Seq[T]) bool {
	nextA, stopA := iter.Pull(iter.Seq[T](a))
	defer stopA()
	nextB, stopB := iter.Pull(iter.Seq[T](b))
	defer stopB()
	for {
		va, okA := nextA()
		vb, okB := nextB()
		if okA != okB {
			return false
		}
		if !okA {
			return true
		}
		if va != vb {
			return false
		}
	}
}

// Compact returns a Seq that drops zero-value elements (lodash compact). A
// zero-value element equals the zero value of T.
func Compact[T comparable](s Seq[T]) Seq[T] {
	var zero T
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			if v == zero {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// Without returns a Seq that excludes any element equal to one of vals
// (lodash without).
func Without[T comparable](s Seq[T], vals ...T) Seq[T] {
	exclude := make(map[T]struct{}, len(vals))
	for _, v := range vals {
		exclude[v] = struct{}{}
	}
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			if _, drop := exclude[v]; drop {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// ToSet returns a set (map[T]struct{}) of the distinct elements of s. An empty
// sequence yields an empty (non-nil) map.
func ToSet[T comparable](s Seq[T]) map[T]struct{} {
	set := make(map[T]struct{})
	for v := range iter.Seq[T](s) {
		set[v] = struct{}{}
	}
	return set
}

// JoinStrings concatenates the elements of a string Seq, separated by sep. An
// empty sequence yields "". This is the free-function form of the Seq.Join
// method specialized for strings (no per-element formatting).
func JoinStrings(s Seq[string], sep string) string {
	var b strings.Builder
	first := true
	for v := range iter.Seq[string](s) {
		if !first {
			b.WriteString(sep)
		}
		first = false
		b.WriteString(v)
	}
	return b.String()
}

// Union returns a Seq of the distinct elements across all seqs, preserving
// first-occurrence order across the concatenation. An empty input yields an
// empty Seq.
func Union[T comparable](seqs ...Seq[T]) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for _, s := range seqs {
			for v := range iter.Seq[T](s) {
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Intersect returns a Seq of elements present in both a and b, preserving a's
// order and dropping duplicates (set intersection).
func Intersect[T comparable](a, b Seq[T]) Seq[T] {
	bset := ToSet(b)
	return Seq[T](func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range iter.Seq[T](a) {
			if _, ok := bset[v]; !ok {
				continue
			}
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	})
}

// Difference returns a Seq of elements in a but not in b, preserving a's order
// and dropping duplicates (set difference a − b).
func Difference[T comparable](a, b Seq[T]) Seq[T] {
	bset := ToSet(b)
	return Seq[T](func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range iter.Seq[T](a) {
			if _, inB := bset[v]; inB {
				continue
			}
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	})
}

// SymmetricDifference returns a Seq of elements in exactly one of a or b
// (lodash xor), preserving first-occurrence order across a then b and dropping
// duplicates.
func SymmetricDifference[T comparable](a, b Seq[T]) Seq[T] {
	aset := ToSet(a)
	bset := ToSet(b)
	return Seq[T](func(yield func(T) bool) {
		seen := make(map[T]struct{})
		emit := func(v T) bool {
			if _, dup := seen[v]; dup {
				return true
			}
			seen[v] = struct{}{}
			return yield(v)
		}
		for v := range iter.Seq[T](a) {
			if _, inB := bset[v]; inB {
				continue
			}
			if !emit(v) {
				return
			}
		}
		for v := range iter.Seq[T](b) {
			if _, inA := aset[v]; inA {
				continue
			}
			if !emit(v) {
				return
			}
		}
	})
}
