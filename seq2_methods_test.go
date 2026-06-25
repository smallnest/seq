package seq

import (
	"slices"
	"testing"
)

// helper: build a Seq2 from a slice of pairs (deterministic order).
func seq2FromPairs[K, V any](pairs []Pair[K, V]) Seq2[K, V] {
	return Entries(pairs)
}

func TestMapValues(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}})
	got := ToMap(src.MapValues(func(v int) int { return v * 10 }))
	if got["a"] != 10 || got["b"] != 20 {
		t.Fatalf("MapValues: got %v", got)
	}
	// type-changing value
	gotStr := ToMap(src.MapValues(func(v int) string { return "v" + itoa(v) }))
	if gotStr["a"] != "v1" || gotStr["b"] != "v2" {
		t.Fatalf("MapValues type change: %v", gotStr)
	}
}

func TestMapKeys(t *testing.T) {
	src := seq2FromPairs([]Pair[int, string]{{1, "a"}, {2, "b"}})
	got := ToMap(src.MapKeys(func(k int) string { return itoa(k) }))
	if got["1"] != "a" || got["2"] != "b" {
		t.Fatalf("MapKeys: got %v", got)
	}
}

func TestMapBoth(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}})
	got := ToMap(src.Map(func(k string, v int) (string, int) { return k + "!", v + 100 }))
	if got["a!"] != 101 || got["b!"] != 102 {
		t.Fatalf("Map both: got %v", got)
	}
}

func TestFilter2(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}})
	got := ToMap(src.Filter(func(k string, v int) bool { return v > 1 }))
	if len(got) != 2 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("Filter2: got %v", got)
	}
}

func TestKeysValues(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}})
	keys := collect(src.Keys())
	// order preserved
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("Keys: got %v", keys)
	}
	vals := collect(src.Values())
	if len(vals) != 3 || vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Fatalf("Values: got %v", vals)
	}
	// Keys()/Values() return Seq, supporting full method chain
	evens := collect(src.Values().Filter(func(v int) bool { return v%2 == 0 }))
	if !slices.Equal(evens, []int{2}) {
		t.Fatalf("Values().Filter chain: got %v", evens)
	}
}

func TestForEach2(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}})
	var got []string
	src.ForEach(func(k string, v int) { got = append(got, k+itoa(v)) })
	if len(got) != 2 || got[0] != "a1" || got[1] != "b2" {
		t.Fatalf("ForEach2: %v", got)
	}
}

func TestFold2(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}})
	sum := src.Fold(0, func(acc int, k string, v int) int { return acc + v })
	if sum != 6 {
		t.Fatalf("Fold2: %d", sum)
	}
	// type-changing fold
	joined := src.Fold("", func(acc string, k string, v int) string {
		return acc + k + itoa(v)
	})
	if joined != "a1b2c3" {
		t.Fatalf("Fold2 type change: %q", joined)
	}
}

func TestCount2(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}})
	if n := src.Count(); n != 3 {
		t.Fatalf("Count2: %d", n)
	}
	if n := seq2FromPairs([]Pair[string, int]{}).Count(); n != 0 {
		t.Fatalf("Count2 empty: %d", n)
	}
}

func TestFind2(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}})
	p, ok := src.Find(func(k string, v int) bool { return v == 2 }).Get()
	if !ok || p.Left != "b" || p.Right != 2 {
		t.Fatalf("Find2: (%+v,%v)", p, ok)
	}
	if src.Find(func(k string, v int) bool { return v == 99 }).IsPresent() {
		t.Fatal("Find2 absent")
	}
}

func TestAnyAll2(t *testing.T) {
	src := seq2FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}})
	if !src.Any(func(k string, v int) bool { return v == 2 }) {
		t.Fatal("Any2 true")
	}
	if src.Any(func(k string, v int) bool { return v == 99 }) {
		t.Fatal("Any2 false")
	}
	if !src.All(func(k string, v int) bool { return v < 10 }) {
		t.Fatal("All2 true")
	}
	if src.All(func(k string, v int) bool { return v < 2 }) {
		t.Fatal("All2 false")
	}
}

// itoa is a tiny local int->string to avoid pulling strconv + keep test self-contained.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf []byte
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
