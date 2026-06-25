package seq

import "testing"

// Tests for the curated operators and predicate builders in funcs.go, including
// their use through Seq operations (which also exercises the Go 1.21+ inference
// that lets seq.Add be passed where func(T, T) T is expected).

func TestOperators(t *testing.T) {
	if got := Add(3, 4); got != 7 {
		t.Errorf("Add(3,4) = %d, want 7", got)
	}
	if got := Sub(10, 4); got != 6 {
		t.Errorf("Sub(10,4) = %d, want 6", got)
	}
	if got := Mul(3, 4); got != 12 {
		t.Errorf("Mul(3,4) = %d, want 12", got)
	}
	if got := Max2(3, 4); got != 4 {
		t.Errorf("Max2(3,4) = %d, want 4", got)
	}
	if got := Max2(4, 3); got != 4 {
		t.Errorf("Max2(4,3) = %d, want 4", got)
	}
	if got := Min2(3, 4); got != 3 {
		t.Errorf("Min2(3,4) = %d, want 3", got)
	}
	if got := Min2(4, 3); got != 3 {
		t.Errorf("Min2(4,3) = %d, want 3", got)
	}
	// Equal operands return that operand for both Max2 and Min2.
	if got := Max2(5, 5); got != 5 {
		t.Errorf("Max2(5,5) = %d, want 5", got)
	}
	if got := Min2(5, 5); got != 5 {
		t.Errorf("Min2(5,5) = %d, want 5", got)
	}
	// Works for float and string (Ordered) too.
	if got := Add(1.5, 2.5); got != 4.0 {
		t.Errorf("Add(1.5,2.5) = %v, want 4", got)
	}
	if got := Max2("a", "b"); got != "b" {
		t.Errorf("Max2(a,b) = %q, want b", got)
	}
}

func TestOperatorsThroughReduce(t *testing.T) {
	// The headline win: pass the named operator, no instantiation, no closure.
	if sum, ok := From([]int{1, 2, 3, 4}).Reduce(Add); !ok || sum != 10 {
		t.Errorf("Reduce(Add) = (%d, %v), want (10, true)", sum, ok)
	}
	if prod, ok := From([]int{1, 2, 3, 4}).Reduce(Mul); !ok || prod != 24 {
		t.Errorf("Reduce(Mul) = (%d, %v), want (24, true)", prod, ok)
	}
	if mx, ok := From([]int{3, 1, 4, 1, 5}).Reduce(Max2); !ok || mx != 5 {
		t.Errorf("Reduce(Max2) = (%d, %v), want (5, true)", mx, ok)
	}
	if mn, ok := From([]int{3, 1, 4, 1, 5}).Reduce(Min2); !ok || mn != 1 {
		t.Errorf("Reduce(Min2) = (%d, %v), want (1, true)", mn, ok)
	}
}

func TestPredicateBuilders(t *testing.T) {
	cases := []struct {
		name string
		pred func(int) bool
		in   int
		want bool
	}{
		{"Eq hit", Eq(5), 5, true},
		{"Eq miss", Eq(5), 6, false},
		{"Ne hit", Ne(5), 6, true},
		{"Ne miss", Ne(5), 5, false},
		{"Gt true", Gt(5), 6, true},
		{"Gt false eq", Gt(5), 5, false},
		{"Ge true eq", Ge(5), 5, true},
		{"Ge false", Ge(5), 4, false},
		{"Lt true", Lt(5), 4, true},
		{"Lt false eq", Lt(5), 5, false},
		{"Le true eq", Le(5), 5, true},
		{"Le false", Le(5), 6, false},
	}
	for _, c := range cases {
		if got := c.pred(c.in); got != c.want {
			t.Errorf("%s: pred(%d) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestNot(t *testing.T) {
	isZero := Eq(0)
	notZero := Not(isZero)
	if notZero(0) {
		t.Errorf("Not(Eq(0))(0) = true, want false")
	}
	if !notZero(1) {
		t.Errorf("Not(Eq(0))(1) = false, want true")
	}
	// Not(Eq(v)) is equivalent to Ne(v).
	for _, x := range []int{-1, 0, 1, 2} {
		if Not(Eq(0))(x) != Ne(0)(x) {
			t.Errorf("Not(Eq(0)) and Ne(0) disagree at %d", x)
		}
	}
}

func TestIdentity(t *testing.T) {
	if Identity(42) != 42 {
		t.Errorf("Identity(42) = %d, want 42", Identity(42))
	}
	if Identity("x") != "x" {
		t.Errorf(`Identity("x") = %q, want "x"`, Identity("x"))
	}
	// As a no-op Map transform.
	got := From([]int{1, 2, 3}).Map(Identity).Collect()
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("Map(Identity) = %v, want [1 2 3]", got)
	}
}

func TestPredicateBuildersThroughFilter(t *testing.T) {
	xs := From([]int{-2, -1, 0, 1, 2})
	if got := xs.Filter(Gt(0)).Collect(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("Filter(Gt(0)) = %v, want [1 2]", got)
	}
	if got := From([]int{0, 1, 0, 2, 0}).Reject(Eq(0)).Collect(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("Reject(Eq(0)) = %v, want [1 2]", got)
	}
}
