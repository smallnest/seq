package seq

import (
	"fmt"
	"strconv"
	"testing"
)

func TestSomeAndNone(t *testing.T) {
	s := Some(42)
	if !s.present || s.value != 42 {
		t.Fatalf("Some(42) = %+v, want present 42", s)
	}

	n := None[int]()
	if n.present {
		t.Fatalf("None() = %+v, want empty", n)
	}

	// Zero value must behave like None.
	var z Optional[int]
	if z.present {
		t.Fatalf("zero Optional = %+v, want empty (equivalent to None)", z)
	}
}

func TestSomeNilPointerIsPresent(t *testing.T) {
	// A present Optional holding a nil pointer is still present: presence is
	// independent of the value being a zero value (PRD US-001/US-002 invariant).
	var p *int
	o := Some(p)
	if !o.IsPresent() {
		t.Fatalf("Some((*int)(nil)).IsPresent() = false, want true")
	}
	if v, ok := o.Get(); !ok || v != nil {
		t.Fatalf("Some((*int)(nil)).Get() = (%v, %v), want (nil, true)", v, ok)
	}
	// OrElse must return the present nil, not the fallback.
	fallback := new(int)
	if got := o.OrElse(fallback); got != nil {
		t.Fatalf("OrElse on present nil pointer returned fallback, want nil")
	}
}

func TestToOptional(t *testing.T) {
	if got := ToOptional(7, true); !got.present || got.value != 7 {
		t.Fatalf("ToOptional(7, true) = %+v, want Some(7)", got)
	}
	if got := ToOptional(7, false); got.present {
		t.Fatalf("ToOptional(7, false) = %+v, want None", got)
	}
}

func TestOptionalString(t *testing.T) {
	if got := Some(3).String(); got != "Some(3)" {
		t.Errorf("Some(3).String() = %q, want %q", got, "Some(3)")
	}
	if got := None[int]().String(); got != "None" {
		t.Errorf("None().String() = %q, want %q", got, "None")
	}
}

func TestOptionalGet(t *testing.T) {
	if v, ok := Some(5).Get(); !ok || v != 5 {
		t.Errorf("Some(5).Get() = (%d, %v), want (5, true)", v, ok)
	}
	if v, ok := None[int]().Get(); ok || v != 0 {
		t.Errorf("None().Get() = (%d, %v), want (0, false)", v, ok)
	}
}

func TestOptionalPresence(t *testing.T) {
	if !Some(1).IsPresent() || Some(1).IsEmpty() {
		t.Error("Some(1): want present, not empty")
	}
	if None[int]().IsPresent() || !None[int]().IsEmpty() {
		t.Error("None: want empty, not present")
	}
}

func TestOptionalOrElseOrZero(t *testing.T) {
	if got := Some(3).OrElse(9); got != 3 {
		t.Errorf("Some(3).OrElse(9) = %d, want 3", got)
	}
	if got := None[int]().OrElse(9); got != 9 {
		t.Errorf("None.OrElse(9) = %d, want 9", got)
	}
	if got := Some(3).OrZero(); got != 3 {
		t.Errorf("Some(3).OrZero() = %d, want 3", got)
	}
	if got := None[int]().OrZero(); got != 0 {
		t.Errorf("None.OrZero() = %d, want 0", got)
	}
}

func TestOptionalUnwrap(t *testing.T) {
	if got := Some(7).Unwrap(); got != 7 {
		t.Errorf("Some(7).Unwrap() = %d, want 7", got)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("None.Unwrap() did not panic")
		}
	}()
	None[int]().Unwrap()
}

func TestOptionalUnwrapOr(t *testing.T) {
	if got := Some(3).UnwrapOr(9); got != 3 {
		t.Errorf("Some(3).UnwrapOr(9) = %d, want 3", got)
	}
	if got := None[int]().UnwrapOr(9); got != 9 {
		t.Errorf("None.UnwrapOr(9) = %d, want 9", got)
	}
}

func TestOptionalUnwrapOrElse(t *testing.T) {
	if got := Some(3).UnwrapOrElse(func() int { panic("should not be called") }); got != 3 {
		t.Errorf("Some(3).UnwrapOrElse = %d, want 3", got)
	}
	if got := None[int]().UnwrapOrElse(func() int { return 42 }); got != 42 {
		t.Errorf("None.UnwrapOrElse = %d, want 42", got)
	}
}

func TestOptionalMap(t *testing.T) {
	got := Some(3).Map(func(x int) int { return x * 2 })
	if v, ok := got.Get(); !ok || v != 6 {
		t.Errorf("Some(3).Map(*2) = %v, want Some(6)", got)
	}
	// None must pass through without calling f.
	got = None[int]().Map(func(int) int { panic("should not be called") })
	if got.IsPresent() {
		t.Errorf("None.Map = %v, want None", got)
	}
}

func TestOptionalFilter(t *testing.T) {
	if got := Some(4).Filter(func(x int) bool { return x%2 == 0 }); !got.IsPresent() {
		t.Errorf("Some(4).Filter(even) = %v, want Some(4)", got)
	}
	if got := Some(3).Filter(func(x int) bool { return x%2 == 0 }); got.IsPresent() {
		t.Errorf("Some(3).Filter(even) = %v, want None", got)
	}
	if got := None[int]().Filter(func(int) bool { panic("should not be called") }); got.IsPresent() {
		t.Errorf("None.Filter = %v, want None", got)
	}
}

func TestOptionalFlatMap(t *testing.T) {
	half := func(x int) Optional[int] {
		if x%2 == 0 {
			return Some(x / 2)
		}
		return None[int]()
	}
	if got := Some(8).FlatMap(half); got.OrZero() != 4 {
		t.Errorf("Some(8).FlatMap(half) = %v, want Some(4)", got)
	}
	if got := Some(3).FlatMap(half); got.IsPresent() {
		t.Errorf("Some(3).FlatMap(half) = %v, want None", got)
	}
	if got := None[int]().FlatMap(func(int) Optional[int] { panic("should not be called") }); got.IsPresent() {
		t.Errorf("None.FlatMap = %v, want None", got)
	}
}

func TestOptionalOr(t *testing.T) {
	if got := Some(1).Or(Some(2)); got.OrZero() != 1 {
		t.Errorf("Some(1).Or(Some(2)) = %v, want Some(1)", got)
	}
	if got := None[int]().Or(Some(2)); got.OrZero() != 2 {
		t.Errorf("None.Or(Some(2)) = %v, want Some(2)", got)
	}
}

func TestOptionalIfPresent(t *testing.T) {
	called := false
	Some(1).IfPresent(func(int) { called = true })
	if !called {
		t.Error("Some(1).IfPresent: f not called")
	}
	None[int]().IfPresent(func(int) { t.Error("None.IfPresent: f must not be called") })
}

func TestOptionalToSlice(t *testing.T) {
	if got := Some(1).ToSlice(); len(got) != 1 || got[0] != 1 {
		t.Errorf("Some(1).ToSlice() = %v, want [1]", got)
	}
	got := None[int]().ToSlice()
	if got == nil || len(got) != 0 {
		t.Errorf("None.ToSlice() = %v, want non-nil empty slice", got)
	}
}

func TestMapOptional(t *testing.T) {
	got := MapOptional(Some(42), strconv.Itoa)
	if v, ok := got.Get(); !ok || v != "42" {
		t.Errorf("MapOptional(Some(42), Itoa) = %v, want Some(\"42\")", got)
	}
	none := MapOptional(None[int](), func(int) string { panic("should not be called") })
	if none.IsPresent() {
		t.Errorf("MapOptional(None, ...) = %v, want None", none)
	}
}

func ExampleToOptional() {
	// ToOptional bridges a legacy (value, ok) pair into an Optional. Here we
	// synthesize one; most Seq terminals (Find, First, Max, ...) already return
	// an Optional directly and need no bridge.
	v, ok := 42, true
	out := ToOptional(v, ok).
		Map(func(x int) int { return x * 10 }).
		OrElse(-1)
	fmt.Println(out)
	// Output: 420
}

func ExampleSeq_Find() {
	s := From([]int{1, 2, 3, 4})
	// Find returns an Optional directly — chain post-processing with no bridge.
	out := s.Find(func(x int) bool { return x > 2 }).
		Map(func(x int) int { return x * 10 }).
		OrElse(-1)
	fmt.Println(out)
	// Output: 30
}
