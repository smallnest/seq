package seq

import (
	"reflect"
	"testing"
)

func TestPartitionBy(t *testing.T) {
	// Group by parity, preserving first-appearance order of keys.
	s := From([]int{1, 2, 3, 4, 5, 6})
	got := map[string][]int{}
	var order []string
	for k, group := range PartitionBy(s, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	}) {
		order = append(order, k)
		got[k] = group
	}
	// "odd" key first (1 appears before 2).
	if !reflect.DeepEqual(order, []string{"odd", "even"}) {
		t.Errorf("key order = %v, want [odd even]", order)
	}
	if !reflect.DeepEqual(got["odd"], []int{1, 3, 5}) {
		t.Errorf("odd group = %v, want [1 3 5]", got["odd"])
	}
	if !reflect.DeepEqual(got["even"], []int{2, 4, 6}) {
		t.Errorf("even group = %v, want [2 4 6]", got["even"])
	}
}

func TestPartitionByOrderByFirstAppearance(t *testing.T) {
	// First key seen is 2 (even), so even must come first.
	s := From([]int{2, 1, 4, 3})
	var order []int
	for k := range PartitionBy(s, func(n int) int { return n % 2 }) {
		order = append(order, k)
	}
	if !reflect.DeepEqual(order, []int{0, 1}) {
		t.Errorf("key order = %v, want [0 1] (even seen first)", order)
	}
}

func TestPartitionByEmpty(t *testing.T) {
	count := 0
	for range PartitionBy(From([]int{}), func(n int) int { return n }) {
		count++
	}
	if count != 0 {
		t.Errorf("PartitionBy(empty) yielded %d pairs, want 0", count)
	}
}

func TestPartitionBySingle(t *testing.T) {
	got := map[int][]int{}
	for k, g := range PartitionBy(From([]int{7}), func(n int) int { return n }) {
		got[k] = g
	}
	if !reflect.DeepEqual(got, map[int][]int{7: {7}}) {
		t.Errorf("PartitionBy([7]) = %v, want {7:[7]}", got)
	}
}
