package runtime

// trap is a compiler hint that this function cannot be executed. It is
// translated into either a trap instruction or a call to abort().
//export llvm.trap
func trap()

// Builtin function panic(msg), used as a compiler intrinsic.
func _panic(message interface{}) {
	//printstring("panic: ")
	//printitf(message)
	//printnl()
	abort()
}

// Cause a runtime panic, which is (currently) always a string.
func runtimePanic(msg string) {
	//printstring("panic: runtime error: ")
	//println(msg)
	abort()
}

// Try to recover a panicking goroutine.
func _recover() interface{} {
	// Deferred functions are currently not executed during panic, so there is
	// no way this can return anything besides nil.
	return nil
}

// Panic when trying to dereference a nil pointer.
func nilPanic() {
	runtimePanic("nil pointer dereference")
}

// Panic when trying to acces an array or slice out of bounds.
func lookupPanic() {
	runtimePanic("index out of range")
}

// Panic when trying to slice a slice out of bounds.
func slicePanic() {
	runtimePanic("slice out of range")
}

// Panic when trying to convert a slice to an array pointer (Go 1.17+) and the
// slice is shorter than the array.
func sliceToArrayPointerPanic() {
	runtimePanic("slice smaller than array")
}

// Panic when calling unsafe.Slice() (Go 1.17+) with a len that's too large
// (which includes if the ptr is nil and len is nonzero).
func unsafeSlicePanic() {
	runtimePanic("unsafe.Slice: len out of range")
}

// Panic when trying to create a new channel that is too big.
func chanMakePanic() {
	runtimePanic("new channel is too big")
}

// Panic when a shift value is negative.
func negativeShiftPanic() {
	runtimePanic("negative shift")
}

// Panic when there is a divide by zero.
func divideByZeroPanic() {
	runtimePanic("divide by zero")
}

func blockingPanic() {
	runtimePanic("trying to do blocking operation in exported function")
}
