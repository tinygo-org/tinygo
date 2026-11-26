package cm

import "unsafe"

// ErrorContext represents the Component Model [error-context] type,
// an immutable, non-deterministic, host-defined value meant to aid in debugging.
//
// [error-context]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#error-context-type
type ErrorContext struct {
	_ HostLayout
	errorContext
}

type errorContext uint32

// Error implements the [error] interface. It returns the debug message associated with err.
func (err errorContext) Error() string {
	return err.DebugMessage()
}

// String implements [fmt.Stringer].
func (err errorContext) String() string {
	return err.DebugMessage()
}

// DebugMessage represents the Canonical ABI [error-context.debug-message] function.
//
// [error-context.debug-message]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#error-contextdebug-message
func (err errorContext) DebugMessage() string {
	var s string
	wasmimport_errorContextDebugMessage(uint32(err), unsafe.Pointer(&s))
	return s
}

// Drop represents the Canonical ABI [error-context.drop] function.
//
// [error-context.drop]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#error-contextdrop
func (err errorContext) Drop() {
	wasmimport_errorContextDrop(uint32(err))
}
