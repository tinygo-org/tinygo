package runtime

import (
	"internal/task"
	"unsafe"
)

// trap is a compiler hint that this function cannot be executed. It is
// translated into either a trap instruction or a call to abort().
//export llvm.trap
func trap()

// Inline assembly stub. It is essentially C longjmp but modified a bit for the
// purposes of TinyGo. It restores the stack pointer and jumps to the given pc.
//export tinygo_longjmp
func tinygo_longjmp(frame *task.DeferFrame)

// Compiler intrinsic.
// Returns whether recover is supported on the current architecture.
func supportsRecover() bool

// Builtin function panic(msg), used as a compiler intrinsic.
func _panic(message interface{}) {
	if supportsRecover() {
		frame := task.Current().DeferFrame
		if frame != nil {
			frame.PanicValue = message
			frame.Panicking = true
			tinygo_longjmp(frame)
			// unreachable
		}
	}
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

// Called at the start of a function that includes a deferred call.
// It gets passed in the stack-allocated defer frame and configures it.
// Note that the frame is not zeroed yet, so we need to initialize all values
// that will be used.
//go:inline
//go:nobounds
func setupDeferFrame(frame *task.DeferFrame, jumpSP unsafe.Pointer) {
	currentTask := task.Current()
	frame.Previous = currentTask.DeferFrame
	frame.JumpSP = jumpSP
	frame.Panicking = false
	currentTask.DeferFrame = frame
}

// Called right before the return instruction. It pops the defer frame from the
// linked list of defer frames. It also re-raises a panic if the goroutine is
// still panicking.
//go:inline
//go:nobounds
func destroyDeferFrame(frame *task.DeferFrame) {
	task.Current().DeferFrame = frame.Previous
	if frame.Panicking {
		// We're still panicking!
		// Re-raise the panic now.
		_panic(frame.PanicValue)
	}
}

// _recover is the built-in recover() function. It tries to recover a currently
// panicking goroutine.
// useParentFrame is set when the caller of runtime._recover has a defer frame
// itself. In that case, recover() shouldn't check that frame but one frame up.
func _recover(useParentFrame bool) interface{} {
	if !supportsRecover() {
		// Compiling without stack unwinding support, so make this a no-op.
		return nil
	}
	// TODO: somehow check that recover() is called directly by a deferred
	// function in a panicking goroutine. Maybe this can be done by comparing
	// the frame pointer?
	frame := task.Current().DeferFrame
	if useParentFrame {
		// Don't recover panic from the current frame (which can't be panicking
		// already), but instead from the previous frame.
		frame = frame.Previous
	}
	if frame != nil && frame.Panicking {
		// Only the first call to recover returns the panic value. It also stops
		// the panicking sequence, hence setting panicking to false.
		frame.Panicking = false
		return frame.PanicValue
	}
	// Not panicking, so return a nil interface.
	return nil
}

// Panic when trying to dereference a nil pointer.
func nilPanic() {
	runtimePanic("nil pointer dereference")
}

// Panic when trying to add an entry to a nil map
func nilMapPanic() {
	runtimePanic("assignment to entry in nil map")
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
