package seq

import (
	"cmp"
	"iter"
)

// Constrained subtypes (FR-8). These pin a constraint onto T at the type
// level, so the constrained operations (Distinct/Max/Sum/...) that must be
// free functions on bare Seq[T any] become methods here, restoring the
// left-to-right chain:
//
//	seq.Numbers(seq.From([]int{1,2,3})).Distinct().Sum()
//
// The subtypes are defined types over iter.Seq[T], so they interconvert with
// iter.Seq[T] at zero cost. Entering a subtype is a free function (constrains
// T); downgrading between subtypes is a method (a stronger constraint already
// satisfies a weaker one — Numeric ⊂ Ordered ⊂ comparable).

// SeqComparable is a Seq whose element T is comparable. Its methods (Distinct,
// Contains, etc.) are the constrained operations that are free functions on
// bare Seq[T].
type SeqComparable[T comparable] iter.Seq[T]

// SeqOrdered is a Seq whose element T is cmp.Ordered. It inherits the
// comparable operations plus Max/Min/Sort.
type SeqOrdered[T cmp.Ordered] iter.Seq[T]

// SeqNumeric is a Seq whose element T is Numeric. It inherits the ordered and
// comparable operations plus Sum/Product/Mean. (The type is named SeqNumeric;
// the entry function is Numbers to avoid clashing with the Numeric constraint.)
type SeqNumeric[T Numeric] iter.Seq[T]

// --- Entry free functions (constrain T, so they must be free functions) ---

// Comparable converts a Seq[T] (T comparable) into a SeqComparable[T], enabling
// Distinct/Contains/... as methods.
func Comparable[T comparable](s Seq[T]) SeqComparable[T] {
	return SeqComparable[T](iter.Seq[T](s))
}

// Ordered converts a Seq[T] (T cmp.Ordered) into a SeqOrdered[T].
func Ordered[T cmp.Ordered](s Seq[T]) SeqOrdered[T] {
	return SeqOrdered[T](iter.Seq[T](s))
}

// Numbers converts a Seq[T] (T Numeric) into a SeqNumeric[T].
func Numbers[T Numeric](s Seq[T]) SeqNumeric[T] {
	return SeqNumeric[T](iter.Seq[T](s))
}

// --- Downgrade methods (strong constraint satisfies weak, so methods) ---

// Ordered downgrades a SeqNumeric[T] to SeqOrdered[T] (Numeric satisfies
// cmp.Ordered).
func (s SeqNumeric[T]) Ordered() SeqOrdered[T] {
	return SeqOrdered[T](iter.Seq[T](s))
}

// Comparable downgrades a SeqOrdered[T] to SeqComparable[T] (Ordered satisfies
// comparable).
func (s SeqOrdered[T]) Comparable() SeqComparable[T] {
	return SeqComparable[T](iter.Seq[T](s))
}

// Seq drops back to the bare Seq[T], e.g. before a Map that changes T (which is
// not offered on the subtypes).
func (s SeqComparable[T]) Seq() Seq[T] {
	return Seq[T](iter.Seq[T](s))
}

// --- SeqComparable[T] methods (constrained: comparable) ---

// Distinct yields each distinct element once, preserving first-occurrence order.
func (s SeqComparable[T]) Distinct() SeqComparable[T] {
	return SeqComparable[T](Distinct(s.Seq()))
}

// Contains reports whether v occurs in s. Short-circuits.
func (s SeqComparable[T]) Contains(v T) bool {
	return Contains(s.Seq(), v)
}

// IndexOf returns the first index of v, or (0, false) if absent.
func (s SeqComparable[T]) IndexOf(v T) (int, bool) {
	return IndexOf(s.Seq(), v)
}

// CountValues returns a map of element -> count.
func (s SeqComparable[T]) CountValues() map[T]int {
	return CountValues(s.Seq())
}

// ToSet returns a set (map[T]struct{}) of the distinct elements.
func (s SeqComparable[T]) ToSet() map[T]struct{} {
	return ToSet(s.Seq())
}

// Union returns the distinct union with others, preserving first-occurrence
// order. Returns a SeqComparable so the chain continues.
func (s SeqComparable[T]) Union(others ...SeqComparable[T]) SeqComparable[T] {
	seqs := make([]Seq[T], 0, len(others)+1)
	seqs = append(seqs, s.Seq())
	for _, o := range others {
		seqs = append(seqs, o.Seq())
	}
	return SeqComparable[T](Union(seqs...))
}

// Intersect returns the intersection with other (set semantics, dedup).
func (s SeqComparable[T]) Intersect(other SeqComparable[T]) SeqComparable[T] {
	return SeqComparable[T](Intersect(s.Seq(), other.Seq()))
}

// Difference returns the set difference s − other.
func (s SeqComparable[T]) Difference(other SeqComparable[T]) SeqComparable[T] {
	return SeqComparable[T](Difference(s.Seq(), other.Seq()))
}

// Equal reports element-wise equality with other (same order).
func (s SeqComparable[T]) Equal(other SeqComparable[T]) bool {
	return Equal(s.Seq(), other.Seq())
}

// --- T-preserving intermediate methods re-exposed on SeqComparable[T] ---

// Filter keeps elements satisfying pred; returns SeqComparable so the chain
// of constrained ops continues.
func (s SeqComparable[T]) Filter(pred func(T) bool) SeqComparable[T] {
	return SeqComparable[T](s.Seq().Filter(pred))
}

// Reject drops elements satisfying pred.
func (s SeqComparable[T]) Reject(pred func(T) bool) SeqComparable[T] {
	return SeqComparable[T](s.Seq().Reject(pred))
}

// Take yields at most the first n elements.
func (s SeqComparable[T]) Take(n int) SeqComparable[T] {
	return SeqComparable[T](s.Seq().Take(n))
}

// Drop skips the first n elements.
func (s SeqComparable[T]) Drop(n int) SeqComparable[T] {
	return SeqComparable[T](s.Seq().Drop(n))
}

// TakeWhile yields until the first element failing pred.
func (s SeqComparable[T]) TakeWhile(pred func(T) bool) SeqComparable[T] {
	return SeqComparable[T](s.Seq().TakeWhile(pred))
}

// DropWhile skips leading elements satisfying pred.
func (s SeqComparable[T]) DropWhile(pred func(T) bool) SeqComparable[T] {
	return SeqComparable[T](s.Seq().DropWhile(pred))
}

// Peek invokes f per element as a side effect, passing elements through.
func (s SeqComparable[T]) Peek(f func(T)) SeqComparable[T] {
	return SeqComparable[T](s.Seq().Peek(f))
}

// Replace lazily replaces the first n elements equal to old with new, passing
// all other elements through unchanged in order. n < 0 means replace every
// match (equivalent to [SeqComparable.ReplaceAll]); n == 0 replaces nothing.
func (s SeqComparable[T]) Replace(old, new T, n int) SeqComparable[T] {
	return SeqComparable[T](iter.Seq[T](func(yield func(T) bool) {
		replaced := 0
		for v := range iter.Seq[T](s) {
			if v == old && (n < 0 || replaced < n) {
				v = new
				replaced++
			}
			if !yield(v) {
				return
			}
		}
	}))
}

// ReplaceAll lazily replaces every element equal to old with new, passing all
// other elements through unchanged. It is Replace(old, new, -1).
func (s SeqComparable[T]) ReplaceAll(old, new T) SeqComparable[T] {
	return s.Replace(old, new, -1)
}

// --- SeqOrdered[T] methods (constrained: cmp.Ordered) ---

// Max returns the maximum element, or (zero, false) if empty.
func (s SeqOrdered[T]) Max() (T, bool) {
	return Max(s.Comparable().Seq())
}

// Min returns the minimum element, or (zero, false) if empty.
func (s SeqOrdered[T]) Min() (T, bool) {
	return Min(s.Comparable().Seq())
}

// Sort returns the elements in ascending order as a SeqOrdered (re-iterable).
func (s SeqOrdered[T]) Sort() SeqOrdered[T] {
	return SeqOrdered[T](Sort(s.Comparable().Seq()))
}

// Distinct yields each distinct element once (inherited from comparable).
func (s SeqOrdered[T]) Distinct() SeqOrdered[T] {
	return SeqOrdered[T](Distinct(s.Comparable().Seq()))
}

// Contains reports whether v occurs (inherited from comparable; short-circuit).
func (s SeqOrdered[T]) Contains(v T) bool {
	return Contains(s.Comparable().Seq(), v)
}

// IndexOf returns the first index of v (inherited from comparable).
func (s SeqOrdered[T]) IndexOf(v T) (int, bool) {
	return IndexOf(s.Comparable().Seq(), v)
}

// CountValues returns element -> count (inherited from comparable).
func (s SeqOrdered[T]) CountValues() map[T]int {
	return CountValues(s.Comparable().Seq())
}

// ToSet returns a set of distinct elements (inherited from comparable).
func (s SeqOrdered[T]) ToSet() map[T]struct{} {
	return ToSet(s.Comparable().Seq())
}

// Union returns the distinct union with others (inherited from comparable).
func (s SeqOrdered[T]) Union(others ...SeqOrdered[T]) SeqOrdered[T] {
	seqs := make([]Seq[T], 0, len(others)+1)
	seqs = append(seqs, s.Comparable().Seq())
	for _, o := range others {
		seqs = append(seqs, o.Comparable().Seq())
	}
	return SeqOrdered[T](Union(seqs...))
}

// Intersect returns the intersection with other (inherited from comparable).
func (s SeqOrdered[T]) Intersect(other SeqOrdered[T]) SeqOrdered[T] {
	return SeqOrdered[T](Intersect(s.Comparable().Seq(), other.Comparable().Seq()))
}

// Difference returns the set difference s − other (inherited from comparable).
func (s SeqOrdered[T]) Difference(other SeqOrdered[T]) SeqOrdered[T] {
	return SeqOrdered[T](Difference(s.Comparable().Seq(), other.Comparable().Seq()))
}

// Equal reports element-wise equality with other (inherited from comparable).
func (s SeqOrdered[T]) Equal(other SeqOrdered[T]) bool {
	return Equal(s.Comparable().Seq(), other.Comparable().Seq())
}

// --- SeqNumeric[T] methods (constrained: Numeric) ---

// Sum returns the sum of all elements (zero for empty).
func (s SeqNumeric[T]) Sum() T {
	return Sum(s.Ordered().Comparable().Seq())
}

// Max returns the maximum element (inherited from ordered), or (zero,false).
func (s SeqNumeric[T]) Max() (T, bool) {
	return Max(s.Ordered().Comparable().Seq())
}

// Min returns the minimum element (inherited from ordered), or (zero,false).
func (s SeqNumeric[T]) Min() (T, bool) {
	return Min(s.Ordered().Comparable().Seq())
}

// Sort returns the elements in ascending order as a SeqNumeric (inherited
// from ordered).
func (s SeqNumeric[T]) Sort() SeqNumeric[T] {
	return SeqNumeric[T](Sort(s.Ordered().Comparable().Seq()))
}

// Distinct yields each distinct element once (inherited from comparable).
func (s SeqNumeric[T]) Distinct() SeqNumeric[T] {
	return SeqNumeric[T](Distinct(s.Ordered().Comparable().Seq()))
}

// Contains reports whether v occurs (inherited from comparable).
func (s SeqNumeric[T]) Contains(v T) bool {
	return Contains(s.Ordered().Comparable().Seq(), v)
}

// IndexOf returns the first index of v (inherited from comparable).
func (s SeqNumeric[T]) IndexOf(v T) (int, bool) {
	return IndexOf(s.Ordered().Comparable().Seq(), v)
}

// CountValues returns element -> count (inherited from comparable).
func (s SeqNumeric[T]) CountValues() map[T]int {
	return CountValues(s.Ordered().Comparable().Seq())
}

// ToSet returns a set of distinct elements (inherited from comparable).
func (s SeqNumeric[T]) ToSet() map[T]struct{} {
	return ToSet(s.Ordered().Comparable().Seq())
}

// Union returns the distinct union with others (inherited from comparable).
func (s SeqNumeric[T]) Union(others ...SeqNumeric[T]) SeqNumeric[T] {
	seqs := make([]Seq[T], 0, len(others)+1)
	seqs = append(seqs, s.Ordered().Comparable().Seq())
	for _, o := range others {
		seqs = append(seqs, o.Ordered().Comparable().Seq())
	}
	return SeqNumeric[T](Union(seqs...))
}

// Intersect returns the intersection with other (inherited from comparable).
func (s SeqNumeric[T]) Intersect(other SeqNumeric[T]) SeqNumeric[T] {
	return SeqNumeric[T](Intersect(s.Ordered().Comparable().Seq(), other.Ordered().Comparable().Seq()))
}

// Difference returns the set difference s − other (inherited from comparable).
func (s SeqNumeric[T]) Difference(other SeqNumeric[T]) SeqNumeric[T] {
	return SeqNumeric[T](Difference(s.Ordered().Comparable().Seq(), other.Ordered().Comparable().Seq()))
}

// Equal reports element-wise equality with other (inherited from comparable).
func (s SeqNumeric[T]) Equal(other SeqNumeric[T]) bool {
	return Equal(s.Ordered().Comparable().Seq(), other.Ordered().Comparable().Seq())
}

// Product returns the product of all elements (1 for empty).
func (s SeqNumeric[T]) Product() T {
	return Product(s.Ordered().Comparable().Seq())
}

// Mean returns the arithmetic mean as float64 (0 for empty).
func (s SeqNumeric[T]) Mean() float64 {
	return Mean(s.Ordered().Comparable().Seq())
}

// --- T-preserving intermediate methods re-exposed on SeqOrdered[T] ---

// Filter keeps elements satisfying pred; returns SeqOrdered.
func (s SeqOrdered[T]) Filter(pred func(T) bool) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().Filter(pred))
}

// Reject drops elements satisfying pred.
func (s SeqOrdered[T]) Reject(pred func(T) bool) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().Reject(pred))
}

// Take yields at most the first n elements.
func (s SeqOrdered[T]) Take(n int) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().Take(n))
}

// Drop skips the first n elements.
func (s SeqOrdered[T]) Drop(n int) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().Drop(n))
}

// TakeWhile yields until the first element failing pred.
func (s SeqOrdered[T]) TakeWhile(pred func(T) bool) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().TakeWhile(pred))
}

// DropWhile skips leading elements satisfying pred.
func (s SeqOrdered[T]) DropWhile(pred func(T) bool) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().DropWhile(pred))
}

// Peek invokes f per element as a side effect.
func (s SeqOrdered[T]) Peek(f func(T)) SeqOrdered[T] {
	return SeqOrdered[T](s.Comparable().Seq().Peek(f))
}

// --- T-preserving intermediate methods re-exposed on SeqNumeric[T] ---

// Filter keeps elements satisfying pred; returns SeqNumeric.
func (s SeqNumeric[T]) Filter(pred func(T) bool) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().Filter(pred))
}

// Reject drops elements satisfying pred.
func (s SeqNumeric[T]) Reject(pred func(T) bool) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().Reject(pred))
}

// Take yields at most the first n elements.
func (s SeqNumeric[T]) Take(n int) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().Take(n))
}

// Drop skips the first n elements.
func (s SeqNumeric[T]) Drop(n int) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().Drop(n))
}

// TakeWhile yields until the first element failing pred.
func (s SeqNumeric[T]) TakeWhile(pred func(T) bool) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().TakeWhile(pred))
}

// DropWhile skips leading elements satisfying pred.
func (s SeqNumeric[T]) DropWhile(pred func(T) bool) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().DropWhile(pred))
}

// Peek invokes f per element as a side effect.
func (s SeqNumeric[T]) Peek(f func(T)) SeqNumeric[T] {
	return SeqNumeric[T](s.Ordered().Comparable().Seq().Peek(f))
}

// Collect materializes the subtype into a slice (terminal convenience).
func (s SeqNumeric[T]) Collect() []T {
	return s.Ordered().Comparable().Seq().Collect()
}

// Collect materializes a SeqOrdered into a slice.
func (s SeqOrdered[T]) Collect() []T {
	return s.Comparable().Seq().Collect()
}

// Collect materializes a SeqComparable into a slice.
func (s SeqComparable[T]) Collect() []T {
	return s.Seq().Collect()
}
