//go:build scheduler.asyncify

package task

import "unsafe"

// The last word of a stack holds the owning task as an integer, right before
// the object header, so the GC finds it without keeping the task alive.
const ownerSize = unsafe.Sizeof(uintptr(0))

func setStackOwner(t *Task) {
	*(*uintptr)(unsafe.Pointer(t.state.stackEnd - ownerSize)) = uintptr(unsafe.Pointer(t))
}

// stackOwner returns the task of the stack that ends before the object header
// at header, or nil if there is none.
func stackOwner(header uintptr) *Task {
	slot := header - ownerSize
	owner := *(*uintptr)(unsafe.Pointer(slot))
	if owner == 0 {
		return nil
	}
	t := (*Task)(unsafe.Pointer(owner))
	if t.state.stackEnd != header {
		return nil
	}
	return t
}

// scanStack marks the save area at the bottom and the C stack frames at the
// top. The gap between them holds stale words that must not be roots.
func scanStack(t *Task, mark func(start, end uintptr)) {
	s := &t.state
	if s.canaryPtr == nil {
		return
	}
	base := uintptr(unsafe.Pointer(s.canaryPtr))
	top := uintptr(s.top)
	saved := uintptr(s.asyncifysp)
	if s.panicReplay != 0 && uintptr(s.panicOrigin.asyncifysp) > saved {
		saved = uintptr(s.panicOrigin.asyncifysp)
	}
	sp := uintptr(s.csp)
	if t == currentTask {
		sp = getCurrentStackPointer()
	} else if s.running {
		// It resumed the current task from a JS callback, so its frames
		// reach below csp. Scan all of it.
		mark(base, top)
		return
	}
	if saved > top {
		saved = top
	}
	if sp < saved {
		// The frames reach into the save area, so scan the whole stack.
		sp = saved
	}
	if saved > base {
		mark(base, saved)
	}
	if top > sp {
		mark(sp, top)
	}
}

// getCurrentStackPointer is runtime.getCurrentStackPointer, with its C
// calling convention.
//
//export tinygo_getCurrentStackPointer
func getCurrentStackPointer() uintptr
