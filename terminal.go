package seq

import (
	"iter"
	"slices"
	"strings"
)

// Terminal methods on Seq[T]. These drive evaluation and return a non-Seq
// result. Several use Go 1.27 method-level type parameters (Fold[U],
// GroupCount[K], GroupBy[K], KeyBy[K], MaxByKey/MinByKey[K], SumBy/MeanBy[U]).

// Collect materializes the Seq into a slice. It drives the full pipeline.
func (s Seq[T]) Collect() []T {
	return slices.Collect(iter.Seq[T](s))
}

// ForEach invokes f for each element.
func (s Seq[T]) ForEach(f func(T)) {
	for v := range iter.Seq[T](s) {
		f(v)
	}
}

// ForEachIndexed invokes f with the zero-based index and each element.
func (s Seq[T]) ForEachIndexed(f func(int, T)) {
	i := 0
	for v := range iter.Seq[T](s) {
		f(i, v)
		i++
	}
}

// Reduce folds the sequence with f, starting from the first element. An empty
// sequence returns None.
func (s Seq[T]) Reduce(f func(T, T) T) Optional[T] {
	var acc T
	first := true
	for v := range iter.Seq[T](s) {
		if first {
			acc = v
			first = false
			continue
		}
		acc = f(acc, v)
	}
	if first {
		return None[T]()
	}
	return Some(acc)
}

// Count returns the number of elements.
func (s Seq[T]) Count() int {
	n := 0
	for range iter.Seq[T](s) {
		n++
	}
	return n
}

// CountBy returns the number of elements satisfying pred.
func (s Seq[T]) CountBy(pred func(T) bool) int {
	n := 0
	for v := range iter.Seq[T](s) {
		if pred(v) {
			n++
		}
	}
	return n
}

// Find returns the first element satisfying pred, or None.
func (s Seq[T]) Find(pred func(T) bool) Optional[T] {
	for v := range iter.Seq[T](s) {
		if pred(v) {
			return Some(v)
		}
	}
	return None[T]()
}

// FindIndex returns the index of the first element satisfying pred, or None.
func (s Seq[T]) FindIndex(pred func(T) bool) Optional[int] {
	i := 0
	for v := range iter.Seq[T](s) {
		if pred(v) {
			return Some(i)
		}
		i++
	}
	return None[int]()
}

// FindLast returns the last element satisfying pred (lodash findLast). It must
// scan the whole sequence to find the last match. Returns None if none match.
func (s Seq[T]) FindLast(pred func(T) bool) Optional[T] {
	var found T
	ok := false
	for v := range iter.Seq[T](s) {
		if pred(v) {
			found = v
			ok = true
		}
	}
	if !ok {
		return None[T]()
	}
	return Some(found)
}

// FindLastIndex returns the index of the last element satisfying pred, or None.
func (s Seq[T]) FindLastIndex(pred func(T) bool) Optional[int] {
	idx := 0
	ok := false
	i := 0
	for v := range iter.Seq[T](s) {
		if pred(v) {
			idx = i
			ok = true
		}
		i++
	}
	if !ok {
		return None[int]()
	}
	return Some(idx)
}

// Any reports whether any element satisfies pred (short-circuits on first
// match; Scala exists).
func (s Seq[T]) Any(pred func(T) bool) bool {
	for v := range iter.Seq[T](s) {
		if pred(v) {
			return true
		}
	}
	return false
}

// All reports whether every element satisfies pred (short-circuits on first
// failure; Scala forall).
func (s Seq[T]) All(pred func(T) bool) bool {
	for v := range iter.Seq[T](s) {
		if !pred(v) {
			return false
		}
	}
	return true
}

// None reports whether no element satisfies pred (short-circuits on first
// match).
func (s Seq[T]) None(pred func(T) bool) bool {
	for v := range iter.Seq[T](s) {
		if pred(v) {
			return false
		}
	}
	return true
}

// First returns the first element, or None if empty.
func (s Seq[T]) First() Optional[T] {
	for v := range iter.Seq[T](s) {
		return Some(v)
	}
	return None[T]()
}

// Last returns the last element. It must scan the whole sequence. Returns
// None if empty.
func (s Seq[T]) Last() Optional[T] {
	var last T
	ok := false
	for v := range iter.Seq[T](s) {
		last = v
		ok = true
	}
	if !ok {
		return None[T]()
	}
	return Some(last)
}

// Nth returns the zero-based nth element, or None if out of range. n < 0 is
// out of range.
func (s Seq[T]) Nth(n int) Optional[T] {
	if n < 0 {
		return None[T]()
	}
	i := 0
	for v := range iter.Seq[T](s) {
		if i == n {
			return Some(v)
		}
		i++
	}
	return None[T]()
}

// IsEmpty reports whether the sequence yields no elements. It short-circuits
// on the first element.
func (s Seq[T]) IsEmpty() bool {
	for range iter.Seq[T](s) {
		return false
	}
	return true
}

// Partition splits the sequence into two slices: those satisfying pred and
// those not.
func (s Seq[T]) Partition(pred func(T) bool) ([]T, []T) {
	var yes, no []T
	for v := range iter.Seq[T](s) {
		if pred(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no
}

// Span splits the sequence at the first element failing pred (Scala span):
// the first run of matching elements, then the rest.
func (s Seq[T]) Span(pred func(T) bool) ([]T, []T) {
	var prefix []T
	var rest []T
	split := false
	for v := range iter.Seq[T](s) {
		if split {
			rest = append(rest, v)
			continue
		}
		if pred(v) {
			prefix = append(prefix, v)
		} else {
			rest = append(rest, v)
			split = true
		}
	}
	return prefix, rest
}

// Join converts each element to a string via str and joins them with sep
// (Scala mkString; formerly JoinFunc). An empty sequence yields "".
func (s Seq[T]) Join(sep string, str func(T) string) string {
	var b strings.Builder
	first := true
	for v := range iter.Seq[T](s) {
		if !first {
			b.WriteString(sep)
		}
		first = false
		b.WriteString(str(v))
	}
	return b.String()
}
