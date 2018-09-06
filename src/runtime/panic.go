package runtime

func _panic(message interface{}) {
	printstring("panic: ")
	printitf(message)
	printnl()
	abort()
}

// Check for bounds in *ssa.Index, *ssa.IndexAddr and *ssa.Lookup.
func lookupBoundsCheck(length, index int) {
	if index < 0 || index >= length {
		// printstring() here is safe as this function is excluded from bounds
		// checking.
		printstring("panic: runtime error: index out of range\n")
		abort()
	}
}

// Check for bounds in *ssa.Slice.
func sliceBoundsCheck(length, low, high uint) {
	if !(0 <= low && low <= high && high <= length) {
		printstring("panic: runtime error: slice out of range\n")
		abort()
	}
}

// Check for bounds in *ssa.MakeSlice.
func sliceBoundsCheckMake(length, capacity uint) {
	if !(0 <= length && length <= capacity) {
		printstring("panic: runtime error: slice size out of range\n")
		abort()
	}
}
