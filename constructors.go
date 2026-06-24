package seq

// Constructors — free functions that build a Seq[T] / Seq2[K,V] from various
// sources. They are free functions because they have no receiver; this is the
// entry point of every pipeline (see design doc "构造入口自由函数").
//
// Lazy semantics: every constructor returns an unevaluated Seq. The underlying
// generator only runs when a terminal operation drives iteration. Infinite
// sources (Generate, Iterate, RepeatInf) are therefore safe as long as a
// downstream Take/Slice/Find bounds the consumption.
//
// Re-iteration: slice-backed sources (From, Of, Empty, Range, RangeStep,
// Repeat) may be iterated more than once. Channel-backed sources (FromChannel)
// are one-shot — iterating again yields nothing once the channel is drained.

// From creates a Seq backed by a slice. The resulting Seq may be iterated
// repeatedly; it does not copy the slice, so mutation of the backing slice
// between iterations is visible.
func From[T any](s []T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	})
}

// Of creates a Seq from a variadic argument list.
func Of[T any](items ...T) Seq[T] {
	return From(items)
}

// Empty creates a Seq that yields no elements.
func Empty[T any]() Seq[T] {
	return Seq[T](func(yield func(T) bool) {})
}

// Range creates a Seq of the integers in [start, end), ascending when
// start < end and descending when start > end. An empty range yields nothing.
// A zero-width range (start == end) is empty.
func Range(start, end int) Seq[int] {
	return Seq[int](func(yield func(int) bool) {
		if start < end {
			for i := start; i < end; i++ {
				if !yield(i) {
					return
				}
			}
		} else if start > end {
			for i := start; i > end; i-- {
				if !yield(i) {
					return
				}
			}
		}
	})
}

// RangeStep creates a Seq of the integers in [start, end) advancing by step.
// A non-positive step yields an empty Seq (the operation is undefined for
// step <= 0; rather than panic we produce nothing, matching the "empty input,
// empty output" convention).
func RangeStep(start, end, step int) Seq[int] {
	return Seq[int](func(yield func(int) bool) {
		if step <= 0 {
			return
		}
		if start < end {
			for i := start; i < end; i += step {
				if !yield(i) {
					return
				}
			}
		} else if start > end {
			// Descending: step still positive, walk downward.
			for i := start; i > end; i -= step {
				if !yield(i) {
					return
				}
			}
		}
	})
}

// Repeat creates a Seq that yields v exactly n times. n <= 0 yields nothing.
func Repeat[T any](n int, v T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for i := 0; i < n; i++ {
			if !yield(v) {
				return
			}
		}
	})
}

// RepeatInf creates a Seq that yields v forever. It must be consumed by a
// bounding operation (e.g. Take); a terminal that drains it will not return.
func RepeatInf[T any](v T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for {
			if !yield(v) {
				return
			}
		}
	})
}

// Generate creates an infinite Seq by calling f to produce each element. It
// must be consumed by a bounding operation.
func Generate[T any](f func() T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for {
			if !yield(f()) {
				return
			}
		}
	})
}

// Iterate creates an infinite Seq: init, f(init), f(f(init)), ... It must be
// consumed by a bounding operation.
func Iterate[T any](init T, f func(T) T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		v := init
		for {
			if !yield(v) {
				return
			}
			v = f(v)
		}
	})
}

// FromChannel creates a Seq that drains a receive-only channel. It is a
// one-shot source: after the channel is closed/drained, iterating again yields
// nothing.
func FromChannel[T any](ch <-chan T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	})
}

// FromMap creates a Seq2 from a map's key/value pairs. Iteration order follows
// Go's map range (non-deterministic), matching maps.All.
func FromMap[K comparable, V any](m map[K]V) Seq2[K, V] {
	return Seq2[K, V](func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	})
}
