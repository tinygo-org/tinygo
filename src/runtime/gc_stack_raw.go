//go:build (gc.conservative || gc.precise || gc.boehm) && !tinygo.wasm

package runtime

import "internal/task"

// markStack marks all root pointers found on the stack.
//
// This implementation is conservative and relies on the stack top (provided by
// the linker) and getting the current stack pointer from a register. Also, it
// assumes a descending stack. Thus, it is not very portable.
func markStack() {
	// Scan the current stack, and all current registers.
	scanCurrentStack()

	if !task.OnSystemStack() {
		// Mark system stack.
		markRoots(getSystemStackPointer(), stackTop)
	}
}

//go:export tinygo_scanCurrentStack
func scanCurrentStack()

//go:export tinygo_scanstack
func scanstack(sp uintptr) {
	// Mark current stack.
	// This function is called by scanCurrentStack, after pushing all registers onto the stack.
	// Callee-saved registers have been pushed onto stack by tinygo_localscan, so this will scan them too.
	if task.OnSystemStack() {
		// This is the system stack.
		// Scan all words on the stack.
		markRoots(sp, stackTop)
	} else {
		// This is a goroutine stack.
		markCurrentGoroutineStack(sp)
	}
}
