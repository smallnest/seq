package seq

import "testing"

// These tests pin the nil-function contract documented in the package doc:
// operations require non-nil function arguments. A nil function given to a
// lazy intermediate operation panics only when a terminal drives iteration; a
// nil function given to an eager terminal operation panics when invoked.

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected panic on nil function, got none", name)
		}
	}()
	f()
}

func TestNilFuncIntermediateIsLazy(t *testing.T) {
	// Constructing the lazy pipeline with a nil predicate must NOT panic.
	pipeline := From([]int{1, 2, 3}).Filter(nil)
	// Driving it with a terminal must panic.
	mustPanic(t, "Filter(nil).Collect", func() {
		pipeline.Collect()
	})
}

func TestNilFuncTerminalPanics(t *testing.T) {
	mustPanic(t, "ForEach(nil)", func() {
		From([]int{1, 2, 3}).ForEach(nil)
	})
	mustPanic(t, "Reduce(nil)", func() {
		From([]int{1, 2, 3}).Reduce(nil)
	})
}
