// Package interrupt provides access to hardware interrupts. It provides a way
// to define interrupts and to enable/disable them.
package interrupt

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
