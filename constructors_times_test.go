package seq

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTimes(t *testing.T) {
	got := Times(4, func(i int) int { return i * i }).Collect()
	want := []int{0, 1, 4, 9}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Times(4, i*i) = %v, want %v", got, want)
	}
}

func TestTimesEmpty(t *testing.T) {
	if got := Times(0, func(int) int { t.Fatal("f must not be called for n=0"); return 0 }).Collect(); len(got) != 0 {
		t.Errorf("Times(0, ...) = %v, want empty", got)
	}
	if got := Times(-3, func(int) int { t.Fatal("f must not be called for n<0"); return 0 }).Collect(); len(got) != 0 {
		t.Errorf("Times(-3, ...) = %v, want empty", got)
	}
}

func TestTimesLazy(t *testing.T) {
	calls := 0
	s := Times(100, func(i int) int { calls++; return i })
	if calls != 0 {
		t.Errorf("Times called f at construction: %d calls, want 0", calls)
	}
	_ = s.Take(3).Collect()
	if calls != 3 {
		t.Errorf("Times f called %d times after Take(3), want 3", calls)
	}
}

func ExampleTimes() {
	fmt.Println(Times(5, func(i int) int { return i * i }).Collect())
	// Output: [0 1 4 9 16]
}
