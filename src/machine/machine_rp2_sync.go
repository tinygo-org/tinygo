//go:build rp2040 || rp2350

package machine

// machine_rp2040_sync.go contains interrupt and
// lock primitives similar to those found in Pico SDK's
// irq.c

// Clears interrupt flag on a pin
func (p Pin) acknowledgeInterrupt(change PinChange) {
	ioBank0.intR[p>>3].Set(p.ioIntBit(change))
}

// Basic interrupt setting via ioBANK0 for GPIO interrupts.
func (p Pin) setInterrupt(change PinChange, enabled bool) {
	// Separate mask/force/status per-core, so check which core called, and
	// set the relevant IRQ controls.
	switch CurrentCore() {
	case 0:
		p.ctrlSetInterrupt(change, enabled, &ioBank0.proc0IRQctrl)
	case 1:
		p.ctrlSetInterrupt(change, enabled, &ioBank0.proc1IRQctrl)
	}
}

// ctrlSetInterrupt acknowledges any pending interrupt and enables or disables
// the interrupt for a given IRQ control bank (IOBANK, DormantIRQ, QSPI).
//
// pico-sdk calls this the _gpio_set_irq_enabled, not to be confused with
// gpio_set_irq_enabled (no leading underscore).
func (p Pin) ctrlSetInterrupt(change PinChange, enabled bool, base *irqCtrl) {
	p.acknowledgeInterrupt(change)
	enReg := &base.intE[p>>3]
	if enabled {
		enReg.SetBits(p.ioIntBit(change))
	} else {
		enReg.ClearBits(p.ioIntBit(change))
	}
}
