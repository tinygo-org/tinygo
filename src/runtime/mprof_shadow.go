//go:build tinygo.wasm && tinygo.shadowstack && (gc.conservative || gc.precise)

package runtime

import (
	"internal/task"
	"unsafe"
)

// With sites, a stack object starts with the parent object, the number of
// slots and the current call site. See transform/shadowstack.go.
type stackChainSiteObject struct {
	parent   *stackChainObject
	numSlots uintptr
	site     uintptr
}

// memProfileCallers stores the call sites of the callers of alloc in stk. It
// skips the stack objects of memProfileAlloc and alloc.
//
//go:noinline
func memProfileCallers(stk []uintptr) int {
	// This function may get a stack object too. Before any call, its site is 0.
	obj := (*stackChainSiteObject)(unsafe.Pointer(stackChainStart))
	if obj != nil && obj.site == 0 {
		obj = (*stackChainSiteObject)(unsafe.Pointer(obj.parent))
	}
	// The objects of the callers are on the current stack, above this frame.
	lo, hi := getCurrentStackPointer(), stackTop
	if !task.OnSystemStack() {
		hi = task.StackTop()
	}
	// The first two sites are the calls of memProfileAlloc and alloc. Skip all
	// leading objects with those sites, as alloc can appear twice.
	n := 0
	var skipped [2]uintptr
	for obj != nil && n < len(stk) {
		if addr := uintptr(unsafe.Pointer(obj)); addr < lo || addr+unsafe.Sizeof(*obj) > hi {
			// The entry of a goroutine links to the stack it was started from.
			break
		}
		// Site 0 is an object of a function that made no call yet. An invalid
		// site is a stale object, whose parent is still the caller.
		if site := obj.site; site == 0 || site > shadowNumSites {
		} else if skipped[0] == 0 {
			skipped[0] = site
		} else if skipped[1] == 0 {
			skipped[1] = site
		} else if n > 0 || (site != skipped[0] && site != skipped[1]) {
			stk[n] = site
			n++
		}
		// Callers are higher up the stack.
		parent := (*stackChainSiteObject)(unsafe.Pointer(obj.parent))
		if uintptr(unsafe.Pointer(parent)) <= uintptr(unsafe.Pointer(obj)) {
			break
		}
		obj = parent
	}
	return n
}
