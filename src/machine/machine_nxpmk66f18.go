// +build nxp,mk66f18

package machine

import (
	"device/nxp"
	"runtime/volatile"
)

type FastPin struct {
	PDOR *volatile.BitRegister
	PSOR *volatile.BitRegister
	PCOR *volatile.BitRegister
	PTOR *volatile.BitRegister
	PDIR *volatile.BitRegister
	PDDR *volatile.BitRegister
}

type pin struct {
	Bit  uint8
	PCR  *volatile.Register32
	GPIO *nxp.GPIO_Type
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	switch config.Mode {
	case PinInput:
		panic("todo")

	case PinOutput:
		r := p.reg()
		r.GPIO.PDDR.SetBits(1 << r.Bit)
		r.PCR.SetBits(nxp.PORT_PCR0_SRE | nxp.PORT_PCR0_DSE | nxp.PORT_PCR0_MUX(1))
		r.PCR.ClearBits(nxp.PORT_PCR0_ODE)
	}
}

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p Pin) Set(value bool) {
	r := p.reg()
	if value {
		r.GPIO.PSOR.Set(1 << r.Bit)
	} else {
		r.GPIO.PCOR.Set(1 << r.Bit)
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	r := p.reg()
	return r.GPIO.PDIR.HasBits(1 << r.Bit)
}

func (p Pin) Control() *volatile.Register32 {
	return p.reg().PCR
}

func (p Pin) Fast() FastPin {
	r := p.reg()
	return FastPin{
		PDOR: r.GPIO.PDOR.Bit(r.Bit),
		PSOR: r.GPIO.PSOR.Bit(r.Bit),
		PCOR: r.GPIO.PCOR.Bit(r.Bit),
		PTOR: r.GPIO.PTOR.Bit(r.Bit),
		PDIR: r.GPIO.PDIR.Bit(r.Bit),
		PDDR: r.GPIO.PDDR.Bit(r.Bit),
	}
}

func (p FastPin) Set()         { p.PSOR.Set(true) }
func (p FastPin) Clear()       { p.PCOR.Set(true) }
func (p FastPin) Toggle()      { p.PTOR.Set(true) }
func (p FastPin) Write(v bool) { p.PDOR.Set(v) }
func (p FastPin) Read() bool   { return p.PDIR.Get() }
