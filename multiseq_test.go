package seq

import (
	"iter"
	"slices"
	"testing"
)

func collect2[A, B any](s Seq2[A, B]) []Pair[A, B] {
	var out []Pair[A, B]
	for k, v := range iter.Seq2[A, B](s) {
		out = append(out, Pair[A, B]{Left: k, Right: v})
	}
	return out
}

func TestZip(t *testing.T) {
	got := collect2(Zip(From([]string{"a", "b", "c"}), From([]int{1, 2, 3})))
	want := []Pair[string, int]{{"a", 1}, {"b", 2}, {"c", 3}}
	if len(got) != len(want) {
		t.Fatalf("Zip len: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Zip[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
	// stops at shorter
	short := collect2(Zip(From([]int{1, 2, 3, 4, 5}), From([]int{10, 20})))
	if len(short) != 2 || short[0].Left != 1 || short[1].Right != 20 {
		t.Fatalf("Zip short: got %+v", short)
	}
	// empty input
	if len(collect2(Zip(From([]int{}), From([]int{1})))) != 0 {
		t.Fatal("Zip empty left")
	}
}

func TestZipWith(t *testing.T) {
	got := collect(ZipWith(From([]int{1, 2, 3}), From([]int{10, 20, 30}), func(a, b int) int { return a + b }))
	if !slices.Equal(got, []int{11, 22, 33}) {
		t.Fatalf("ZipWith: got %v", got)
	}
	// type-changing merge
	gotStr := collect(ZipWith(From([]int{1, 2}), From([]string{"x", "y"}), func(a int, b string) string {
		return b + string(rune('0'+a))
	}))
	if !slices.Equal(gotStr, []string{"x1", "y2"}) {
		t.Fatalf("ZipWith type change: got %v", gotStr)
	}
	// stops at shorter
	short := collect(ZipWith(From([]int{1, 2, 3}), From([]int{1}), func(a, b int) int { return a * b }))
	if !slices.Equal(short, []int{1}) {
		t.Fatalf("ZipWith short: got %v", short)
	}
}

func TestZipMap(t *testing.T) {
	got := ZipMap(From([]string{"a", "b", "c"}), From([]int{1, 2, 3}))
	want := map[string]int{"a": 1, "b": 2, "c": 3}
	if len(got) != 3 || got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("ZipMap: got %v, want %v", got, want)
	}
	// duplicate key: last wins
	dup := ZipMap(From([]string{"a", "a"}), From([]int{1, 2}))
	if dup["a"] != 2 {
		t.Fatalf("ZipMap dup last-wins: got %v", dup)
	}
	// stops at shorter (vals shorter)
	short := ZipMap(From([]string{"a", "b", "c"}), From([]int{1}))
	if len(short) != 1 || short["a"] != 1 {
		t.Fatalf("ZipMap short: got %v", short)
	}
	// empty -> empty non-nil map
	empty := ZipMap(From([]string{}), From([]int{}))
	if empty == nil || len(empty) != 0 {
		t.Fatalf("ZipMap empty: got %v", empty)
	}
}

func TestZip3(t *testing.T) {
	got := collect(Zip3(From([]int{1, 2}), From([]string{"a", "b"}), From([]float64{1.5, 2.5})))
	if len(got) != 2 {
		t.Fatalf("Zip3 len: %d", len(got))
	}
	if got[0].First != 1 || got[0].Second != "a" || got[0].Third != 1.5 {
		t.Fatalf("Zip3[0]: %+v", got[0])
	}
	if got[1].First != 2 || got[1].Second != "b" || got[1].Third != 2.5 {
		t.Fatalf("Zip3[1]: %+v", got[1])
	}
	// stops at shortest (third shortest)
	short := collect(Zip3(From([]int{1, 2, 3}), From([]int{4, 5, 6}), From([]int{7})))
	if len(short) != 1 || short[0].Third != 7 {
		t.Fatalf("Zip3 short: %+v", short)
	}
}

func TestZip4(t *testing.T) {
	got := collect(Zip4(From([]int{1, 2}), From([]int{3, 4}), From([]int{5, 6}), From([]int{7, 8})))
	if len(got) != 2 {
		t.Fatalf("Zip4 len: %d", len(got))
	}
	if got[0].First != 1 || got[0].Second != 3 || got[0].Third != 5 || got[0].Fourth != 7 {
		t.Fatalf("Zip4[0]: %+v", got[0])
	}
	// stops at shortest (fourth shortest)
	short := collect(Zip4(From([]int{1, 2, 3}), From([]int{4, 5, 6}), From([]int{7, 8, 9}), From([]int{0})))
	if len(short) != 1 {
		t.Fatalf("Zip4 short: %d", len(short))
	}
}

func TestUnzip(t *testing.T) {
	pairs := Zip(From([]string{"a", "b", "c"}), From([]int{1, 2, 3}))
	as, bs := Unzip(pairs)
	if !slices.Equal(collect(as), []string{"a", "b", "c"}) {
		t.Fatalf("Unzip keys: %v", collect(as))
	}
	if !slices.Equal(collect(bs), []int{1, 2, 3}) {
		t.Fatalf("Unzip vals: %v", collect(bs))
	}
	// both sides independent, source drained once
	if !slices.Equal(collect(as), []string{"a", "b", "c"}) || !slices.Equal(collect(bs), []int{1, 2, 3}) {
		t.Fatal("Unzip re-iterable both sides")
	}
	// empty
	ea, eb := Unzip(Zip(From([]int{}), From([]int{})))
	if len(collect(ea)) != 0 || len(collect(eb)) != 0 {
		t.Fatal("Unzip empty")
	}
}

func TestFlatten(t *testing.T) {
	nested := From([]Seq[int]{
		From([]int{1, 2}),
		From([]int{3}),
		From([]int{4, 5, 6}),
	})
	got := collect(Flatten(nested))
	if !slices.Equal(got, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("Flatten: got %v", got)
	}
	// one level only — inner Seq stays as-is if not flattened again
	nested2 := From([]Seq[int]{From([]int{1}), Empty[int]()})
	if !slices.Equal(collect(Flatten(nested2)), []int{1}) {
		t.Fatal("Flatten with empty inner")
	}
	// empty outer
	if len(collect(Flatten(From([]Seq[int]{})))) != 0 {
		t.Fatal("Flatten empty outer")
	}
}

func TestConcat(t *testing.T) {
	got := collect(Concat(From([]int{1, 2}), From([]int{3}), From([]int{4, 5})))
	if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("Concat: got %v", got)
	}
	// no args
	if len(collect(Concat[int]())) != 0 {
		t.Fatal("Concat no args")
	}
	// with empty segments
	got2 := collect(Concat(From([]int{1}), Empty[int](), From([]int{2, 3})))
	if !slices.Equal(got2, []int{1, 2, 3}) {
		t.Fatalf("Concat with empty: %v", got2)
	}
}

func TestInterleave(t *testing.T) {
	got := collect(Interleave(From([]int{1, 4, 7}), From([]int{2, 5, 8}), From([]int{3, 6, 9})))
	if !slices.Equal(got, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Fatalf("Interleave equal length: got %v", got)
	}
	// uneven lengths: round-robin, skipping exhausted
	got2 := collect(Interleave(From([]int{1, 2, 3, 4}), From([]int{10}), From([]int{100, 200, 300})))
	// round: 1, 10, 100; then a exhausted -> 2, _, 200; then 3, _, 300; then 4
	want := []int{1, 10, 100, 2, 200, 3, 300, 4}
	if !slices.Equal(got2, want) {
		t.Fatalf("Interleave uneven: got %v, want %v", got2, want)
	}
	// no args
	if len(collect(Interleave[int]())) != 0 {
		t.Fatal("Interleave no args")
	}
}
