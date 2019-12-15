// +build cortexm

package interrupt

import (
	"device/arm"
)

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	arm.EnableIRQ(uint32(irq.num))
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
