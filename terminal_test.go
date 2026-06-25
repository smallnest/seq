package seq

import (
	"iter"
	"strconv"
	"testing"
)

func TestCollect(t *testing.T) {
	got := From([]int{1, 2, 3}).Collect()
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("Collect: got %v", got)
	}
}

func TestForEach(t *testing.T) {
	var sum int
	From([]int{1, 2, 3}).ForEach(func(x int) { sum += x })
	if sum != 6 {
		t.Fatalf("ForEach sum: %d", sum)
	}
}

func TestForEachIndexed(t *testing.T) {
	var pairs []string
	From([]string{"a", "b"}).ForEachIndexed(func(i int, v string) {
		pairs = append(pairs, strconv.Itoa(i)+v)
	})
	if len(pairs) != 2 || pairs[0] != "0a" || pairs[1] != "1b" {
		t.Fatalf("ForEachIndexed: %v", pairs)
	}
}

func TestFold(t *testing.T) {
	sum := From([]int{1, 2, 3, 4}).Fold(10, func(acc, x int) int { return acc + x })
	if sum != 20 {
		t.Fatalf("Fold: %d", sum)
	}
	// type-changing fold: int -> string
	joined := From([]int{1, 2, 3}).Fold("", func(acc string, x int) string {
		return acc + strconv.Itoa(x)
	})
	if joined != "123" {
		t.Fatalf("Fold type change: %q", joined)
	}
	if got := From([]int{}).Fold(7, func(a, x int) int { return a + x }); got != 7 {
		t.Fatalf("Fold empty: %d", got)
	}
}

func TestReduce(t *testing.T) {
	r, ok := From([]int{1, 2, 3, 4}).Reduce(func(a, b int) int { return a + b }).Get()
	if !ok || r != 10 {
		t.Fatalf("Reduce: (%d,%v)", r, ok)
	}
	if From([]int{}).Reduce(func(a, b int) int { return a + b }).IsPresent() {
		t.Fatal("Reduce empty should be None")
	}
	if r, ok := From([]int{42}).Reduce(func(a, b int) int { return a + b }).Get(); !ok || r != 42 {
		t.Fatalf("Reduce single: (%d,%v)", r, ok)
	}
}

func TestCountCountBy(t *testing.T) {
	if n := From([]int{1, 2, 3, 4, 5}).Count(); n != 5 {
		t.Fatalf("Count: %d", n)
	}
	if n := From([]int{}).Count(); n != 0 {
		t.Fatalf("Count empty: %d", n)
	}
	if n := From([]int{1, 2, 3, 4, 5, 6}).CountBy(func(x int) bool { return x%2 == 0 }); n != 3 {
		t.Fatalf("CountBy: %d", n)
	}
}

func TestGroupCount(t *testing.T) {
	got := From([]string{"a", "b", "a", "c", "b", "a"}).GroupCount(func(s string) string { return s })
	want := map[string]int{"a": 3, "b": 2, "c": 1}
	if len(got) != 3 || got["a"] != 3 || got["b"] != 2 || got["c"] != 1 {
		t.Fatalf("GroupCount: got %v, want %v", got, want)
	}
	// by parity
	byParity := From([]int{1, 2, 3, 4, 5}).GroupCount(func(x int) int { return x % 2 })
	if byParity[0] != 2 || byParity[1] != 3 {
		t.Fatalf("GroupCount parity: %v", byParity)
	}
}

func TestGroupBy(t *testing.T) {
	got := From([]int{1, 2, 3, 4, 5, 6}).GroupBy(func(x int) string {
		if x%2 == 0 {
			return "even"
		}
		return "odd"
	})
	if len(got["even"]) != 3 || got["even"][0] != 2 || got["even"][2] != 6 {
		t.Fatalf("GroupBy even: %v", got["even"])
	}
	if len(got["odd"]) != 3 || got["odd"][0] != 1 {
		t.Fatalf("GroupBy odd: %v", got["odd"])
	}
}

func TestKeyBy(t *testing.T) {
	got := From([]string{"apple", "banana", "cherry"}).KeyBy(func(s string) byte { return s[0] })
	if len(got) != 3 || got['a'] != "apple" || got['c'] != "cherry" {
		t.Fatalf("KeyBy: %v", got)
	}
	// dup key last wins
	dup := From([]string{"a", "b", "c"}).KeyBy(func(s string) string { return "k" })
	if dup["k"] != "c" {
		t.Fatalf("KeyBy dup: %v", dup)
	}
}

func TestFindFamily(t *testing.T) {
	v, ok := From([]int{1, 2, 3, 4}).Find(func(x int) bool { return x > 2 }).Get()
	if !ok || v != 3 {
		t.Fatalf("Find: (%d,%v)", v, ok)
	}
	if From([]int{1, 2}).Find(func(x int) bool { return x > 9 }).IsPresent() {
		t.Fatal("Find absent")
	}
	idx, ok := From([]int{1, 2, 3, 4}).FindIndex(func(x int) bool { return x > 2 }).Get()
	if !ok || idx != 2 {
		t.Fatalf("FindIndex: (%d,%v)", idx, ok)
	}
	// FindLast: last even
	lv, ok := From([]int{1, 2, 3, 4, 5, 6}).FindLast(func(x int) bool { return x%2 == 0 }).Get()
	if !ok || lv != 6 {
		t.Fatalf("FindLast: (%d,%v)", lv, ok)
	}
	if From([]int{1, 3, 5}).FindLast(func(x int) bool { return x%2 == 0 }).IsPresent() {
		t.Fatal("FindLast absent")
	}
	li, ok := From([]int{1, 2, 3, 2}).FindLastIndex(func(x int) bool { return x == 2 }).Get()
	if !ok || li != 3 {
		t.Fatalf("FindLastIndex: (%d,%v)", li, ok)
	}
}

func TestAnyAllNone(t *testing.T) {
	if !From([]int{1, 2, 3}).Any(func(x int) bool { return x == 2 }) {
		t.Fatal("Any true")
	}
	if From([]int{1, 3, 5}).Any(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("Any false")
	}
	if !From([]int{2, 4, 6}).All(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("All true")
	}
	if From([]int{2, 4, 5}).All(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("All false")
	}
	if !From([]int{1, 3, 5}).None(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("None true")
	}
	if From([]int{1, 2, 3}).None(func(x int) bool { return x == 2 }) {
		t.Fatal("None false")
	}
}

func TestFirstLastNth(t *testing.T) {
	v, ok := From([]int{1, 2, 3}).First().Get()
	if !ok || v != 1 {
		t.Fatalf("First: (%d,%v)", v, ok)
	}
	if From([]int{}).First().IsPresent() {
		t.Fatal("First empty")
	}
	v, ok = From([]int{1, 2, 3}).Last().Get()
	if !ok || v != 3 {
		t.Fatalf("Last: (%d,%v)", v, ok)
	}
	if From([]int{}).Last().IsPresent() {
		t.Fatal("Last empty")
	}
	v, ok = From([]int{10, 20, 30}).Nth(1).Get()
	if !ok || v != 20 {
		t.Fatalf("Nth: (%d,%v)", v, ok)
	}
	if From([]int{1, 2}).Nth(5).IsPresent() {
		t.Fatal("Nth out of range")
	}
	if From([]int{1, 2}).Nth(-1).IsPresent() {
		t.Fatal("Nth negative")
	}
}

func TestIsEmpty(t *testing.T) {
	if !From([]int{}).IsEmpty() {
		t.Fatal("IsEmpty empty")
	}
	if From([]int{1}).IsEmpty() {
		t.Fatal("IsEmpty non-empty")
	}
}

func TestPartitionSpan(t *testing.T) {
	yes, no := From([]int{1, 2, 3, 4, 5}).Partition(func(x int) bool { return x%2 == 0 })
	if len(yes) != 2 || yes[1] != 4 || len(no) != 3 || no[0] != 1 {
		t.Fatalf("Partition: yes=%v no=%v", yes, no)
	}
	prefix, rest := From([]int{2, 4, 5, 6}).Span(func(x int) bool { return x%2 == 0 })
	if len(prefix) != 2 || prefix[1] != 4 || len(rest) != 2 || rest[0] != 5 || rest[1] != 6 {
		t.Fatalf("Span: prefix=%v rest=%v", prefix, rest)
	}
	// all match -> rest empty
	p, r := From([]int{2, 4, 6}).Span(func(x int) bool { return x%2 == 0 })
	if len(p) != 3 || len(r) != 0 {
		t.Fatalf("Span all-match: p=%v r=%v", p, r)
	}
}

func TestMaxByMinBy(t *testing.T) {
	mx, ok := From([]string{"apple", "banana", "cherry"}).MaxBy(func(a, b string) bool { return a < b }).Get()
	if !ok || mx != "cherry" {
		t.Fatalf("MaxBy: (%q,%v)", mx, ok)
	}
	mn, ok := From([]string{"apple", "banana", "cherry"}).MinBy(func(a, b string) bool { return a < b }).Get()
	if !ok || mn != "apple" {
		t.Fatalf("MinBy: (%q,%v)", mn, ok)
	}
	if From([]string{}).MaxBy(func(a, b string) bool { return a < b }).IsPresent() {
		t.Fatal("MaxBy empty")
	}
}

type person struct {
	name string
	age  int
}

func TestMaxByKeyMinByKey(t *testing.T) {
	people := From([]person{
		{"a", 30}, {"b", 25}, {"c", 35},
	})
	oldest, ok := people.MaxByKey(func(p person) int { return p.age }).Get()
	if !ok || oldest.age != 35 || oldest.name != "c" {
		t.Fatalf("MaxByKey: %+v", oldest)
	}
	// reuse a fresh source (Seq is lazy, re-iterable if slice-backed)
	people2 := From([]person{{"a", 30}, {"b", 25}, {"c", 35}})
	youngest, ok := people2.MinByKey(func(p person) int { return p.age }).Get()
	if !ok || youngest.age != 25 || youngest.name != "b" {
		t.Fatalf("MinByKey: %+v", youngest)
	}
}

func TestSumByMeanBy(t *testing.T) {
	sum := From([]person{{"", 30}, {"", 25}, {"", 35}}).SumBy(func(p person) int { return p.age })
	if sum != 90 {
		t.Fatalf("SumBy: %d", sum)
	}
	mean := From([]person{{"", 30}, {"", 25}, {"", 35}}).MeanBy(func(p person) int { return p.age })
	if mean != 30.0 {
		t.Fatalf("MeanBy: %v", mean)
	}
	// empty -> 0
	if got := From([]person{}).SumBy(func(p person) int { return p.age }); got != 0 {
		t.Fatalf("SumBy empty: %d", got)
	}
	if got := From([]person{}).MeanBy(func(p person) int { return p.age }); got != 0 {
		t.Fatalf("MeanBy empty: %v", got)
	}
}

func TestJoin(t *testing.T) {
	got := From([]int{1, 2, 3}).Join(", ", strconv.Itoa)
	if got != "1, 2, 3" {
		t.Fatalf("Join: %q", got)
	}
	if got := From([]int{7}).Join(", ", strconv.Itoa); got != "7" {
		t.Fatalf("Join single: %q", got)
	}
	if got := From([]int{}).Join(", ", strconv.Itoa); got != "" {
		t.Fatalf("Join empty: %q", got)
	}
}

// shortCircuitCounter wraps a source and counts pulls, to prove Any/All/None
// short-circuit (they must stop at the first decisive element).
func shortCircuitCounter[T any](src Seq[T]) (Seq[T], *int) {
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

func TestAnyAllNoneShortCircuit(t *testing.T) {
	src, counter := shortCircuitCounter(From([]int{1, 2, 3, 4, 5}))
	// Any should stop at the first match (2 at index 1) -> 2 pulls
	if !src.Any(func(x int) bool { return x == 2 }) {
		t.Fatal("Any")
	}
	if *counter != 2 {
		t.Fatalf("Any short-circuit: pulled %d, expected 2", *counter)
	}

	src2, counter2 := shortCircuitCounter(From([]int{2, 4, 6, 8}))
	// All should stop at the first failure; here all even so it scans all -> 4
	if !src2.All(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("All")
	}
	if *counter2 != 4 {
		t.Fatalf("All all-match: pulled %d, expected 4", *counter2)
	}

	src3, counter3 := shortCircuitCounter(From([]int{2, 4, 5, 6}))
	// All should stop at 5 (index 2) -> 3 pulls
	if src3.All(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("All should be false")
	}
	if *counter3 != 3 {
		t.Fatalf("All short-circuit: pulled %d, expected 3", *counter3)
	}

	// None short-circuits at first match.
	src4, counter4 := shortCircuitCounter(From([]int{1, 3, 4, 5}))
	if !src4.None(func(x int) bool { return x > 10 }) {
		t.Fatal("None should be true")
	}
	// None true: scans all (no match) -> 4
	if *counter4 != 4 {
		t.Fatalf("None no-match: pulled %d, expected 4", *counter4)
	}
}
