package seq

import (
	"math"
	"testing"
)

// Direct edge-case coverage for the numeric/ordered free functions in
// numeric.go: empty, single-element, and all-equal inputs, which the subtype
// tests only exercised transitively.

func TestMaxEdgeCases(t *testing.T) {
	if _, ok := Max(Empty[int]()); ok {
		t.Fatalf("Max(empty): ok = true, want false")
	}
	if v, ok := Max(Of(42)); !ok || v != 42 {
		t.Fatalf("Max(single): got (%d, %v), want (42, true)", v, ok)
	}
	if v, ok := Max(Of(7, 7, 7)); !ok || v != 7 {
		t.Fatalf("Max(all-equal): got (%d, %v), want (7, true)", v, ok)
	}
}

func TestMinEdgeCases(t *testing.T) {
	if _, ok := Min(Empty[int]()); ok {
		t.Fatalf("Min(empty): ok = true, want false")
	}
	if v, ok := Min(Of(42)); !ok || v != 42 {
		t.Fatalf("Min(single): got (%d, %v), want (42, true)", v, ok)
	}
	if v, ok := Min(Of(7, 7, 7)); !ok || v != 7 {
		t.Fatalf("Min(all-equal): got (%d, %v), want (7, true)", v, ok)
	}
}

func TestSumEdgeCases(t *testing.T) {
	if got := Sum(Empty[int]()); got != 0 {
		t.Fatalf("Sum(empty): got %d, want 0", got)
	}
	if got := Sum(Of(42)); got != 42 {
		t.Fatalf("Sum(single): got %d, want 42", got)
	}
	if got := Sum(Of(7, 7, 7)); got != 21 {
		t.Fatalf("Sum(all-equal): got %d, want 21", got)
	}
}

func TestProductEdgeCases(t *testing.T) {
	// Empty yields the multiplicative identity, 1.
	if got := Product(Empty[int]()); got != 1 {
		t.Fatalf("Product(empty): got %d, want 1", got)
	}
	if got := Product(Of(42)); got != 42 {
		t.Fatalf("Product(single): got %d, want 42", got)
	}
	if got := Product(Of(2, 2, 2)); got != 8 {
		t.Fatalf("Product(all-equal): got %d, want 8", got)
	}
}

func TestMeanEdgeCases(t *testing.T) {
	if got := Mean(Empty[int]()); got != 0 {
		t.Fatalf("Mean(empty): got %v, want 0", got)
	}
	if got := Mean(Of(42)); got != 42 {
		t.Fatalf("Mean(single): got %v, want 42", got)
	}
	if got := Mean(Of(7, 7, 7)); got != 7 {
		t.Fatalf("Mean(all-equal): got %v, want 7", got)
	}
	if got := Mean(Of(1, 2)); math.Abs(got-1.5) > 1e-9 {
		t.Fatalf("Mean(1,2): got %v, want 1.5", got)
	}
}

func TestSortEdgeCases(t *testing.T) {
	if got := Sort(Empty[int]()).Collect(); len(got) != 0 {
		t.Fatalf("Sort(empty): got %v, want []", got)
	}
	if got := Sort(Of(42)).Collect(); len(got) != 1 || got[0] != 42 {
		t.Fatalf("Sort(single): got %v, want [42]", got)
	}
	if got := Sort(Of(7, 7, 7)).Collect(); len(got) != 3 || got[0] != 7 || got[2] != 7 {
		t.Fatalf("Sort(all-equal): got %v, want [7 7 7]", got)
	}
	// The materialized result is re-iterable.
	sorted := Sort(Of(3, 1, 2))
	first := sorted.Collect()
	second := sorted.Collect()
	if len(first) != 3 || first[0] != 1 || first[2] != 3 {
		t.Fatalf("Sort: got %v, want [1 2 3]", first)
	}
	if len(second) != len(first) || second[0] != first[0] {
		t.Fatalf("Sort: result not re-iterable: %v vs %v", first, second)
	}
}
