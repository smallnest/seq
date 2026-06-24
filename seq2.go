package seq

import (
	"iter"
)

// Seq2 free functions. These materialize or build a Seq2[K,V]. They are free
// functions because they constrain K to comparable (a map key must be
// comparable) or project from Seq[T] into Seq2[K,V] — neither can be a method
// on Seq2[K,V any] (whose K is any).

// ToMap materializes a Seq2 into a map[K]V. On duplicate keys the later pair
// overwrites the earlier. An empty Seq2 yields an empty (non-nil) map.
func ToMap[K comparable, V any](s Seq2[K, V]) map[K]V {
	out := make(map[K]V)
	for k, v := range iter.Seq2[K, V](s) {
		out[k] = v
	}
	return out
}

// CollectPairs materializes a Seq2 into a slice of Pair[K, V], preserving
// iteration order.
func CollectPairs[K, V any](s Seq2[K, V]) []Pair[K, V] {
	var out []Pair[K, V]
	for k, v := range iter.Seq2[K, V](s) {
		out = append(out, Pair[K, V]{Left: k, Right: v})
	}
	return out
}

// Entries creates a Seq2 from a slice of Pair[K, V]. The result is re-iterable
// (slice-backed).
func Entries[K, V any](pairs []Pair[K, V]) Seq2[K, V] {
	return Seq2[K, V](func(yield func(K, V) bool) {
		for _, p := range pairs {
			if !yield(p.Left, p.Right) {
				return
			}
		}
	})
}

// Associate projects a Seq[T] into a Seq2[K, V] via f. It is a free function
// because K must be constrained comparable-ready in the caller's mind (the
// Seq2 it returns is typically fed to ToMap), and the projection introduces
// independent type parameters K and V beyond the receiver's T.
func Associate[T any, K comparable, V any](s Seq[T], f func(T) (K, V)) Seq2[K, V] {
	return Seq2[K, V](func(yield func(K, V) bool) {
		for v := range iter.Seq[T](s) {
			k, val := f(v)
			if !yield(k, val) {
				return
			}
		}
	})
}
