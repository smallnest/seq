package seq

import "cmp"

// Reusable function values that cut closure boilerplate at the most common call
// sites (Reduce, Fold, ZipWith, and the comparison predicates fed to Filter /
// Reject / TakeWhile). Until Go's lightweight anonymous function syntax lands
// (golang/go#21498), a closure like func(a, b int) int { return a + b } cannot
// be written more tersely — but it can be referenced by name.
//
// These are free functions because they constrain T itself (Numeric /
// cmp.Ordered / comparable), per the library's 划分铁律. Go 1.21+ type inference
// infers each generic operator's type argument from the expected parameter
// type, so callers write seq.Add, not seq.Add[int]:
//
//	sum, _ := seq.From([]int{1, 2, 3, 4}).Reduce(seq.Add)
//	pos    := seq.From(xs).Filter(seq.Gt(0))

// Add returns a + b. It is the named form of func(a, b T) T { return a + b },
// suitable as the reducer for Reduce / Fold / ZipWith.
func Add[T Numeric](a, b T) T { return a + b }

// Sub returns a - b.
func Sub[T Numeric](a, b T) T { return a - b }

// Mul returns a * b. As a Reduce reducer it computes a product.
func Mul[T Numeric](a, b T) T { return a * b }

// Max2 returns the larger of a and b by natural ordering. It is the binary
// operator form, distinct from [Max], which reduces a whole Seq.
func Max2[T cmp.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}

// Min2 returns the smaller of a and b by natural ordering. It is the binary
// operator form, distinct from [Min], which reduces a whole Seq.
func Min2[T cmp.Ordered](a, b T) T {
	if a > b {
		return b
	}
	return a
}

// Eq returns a predicate reporting whether its argument equals v.
func Eq[T comparable](v T) func(T) bool {
	return func(x T) bool { return x == v }
}

// Ne returns a predicate reporting whether its argument differs from v.
func Ne[T comparable](v T) func(T) bool {
	return func(x T) bool { return x != v }
}

// Gt returns a predicate reporting whether its argument is greater than v.
func Gt[T cmp.Ordered](v T) func(T) bool {
	return func(x T) bool { return x > v }
}

// Ge returns a predicate reporting whether its argument is greater than or
// equal to v.
func Ge[T cmp.Ordered](v T) func(T) bool {
	return func(x T) bool { return x >= v }
}

// Lt returns a predicate reporting whether its argument is less than v.
func Lt[T cmp.Ordered](v T) func(T) bool {
	return func(x T) bool { return x < v }
}

// Le returns a predicate reporting whether its argument is less than or equal
// to v.
func Le[T cmp.Ordered](v T) func(T) bool {
	return func(x T) bool { return x <= v }
}

// Not returns a predicate that negates p. It composes with the other predicate
// builders, e.g. Not(Eq(0)) is the same as Ne(0).
func Not[T any](p func(T) bool) func(T) bool {
	return func(x T) bool { return !p(x) }
}

// Identity returns its argument unchanged. It is useful as a no-op transform
// (Map(Identity)) or as a key function that keys by the element itself.
func Identity[T any](x T) T { return x }
