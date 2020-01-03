// +build scheduler.tasks

package runtime

// getSystemStackPointer returns the current stack pointer of the system stack.
// This is not necessarily the same as the current stack pointer.
//export tinygo_getSystemStackPointer
func getSystemStackPointer() uintptr
