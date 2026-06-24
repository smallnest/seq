package seq

import (
	"iter"
	"slices"
	"testing"
)

// collectC collects a SeqComparable via its zero-cost conversion to Seq.
func collectC[T comparable](s SeqComparable[T]) []T {
	return slices.Collect(iter.Seq[T](s))
}

func TestEntryFunctionsAndDowngrades(t *testing.T) {
	// Entry + downgrade chain compiles and round-trips.
	num := Numbers(From([]int{1, 2, 3}))
	ord := num.Ordered()
	cmp2 := ord.Comparable()
	bare := cmp2.Seq()
	if got := collect(bare); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("downgrade round-trip: got %v", got)
	}
	// Comparable/Ordered entry on appropriate types
	sc := Comparable(From([]string{"a", "b"}))
	so := Ordered(From([]int{3, 1, 2}))
	if got := collectC(sc); len(got) != 2 {
		t.Fatalf("Comparable entry: %v", got)
	}
	_ = so
}

func TestHeadlineChainNumbersDistinctSum(t *testing.T) {
	// The design-doc headline: Numbers(From(xs)).Distinct().Sum()
	got := Numbers(From([]int{1, 2, 2, 3, 3, 3})).Distinct().Sum()
	if got != 6 { // 1+2+3
		t.Fatalf("headline chain: got %d, want 6", got)
	}
}

func TestComparableStringDistinct(t *testing.T) {
	// string via Comparable().Distinct()
	got := collectC(Comparable(From([]string{"a", "b", "a", "c", "b"})).Distinct())
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("string Distinct: got %v", got)
	}
}

func TestNumbersFloatMean(t *testing.T) {
	// float via Numbers().Mean()
	got := Numbers(From([]float64{1.0, 2.0, 3.0, 4.0})).Mean()
	if got != 2.5 {
		t.Fatalf("float Mean: got %v, want 2.5", got)
	}
}

func TestSeqOrderedMinMaxSort(t *testing.T) {
	so := Ordered(From([]int{3, 1, 4, 1, 5, 9, 2, 6}))
	mx, ok := so.Max()
	if !ok || mx != 9 {
		t.Fatalf("Max: (%d,%v)", mx, ok)
	}
	mn, ok := so.Min()
	if !ok || mn != 1 {
		t.Fatalf("Min: (%d,%v)", mn, ok)
	}
	// Sort returns SeqOrdered, collectable
	sorted := so.Sort()
	// re-derive a fresh source since Sort is fine on lazy seq
	got := slices.Collect(iter.Seq[int](sorted))
	if !slices.Equal(got, []int{1, 1, 2, 3, 4, 5, 6, 9}) {
		t.Fatalf("Sort: got %v", got)
	}
}

func TestSeqComparableContainsIndexOfCountValues(t *testing.T) {
	sc := Comparable(From([]int{1, 2, 3, 2, 1}))
	if !sc.Contains(2) {
		t.Fatal("Contains 2")
	}
	if sc.Contains(9) {
		t.Fatal("Contains 9")
	}
	if idx, ok := sc.IndexOf(2); !ok || idx != 1 {
		t.Fatalf("IndexOf 2: (%d,%v)", idx, ok)
	}
	cv := sc.CountValues()
	if cv[1] != 2 || cv[2] != 2 || cv[3] != 1 {
		t.Fatalf("CountValues: %v", cv)
	}
}

func TestSeqComparableSetOps(t *testing.T) {
	a := Comparable(From([]int{1, 2, 3}))
	b := Comparable(From([]int{2, 3, 4}))
	if got := collectC(a.Union(b)); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Union: %v", got)
	}
	if got := collectC(a.Intersect(b)); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("Intersect: %v", got)
	}
	if got := collectC(a.Difference(b)); !slices.Equal(got, []int{1}) {
		t.Fatalf("Difference: %v", got)
	}
	if !a.Equal(Comparable(From([]int{1, 2, 3}))) {
		t.Fatal("Equal true")
	}
	if a.Equal(Comparable(From([]int{1, 2}))) {
		t.Fatal("Equal false")
	}
}

func TestSubtypeTPreservingChain(t *testing.T) {
	// Filter/Take etc. on each subtype return the SAME subtype, chain intact.
	got := Numbers(From([]int{1, 2, 3, 4, 5, 6})).
		Filter(func(x int) bool { return x%2 == 0 }).
		Take(2).
		Sum()
	if got != 6 { // 2+4
		t.Fatalf("SeqNumeric chain: got %d, want 6", got)
	}
	// SeqOrdered chain: Sort then Take
	got2 := slices.Collect(iter.Seq[int](Ordered(From([]int{3, 1, 2})).Sort().Take(2)))
	if !slices.Equal(got2, []int{1, 2}) {
		t.Fatalf("SeqOrdered chain: got %v", got2)
	}
	// SeqComparable chain: Distinct then Drop
	got3 := collectC(Comparable(From([]int{1, 2, 2, 3})).Distinct().Drop(1))
	if !slices.Equal(got3, []int{2, 3}) {
		t.Fatalf("SeqComparable chain: got %v", got3)
	}
}

func TestSubtypePeekRejectTakeWhileDropWhile(t *testing.T) {
	var seen []int
	got := Numbers(From([]int{1, 2, 3, 4, 5})).
		Peek(func(x int) { seen = append(seen, x) }).
		Reject(func(x int) bool { return x%2 == 0 }).
		TakeWhile(func(x int) bool { return x < 4 }).
		Sum()
	if got != 4 { // odds <4: 1,3
		t.Fatalf("peek/reject/takewhile chain: got %d, want 4", got)
	}
	if len(seen) != 5 {
		t.Fatalf("peek saw %d, want 5", len(seen))
	}
}

func TestSubtypeDowngradeFullChain(t *testing.T) {
	// Numbers().Ordered().Comparable() then Distinct — full downgrade chain.
	got := collectC(
		Numbers(From([]int{3, 1, 3, 2, 1})).
			Ordered().    // SeqNumeric -> SeqOrdered
			Comparable(). // SeqOrdered -> SeqComparable
			Distinct(),
	)
	if !slices.Equal(got, []int{3, 1, 2}) {
		t.Fatalf("full downgrade chain Distinct: got %v", got)
	}
}

func TestSubtypeMapDropsToBareSeq(t *testing.T) {
	// Map changes T, so it's not on the subtype; use Seq() to drop back.
	got := collect(
		Numbers(From([]int{1, 2, 3})).
			Ordered().
			Comparable().
			Seq(). // drop to bare Seq[T] before Map
			Map(func(x int) int { return x * 10 }),
	)
	if !slices.Equal(got, []int{10, 20, 30}) {
		t.Fatalf("Map after downgrade: got %v", got)
	}
}

func TestSubtypeCollectTerminals(t *testing.T) {
	if got := Numbers(From([]int{1, 2, 3})).Collect(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("SeqNumeric.Collect: %v", got)
	}
	if got := Ordered(From([]int{1, 2, 3})).Collect(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("SeqOrdered.Collect: %v", got)
	}
	if got := Comparable(From([]int{1, 2, 3})).Collect(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("SeqComparable.Collect: %v", got)
	}
}

func TestSubtypeProductEmpty(t *testing.T) {
	// Product empty -> 1; Sum empty -> 0
	if got := Numbers(From([]int{})).Product(); got != 1 {
		t.Fatalf("Product empty: %d", got)
	}
	if got := Numbers(From([]int{})).Sum(); got != 0 {
		t.Fatalf("Sum empty: %d", got)
	}
}
