package runtime

// The Error interface identifies a run time error.
type Error interface {
	error

	RuntimeError()
}

// plainError is a runtime.Error implementation for plain string messages.
type plainError string

func (e plainError) Error() string { return string(e) }
func (e plainError) RuntimeError() {}

const runtimeErrorPrefix = "runtime error: "

const (
	errNilPointer          = plainError(runtimeErrorPrefix + "invalid memory address or nil pointer dereference")
	errNilMap              = plainError("assignment to entry in nil map")
	errIndexOutOfRange     = plainError(runtimeErrorPrefix + "index out of range")
	errSliceOutOfRange     = plainError(runtimeErrorPrefix + "slice out of range")
	errSliceToArray        = plainError(runtimeErrorPrefix + "slice smaller than array")
	errUnsafeSliceLength   = plainError(runtimeErrorPrefix + "unsafe.Slice/String: len out of range")
	errChannelTooBig       = plainError("new channel is too big")
	errNegativeShift       = plainError(runtimeErrorPrefix + "negative shift")
	errDivideByZero        = plainError(runtimeErrorPrefix + "integer divide by zero")
	errBlockingExported    = plainError("trying to do blocking operation in exported function")
	errTimersUnsupported   = plainError("timers not supported without a scheduler")
	errIntegerOverflow     = plainError(runtimeErrorPrefix + "integer overflow")
	errUnsupportedSignal   = plainError("unsupported signal number")
	errUnsupportedExit     = plainError("unsupported: syscall.Exit")
	errSendOnClosedChannel = plainError("send on closed channel")
	errCloseNilChannel     = plainError("close of nil channel")
	errCloseClosedChannel  = plainError("close of closed channel")
	errWasmBeforeInit      = plainError("//go:wasmexport function called before runtime initialization")
	errWasmAfterMain       = plainError("//go:wasmexport function called after main.main returned")
	errWasmDidNotFinish    = plainError("//go:wasmexport function did not finish")
	errUncomparable        = plainError(runtimeErrorPrefix + "comparing un-comparable type")
	errTypeAssert          = plainError("type assert failed")
	errSchedulerDisabled   = plainError("scheduler is disabled")
	errOutOfMemory         = plainError("out of memory")
)
