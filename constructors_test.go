package seq

import (
	"iter"
	"slices"
	"testing"
)

// collect drains a Seq[T] into a slice. We use slices.Collect over the
// zero-cost iter.Seq[T] conversion so this test file does not depend on the
// Collect method (issue #7). Once Collect lands, these tests remain valid.
func collect[T any](s Seq[T]) []T {
	return slices.Collect(iter.Seq[T](s))
}

func TestFrom(t *testing.T) {
	got := collect(From([]int{1, 2, 3}))
	want := []int{1, 2, 3}
	if !slices.Equal(got, want) {
		t.Fatalf("From: got %v, want %v", got, want)
	}
	// empty slice -> empty seq
	if len(collect(From([]int{}))) != 0 {
		t.Fatal("From empty slice should yield nothing")
	}
	// slice-backed source is re-iterable
	s := From([]int{7, 8})
	if !slices.Equal(collect(s), []int{7, 8}) || !slices.Equal(collect(s), []int{7, 8}) {
		t.Fatal("From should be re-iterable")
	}
}

func TestOf(t *testing.T) {
	got := collect(Of("a", "b", "c"))
	want := []string{"a", "b", "c"}
	if !slices.Equal(got, want) {
		t.Fatalf("Of: got %v, want %v", got, want)
	}
	if len(collect(Of[int]())) != 0 {
		t.Fatal("Of with no args should be empty")
	}
}

func TestEmpty(t *testing.T) {
	if len(collect(Empty[int]())) != 0 {
		t.Fatal("Empty should yield nothing")
	}
}

func TestRange(t *testing.T) {
	cases := []struct {
		name       string
		start, end int
		want       []int
	}{
		{"ascending", 0, 3, []int{0, 1, 2}},
		{"single", 5, 6, []int{5}},
		{"descending", 3, 0, []int{3, 2, 1}},
		{"equal empty", 2, 2, []int{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := collect(Range(c.start, c.end))
			if !slices.Equal(got, c.want) {
				t.Fatalf("Range(%d,%d): got %v, want %v", c.start, c.end, got, c.want)
			}
		})
	}
}

func TestRangeStep(t *testing.T) {
	cases := []struct {
		name             string
		start, end, step int
		want             []int
	}{
		{"asc step 2", 0, 10, 2, []int{0, 2, 4, 6, 8}},
		{"desc step 3", 10, 0, 3, []int{10, 7, 4, 1}},
		{"non-positive step empty", 0, 5, 0, []int{}},
		{"negative step empty", 0, 5, -1, []int{}},
		{"equal empty", 3, 3, 2, []int{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := collect(RangeStep(c.start, c.end, c.step))
			if !slices.Equal(got, c.want) {
				t.Fatalf("RangeStep(%d,%d,%d): got %v, want %v", c.start, c.end, c.step, got, c.want)
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	got := collect(Repeat(3, "x"))
	if !slices.Equal(got, []string{"x", "x", "x"}) {
		t.Fatalf("Repeat: got %v", got)
	}
	if len(collect(Repeat(0, "x"))) != 0 {
		t.Fatal("Repeat n=0 should be empty")
	}
	if len(collect(Repeat(-2, "x"))) != 0 {
		t.Fatal("Repeat n<0 should be empty")
	}
}

// takeN bounds an (possibly infinite) source for testing without depending on
// the Take method (issue #6).
func takeN[T any](s Seq[T], n int) []T {
	var out []T
	for v := range iter.Seq[T](s) {
		if len(out) >= n {
			break
		}
		out = append(out, v)
	}
	return out
}

func TestRepeatInfBounded(t *testing.T) {
	// Infinite source must not hang when bounded.
	got := takeN(RepeatInf(7), 4)
	if !slices.Equal(got, []int{7, 7, 7, 7}) {
		t.Fatalf("RepeatInf bounded: got %v", got)
	}
}

func TestGenerateBounded(t *testing.T) {
	i := 0
	src := Generate(func() int { i++; return i })
	got := takeN(src, 3)
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("Generate bounded: got %v", got)
	}
}

func TestIterateBounded(t *testing.T) {
	// init, f(init), f(f(init))...
	src := Iterate(1, func(x int) int { return x * 2 })
	got := takeN(src, 5)
	if !slices.Equal(got, []int{1, 2, 4, 8, 16}) {
		t.Fatalf("Iterate bounded: got %v", got)
	}
}

func TestFromChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	got := collect(FromChannel(ch))
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("FromChannel: got %v", got)
	}
	// one-shot: iterating again yields nothing.
	if len(collect(FromChannel(ch))) != 0 {
		t.Fatal("FromChannel should be a one-shot source after drain")
	}

	// early break does not block (channel stays open with remaining items).
	ch2 := make(chan int, 5)
	for _, v := range []int{10, 11, 12, 13, 14} {
		ch2 <- v
	}
	got2 := takeN(FromChannel(ch2), 2)
	if !slices.Equal(got2, []int{10, 11}) {
		t.Fatalf("FromChannel early break: got %v", got2)
	}
}

func TestFromMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	s2 := FromMap(m)
	got := map[string]int{}
	for k, v := range iter.Seq2[string, int](s2) {
		got[k] = v
	}
	if len(got) != 3 || got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("FromMap: got %v", got)
	}
}
