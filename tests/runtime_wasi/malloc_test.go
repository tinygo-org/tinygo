//go:build tinygo.wasm

package runtime_wasi

import (
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"unsafe"
)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer

//export free
func libc_free(ptr unsafe.Pointer)

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer

//export realloc
func libc_realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer

func getFilledBuffer_malloc() uintptr {
	ptr := libc_malloc(5)
	fillPanda(ptr)
	return uintptr(ptr)
}

func getFilledBuffer_calloc() uintptr {
	ptr := libc_calloc(2, 5)
	fillPanda(ptr)
	*(*byte)(unsafe.Add(ptr, 5)) = 'b'
	*(*byte)(unsafe.Add(ptr, 6)) = 'e'
	*(*byte)(unsafe.Add(ptr, 7)) = 'a'
	*(*byte)(unsafe.Add(ptr, 8)) = 'r'
	*(*byte)(unsafe.Add(ptr, 9)) = 's'
	return uintptr(ptr)
}

func getFilledBuffer_realloc() uintptr {
	origPtr := getFilledBuffer_malloc()
	ptr := libc_realloc(unsafe.Pointer(origPtr), 9)
	*(*byte)(unsafe.Add(ptr, 5)) = 'b'
	*(*byte)(unsafe.Add(ptr, 6)) = 'e'
	*(*byte)(unsafe.Add(ptr, 7)) = 'a'
	*(*byte)(unsafe.Add(ptr, 8)) = 'r'
	return uintptr(ptr)
}

func getFilledBuffer_reallocNil() uintptr {
	ptr := libc_realloc(nil, 5)
	fillPanda(ptr)
	return uintptr(ptr)
}

func fillPanda(ptr unsafe.Pointer) {
	*(*byte)(unsafe.Add(ptr, 0)) = 'p'
	*(*byte)(unsafe.Add(ptr, 1)) = 'a'
	*(*byte)(unsafe.Add(ptr, 2)) = 'n'
	*(*byte)(unsafe.Add(ptr, 3)) = 'd'
	*(*byte)(unsafe.Add(ptr, 4)) = 'a'
}

func checkFilledBuffer(t *testing.T, ptr uintptr, content string) {
	t.Helper()
	buf := *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: ptr,
		Len:  len(content),
	}))
	if buf != content {
		t.Errorf("expected %q, got %q", content, buf)
	}
}

func TestMallocFree(t *testing.T) {
	tests := []struct {
		name      string
		getBuffer func() uintptr
		content   string
	}{
		{
			name:      "malloc",
			getBuffer: getFilledBuffer_malloc,
			content:   "panda",
		},
		{
			name:      "calloc",
			getBuffer: getFilledBuffer_calloc,
			content:   "pandabears",
		},
		{
			name:      "realloc",
			getBuffer: getFilledBuffer_realloc,
			content:   "pandabear",
		},
		{
			name:      "realloc nil",
			getBuffer: getFilledBuffer_reallocNil,
			content:   "panda",
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			bufPtr := tt.getBuffer()
			// Don't use defer to free the buffer as it seems to cause the GC to track it.

			// Churn GC, the pointer should still be valid until free is called.
			for i := 0; i < 1000; i++ {
				a := "hello" + strconv.Itoa(i)
				// Some conditional logic to ensure optimization doesn't remove the loop completely.
				if len(a) < 0 {
					break
				}
				runtime.GC()
			}

			checkFilledBuffer(t, bufPtr, tt.content)

			libc_free(unsafe.Pointer(bufPtr))
		})
	}
}

func TestMallocEmpty(t *testing.T) {
	ptr := libc_malloc(0)
	if ptr != nil {
		t.Errorf("expected nil pointer, got %p", ptr)
	}
}

func TestCallocEmpty(t *testing.T) {
	ptr := libc_calloc(0, 1)
	if ptr != nil {
		t.Errorf("expected nil pointer, got %p", ptr)
	}
	ptr = libc_calloc(1, 0)
	if ptr != nil {
		t.Errorf("expected nil pointer, got %p", ptr)
	}
}

func TestReallocEmpty(t *testing.T) {
	ptr := libc_malloc(1)
	if ptr == nil {
		t.Error("expected pointer but was nil")
	}
	ptr = libc_realloc(ptr, 0)
	if ptr != nil {
		t.Errorf("expected nil pointer, got %p", ptr)
	}
}
