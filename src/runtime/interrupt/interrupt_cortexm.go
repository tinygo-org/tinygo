//go:build cortexm

package interrupt

import (
	"device/arm"
)

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	arm.EnableIRQ(uint32(irq.num))
}

// Disable disables this interrupt.
func (irq Interrupt) Disable() {
	arm.DisableIRQ(uint32(irq.num))
}

// SetPriority sets the interrupt priority for this interrupt. A lower number
// means a higher priority. Additionally, most hardware doesn't implement all
// priority bits (only the uppoer bits).
//
// Examples: 0xff (lowest priority), 0xc0 (low priority), 0x00 (highest possible
// priority).
func (irq Interrupt) SetPriority(priority uint8) {
	arm.SetPriority(uint32(irq.num), uint32(priority))
}

// State represents the previous global interrupt state.
type State uintptr

// Disable disables all interrupts and returns the previous interrupt state. It
// can be used in a critical section like this:
//
//	state := interrupt.Disable()
//	// critical section
//	interrupt.Restore(state)
//
// Critical sections can be nested. Make sure to call Restore in the same order
// as you called Disable (this happens naturally with the pattern above).
func Disable() (state State) {
	return State(arm.DisableInterrupts())
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// critical sections.
func Restore(state State) {
	arm.EnableInterrupts(uintptr(state))
}

// In returns whether the system is currently in an interrupt.
func In() bool {
	// The VECTACTIVE field gives the instruction vector that is currently
	// active (in handler mode), or 0 if not in an interrupt.
	// Documentation:
	// https://developer.arm.com/documentation/dui0497/a/cortex-m0-peripherals/system-control-block/interrupt-control-and-state-register
	vectactive := uint8(arm.SCB.ICSR.Get())
	return vectactive != 0
}
