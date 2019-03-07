package runtime

// trap is a compiler hint that this function cannot be executed. It is
// translated into either a trap instruction or a call to abort().
//go:export llvm.trap
func trap()

// Builtin function panic(msg), used as a compiler intrinsic.
func _panic(message interface{}) {
	printstring("panic: ")
	printitf(message)
	printnl()
	abort()
}

// Cause a runtime panic, which is (currently) always a string.
func runtimePanic(msg string) {
	printstring("panic: runtime error: ")
	println(msg)
	abort()
}

// Try to recover a panicking goroutine.
func _recover() interface{} {
	// Deferred functions are currently not executed during panic, so there is
	// no way this can return anything besides nil.
	return nil
}

// Panic when trying to dereference a nil pointer.
func nilpanic() {
	runtimePanic("nil pointer dereference")
}

// Panic when trying to acces an array or slice out of bounds.
func lookuppanic() {
	runtimePanic("index out of range")
}

// Panic when trying to slice a slice out of bounds.
func slicepanic() {
	runtimePanic("slice out of range")
}

// Check for bounds in *ssa.MakeSlice.
func sliceBoundsCheckMake(length, capacity uintptr, max uintptr) {
	if length > capacity || capacity > max {
		runtimePanic("slice size out of range")
	}
}

// Check for bounds in *ssa.MakeSlice. Supports 64-bit indexes.
func sliceBoundsCheckMake64(length, capacity uint64, max uintptr) {
	if length > capacity || capacity > uint64(max) {
		runtimePanic("slice size out of range")
	}
}
