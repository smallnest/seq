package seq

import (
	"iter"
	"slices"
)

// Materializing intermediate methods. These need to inspect the whole sequence
// (or its tail) to produce their output, so they collect once and return a
// slice-backed (re-iterable) Seq. Per the PRD they are marked "内部物化".

// TakeRight yields the last n elements. n <= 0 yields nothing; n greater than
// the length yields everything. Materializes internally.
func (s Seq[T]) TakeRight(n int) Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	if n <= 0 {
		return Empty[T]()
	}
	if n >= len(all) {
		return From(all)
	}
	return From(all[len(all)-n:])
}

// DropRight drops the last n elements and yields the rest. n <= 0 yields
// everything; n >= length yields nothing. Materializes internally.
func (s Seq[T]) DropRight(n int) Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	if n <= 0 {
		return From(all)
	}
	if n >= len(all) {
		return Empty[T]()
	}
	return From(all[:len(all)-n])
}

// Slice yields the half-open sub-range [start, end) of the materialized
// sequence, clamped to bounds (lodash slice semantics). start is clamped to
// [0, len]; end is clamped to [start, len].
func (s Seq[T]) Slice(start, end int) Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	if start < 0 {
		start = 0
	}
	if start > len(all) {
		start = len(all)
	}
	if end < start {
		end = start
	}
	if end > len(all) {
		end = len(all)
	}
	return From(all[start:end])
}

// Init yields all elements except the last (Scala init). An empty sequence
// stays empty. Materializes internally.
func (s Seq[T]) Init() Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	if len(all) == 0 {
		return Empty[T]()
	}
	return From(all[:len(all)-1])
}

// Tail yields all elements except the first (Scala tail, equivalent to
// Drop(1)). An empty sequence stays empty. Kept as a method for parity with
// Init; materializes internally for a stable re-iterable result.
func (s Seq[T]) Tail() Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	if len(all) == 0 {
		return Empty[T]()
	}
	return From(all[1:])
}

// SortBy returns a Seq sorted by the less comparator (ascending when less is
// a < b). Materializes internally; stable (uses slices.SortStableFunc). The
// bool comparator is adapted to slices' cmp-int form.
func (s Seq[T]) SortBy(less func(a, b T) bool) Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	slices.SortStableFunc(all, func(a, b T) int {
		switch {
		case less(a, b):
			return -1
		case less(b, a):
			return 1
		default:
			return 0
		}
	})
	return From(all)
}

// Reverse yields the elements in reverse order. Materializes internally.
func (s Seq[T]) Reverse() Seq[T] {
	all := slices.Collect(iter.Seq[T](s))
	slices.Reverse(all)
	return From(all)
}

// Enumerate pairs each element with its zero-based index, yielding a
// Seq2[int, T] (Scala zipWithIndex).
func (s Seq[T]) Enumerate() Seq2[int, T] {
	return Seq2[int, T](func(yield func(int, T) bool) {
		i := 0
		for v := range iter.Seq[T](s) {
			if !yield(i, v) {
				return
			}
			i++
		}
	})
}
