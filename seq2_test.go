package seq

import (
	"iter"
	"testing"
)

func TestToMap(t *testing.T) {
	s2 := Zip(From([]string{"a", "b", "c"}), From([]int{1, 2, 3}))
	got := ToMap(s2)
	want := map[string]int{"a": 1, "b": 2, "c": 3}
	if len(got) != 3 || got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("ToMap: got %v, want %v", got, want)
	}
	// duplicate key: last wins
	dup := ToMap(Zip(From([]string{"a", "a", "b"}), From([]int{1, 2, 3})))
	if dup["a"] != 2 || dup["b"] != 3 {
		t.Fatalf("ToMap dup last-wins: got %v", dup)
	}
	// empty -> non-nil empty map
	empty := ToMap(Zip(From([]string{}), From([]int{})))
	if empty == nil || len(empty) != 0 {
		t.Fatalf("ToMap empty: got %v", empty)
	}
}

func TestCollectPairs(t *testing.T) {
	s2 := Zip(From([]string{"a", "b"}), From([]int{1, 2}))
	got := CollectPairs(s2)
	want := []Pair[string, int]{{"a", 1}, {"b", 2}}
	if len(got) != len(want) {
		t.Fatalf("CollectPairs len: %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("CollectPairs[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
	// preserves order
	got2 := CollectPairs(Zip(From([]int{3, 1, 2}), From([]string{"c", "a", "b"})))
	if got2[0].Left != 3 || got2[1].Left != 1 || got2[2].Left != 2 {
		t.Fatalf("CollectPairs order: %+v", got2)
	}
	// empty
	if len(CollectPairs(Zip(From([]int{}), From([]int{})))) != 0 {
		t.Fatal("CollectPairs empty")
	}
}

func TestEntries(t *testing.T) {
	pairs := []Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}}
	s2 := Entries(pairs)
	// round-trips with CollectPairs
	got := CollectPairs(s2)
	if len(got) != 3 || got[0] != pairs[0] || got[2] != pairs[2] {
		t.Fatalf("Entries round-trip: %+v", got)
	}
	// re-iterable (slice-backed)
	if len(CollectPairs(s2)) != 3 {
		t.Fatal("Entries should be re-iterable")
	}
	// empty slice
	if len(CollectPairs(Entries([]Pair[int, int]{}))) != 0 {
		t.Fatal("Entries empty")
	}
}

func TestAssociate(t *testing.T) {
	// associate words to their lengths
	src := From([]string{"go", "rust", "c"})
	s2 := Associate(src, func(w string) (string, int) { return w, len(w) })
	got := ToMap(s2)
	want := map[string]int{"go": 2, "rust": 4, "c": 1}
	if len(got) != 3 || got["go"] != 2 || got["rust"] != 4 || got["c"] != 1 {
		t.Fatalf("Associate: got %v, want %v", got, want)
	}
	// order preserved via CollectPairs
	pairs := CollectPairs(Associate(From([]int{10, 20, 30}), func(n int) (int, string) { return n, "x" }))
	if len(pairs) != 3 || pairs[0].Left != 10 || pairs[1].Left != 20 || pairs[2].Left != 30 {
		t.Fatalf("Associate order: %+v", pairs)
	}
	// empty source -> empty Seq2 (drains to nothing)
	empty := Associate(From([]int{}), func(n int) (int, int) { return n, n })
	count := 0
	for range iter.Seq2[int, int](empty) {
		count++
	}
	if count != 0 {
		t.Fatalf("Associate empty should yield nothing, got %d", count)
	}
}
