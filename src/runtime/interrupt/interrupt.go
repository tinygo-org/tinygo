// Package interrupt provides access to hardware interrupts. It provides a way
// to define interrupts and to enable/disable them.
package interrupt

import "unsafe"

// Interrupt provides direct access to hardware interrupts. You can configure
// this interrupt through this interface.
//
// Do not use the zero value of an Interrupt object. Instead, call New to obtain
// an interrupt handle.
type Interrupt struct {
	// Make this number unexported so it cannot be set directly. This provides
	// some encapsulation.
	num int
}

// New is a compiler intrinsic that creates a new Interrupt object. You may call
// it only once, and must pass constant parameters to it. That means that the
// interrupt ID must be a Go constant and that the handler must be a simple
// function: closures are not supported.
func New(id int, handler func(Interrupt)) Interrupt

// handle is used internally, between IR generation and interrupt lowering. The
// frontend will create runtime/interrupt.handle objects, cast them to an int,
// and use that in an Interrupt object. That way the compiler will be able to
// optimize away all interrupt handles that are never used in a program.
// This system only works when interrupts need to be enabled before use and this
// is done only through calling Enable() on this object. If interrupts cannot
// individually be enabled/disabled, the compiler should create a pseudo-call
// (like runtime/interrupt.use()) that keeps the interrupt alive.
type handle struct {
	context unsafe.Pointer
	funcPtr uintptr
	Interrupt
}
