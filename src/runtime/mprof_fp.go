//go:build (darwin || linux) && !baremetal && (amd64 || arm64) && scheduler.threads && (gc.conservative || gc.precise)

package runtime

// Stack walking for the memory profiler with frame pointers. A frame record
// is the caller's frame pointer and then the return address.

import (
	"internal/task"
	"unsafe"
)

//export llvm.frameaddress.p0
func frameAddress(level uint32) unsafe.Pointer

// memProfileCallers stores the return addresses of the callers of alloc in
// stk. It skips its own frame and that of memProfileAlloc.
//
//go:noinline
func memProfileCallers(stk []uintptr) int {
	const ptrSize = unsafe.Sizeof(uintptr(0))
	// Stay within the goroutine's stack. The thread start code above it may
	// not have frame records.
	top := task.StackTop()
	fp := uintptr(frameAddress(0))
	n := 0
	for skip := 2; n < len(stk); skip-- {
		if fp == 0 || fp%ptrSize != 0 || fp+2*ptrSize > top {
			break
		}
		pc := *(*uintptr)(unsafe.Pointer(fp + ptrSize))
		if pc == 0 {
			break
		}
		if skip <= 0 {
			stk[n] = pc
			n++
		}
		next := *(*uintptr)(unsafe.Pointer(fp))
		if next <= fp {
			break
		}
		fp = next
	}
	return n
}
