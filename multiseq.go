package seq

import (
	"iter"
)

// Multi-sequence combinators, kept as free functions. Of these, only Flatten is
// forced out by the language: a Seq[Seq[T]] receiver is illegal — a receiver's
// type argument must be an identifier, not Seq[T]. The Zip family pairs two
// independent sources; its natural method form (Seq[T]).Zip[R](Seq[R])
// Seq[Pair[T,R]] hits an instantiation cycle (re-instantiating Seq with an
// argument that deforms T — Pair[T,R] — never terminates), so it cannot return
// the same defined type — returning Seq2 / the underlying iter.Seq would compile
// and stay lazy.
//
// Two distinct Go "cycle" restrictions are in play, both still rejecting this on
// go1.27rc1: (1) #75883 (Go 1.26) LIFTED only the self-referential *constraint*
// rule (type T[P T[P]]); it does not help here, since Seq's T is constrained to
// any. (2) The monomorphization circuit breaker: Go generates all methods for
// each instantiation, so a method returning A[A[T]] forces A[A[T]]'s methods →
// A[A[A[int]]] → … ad infinitum. #80109 closed that self-nesting form
// (A[A[T]]) as "working as intended". Zip's Seq[Pair[T,R]] is the *derived but
// non-self-nesting* form tracked by #80172 (still OPEN, no Go-team ruling): the
// equivalent free function compiles, so whether this sub-case is truly intended
// is unsettled. Either way it does not compile today; the Zip family stays free
// for source symmetry (划分铁律: 多源组合 / 嵌套实例化).

// Zip pairs elements of a and b position-wise, stopping when the shorter
// sequence is exhausted. Yields Pair[A, B].
func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B] {
	return Seq2[A, B](func(yield func(A, B) bool) {
		nextA, stopA := iter.Pull(iter.Seq[A](a))
		defer stopA()
		nextB, stopB := iter.Pull(iter.Seq[B](b))
		defer stopB()
		for {
			va, okA := nextA()
			if !okA {
				return
			}
			vb, okB := nextB()
			if !okB {
				return
			}
			if !yield(va, vb) {
				return
			}
		}
	})
}

// ZipWith pairs a and b and applies f to each pair, stopping at the shorter
// sequence. Yields Seq[C] (lodash zipWith).
func ZipWith[A, B, C any](a Seq[A], b Seq[B], f func(A, B) C) Seq[C] {
	return Seq[C](func(yield func(C) bool) {
		nextA, stopA := iter.Pull(iter.Seq[A](a))
		defer stopA()
		nextB, stopB := iter.Pull(iter.Seq[B](b))
		defer stopB()
		for {
			va, okA := nextA()
			if !okA {
				return
			}
			vb, okB := nextB()
			if !okB {
				return
			}
			if !yield(f(va, vb)) {
				return
			}
		}
	})
}

// ZipMap pairs keys and values position-wise into a map, stopping at the
// shorter sequence. On duplicate keys the later pair overwrites the earlier
// (lodash zipObject). An empty input yields an empty (non-nil) map.
func ZipMap[K comparable, V any](keys Seq[K], vals Seq[V]) map[K]V {
	out := make(map[K]V)
	nextK, stopK := iter.Pull(iter.Seq[K](keys))
	defer stopK()
	nextV, stopV := iter.Pull(iter.Seq[V](vals))
	defer stopV()
	for {
		k, okK := nextK()
		if !okK {
			return out
		}
		v, okV := nextV()
		if !okV {
			return out
		}
		out[k] = v
	}
}

// Zip3 triples a, b, c position-wise, stopping at the shortest. Yields
// Seq[Tuple3[A,B,C]].
func Zip3[A, B, C any](a Seq[A], b Seq[B], c Seq[C]) Seq[Tuple3[A, B, C]] {
	return Seq[Tuple3[A, B, C]](func(yield func(Tuple3[A, B, C]) bool) {
		nextA, stopA := iter.Pull(iter.Seq[A](a))
		defer stopA()
		nextB, stopB := iter.Pull(iter.Seq[B](b))
		defer stopB()
		nextC, stopC := iter.Pull(iter.Seq[C](c))
		defer stopC()
		for {
			va, okA := nextA()
			if !okA {
				return
			}
			vb, okB := nextB()
			if !okB {
				return
			}
			vc, okC := nextC()
			if !okC {
				return
			}
			if !yield(Tuple3[A, B, C]{First: va, Second: vb, Third: vc}) {
				return
			}
		}
	})
}

// Zip4 quadruples a, b, c, d position-wise, stopping at the shortest. Yields
// Seq[Tuple4[A,B,C,D]].
func Zip4[A, B, C, D any](a Seq[A], b Seq[B], c Seq[C], d Seq[D]) Seq[Tuple4[A, B, C, D]] {
	return Seq[Tuple4[A, B, C, D]](func(yield func(Tuple4[A, B, C, D]) bool) {
		nextA, stopA := iter.Pull(iter.Seq[A](a))
		defer stopA()
		nextB, stopB := iter.Pull(iter.Seq[B](b))
		defer stopB()
		nextC, stopC := iter.Pull(iter.Seq[C](c))
		defer stopC()
		nextD, stopD := iter.Pull(iter.Seq[D](d))
		defer stopD()
		for {
			va, okA := nextA()
			if !okA {
				return
			}
			vb, okB := nextB()
			if !okB {
				return
			}
			vc, okC := nextC()
			if !okC {
				return
			}
			vd, okD := nextD()
			if !okD {
				return
			}
			if !yield(Tuple4[A, B, C, D]{First: va, Second: vb, Third: vc, Fourth: vd}) {
				return
			}
		}
	})
}

// Unzip splits a Seq2 of (A, B) pairs into two separate Seqs. Each result may
// be iterated independently; the source is drained once into buffers on first
// use of either (so both sides see the same elements).
func Unzip[A, B any](s Seq2[A, B]) (Seq[A], Seq[B]) {
	var as []A
	var bs []B
	drained := false
	drain := func() {
		if drained {
			return
		}
		drained = true
		for k, v := range iter.Seq2[A, B](s) {
			as = append(as, k)
			bs = append(bs, v)
		}
	}
	return Seq[A](func(yield func(A) bool) {
			drain()
			for _, v := range as {
				if !yield(v) {
					return
				}
			}
		}),
		Seq[B](func(yield func(B) bool) {
			drain()
			for _, v := range bs {
				if !yield(v) {
					return
				}
			}
		})
}

// Flatten flattens one level of a Seq[Seq[T]] into a Seq[T]. Deeper nesting
// requires explicit repeated calls (Go's type system cannot express arbitrary
// depth).
func Flatten[T any](s Seq[Seq[T]]) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for inner := range iter.Seq[Seq[T]](s) {
			for v := range iter.Seq[T](inner) {
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Concat concatenates the given sequences in order (variadic form). With no
// arguments it yields nothing.
func Concat[T any](seqs ...Seq[T]) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for _, s := range seqs {
			for v := range iter.Seq[T](s) {
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Interleave round-robins elements from the given sequences: one element from
// each in turn, repeatedly, skipping any sequence that has been exhausted
// until all are exhausted.
func Interleave[T any](seqs ...Seq[T]) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		if len(seqs) == 0 {
			return
		}
		pullers := make([]func() (T, bool), len(seqs))
		stops := make([]func(), len(seqs))
		for i, s := range seqs {
			next, stop := iter.Pull(iter.Seq[T](s))
			pullers[i] = next
			stops[i] = stop
		}
		defer func() {
			for _, stop := range stops {
				stop()
			}
		}()
		active := len(pullers)
		for active > 0 {
			for i, next := range pullers {
				if next == nil {
					continue
				}
				v, ok := next()
				if !ok {
					pullers[i] = nil
					active--
					continue
				}
				if !yield(v) {
					return
				}
			}
		}
	})
}
