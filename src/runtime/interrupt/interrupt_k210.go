// +build k210

package interrupt

import "device/kendryte"

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	// TODO: Use current hartid
	kendryte.PLIC.TARGET_ENABLES[0].ENABLE[irq.num/32].SetBits(1 << (uint(irq.num) % 32))
}

// SetPriority sets the interrupt priority for this interrupt. A higher priority
// number means a higher priority (unlike Cortex-M). Priority 0 effectively
// disables the interrupt.
func (irq Interrupt) SetPriority(priority uint8) {
	kendryte.PLIC.PRIORITY[irq.num].Set(uint32(priority))
}
