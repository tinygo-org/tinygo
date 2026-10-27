//go:build scheduler.asyncify && (gc.conservative || gc.precise)

package task

import (
	"internal/gclayout"
	"unsafe"
)

// The GC does not scan stacks as objects. It calls StackReached for each stack
// it finds, which marks the live parts.
func stackLayout() unsafe.Pointer { return gclayout.Stack.AsPtr() }

//go:linkname markRoots runtime.markRoots
func markRoots(start, end uintptr)

//go:linkname markRoot runtime.markRoot
func markRoot(addr, root uintptr)

//go:linkname stackBodySize runtime.stackBodySize
func stackBodySize(size uintptr) uintptr

// StackReached is called by the GC when it marks a stack. The stack ends right
// before the object header at header. It marks the owning task and the live
// part of the stack.
func StackReached(header uintptr) {
	t := stackOwner(header)
	if t == nil {
		return
	}
	markRoot(0, uintptr(unsafe.Pointer(t)))
	scanStack(t, markRoots)
}
