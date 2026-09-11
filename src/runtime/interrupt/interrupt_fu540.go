//go:build sifive && fu540

package interrupt

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
}

// SetPriority sets the interrupt priority for this interrupt. A higher priority
// number means a higher priority (unlike Cortex-M). Priority 0 effectively
// disables the interrupt.
func (irq Interrupt) SetPriority(priority uint8) {
}
