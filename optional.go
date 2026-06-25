package seq

import "fmt"

// Optional is a lightweight, zero-dependency wrapper over the (value, present)
// pair that the standard library and this package express as a (T, bool)
// return. It exists purely as an opt-in, caller-side convenience: no method on
// [Seq], [Seq2] or the constrained subtypes takes or returns an Optional, so
// the package keeps its zero-cost interoperability with iter.Seq and the rest
// of the standard-library iterator ecosystem. Bridge from an existing (T, bool)
// method with [ToOptional]:
//
//	seq.ToOptional(s.Find(pred)).Map(f).OrElse(fallback)
//
// The zero value Optional[T]{} is a valid empty Optional, equivalent to
// [None].
//
// Because Go methods cannot introduce new type parameters, [Optional.Map] can
// only return Optional[T] (the same element type); to change the element type
// use the package-level [MapOptional].
type Optional[T any] struct {
	value   T
	present bool
}

// Some returns an Optional that holds v.
func Some[T any](v T) Optional[T] {
	return Optional[T]{value: v, present: true}
}

// None returns an empty Optional. It is equivalent to the zero value
// Optional[T]{}.
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// ToOptional bridges a (value, ok) pair — the convention used by [Seq.Find],
// [Seq.First] and the other partial-result methods — into an Optional. When ok
// is true it is equivalent to Some(v); otherwise it is None[T]().
func ToOptional[T any](v T, ok bool) Optional[T] {
	if ok {
		return Some(v)
	}
	return None[T]()
}

// Get returns the contained value and whether it is present, mirroring the
// (T, bool) convention used throughout the package so an Optional flows back
// into existing call sites: if v, ok := o.Get(); ok { ... }.
func (o Optional[T]) Get() (T, bool) {
	return o.value, o.present
}

// IsPresent reports whether a value is present.
func (o Optional[T]) IsPresent() bool {
	return o.present
}

// IsEmpty reports whether no value is present. It is the negation of
// [Optional.IsPresent].
func (o Optional[T]) IsEmpty() bool {
	return !o.present
}

// OrElse returns the contained value when present, otherwise fallback.
func (o Optional[T]) OrElse(fallback T) T {
	if o.present {
		return o.value
	}
	return fallback
}

// Unwrap returns the contained value, or panics if the Optional is empty. It
// mirrors Rust's Option::unwrap and is for call sites that have already
// established presence; prefer [Optional.Get] or [Optional.OrElse] when
// absence is possible.
func (o Optional[T]) Unwrap() T {
	if !o.present {
		panic("seq: Unwrap called on an empty Optional")
	}
	return o.value
}

// UnwrapOr returns the contained value when present, otherwise fallback. It is
// the Rust-named alias of [Optional.OrElse].
func (o Optional[T]) UnwrapOr(fallback T) T {
	return o.OrElse(fallback)
}

// UnwrapOrElse returns the contained value when present, otherwise the result
// of calling f. Unlike [Optional.UnwrapOr], the fallback is computed lazily,
// only when empty (Rust's Option::unwrap_or_else). f must be non-nil when the
// Optional is empty (it is not called when present).
func (o Optional[T]) UnwrapOrElse(f func() T) T {
	if o.present {
		return o.value
	}
	return f()
}

// OrZero returns the contained value when present, otherwise the zero value
// of T.
func (o Optional[T]) OrZero() T {
	var zero T
	if o.present {
		return o.value
	}
	return zero
}

// Map returns Some(f(value)) when a value is present, or None unchanged when
// empty. Because Go methods cannot introduce new type parameters, Map cannot
// change the element type; use the package-level [MapOptional] for that. f
// must be non-nil when a value is present (it is not called when empty).
func (o Optional[T]) Map(f func(T) T) Optional[T] {
	if !o.present {
		return o
	}
	return Some(f(o.value))
}

// Filter returns the Optional unchanged when present and pred(value) is true,
// otherwise None. pred must be non-nil when a value is present.
func (o Optional[T]) Filter(pred func(T) bool) Optional[T] {
	if o.present && pred(o.value) {
		return o
	}
	return None[T]()
}

// FlatMap returns f(value) when present, or None when empty. It is the
// monadic bind that lets f itself decide presence. f must be non-nil when a
// value is present.
func (o Optional[T]) FlatMap(f func(T) Optional[T]) Optional[T] {
	if !o.present {
		return None[T]()
	}
	return f(o.value)
}

// Or returns the receiver when it is present, otherwise other.
func (o Optional[T]) Or(other Optional[T]) Optional[T] {
	if o.present {
		return o
	}
	return other
}

// IfPresent calls f with the contained value when present, and does nothing
// when empty. f must be non-nil when a value is present.
func (o Optional[T]) IfPresent(f func(T)) {
	if o.present {
		f(o.value)
	}
}

// ToSlice materializes the Optional: a single-element slice when present, or
// an empty (non-nil, length-zero) slice when empty.
func (o Optional[T]) ToSlice() []T {
	if o.present {
		return []T{o.value}
	}
	return []T{}
}

// String implements [fmt.Stringer]: "Some(<value>)" when present, "None"
// otherwise.
func (o Optional[T]) String() string {
	if o.present {
		return fmt.Sprintf("Some(%v)", o.value)
	}
	return "None"
}

// MapOptional transforms an Optional[T] into an Optional[U] by applying f to
// the contained value: Some(f(value)) when present, None[U]() when empty. It
// is the package-level, type-changing counterpart to [Optional.Map], required
// because Go methods cannot introduce the new type parameter U. f must be
// non-nil when a value is present (it is not called when empty).
func MapOptional[T, U any](o Optional[T], f func(T) U) Optional[U] {
	if v, ok := o.Get(); ok {
		return Some(f(v))
	}
	return None[U]()
}
