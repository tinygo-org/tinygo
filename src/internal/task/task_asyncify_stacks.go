//go:build scheduler.asyncify

package task

import "unsafe"

// stackTasks lists every task that has a heap stack, newest first.
var stackTasks *Task

func linkStack(t *Task) {
	t.state.nextStack = stackTasks
	if stackTasks != nil {
		stackTasks.state.prevStack = t
	}
	stackTasks = t
}

// DetachStack takes the current task off the stack list. A //go:wasmexport call
// does this before its last pause, since it never resumes or exits.
func DetachStack() { unlinkStack(currentTask) }

func unlinkStack(t *Task) {
	if t.state.prevStack != nil {
		t.state.prevStack.state.nextStack = t.state.nextStack
	} else if stackTasks == t {
		stackTasks = t.state.nextStack
	}
	if t.state.nextStack != nil {
		t.state.nextStack.state.prevStack = t.state.prevStack
	}
	t.state.nextStack, t.state.prevStack = nil, nil
}

// liveStackRanges calls mark with the parts of each stack that hold data: the
// save area at the bottom and the C stack frames at the top. The gap between
// them keeps stale words from deeper calls, which must not be treated as roots.
func liveStackRanges(mark func(start, end uintptr)) {
	if t := currentTask; t != nil && t.state.prevStack == nil && stackTasks != t {
		// It left the list on exit or detach but still runs until its final pause.
		mark(getCurrentStackPointer(), uintptr(t.state.top))
	}
	for t := stackTasks; t != nil; t = t.state.nextStack {
		s := &t.state
		if s.canaryPtr == nil {
			continue
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
			continue
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
}

// getCurrentStackPointer is runtime.getCurrentStackPointer, with its C
// calling convention.
//
//export tinygo_getCurrentStackPointer
func getCurrentStackPointer() uintptr
