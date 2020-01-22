// +build nxp,mk66f18,teensy36

package machine

import (
	"device/nxp"
)

//go:keep
//go:section .flashconfig
var FlashConfig = [16]byte{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xDE, 0xF9, 0xFF, 0xFF,
}

func CPUFrequency() uint32 {
	return 180000000
}

// LED on the Teensy
const LED Pin = 13

var _pinRegisters [64]pinRegisters

func init() {
	_pinRegisters[13].Bit = 5
	_pinRegisters[13].PCR = &nxp.PORTC.PCR5
	_pinRegisters[13].PDOR = nxp.GPIOC.PDOR.Bit(5)
	_pinRegisters[13].PSOR = nxp.GPIOC.PSOR.Bit(5)
	_pinRegisters[13].PCOR = nxp.GPIOC.PCOR.Bit(5)
	_pinRegisters[13].PTOR = nxp.GPIOC.PTOR.Bit(5)
	_pinRegisters[13].PDIR = nxp.GPIOC.PDIR.Bit(5)
	_pinRegisters[13].PDDR = nxp.GPIOC.PDDR.Bit(5)
}

//go:inline
func (p Pin) registers() pinRegisters {
	return _pinRegisters[p]
}
