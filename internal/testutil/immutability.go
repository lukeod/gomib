package testutil

import (
	"reflect"
	"slices"
	"testing"
)

// Isolated checks that a read returns want before and after mutating its result.
// The caller must supply an independent expectation and a fresh fixture for
// each mutation case.
func Isolated[T any](t testing.TB, want T, read func() T, mutate func(T)) {
	t.Helper()
	got := read()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("initial read = %#v, want %#v", got, want)
	}
	mutate(got)
	if got := read(); !reflect.DeepEqual(want, got) {
		t.Fatalf("read after mutation = %#v, want %#v", got, want)
	}
}

// SliceIsolated checks that replacing an element cannot affect a later read.
// It only checks ownership of the slice; nested mutable values need their own
// mutation cases. Pointer elements refer to shared model objects and are replaced
// rather than modified.
func SliceIsolated[S ~[]E, E any](t testing.TB, read func() S) {
	t.Helper()
	want := slices.Clone(read())
	var zero E
	if len(want) == 0 || reflect.DeepEqual(want[0], zero) {
		t.Fatal("fixture must contain a nonzero first element")
	}
	Isolated(t, want, read, func(got S) { got[0] = zero })
}
