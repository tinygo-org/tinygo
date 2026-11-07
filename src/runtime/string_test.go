package runtime_test

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestStringConcatOverflow(t *testing.T) {
	maxInt := uintptr(math.MaxInt)
	for _, tc := range []struct{ x, y uintptr }{
		{maxInt, 1},
		{1, maxInt},
		{maxInt, maxInt},
		{^uintptr(0), 1},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("StringConcatLen(%d, %d) did not panic", tc.x, tc.y)
				}
			}()
			n := runtime.StringConcatLen(tc.x, tc.y)
			t.Errorf("StringConcatLen(%d, %d) = %d", tc.x, tc.y, n)
		}()
	}
}

func TestStringConcatSizeBoundary(t *testing.T) {
	max := uintptr(math.MaxInt)
	for _, tc := range []struct{ x, y, want uintptr }{
		{0, 0, 0},
		{1, 2, 3},
		{max - 1, 1, max},
		{max / 2, max - max/2, max},
	} {
		if got := runtime.StringConcatSize(tc.x, tc.y); got != tc.want {
			t.Errorf("StringConcatSize(%d, %d) = %d, want %d", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestStringConcatOverflowDirect(t *testing.T) {
	var b byte
	x := unsafe.String(&b, math.MaxInt)
	defer func() {
		if recover() == nil {
			t.Error("string concatenation did not panic")
		}
	}()
	_ = x + "x"
	t.Error("string concatenation returned")
}

func TestStringConcatOverflowIndirect(t *testing.T) {
	var b byte
	x := unsafe.String(&b, math.MaxInt)
	defer func() {
		if recover() == nil {
			t.Error("string concatenation did not panic")
		}
	}()
	_ = concatStrings(x, "x")
	t.Error("string concatenation returned")
}

//go:noinline
func concatStrings(x, y string) string {
	return x + y
}

func TestStringConcat(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	x, y := "foo", "bar"
	for _, tc := range []struct{ x, y, want string }{
		{x, y, "foobar"},
		{x, "", "foo"},
		{"", y, "bar"},
		{"", "", ""},
	} {
		if got := tc.x + tc.y; got != tc.want {
			t.Errorf("%q + %q = %q, want %q", tc.x, tc.y, got, tc.want)
		}
	}
}
