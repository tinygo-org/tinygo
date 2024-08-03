package runtime

import (
	"internal/task"
	"unsafe"
)

// trap is a compiler hint that this function cannot be executed. It is
// translated into either a trap instruction or a call to abort().
//
//export llvm.trap
func trap()

// Inline assembly stub. It is essentially C longjmp but modified a bit for the
// purposes of TinyGo. It restores the stack pointer and jumps to the given pc.
//
//export tinygo_longjmp
func tinygo_longjmp(frame *deferFrameInlineAsm)

// Compiler intrinsic.
// Returns whether recover is supported on the current architecture.
func supportsRecover() recoverType

type recoverType uint8

const (
	// Note: these contants match recoverSupport in compiler/defer.go
	recoverNone recoverType = iota
	recoverInlineAsm
	recoverWasmEH
)

const (
	panicStrategyPrint = 1
	panicStrategyTrap  = 2
)

// Compile intrinsic.
// Returns which strategy is used. This is usually "print" but can be changed
// using the -panic= compiler flag.
func panicStrategy() uint8

// LLVM compiler intrinsic: this inserts a wasm 'throw' instruction.
//
//export llvm.wasm.throw
func wasm_throw(uint32, unsafe.Pointer)

// DeferFrameInlineAsm is a stack allocated object that stores information for
// the current "defer frame", which is used in functions that use the `defer`
// keyword.
// The compiler knows about the JumpPC struct offset, so it should not be moved
// without also updating compiler/defer.go.
type deferFrameInlineAsm struct {
	JumpSP     unsafe.Pointer                 // stack pointer to return to
	JumpPC     unsafe.Pointer                 // pc to return to
	ExtraRegs  [deferExtraRegs]unsafe.Pointer // extra registers (depending on the architecture)
	Previous   *deferFrameInlineAsm           // previous recover buffer pointer
	Panicking  bool                           // true iff this defer frame is panicking
	PanicValue interface{}                    // panic value, might be nil for panic(nil) for example
}

// DeferFrameWasmEH is like deferFrameInlineAsm but for WebAssembly exception
// handling.
type deferFrameWasmEH struct {
	Previous   *deferFrameWasmEH // previous recover buffer pointer
	Panicking  bool              // true iff this defer frame is panicking
	PanicValue interface{}       // panic value, might be nil for panic(nil) for example
}

// Builtin function panic(msg), used as a compiler intrinsic.
func _panic(message interface{}) {
	if panicStrategy() == panicStrategyTrap {
		trap()
	}
	switch supportsRecover() {
	case recoverInlineAsm:
		frame := (*deferFrameInlineAsm)(task.Current().DeferFrame)
		if frame != nil {
			frame.PanicValue = message
			frame.Panicking = true
			tinygo_longjmp(frame)
			// unreachable
		}
	case recoverWasmEH:
		frame := (*deferFrameWasmEH)(task.Current().DeferFrame)
		if frame != nil {
			frame.PanicValue = message
			frame.Panicking = true
			wasm_throw(0, nil)
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
	// As long as this function is inined, llvm.returnaddress(0) will return
	// something sensible.
	runtimePanicAt(returnAddress(0), msg)
}

func runtimePanicAt(addr unsafe.Pointer, msg string) {
	if panicStrategy() == panicStrategyTrap {
		trap()
	}
	if hasReturnAddr {
		printstring("panic: runtime error at ")
		printptr(uintptr(addr) - callInstSize)
		printstring(": ")
	} else {
		printstring("panic: runtime error: ")
	}
	println(msg)
	abort()
}

// Called at the start of a function that includes a deferred call.
// It gets passed in the stack-allocated defer frame and configures it.
// Note that the frame is not zeroed yet, so we need to initialize all values
// that will be used.
//
//go:inline
//go:nobounds
func setupDeferFrameInlineAsm(frame *deferFrameInlineAsm, jumpSP unsafe.Pointer) {
	currentTask := task.Current()
	frame.Previous = (*deferFrameInlineAsm)(currentTask.DeferFrame)
	frame.JumpSP = jumpSP
	frame.Panicking = false
	currentTask.DeferFrame = unsafe.Pointer(frame)
}

//go:inline
//go:nobounds
func setupDeferFrameWasmEH(frame *deferFrameWasmEH) {
	currentTask := task.Current()
	frame.Previous = (*deferFrameWasmEH)(currentTask.DeferFrame)
	frame.Panicking = false
	currentTask.DeferFrame = unsafe.Pointer(frame)
}

// Called right before the return instruction. It pops the defer frame from the
// linked list of defer frames. It also re-raises a panic if the goroutine is
// still panicking.
//
//go:inline
//go:nobounds
func destroyDeferFrameInlineAsm(frame *deferFrameInlineAsm) {
	task.Current().DeferFrame = unsafe.Pointer(frame.Previous)
	if frame.Panicking {
		// We're still panicking!
		// Re-raise the panic now.
		_panic(frame.PanicValue)
	}
}

// Like destroyDeferFrameInlineAsm but for wasm.
//
//go:inline
//go:nobounds
func destroyDeferFrameWasmEH(frame *deferFrameWasmEH) {
	task.Current().DeferFrame = unsafe.Pointer(frame.Previous)
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
	// TODO: somehow check that recover() is called directly by a deferred
	// function in a panicking goroutine. Maybe this can be done by comparing
	// the frame pointer?
	switch supportsRecover() {
	case recoverInlineAsm:
		frame := (*deferFrameInlineAsm)(task.Current().DeferFrame)
		if useParentFrame {
			// Don't recover panic from the current frame (which can't be
			// panicking already), but instead from the previous frame.
			frame = frame.Previous
		}
		if frame != nil && frame.Panicking {
			// Only the first call to recover returns the panic value. It also
			// stops the panicking sequence, hence setting panicking to false.
			frame.Panicking = false
			return frame.PanicValue
		}
		// Not panicking, so return a nil interface.
		return nil
	case recoverWasmEH:
		frame := (*deferFrameWasmEH)(task.Current().DeferFrame)
		if useParentFrame {
			// Don't recover panic from the current frame (which can't be
			// panicking already), but instead from the previous frame.
			frame = frame.Previous
		}
		if frame != nil && frame.Panicking {
			// Only the first call to recover returns the panic value. It also
			// stops the panicking sequence, hence setting panicking to false.
			frame.Panicking = false
			return frame.PanicValue
		}
		// Not panicking, so return a nil interface.
		return nil
	default:
		// Compiling without stack unwinding support, so make this a no-op.
		return nil
	}
}

// Panic when trying to dereference a nil pointer.
func nilPanic() {
	runtimePanicAt(returnAddress(0), "nil pointer dereference")
}

// Panic when trying to add an entry to a nil map
func nilMapPanic() {
	runtimePanicAt(returnAddress(0), "assignment to entry in nil map")
}

// Panic when trying to acces an array or slice out of bounds.
func lookupPanic() {
	runtimePanicAt(returnAddress(0), "index out of range")
}

// Panic when trying to slice a slice out of bounds.
func slicePanic() {
	runtimePanicAt(returnAddress(0), "slice out of range")
}

// Panic when trying to convert a slice to an array pointer (Go 1.17+) and the
// slice is shorter than the array.
func sliceToArrayPointerPanic() {
	runtimePanicAt(returnAddress(0), "slice smaller than array")
}

// Panic when calling unsafe.Slice() (Go 1.17+) or unsafe.String() (Go 1.20+)
// with a len that's too large (which includes if the ptr is nil and len is
// nonzero).
func unsafeSlicePanic() {
	runtimePanicAt(returnAddress(0), "unsafe.Slice/String: len out of range")
}

// Panic when trying to create a new channel that is too big.
func chanMakePanic() {
	runtimePanicAt(returnAddress(0), "new channel is too big")
}

// Panic when a shift value is negative.
func negativeShiftPanic() {
	runtimePanicAt(returnAddress(0), "negative shift")
}

// Panic when there is a divide by zero.
func divideByZeroPanic() {
	runtimePanicAt(returnAddress(0), "divide by zero")
}

func blockingPanic() {
	runtimePanicAt(returnAddress(0), "trying to do blocking operation in exported function")
}
