package runtime

// trap is a compiler hint that this function cannot be executed. It is
// translated into either a trap instruction or a call to abort().
//export llvm.trap
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

// Panic when trying to create a new channel that is too big.
func chanMakePanic() {
	runtimePanic("new channel is too big")
}

// Panic when a shift value is negative.
func negativeShiftPanic() {
	runtimePanic("negative shift")
}

func blockingPanic() {
	runtimePanic("trying to do blocking operation in exported function")
}
