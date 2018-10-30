package runtime

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

// Check for bounds in *ssa.Index, *ssa.IndexAddr and *ssa.Lookup.
func lookupBoundsCheck(length lenType, index int) {
	if index < 0 || index >= int(length) {
		runtimePanic("index out of range")
	}
}

// Check for bounds in *ssa.Index, *ssa.IndexAddr and *ssa.Lookup.
// Supports 64-bit indexes.
func lookupBoundsCheckLong(length lenType, index int64) {
	if index < 0 || index >= int64(length) {
		runtimePanic("index out of range")
	}
}

// Check for bounds in *ssa.Slice.
func sliceBoundsCheck(capacity lenType, low, high uint) {
	if !(0 <= low && low <= high && high <= uint(capacity)) {
		runtimePanic("slice out of range")
	}
}

// Check for bounds in *ssa.Slice. Supports 64-bit indexes.
func sliceBoundsCheckLong(capacity lenType, low, high uint64) {
	if !(0 <= low && low <= high && high <= uint64(capacity)) {
		runtimePanic("slice out of range")
	}
}

// Check for bounds in *ssa.MakeSlice.
func sliceBoundsCheckMake(length, capacity uint) {
	if !(0 <= length && length <= capacity) {
		runtimePanic("slice size out of range")
	}
}
