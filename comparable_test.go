package seq

import (
	"slices"
	"testing"
)

func TestDistinct(t *testing.T) {
	got := collect(Distinct(From([]int{1, 2, 1, 3, 2, 4})))
	if !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Distinct: got %v", got)
	}
	// preserves first-occurrence order with non-numeric comparable
	gotStr := collect(Distinct(From([]string{"b", "a", "b", "c", "a"})))
	if !slices.Equal(gotStr, []string{"b", "a", "c"}) {
		t.Fatalf("Distinct strings: got %v", gotStr)
	}
	if len(collect(Distinct(From([]int{})))) != 0 {
		t.Fatal("Distinct empty")
	}
}

func TestContains(t *testing.T) {
	if !Contains(From([]int{1, 2, 3}), 2) {
		t.Fatal("Contains 2 should be true")
	}
	if Contains(From([]int{1, 2, 3}), 9) {
		t.Fatal("Contains 9 should be false")
	}
	if Contains(From([]int{}), 1) {
		t.Fatal("Contains empty should be false")
	}
	// short-circuit: stop early (no panic on infinite source via bounded)
}

func TestIndexOf(t *testing.T) {
	if idx, ok := IndexOf(From([]string{"a", "b", "a"}), "a").Get(); !ok || idx != 0 {
		t.Fatalf("IndexOf first a: (%d,%v)", idx, ok)
	}
	if idx, ok := IndexOf(From([]string{"a", "b", "c"}), "b").Get(); !ok || idx != 1 {
		t.Fatalf("IndexOf b: (%d,%v)", idx, ok)
	}
	if IndexOf(From([]string{"a", "b"}), "z").IsPresent() {
		t.Fatal("IndexOf absent should be None")
	}
	if IndexOf(From([]int{}), 1).IsPresent() {
		t.Fatal("IndexOf empty should be None")
	}
}

func TestLastIndexOf(t *testing.T) {
	if idx, ok := LastIndexOf(From([]string{"a", "b", "a", "c", "a"}), "a").Get(); !ok || idx != 4 {
		t.Fatalf("LastIndexOf a: (%d,%v)", idx, ok)
	}
	if LastIndexOf(From([]string{"a", "b"}), "z").IsPresent() {
		t.Fatal("LastIndexOf absent")
	}
	if LastIndexOf(From([]int{}), 1).IsPresent() {
		t.Fatal("LastIndexOf empty")
	}
}

func TestCountValues(t *testing.T) {
	got := CountValues(From([]string{"a", "b", "a", "c", "b", "a"}))
	want := map[string]int{"a": 3, "b": 2, "c": 1}
	if len(got) != len(want) || got["a"] != 3 || got["b"] != 2 || got["c"] != 1 {
		t.Fatalf("CountValues: got %v, want %v", got, want)
	}
	if len(CountValues(From([]int{}))) != 0 {
		t.Fatal("CountValues empty should be empty map")
	}
}

func TestEqual(t *testing.T) {
	if !Equal(From([]int{1, 2, 3}), From([]int{1, 2, 3})) {
		t.Fatal("Equal same")
	}
	if Equal(From([]int{1, 2, 3}), From([]int{1, 2})) {
		t.Fatal("Equal different length")
	}
	if Equal(From([]int{1, 2, 3}), From([]int{1, 3, 2})) {
		t.Fatal("Equal different order")
	}
	if !Equal(From([]int{}), From([]int{})) {
		t.Fatal("Equal two empty")
	}
	if Equal(From([]int{}), From([]int{1})) {
		t.Fatal("Equal empty vs non-empty")
	}
}

func TestCompact(t *testing.T) {
	got := collect(Compact(From([]int{0, 1, 0, 2, 0, 3})))
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("Compact: got %v", got)
	}
	// string zero value is ""
	gotStr := collect(Compact(From([]string{"", "a", "", "b"})))
	if !slices.Equal(gotStr, []string{"a", "b"}) {
		t.Fatalf("Compact strings: got %v", gotStr)
	}
	if len(collect(Compact(From([]int{0, 0, 0})))) != 0 {
		t.Fatal("Compact all-zero")
	}
}

func TestWithout(t *testing.T) {
	got := collect(Without(From([]int{1, 2, 3, 2, 4, 2}), 2, 4))
	if !slices.Equal(got, []int{1, 3}) {
		t.Fatalf("Without: got %v", got)
	}
	// without absent values is identity
	got2 := collect(Without(From([]int{1, 2, 3}), 9))
	if !slices.Equal(got2, []int{1, 2, 3}) {
		t.Fatalf("Without absent: got %v", got2)
	}
}

func TestToSet(t *testing.T) {
	set := ToSet(From([]int{1, 2, 2, 3, 1}))
	if len(set) != 3 || set[1] != struct{}{} || set[2] != struct{}{} || set[3] != struct{}{} {
		t.Fatalf("ToSet: got %v", set)
	}
	if len(ToSet(From([]int{}))) != 0 {
		t.Fatal("ToSet empty")
	}
}

func TestJoinStrings(t *testing.T) {
	if got := JoinStrings(From([]string{"a", "b", "c"}), ", "); got != "a, b, c" {
		t.Fatalf("JoinStrings: got %q", got)
	}
	if got := JoinStrings(From([]string{"only"}), ", "); got != "only" {
		t.Fatalf("JoinStrings single: got %q", got)
	}
	if got := JoinStrings(From([]string{}), ", "); got != "" {
		t.Fatalf("JoinStrings empty: got %q", got)
	}
}

func TestUnion(t *testing.T) {
	got := collect(Union(From([]int{1, 2}), From([]int{2, 3}), From([]int{3, 4})))
	if !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Union: got %v", got)
	}
	if len(collect(Union[int]())) != 0 {
		t.Fatal("Union no args")
	}
}

func TestIntersect(t *testing.T) {
	got := collect(Intersect(From([]int{1, 2, 3, 2, 4}), From([]int{2, 4, 5})))
	if !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("Intersect: got %v", got)
	}
	if len(collect(Intersect(From([]int{1, 2}), From([]int{3, 4})))) != 0 {
		t.Fatal("Intersect disjoint")
	}
}

func TestDifference(t *testing.T) {
	got := collect(Difference(From([]int{1, 2, 3, 4}), From([]int{2, 4})))
	if !slices.Equal(got, []int{1, 3}) {
		t.Fatalf("Difference: got %v", got)
	}
	// a minus b where a subset is in b -> empty
	if len(collect(Difference(From([]int{2, 4}), From([]int{1, 2, 3, 4, 5})))) != 0 {
		t.Fatal("Difference a subset of b")
	}
}

func TestSymmetricDifference(t *testing.T) {
	got := collect(SymmetricDifference(From([]int{1, 2, 3}), From([]int{2, 3, 4})))
	if !slices.Equal(got, []int{1, 4}) {
		t.Fatalf("SymmetricDifference: got %v", got)
	}
	// identical sets -> empty
	if len(collect(SymmetricDifference(From([]int{1, 2}), From([]int{1, 2})))) != 0 {
		t.Fatal("SymmetricDifference identical")
	}
}

func TestMaxMin(t *testing.T) {
	mx, ok := Max(From([]int{3, 1, 4, 1, 5, 9, 2, 6})).Get()
	if !ok || mx != 9 {
		t.Fatalf("Max: (%d,%v)", mx, ok)
	}
	mn, ok := Min(From([]int{3, 1, 4, 1, 5, 9, 2, 6})).Get()
	if !ok || mn != 1 {
		t.Fatalf("Min: (%d,%v)", mn, ok)
	}
	if Max(From([]int{})).IsPresent() {
		t.Fatal("Max empty should be None")
	}
	if Min(From([]int{})).IsPresent() {
		t.Fatal("Min empty should be None")
	}
	// floats
	fmx := Max(From([]float64{1.5, 2.5, 0.5})).OrZero()
	if fmx != 2.5 {
		t.Fatalf("Max float: %v", fmx)
	}
	// strings (cmp.Ordered)
	smax := Max(From([]string{"apple", "banana", "cherry"})).OrZero()
	if smax != "cherry" {
		t.Fatalf("Max string: %q", smax)
	}
}

func TestSumProductMean(t *testing.T) {
	if got := Sum(From([]int{1, 2, 3, 4})); got != 10 {
		t.Fatalf("Sum: %d", got)
	}
	if got := Product(From([]int{1, 2, 3, 4})); got != 24 {
		t.Fatalf("Product: %d", got)
	}
	if got := Mean(From([]int{1, 2, 3, 4})); got != 2.5 {
		t.Fatalf("Mean: %v", got)
	}
	// empty-sequence conventions
	if got := Sum(From([]int{})); got != 0 {
		t.Fatalf("Sum empty: %d", got)
	}
	if got := Product(From([]int{})); got != 1 {
		t.Fatalf("Product empty: %d", got)
	}
	if got := Mean(From([]int{})); got != 0 {
		t.Fatalf("Mean empty: %v", got)
	}
	// floats
	if got := Sum(From([]float64{1.5, 2.5})); got != 4.0 {
		t.Fatalf("Sum float: %v", got)
	}
	if got := Product(From([]float64{2.0, 3.0})); got != 6.0 {
		t.Fatalf("Product float: %v", got)
	}
}

func TestSort(t *testing.T) {
	got := collect(Sort(From([]int{3, 1, 4, 1, 5, 9, 2, 6})))
	if !slices.Equal(got, []int{1, 1, 2, 3, 4, 5, 6, 9}) {
		t.Fatalf("Sort: got %v", got)
	}
	gotStr := collect(Sort(From([]string{"banana", "apple", "cherry"})))
	if !slices.Equal(gotStr, []string{"apple", "banana", "cherry"}) {
		t.Fatalf("Sort strings: got %v", gotStr)
	}
	if len(collect(Sort(From([]int{})))) != 0 {
		t.Fatal("Sort empty")
	}
	// Sort materializes, so result is re-iterable
	s := Sort(From([]int{2, 1}))
	if !slices.Equal(collect(s), []int{1, 2}) || !slices.Equal(collect(s), []int{1, 2}) {
		t.Fatal("Sort result should be re-iterable")
	}
}
