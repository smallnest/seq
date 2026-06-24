package seq

import (
	"iter"
)

// Seq2[K,V] methods (FR-7a). These operate on key/value sequences. The
// type-changing transforms (MapValues[U], MapKeys[J], Map[J,U]) use Go 1.27
// method-level type parameters. Keys()/Values() project to Seq[K]/Seq[V].

// MapValues transforms each value via f, keeping keys. U is the method's own
// type parameter.
func (s Seq2[K, V]) MapValues[U any](f func(V) U) Seq2[K, U] {
	return Seq2[K, U](func(yield func(K, U) bool) {
		for k, v := range iter.Seq2[K, V](s) {
			if !yield(k, f(v)) {
				return
			}
		}
	})
}

// MapKeys transforms each key via f, keeping values. J is the method's own
// type parameter. Note: distinct source keys may collide after mapping; later
// pairs overwrite earlier (the Seq2 still yields all pairs, but downstream
// ToMap would dedup).
func (s Seq2[K, V]) MapKeys[J any](f func(K) J) Seq2[J, V] {
	return Seq2[J, V](func(yield func(J, V) bool) {
		for k, v := range iter.Seq2[K, V](s) {
			if !yield(f(k), v) {
				return
			}
		}
	})
}

// Map transforms both key and value via f. J and U are the method's own type
// parameters.
func (s Seq2[K, V]) Map[J, U any](f func(K, V) (J, U)) Seq2[J, U] {
	return Seq2[J, U](func(yield func(J, U) bool) {
		for k, v := range iter.Seq2[K, V](s) {
			jk, uv := f(k, v)
			if !yield(jk, uv) {
				return
			}
		}
	})
}

// Filter keeps pairs satisfying pred.
func (s Seq2[K, V]) Filter(pred func(K, V) bool) Seq2[K, V] {
	return Seq2[K, V](func(yield func(K, V) bool) {
		for k, v := range iter.Seq2[K, V](s) {
			if !pred(k, v) {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	})
}

// Keys projects the keys into a Seq[K].
func (s Seq2[K, V]) Keys() Seq[K] {
	return Seq[K](func(yield func(K) bool) {
		for k := range iter.Seq2[K, V](s) {
			if !yield(k) {
				return
			}
		}
	})
}

// Values projects the values into a Seq[V].
func (s Seq2[K, V]) Values() Seq[V] {
	return Seq[V](func(yield func(V) bool) {
		for _, v := range iter.Seq2[K, V](s) {
			if !yield(v) {
				return
			}
		}
	})
}

// ForEach invokes f for each (key, value) pair.
func (s Seq2[K, V]) ForEach(f func(K, V)) {
	for k, v := range iter.Seq2[K, V](s) {
		f(k, v)
	}
}

// Fold left-folds the pairs into an accumulator starting from init. U is the
// method's own type parameter.
func (s Seq2[K, V]) Fold[U any](init U, f func(U, K, V) U) U {
	acc := init
	for k, v := range iter.Seq2[K, V](s) {
		acc = f(acc, k, v)
	}
	return acc
}

// Count returns the number of pairs.
func (s Seq2[K, V]) Count() int {
	n := 0
	for range iter.Seq2[K, V](s) {
		n++
	}
	return n
}

// Find returns the first (key, value) pair satisfying pred, or (zero, zero,
// false) if none match.
func (s Seq2[K, V]) Find(pred func(K, V) bool) (K, V, bool) {
	for k, v := range iter.Seq2[K, V](s) {
		if pred(k, v) {
			return k, v, true
		}
	}
	var zk K
	var zv V
	return zk, zv, false
}

// Any reports whether any pair satisfies pred (short-circuits on first match).
func (s Seq2[K, V]) Any(pred func(K, V) bool) bool {
	for k, v := range iter.Seq2[K, V](s) {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// All reports whether every pair satisfies pred (short-circuits on first
// failure).
func (s Seq2[K, V]) All(pred func(K, V) bool) bool {
	for k, v := range iter.Seq2[K, V](s) {
		if !pred(k, v) {
			return false
		}
	}
	return true
}
