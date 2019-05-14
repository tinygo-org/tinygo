// +build gc.conservative
// +build cortexm

package runtime

// markStack marks all root pointers found on the stack.
//
// This implementation is conservative and relies on the stack top (provided by
// the linker) and getting the current stack pointer from a register. Also, it
// assumes a descending stack. Thus, it is not very portable.
func markStack() {
	markRoots(getCurrentStackPointer(), stackTop)
}
