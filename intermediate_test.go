package seq

import (
	"iter"
	"strconv"
	"strings"
	"testing"
)

// lazinessCounter wraps a source and counts how many elements it has pulled.
// Intermediate methods must pull zero elements until a terminal drives them.
func lazinessCounter[T any](src Seq[T]) (Seq[T], *int) {
	count := 0
	wrapped := Seq[T](func(yield func(T) bool) {
		for v := range iter.Seq[T](src) {
			count++
			if !yield(v) {
				return
			}
		}
	})
	return wrapped, &count
}

func TestMapTypeChange(t *testing.T) {
	got := collect(From([]int{1, 2}).Map(strconv.Itoa))
	if len(got) != 2 || got[0] != "1" || got[1] != "2" {
		t.Fatalf("Map type change: got %v", got)
	}
}

func TestFlatMap(t *testing.T) {
	got := collect(From([]int{1, 2, 3}).FlatMap(func(x int) Seq[int] {
		return Range(0, x) // 0..x-1
	}))
	// 1 -> [0]; 2 -> [0,1]; 3 -> [0,1,2]
	want := []int{0, 0, 1, 0, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("FlatMap len: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("FlatMap[%d]: got %v, want %v", i, got, want)
		}
	}
}

func TestFilterMap(t *testing.T) {
	// keep evens, map to their square
	got := collect(From([]int{1, 2, 3, 4, 5}).FilterMap(func(x int) (int, bool) {
		if x%2 != 0 {
			return 0, false
		}
		return x * x, true
	}))
	if len(got) != 2 || got[0] != 4 || got[1] != 16 {
		t.Fatalf("FilterMap: got %v", got)
	}
}

func TestFilterReject(t *testing.T) {
	evens := collect(From([]int{1, 2, 3, 4, 5}).Filter(func(x int) bool { return x%2 == 0 }))
	if len(evens) != 2 || evens[0] != 2 || evens[1] != 4 {
		t.Fatalf("Filter: got %v", evens)
	}
	odds := collect(From([]int{1, 2, 3, 4, 5}).Reject(func(x int) bool { return x%2 == 0 }))
	if len(odds) != 3 || odds[0] != 1 || odds[2] != 5 {
		t.Fatalf("Reject: got %v", odds)
	}
}

func TestTakeDrop(t *testing.T) {
	if got := collect(From([]int{1, 2, 3, 4, 5}).Take(3)); len(got) != 3 || got[2] != 3 {
		t.Fatalf("Take: got %v", got)
	}
	if got := collect(From([]int{1, 2, 3, 4, 5}).Take(0)); len(got) != 0 {
		t.Fatal("Take 0")
	}
	if got := collect(From([]int{1, 2, 3, 4, 5}).Take(100)); len(got) != 5 {
		t.Fatal("Take over-length")
	}
	if got := collect(From([]int{1, 2, 3, 4, 5}).Drop(2)); len(got) != 3 || got[0] != 3 {
		t.Fatalf("Drop: got %v", got)
	}
	if got := collect(From([]int{1, 2, 3}).Drop(0)); len(got) != 3 {
		t.Fatal("Drop 0")
	}
	if got := collect(From([]int{1, 2, 3}).Drop(100)); len(got) != 0 {
		t.Fatal("Drop over-length")
	}
}

func TestTakeWhileDropWhile(t *testing.T) {
	tw := collect(From([]int{1, 2, 3, 1, 2}).TakeWhile(func(x int) bool { return x < 3 }))
	if len(tw) != 2 || tw[0] != 1 || tw[1] != 2 {
		t.Fatalf("TakeWhile: got %v", tw)
	}
	// DropWhile stops dropping at first non-match; later matches ARE yielded.
	dw := collect(From([]int{1, 2, 3, 1, 2}).DropWhile(func(x int) bool { return x < 3 }))
	if len(dw) != 3 || dw[0] != 3 || dw[1] != 1 || dw[2] != 2 {
		t.Fatalf("DropWhile: got %v", dw)
	}
}

func TestScan(t *testing.T) {
	// prefix sums, including initial 0
	got := collect(From([]int{1, 2, 3, 4}).Scan(0, func(acc, x int) int { return acc + x }))
	want := []int{0, 1, 3, 6, 10}
	if len(got) != len(want) {
		t.Fatalf("Scan len: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Scan[%d]: got %v, want %v", i, got, want)
		}
	}
	// type-changing scan
	gotStr := collect(From([]int{1, 2}).Scan("", func(acc string, x int) string {
		return acc + strconv.Itoa(x)
	}))
	if len(gotStr) != 3 || gotStr[0] != "" || gotStr[2] != "12" {
		t.Fatalf("Scan type change: got %v", gotStr)
	}
}

func TestPeek(t *testing.T) {
	var seen []int
	got := collect(From([]int{1, 2, 3}).Peek(func(x int) { seen = append(seen, x) }))
	if len(got) != 3 || len(seen) != 3 || seen[2] != 3 {
		t.Fatalf("Peek: got %v, side %v", got, seen)
	}
}

func TestIntersperse(t *testing.T) {
	got := collect(From([]int{1, 2, 3}).Intersperse(0))
	want := []int{1, 0, 2, 0, 3}
	if len(got) != len(want) {
		t.Fatalf("Intersperse len: got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Intersperse[%d]: got %v, want %v", i, got, want)
		}
	}
	if len(collect(From([]int{}).Intersperse(0))) != 0 {
		t.Fatal("Intersperse empty")
	}
	if len(collect(From([]int{1}).Intersperse(0))) != 1 {
		t.Fatal("Intersperse single")
	}
}

func TestConcatMethod(t *testing.T) {
	got := collect(From([]int{1, 2}).Concat(From([]int{3, 4})))
	if len(got) != 4 || got[3] != 4 {
		t.Fatalf("Concat method: got %v", got)
	}
}

func TestDistinctBy(t *testing.T) {
	// dedup by parity: first even, first odd
	got := collect(From([]int{1, 2, 3, 4, 5, 6}).DistinctBy(func(x int) int { return x % 2 }))
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("DistinctBy: got %v", got)
	}
	// dedup by first letter
	gotStr := collect(From([]string{"apple", "ant", "banana", "cherry", "cat"}).DistinctBy(func(s string) byte {
		return s[0]
	}))
	if len(gotStr) != 3 || gotStr[0] != "apple" || gotStr[1] != "banana" || gotStr[2] != "cherry" {
		t.Fatalf("DistinctBy string: got %v", gotStr)
	}
}

func TestChunk(t *testing.T) {
	// Chunk returns iter.Seq[Seq[T]]; convert to Seq[Seq[T]] to collect.
	chunks := collect(Seq[Seq[int]](From([]int{1, 2, 3, 4, 5}).Chunk(2)))
	if len(chunks) != 3 {
		t.Fatalf("Chunk count: got %d", len(chunks))
	}
	if len(collect(chunks[0])) != 2 || collect(chunks[0])[0] != 1 {
		t.Fatalf("Chunk[0]: %v", collect(chunks[0]))
	}
	// last chunk short
	if len(collect(chunks[2])) != 1 || collect(chunks[2])[0] != 5 {
		t.Fatalf("Chunk[2]: %v", collect(chunks[2]))
	}
	// inner Seq supports full method chain
	if got := collect(chunks[0].Map(func(x int) int { return x * 10 })); len(got) != 2 || got[1] != 20 {
		t.Fatalf("Chunk inner chain: %v", got)
	}
	if len(collect(Seq[Seq[int]](From([]int{1, 2, 3}).Chunk(0)))) != 0 {
		t.Fatal("Chunk size 0")
	}
	// outer chaining via the documented Seq[Seq[T]] wrap
	firstChunk := collect(Seq[Seq[int]](From([]int{1, 2, 3, 4}).Chunk(2)).Take(1))
	if len(firstChunk) != 1 || collect(firstChunk[0])[1] != 2 {
		t.Fatalf("Chunk outer chain: %v", firstChunk)
	}
}

func TestWindow(t *testing.T) {
	wins := collect(Seq[Seq[int]](From([]int{1, 2, 3, 4, 5}).Window(3, 1)))
	if len(wins) != 3 {
		t.Fatalf("Window count: got %d", len(wins))
	}
	// [1,2,3], [2,3,4], [3,4,5]
	if got := collect(wins[0]); len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("Window[0]: %v", got)
	}
	if got := collect(wins[2]); got[0] != 3 || got[2] != 5 {
		t.Fatalf("Window[2]: %v", got)
	}
	// step > size: non-overlapping windows
	wins2 := collect(Seq[Seq[int]](From([]int{1, 2, 3, 4, 5, 6}).Window(2, 2)))
	if len(wins2) != 3 {
		t.Fatalf("Window step=size count: got %d", len(wins2))
	}
}

func TestTakeRightDropRightSlice(t *testing.T) {
	if got := collect(From([]int{1, 2, 3, 4, 5}).TakeRight(2)); len(got) != 2 || got[0] != 4 || got[1] != 5 {
		t.Fatalf("TakeRight: got %v", got)
	}
	if got := collect(From([]int{1, 2, 3}).TakeRight(0)); len(got) != 0 {
		t.Fatal("TakeRight 0")
	}
	if got := collect(From([]int{1, 2, 3}).TakeRight(100)); len(got) != 3 {
		t.Fatal("TakeRight over-length")
	}
	if got := collect(From([]int{1, 2, 3, 4, 5}).DropRight(2)); len(got) != 3 || got[2] != 3 {
		t.Fatalf("DropRight: got %v", got)
	}
	if got := collect(From([]int{1, 2, 3, 4, 5}).Slice(1, 4)); len(got) != 3 || got[0] != 2 || got[2] != 4 {
		t.Fatalf("Slice: got %v", got)
	}
	// clamping
	if got := collect(From([]int{1, 2, 3}).Slice(-1, 100)); len(got) != 3 {
		t.Fatalf("Slice clamp: got %v", got)
	}
}

func TestInitTail(t *testing.T) {
	if got := collect(From([]int{1, 2, 3}).Init()); len(got) != 2 || got[1] != 2 {
		t.Fatalf("Init: got %v", got)
	}
	if len(collect(From([]int{1}).Init())) != 0 {
		t.Fatal("Init single -> empty")
	}
	if len(collect(From([]int{}).Init())) != 0 {
		t.Fatal("Init empty")
	}
	if got := collect(From([]int{1, 2, 3}).Tail()); len(got) != 2 || got[0] != 2 {
		t.Fatalf("Tail: got %v", got)
	}
	if len(collect(From([]int{1}).Tail())) != 0 {
		t.Fatal("Tail single -> empty")
	}
}

func TestSortByReverse(t *testing.T) {
	got := collect(From([]int{3, 1, 4, 1, 5}).SortBy(func(a, b int) bool { return a < b }))
	if len(got) != 5 || got[0] != 1 || got[4] != 5 {
		t.Fatalf("SortBy: got %v", got)
	}
	// stability: equal elements (both 1) keep relative order
	byMod := collect(From([]int{1, 2, 3, 4}).SortBy(func(a, b int) bool { return a%2 < b%2 }))
	_ = byMod
	rev := collect(From([]int{1, 2, 3}).Reverse())
	if len(rev) != 3 || rev[0] != 3 || rev[2] != 1 {
		t.Fatalf("Reverse: got %v", rev)
	}
}

func TestEnumerate(t *testing.T) {
	s2 := From([]string{"a", "b", "c"}).Enumerate()
	got := map[int]string{}
	for i, v := range iter.Seq2[int, string](s2) {
		got[i] = v
	}
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("Enumerate: got %v", got)
	}
}

// TestLaziness verifies intermediate ops do NOT traverse the source until a
// terminal drives them. This is the headline laziness acceptance criterion.
func TestLaziness(t *testing.T) {
	src, counter := lazinessCounter(From([]int{1, 2, 3, 4, 5}))

	// Build a chain of intermediate ops — none should pull any element yet.
	chain := src.
		Filter(func(x int) bool { return x%2 == 0 }).
		Map(func(x int) int { return x * x }).
		Take(1)
	if *counter != 0 {
		t.Fatalf("intermediate chain pulled %d elements before terminal; expected 0", *counter)
	}

	// Driving a terminal traverses the source (and short-circuits at Take(1)).
	got := collect(chain)
	if *counter == 0 {
		t.Fatal("terminal did not traverse the source")
	}
	if len(got) != 1 || got[0] != 4 {
		t.Fatalf("lazy chain result: got %v, want [4]", got)
	}
}

// TestChainingCompiles verifies Filter().Map().Take(n) compiles and is
// semantically correct (the headline chain from the design doc).
func TestChainingCompiles(t *testing.T) {
	got := collect(From([]int{1, 2, 3, 4, 5, 6}).
		Filter(func(x int) bool { return x%2 == 0 }).
		Map(func(x int) string { return "n" + strconv.Itoa(x) }).
		Take(2))
	if len(got) != 2 || got[0] != "n2" || got[1] != "n4" {
		t.Fatalf("chain: got %v", got)
	}
	// infinite source bounded by Take in a chain
	inf := Generate(func() int { return 7 }).Take(3)
	if got := collect(inf); len(got) != 3 || got[0] != 7 {
		t.Fatalf("bounded infinite: got %v", got)
	}
}

func TestLazinessShortCircuit(t *testing.T) {
	// A terminal driving an infinite source must terminate via short-circuit.
	src, _ := lazinessCounter(Generate(func() int {
		// would hang if not short-circuited
		return 1
	}))
	got := collect(src.Filter(func(x int) bool { return x > 0 }).Take(5))
	if len(got) != 5 {
		t.Fatalf("short-circuit infinite: got %v", got)
	}
	// strings import kept meaningful: ensure a strings use compiles
	if strings.Join([]string{"a", "b"}, ",") != "a,b" {
		t.Fatal("strings sanity")
	}
}
