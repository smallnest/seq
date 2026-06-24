package seq

import (
	"iter"
	"slices"
)

// Intermediate transform methods on Seq[T]. These are lazy: each returns a new
// Seq that is not evaluated until a terminal operation drives iteration. The
// type-changing ones (Map, FlatMap, FilterMap, Scan) use Go 1.27 method-level
// type parameters (golang/go#77273); the T-preserving ones re-use T.

// Map transforms each element T into U via f, returning a Seq[U]. This is the
// headline Go 1.27 generic method: U is the method's own type parameter.
func (s Seq[T]) Map[U any](f func(T) U) Seq[U] {
	return Seq[U](func(yield func(U) bool) {
		for v := range iter.Seq[T](s) {
			if !yield(f(v)) {
				return
			}
		}
	})
}

// FlatMap maps each element to a sub-Seq[U] via f and flattens one level.
func (s Seq[T]) FlatMap[U any](f func(T) Seq[U]) Seq[U] {
	return Seq[U](func(yield func(U) bool) {
		for v := range iter.Seq[T](s) {
			for u := range iter.Seq[U](f(v)) {
				if !yield(u) {
					return
				}
			}
		}
	})
}

// FilterMap applies f to each element; pairs whose f returns (u, true) are
// kept and mapped to u, others are dropped. Combines Filter + Map in one pass.
func (s Seq[T]) FilterMap[U any](f func(T) (U, bool)) Seq[U] {
	return Seq[U](func(yield func(U) bool) {
		for v := range iter.Seq[T](s) {
			u, ok := f(v)
			if !ok {
				continue
			}
			if !yield(u) {
				return
			}
		}
	})
}

// Filter keeps elements satisfying pred.
func (s Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			if !pred(v) {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// Reject drops elements satisfying pred (the inverse of Filter).
func (s Seq[T]) Reject(pred func(T) bool) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			if pred(v) {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// Take yields at most the first n elements. n <= 0 yields nothing.
func (s Seq[T]) Take(n int) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		count := 0
		for v := range iter.Seq[T](s) {
			if !yield(v) {
				return
			}
			count++
			if count >= n {
				return
			}
		}
	})
}

// Drop skips the first n elements and yields the rest. n <= 0 yields all.
func (s Seq[T]) Drop(n int) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		skipped := 0
		for v := range iter.Seq[T](s) {
			if skipped < n {
				skipped++
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// TakeWhile yields elements until the first one that fails pred (which is not
// yielded), then stops.
func (s Seq[T]) TakeWhile(pred func(T) bool) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			if !pred(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	})
}

// DropWhile skips leading elements that satisfy pred, then yields the rest
// (including any later elements that happen to satisfy pred).
func (s Seq[T]) DropWhile(pred func(T) bool) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		dropping := true
		for v := range iter.Seq[T](s) {
			if dropping && pred(v) {
				continue
			}
			dropping = false
			if !yield(v) {
				return
			}
		}
	})
}

// Scan emits each running accumulation, including the initial value: init,
// f(init, x0), f(f(init, x0), x1), ... This is a left scan (prefix sums).
func (s Seq[T]) Scan[U any](init U, f func(U, T) U) Seq[U] {
	return Seq[U](func(yield func(U) bool) {
		acc := init
		if !yield(acc) {
			return
		}
		for v := range iter.Seq[T](s) {
			acc = f(acc, v)
			if !yield(acc) {
				return
			}
		}
	})
}

// Peek invokes f for each element as a side effect (e.g. logging) and passes
// the element through unchanged. f's return value is ignored.
func (s Seq[T]) Peek(f func(T)) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			f(v)
			if !yield(v) {
				return
			}
		}
	})
}

// Intersperse inserts sep between each pair of elements.
func (s Seq[T]) Intersperse(sep T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		first := true
		for v := range iter.Seq[T](s) {
			if !first {
				if !yield(sep) {
					return
				}
			}
			first = false
			if !yield(v) {
				return
			}
		}
	})
}

// Concat appends another Seq[T] after this one. (Variadic multi-source
// concatenation is the free function Concat; this is the 2-source method form.)
func (s Seq[T]) Concat(other Seq[T]) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](s) {
			if !yield(v) {
				return
			}
		}
		for v := range iter.Seq[T](other) {
			if !yield(v) {
				return
			}
		}
	})
}

// DistinctBy deduplicates by a key extracted from each element. K is the
// method's own constrained type parameter (comparable); first occurrence wins.
func (s Seq[T]) DistinctBy[K comparable](key func(T) K) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		seen := make(map[K]struct{})
		for v := range iter.Seq[T](s) {
			k := key(v)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	})
}

// Chunk splits the sequence into non-overlapping runs of size `size` (the last
// may be shorter). size <= 0 yields nothing.
//
// The return type is iter.Seq[Seq[T]] rather than Seq[Seq[T]] because Go 1.27's
// generic methods forbid a method on Seq[T] from instantiating Seq[Seq[T]]
// (instantiation cycle: T would be instantiated as Seq[T]). The yielded inner
// values are full Seq[T], so each chunk retains the complete method set — the
// only thing lost is chaining on the *outer* sequence without an explicit
// Seq[Seq[T]](s.Chunk(n)) conversion. Wrap with Seq[Seq[T]](s.Chunk(n)) if you
// need outer chaining.
func (s Seq[T]) Chunk(size int) iter.Seq[Seq[T]] {
	return func(yield func(Seq[T]) bool) {
		if size <= 0 {
			return
		}
		var buf []T
		for v := range iter.Seq[T](s) {
			buf = append(buf, v)
			if len(buf) == size {
				if !yield(From(buf)) {
					return
				}
				buf = nil
			}
		}
		if len(buf) > 0 {
			yield(From(buf))
		}
	}
}

// Window yields sliding windows of `size` elements advancing by `step`.
// step <= 0 yields nothing; size <= 0 yields nothing. A window is emitted as
// long as a full `size` elements are available from the current start.
//
// Like Chunk, Window returns iter.Seq[Seq[T]] (not Seq[Seq[T]]) due to the
// Go 1.27 instantiation-cycle limit on generic methods. Inner windows are full
// Seq[T]; wrap with Seq[Seq[T]](s.Window(...)) for outer chaining.
func (s Seq[T]) Window(size, step int) iter.Seq[Seq[T]] {
	return func(yield func(Seq[T]) bool) {
		if size <= 0 || step <= 0 {
			return
		}
		// Materialize once so we can re-scan for overlapping windows.
		all := slices.Collect(iter.Seq[T](s))
		for start := 0; start+size <= len(all); start += step {
			if !yield(From(all[start : start+size])) {
				return
			}
		}
	}
}
