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

const (
	errNilPointer          = plainError("nil pointer dereference")
	errNilMap              = plainError("assignment to entry in nil map")
	errIndexOutOfRange     = plainError("index out of range")
	errSliceOutOfRange     = plainError("slice out of range")
	errSliceToArray        = plainError("slice smaller than array")
	errUnsafeSliceLength   = plainError("unsafe.Slice/String: len out of range")
	errChannelTooBig       = plainError("new channel is too big")
	errNegativeShift       = plainError("negative shift")
	errDivideByZero        = plainError("divide by zero")
	errBlockingExported    = plainError("trying to do blocking operation in exported function")
	errTimersUnsupported   = plainError("timers not supported without a scheduler")
	errIntegerOverflow     = plainError("integer overflow")
	errUnsupportedSignal   = plainError("unsupported signal number")
	errUnsupportedExit     = plainError("unsupported: syscall.Exit")
	errSendOnClosedChannel = plainError("send on closed channel")
	errCloseNilChannel     = plainError("close of nil channel")
	errCloseClosedChannel  = plainError("close of closed channel")
	errWasmBeforeInit      = plainError("//go:wasmexport function called before runtime initialization")
	errWasmAfterMain       = plainError("//go:wasmexport function called after main.main returned")
	errWasmDidNotFinish    = plainError("//go:wasmexport function did not finish")
	errUncomparable        = plainError("comparing un-comparable type")
	errTypeAssert          = plainError("type assert failed")
	errSchedulerDisabled   = plainError("scheduler is disabled")
	errOutOfMemory         = plainError("out of memory")
)
