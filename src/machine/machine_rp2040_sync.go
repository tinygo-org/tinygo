//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
)

// machine_rp2040_sync.go contains interrupt and
// lock primitives similar to those found in Pico SDK's
// irq.c

const (
	// Number of spin locks available
	_NUMSPINLOCKS = 32
	// Number of interrupt handlers available
	_NUMIRQ               = 32
	_PICO_SPINLOCK_ID_IRQ = 9
	_NUMBANK0_GPIOS       = 30
)

// Sets raspberry pico in dormant state until interrupt event is triggered
// usage:
//  err := machine.Sleep(interruptPin, machine.PinRising)
// Above code will cause program to sleep until a high level signal
// is found on interruptPin. Be sure to configure pin as input beforehand.
// Taken from https://github.com/raspberrypi/pico-extras/ -> pico/sleep.h
func Sleep(pin Pin, change PinChange) error {
	if pin > 31 {
		return ErrInvalidInputPin
	}

	base := &ioBank0.dormantWakeIRQctrl
	pin.ctrlSetInterrupt(change, true, base) // gpio_set_dormant_irq_enabled

	xosc.Dormant() // Execution stops here until woken up

	// Clear the irq so we can go back to dormant mode again if we want
	pin.acknowledgeInterrupt(change)
	return nil
}

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

// Enable or disable a specific interrupt on the executing core.
// num is the interrupt number which must be in [0,31].
func irqSet(num uint32, enabled bool) {
	if num >= _NUMIRQ {
		return
	}
	irqSetMask(1<<num, enabled)
}

func irqSetMask(mask uint32, enabled bool) {
	if enabled {
		// Clear pending before enable
		// (if IRQ is actually asserted, it will immediately re-pend)
		rp.PPB.NVIC_ICPR.Set(mask)
		rp.PPB.NVIC_ISER.Set(mask)
	} else {
		rp.PPB.NVIC_ICER.Set(mask)
	}
}
