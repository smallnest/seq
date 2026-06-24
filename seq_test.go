package seq

import (
	"iter"
	"slices"
	"testing"
)

// TestSeqInterconvert verifies Seq[T] interconverts with iter.Seq[T] at zero
// cost: a Seq[T] can be passed to standard library iterator consumers after a
// compile-time conversion, and an iter.Seq[T] can be wrapped as a Seq[T].
func TestSeqInterconvert(t *testing.T) {
	var it iter.Seq[int] = func(yield func(int) bool) {
		for i := 1; i <= 3; i++ {
			if !yield(i) {
				return
			}
		}
	}

	// iter.Seq[T] -> Seq[T] (zero-cost conversion).
	s := Seq[int](it)
	got := slices.Collect(iter.Seq[int](s))
	want := []int{1, 2, 3}
	if !slices.Equal(got, want) {
		t.Fatalf("Seq[int] conversion: got %v, want %v", got, want)
	}

	// Seq[T] -> iter.Seq[T] (zero-cost conversion) feeds slices.Collect.
	got2 := slices.Collect(iter.Seq[int](s))
	if !slices.Equal(got2, want) {
		t.Fatalf("iter.Seq[int] conversion: got %v, want %v", got2, want)
	}
}

// TestSeq2Interconvert verifies Seq2[K,V] interconverts with iter.Seq2[K,V].
func TestSeq2Interconvert(t *testing.T) {
	var it2 iter.Seq2[string, int] = func(yield func(string, int) bool) {
		yield("a", 1)
		yield("b", 2)
	}
	s2 := Seq2[string, int](it2)

	got := map[string]int{}
	for k, v := range iter.Seq2[string, int](s2) {
		got[k] = v
	}
	if got["a"] != 1 || got["b"] != 2 {
		t.Fatalf("Seq2 conversion: got %v", got)
	}
}

func TestPairAndTuples(t *testing.T) {
	p := Pair[string, int]{Left: "x", Right: 9}
	if p.Left != "x" || p.Right != 9 {
		t.Fatalf("Pair fields: %+v", p)
	}

	t3 := Tuple3[int, string, float64]{First: 1, Second: "two", Third: 3.5}
	if t3.First != 1 || t3.Second != "two" || t3.Third != 3.5 {
		t.Fatalf("Tuple3 fields: %+v", t3)
	}

	t4 := Tuple4[int, int, int, int]{First: 1, Second: 2, Third: 3, Fourth: 4}
	if t4.First+t4.Second+t4.Third+t4.Fourth != 10 {
		t.Fatalf("Tuple4 fields: %+v", t4)
	}
}

// myInt is a defined type with underlying type int; it must satisfy Numeric.
type myInt int

// numericSum exercises the Numeric constraint with a mix of concrete and
// defined numeric types. If Numeric were missing a kind, this would fail to
// compile.
func numericSum[T Numeric](vals ...T) T {
	var sum T
	for _, v := range vals {
		sum += v
	}
	return sum
}

func TestNumericConstraint(t *testing.T) {
	if numericSum(1, 2, 3) != 6 {
		t.Fatal("int Numeric")
	}
	if numericSum(1.5, 2.5) != 4.0 {
		t.Fatal("float64 Numeric")
	}
	if numericSum(myInt(1), myInt(2), myInt(3)) != myInt(6) {
		t.Fatal("myInt (~int) Numeric")
	}
	if numericSum(uint(1), uint(2)) != uint(3) {
		t.Fatal("uint Numeric")
	}
}
