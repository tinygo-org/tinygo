// +build arm,baremetal,!avr

package task

import "device/arm"

// nonest is a sync.Locker that blocks nested interrupts while held.
type nonest struct {
	state uintptr
}

//go:inline
func (n *nonest) Lock() {
	n.state = arm.DisableInterrupts()
}

//go:inline
func (n *nonest) Unlock() {
	arm.EnableInterrupts(n.state)
}
