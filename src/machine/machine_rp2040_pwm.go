//go:build rp2040

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

type pwmIRQRegisters struct {
	INTR volatile.Register32
	INTE volatile.Register32
	INTF volatile.Register32
	INTS volatile.Register32
}

// Atomic aliases and PWM layout, RP2040 datasheet sections 2.1.2 and 4.5.3.
// https://datasheets.raspberrypi.com/rp2040/rp2040-datasheet.pdf
var (
	pwmIRQ      = (*pwmIRQRegisters)(unsafe.Add(unsafe.Pointer(rp.PWM), 0xa4))
	pwmIRQSet   = (*pwmIRQRegisters)(unsafe.Add(unsafe.Pointer(rp.PWM), 0x20a4))
	pwmIRQClear = (*pwmIRQRegisters)(unsafe.Add(unsafe.Pointer(rp.PWM), 0x30a4))
)

// SetIRQ enables or disables this PWM slice's wrap interrupt.
func (pwm *pwmGroup) SetIRQ(enabled bool) {
	mask := uint32(1) << pwm.peripheral()
	if enabled {
		pwmIRQSet.INTE.Set(mask)
	} else {
		pwmIRQClear.INTE.Set(mask)
	}
}

// ClearIRQ clears this PWM slice's pending wrap and forced interrupts.
func (pwm *pwmGroup) ClearIRQ() {
	mask := uint32(1) << pwm.peripheral()
	pwmIRQClear.INTF.Set(mask)
	pwmIRQ.INTR.Set(mask)
}

// ForceIRQ forces an interrupt request for this PWM slice.
func (pwm *pwmGroup) ForceIRQ() {
	pwmIRQSet.INTF.Set(uint32(1) << pwm.peripheral())
}

// IsIRQSet reports whether this slice has a pending enabled or forced interrupt.
func (pwm *pwmGroup) IsIRQSet() bool {
	return pwmIRQ.INTS.Get()&(uint32(1)<<pwm.peripheral()) != 0
}
