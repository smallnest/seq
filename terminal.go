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
// sequence returns (zero, false).
func (s Seq[T]) Reduce(f func(T, T) T) (T, bool) {
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
	return acc, !first
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

// Find returns the first element satisfying pred, or (zero, false).
func (s Seq[T]) Find(pred func(T) bool) (T, bool) {
	for v := range iter.Seq[T](s) {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first element satisfying pred, or
// (0, false).
func (s Seq[T]) FindIndex(pred func(T) bool) (int, bool) {
	i := 0
	for v := range iter.Seq[T](s) {
		if pred(v) {
			return i, true
		}
		i++
	}
	return 0, false
}

// FindLast returns the last element satisfying pred (lodash findLast). It must
// scan the whole sequence to find the last match. Returns (zero, false) if
// none match.
func (s Seq[T]) FindLast(pred func(T) bool) (T, bool) {
	var found T
	ok := false
	for v := range iter.Seq[T](s) {
		if pred(v) {
			found = v
			ok = true
		}
	}
	return found, ok
}

// FindLastIndex returns the index of the last element satisfying pred, or
// (0, false).
func (s Seq[T]) FindLastIndex(pred func(T) bool) (int, bool) {
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
	return idx, ok
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

// First returns the first element, or (zero, false) if empty.
func (s Seq[T]) First() (T, bool) {
	for v := range iter.Seq[T](s) {
		return v, true
	}
	var zero T
	return zero, false
}

// Last returns the last element. It must scan the whole sequence. Returns
// (zero, false) if empty.
func (s Seq[T]) Last() (T, bool) {
	var last T
	ok := false
	for v := range iter.Seq[T](s) {
		last = v
		ok = true
	}
	return last, ok
}

// Nth returns the zero-based nth element, or (zero, false) if out of range.
// n < 0 is out of range.
func (s Seq[T]) Nth(n int) (T, bool) {
	if n < 0 {
		var zero T
		return zero, false
	}
	i := 0
	for v := range iter.Seq[T](s) {
		if i == n {
			return v, true
		}
		i++
	}
	var zero T
	return zero, false
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
