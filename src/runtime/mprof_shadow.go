//go:build tinygo.wasm && tinygo.shadowstack && (gc.conservative || gc.precise)

package runtime

import "unsafe"

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
	obj := (*stackChainSiteObject)(unsafe.Pointer(stackChainStart))
	n := 0
	skip := 2
	for obj != nil && n < len(stk) {
		// Site 0 is an object of a function that made no call yet, like this
		// one, or of compiler glue.
		if site := obj.site; site > shadowNumSites {
			break
		} else if site != 0 && skip > 0 {
			skip--
		} else if site != 0 {
			stk[n] = site
			n++
		}
		// Callers are higher up the stack. Stop at a goroutine's entry.
		parent := (*stackChainSiteObject)(unsafe.Pointer(obj.parent))
		if uintptr(unsafe.Pointer(parent)) <= uintptr(unsafe.Pointer(obj)) {
			break
		}
		obj = parent
	}
	return n
}
