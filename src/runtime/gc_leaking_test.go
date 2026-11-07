//go:build gc.leaking

package runtime_test

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestSliceGrowAllocAlignmentOverflow(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) < unsafe.Sizeof(int(0)) {
		t.Skip("uintptr is narrower than int")
	}

	var s [][3]byte
	n := int(^uintptr(0) / 3)
	defer func() {
		if recover() == nil {
			t.Errorf("growing by %d did not panic", n)
		}
	}()
	reflect.ValueOf(&s).Elem().Grow(n)
	t.Errorf("growing by %d returned capacity %d", n, cap(s))
}

func TestSliceGrowAllocPointerOverflow(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) < unsafe.Sizeof(int(0)) {
		t.Skip("uintptr is narrower than int")
	}

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	n := int(^uintptr(0)/3 - 16)
	func() {
		var s [][3]byte
		defer func() {
			if recover() == nil {
				t.Errorf("growing by %d did not panic", n)
			}
		}()
		reflect.ValueOf(&s).Elem().Grow(n)
		t.Errorf("growing by %d returned capacity %d", n, cap(s))
	}()
	runtime.ReadMemStats(&after)
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 1<<20 {
		t.Errorf("failed growth increased TotalAlloc by %d", allocated)
	}
}
