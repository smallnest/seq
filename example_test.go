package seq

import (
	"fmt"
	"strconv"
)

// Example shows the headline flow: a lazy Filter -> Map pipeline read left to
// right, driven by a terminal Collect.
func Example() {
	result := From([]int{1, 2, 3, 4, 5, 6}).
		Filter(func(x int) bool { return x%2 == 0 }).
		Map(func(x int) int { return x * x }).
		Collect()
	fmt.Println(result)
	// Output: [4 16 36]
}

// ExampleSeq_Map demonstrates the type-changing generic Map method, turning a
// Seq[int] into a Seq[string].
func ExampleSeq_Map() {
	words := From([]int{1, 2, 3}).Map(strconv.Itoa).Collect()
	fmt.Println(words)
	// Output: [1 2 3]
}

// ExampleRange builds a sequence of integers and reduces it to their sum.
func ExampleRange() {
	total, _ := Range(1, 5).Reduce(func(a, b int) int { return a + b })
	fmt.Println(total)
	// Output: 10
}

// ExampleDistinct shows a constraint-gated free function: Distinct requires
// T to be comparable, so it cannot be a method on Seq[T any].
func ExampleDistinct() {
	out := Distinct(From([]int{1, 1, 2, 3, 3, 3})).Collect()
	fmt.Println(out)
	// Output: [1 2 3]
}

// ExampleComparable recovers comparable operations as chainable methods via the
// SeqComparable subtype.
func ExampleComparable() {
	n := Comparable(From([]string{"a", "b", "a", "c", "b"})).
		Distinct().
		Seq().
		Count()
	fmt.Println(n)
	// Output: 3
}

// ExampleZip pairs two sequences position-wise and folds the pairs into a sum
// of products.
func ExampleZip() {
	a := From([]int{1, 2, 3})
	b := From([]int{10, 20, 30})
	dot := Zip(a, b).Fold(0, func(acc int, x, y int) int { return acc + x*y })
	fmt.Println(dot)
	// Output: 140
}
