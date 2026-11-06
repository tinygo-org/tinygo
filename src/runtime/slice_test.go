package runtime_test

import (
	"math"
	"runtime"
	"testing"
)

func TestAppendZeroSized(t *testing.T) {
	for _, n := range []int{0, 1, 7, 8, math.MaxInt/2 - 1, math.MaxInt / 2, math.MaxInt/2 + 1, math.MaxInt - 1} {
		s := make([]struct{}, n)
		s = append(s, struct{}{})
		if len(s) != n+1 {
			t.Errorf("append to len %d: len = %d, want %d", n, len(s), n+1)
		}
		if cap(s) < len(s) {
			t.Errorf("append to len %d: cap = %d, want at least %d", n, cap(s), len(s))
		}
	}
}

func TestAppendZeroSizedMaxInt(t *testing.T) {
	s := make([]struct{}, math.MaxInt)
	defer func() {
		if recover() == nil {
			t.Errorf("append to len %d did not panic, len = %d, cap = %d", math.MaxInt, len(s), cap(s))
		}
	}()
	s = append(s, struct{}{})
}

func TestAppendZeroSizedMaxIntIndirect(t *testing.T) {
	s := make([]struct{}, math.MaxInt)
	values := make([]struct{}, 1)
	defer func() {
		if recover() == nil {
			t.Error("append did not panic")
		}
	}()
	s = appendZeroSized(s, values)
	t.Errorf("append returned len %d, cap %d", len(s), cap(s))
}

//go:noinline
func appendZeroSized(s, values []struct{}) []struct{} {
	return append(s, values...)
}

func TestSliceGrowSizeOverflow(t *testing.T) {
	for _, size := range []uintptr{2, 3, 8, 24} {
		n := ^uintptr(0)/size + 1
		if n > math.MaxInt {
			continue
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("SliceGrow(%d, %d) did not panic", n, size)
				}
			}()
			c := runtime.SliceGrow(n, size)
			t.Errorf("SliceGrow(%d, %d) = %d", n, size, c)
		}()
	}
}
