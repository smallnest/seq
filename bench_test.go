package seq

import (
	"testing"
)

// Benchmarks guard the library's core performance claims: a lazy streaming
// chain should not allocate per element, materializing operators have a known
// O(n) memory profile, and iter.Pull-based operators carry the pull-iterator
// overhead. Run with `go test -bench . -benchmem`.

// benchInput is a moderately sized slice reused across benchmarks. It is the
// backing slice for a From source, so iterating it does not allocate.
var benchInput = func() []int {
	xs := make([]int, 10000)
	for i := range xs {
		xs[i] = i
	}
	return xs
}()

// sink prevents the compiler from optimizing benchmark work away.
var sink int

// BenchmarkStreamingChain measures a fully lazy, single-pass pipeline
// (Filter -> Map -> terminal). It should run in O(n) time with no per-element
// heap allocation; -benchmem alloc counts that grow with n signal a leak of a
// hidden materialization into the streaming path.
func BenchmarkStreamingChain(b *testing.B) {
	src := From(benchInput)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc := src.
			Filter(func(x int) bool { return x%2 == 0 }).
			Map(func(x int) int { return x * 2 }).
			Reduce(func(a, b int) int { return a + b }).OrZero()
		sink += acc
	}
}

// BenchmarkCollect measures the cost of materializing a streaming chain into a
// slice. Allocation here is expected and proportional to the surviving element
// count; the benchmark pins that cost.
func BenchmarkCollect(b *testing.B) {
	src := From(benchInput)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := src.Filter(func(x int) bool { return x%2 == 0 }).Collect()
		sink += len(out)
	}
}

// BenchmarkSortMaterializing measures a materializing operator (Sort). It
// collects the whole input and sorts it, so both time and memory are O(n);
// this benchmark documents that floor.
func BenchmarkSortMaterializing(b *testing.B) {
	src := From(benchInput)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += Sort(src).Count()
	}
}

// BenchmarkDistinctMap measures the map-backed membership path (Distinct),
// confirming it stays O(n) with a single map allocation rather than degrading
// to an O(n^2) linear scan.
func BenchmarkDistinctMap(b *testing.B) {
	src := From(benchInput)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += Distinct(src).Count()
	}
}

// BenchmarkZipPull measures an iter.Pull-based operator (Zip). Pull iterators
// allocate a coroutine per side, so this is the most allocation-heavy shape in
// the library; the benchmark tracks that overhead.
func BenchmarkZipPull(b *testing.B) {
	a := From(benchInput)
	c := From(benchInput)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += Zip(a, c).Count()
	}
}
