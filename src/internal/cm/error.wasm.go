package cm

import "unsafe"

// msg uses unsafe.Pointer for compatibility with go1.23 and lower.
//
//go:wasmimport canon error-context.debug-message
//go:noescape
func wasmimport_errorContextDebugMessage(err uint32, msg unsafe.Pointer)

//go:wasmimport canon error-context.drop
//go:noescape
func wasmimport_errorContextDrop(err uint32)
