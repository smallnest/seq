package seq

import (
	"reflect"
	"testing"
)

func TestReplace(t *testing.T) {
	// Replace first 2 occurrences of 1 with 9.
	got := Comparable(From([]int{1, 2, 1, 3, 1})).Replace(1, 9, 2).Collect()
	want := []int{9, 2, 9, 3, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Replace(1,9,2) = %v, want %v", got, want)
	}
}

func TestReplaceN0(t *testing.T) {
	got := Comparable(From([]int{1, 1, 1})).Replace(1, 9, 0).Collect()
	if !reflect.DeepEqual(got, []int{1, 1, 1}) {
		t.Errorf("Replace(1,9,0) = %v, want [1 1 1] (no replacement)", got)
	}
}

func TestReplaceNegativeIsAll(t *testing.T) {
	got := Comparable(From([]int{1, 2, 1, 1})).Replace(1, 9, -1).Collect()
	if !reflect.DeepEqual(got, []int{9, 2, 9, 9}) {
		t.Errorf("Replace(1,9,-1) = %v, want [9 2 9 9]", got)
	}
}

func TestReplaceAll(t *testing.T) {
	got := Comparable(From([]int{1, 2, 1, 3, 1})).ReplaceAll(1, 9).Collect()
	want := []int{9, 2, 9, 3, 9}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReplaceAll(1,9) = %v, want %v", got, want)
	}
}

func TestReplaceNoMatch(t *testing.T) {
	got := Comparable(From([]int{1, 2, 3})).ReplaceAll(9, 0).Collect()
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("ReplaceAll(9,0) = %v, want [1 2 3] (unchanged)", got)
	}
}

func TestReplaceEmpty(t *testing.T) {
	got := Comparable(From([]int{})).ReplaceAll(1, 9).Collect()
	if len(got) != 0 {
		t.Errorf("ReplaceAll on empty = %v, want empty", got)
	}
}

func TestReplaceLazyAndChains(t *testing.T) {
	// Replace returns SeqComparable, so the subtype chain continues.
	got := Comparable(From([]int{1, 1, 2, 2})).ReplaceAll(1, 2).Distinct().Collect()
	if !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("ReplaceAll(1,2).Distinct() = %v, want [2]", got)
	}
}
