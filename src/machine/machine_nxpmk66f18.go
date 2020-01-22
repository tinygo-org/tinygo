// +build nxp,mk66f18

package machine

import (
	"device/nxp"
	"runtime/volatile"
)

const (
	PortControlRegisterSRE = nxp.PORT_PCR0_SRE
	PortControlRegisterDSE = nxp.PORT_PCR0_DSE
	PortControlRegisterODE = nxp.PORT_PCR0_ODE
)

func PortControlRegisterMUX(v uint8) uint32 {
	return (uint32(v) << nxp.PORT_PCR0_MUX_Pos) & nxp.PORT_PCR0_MUX_Msk
}

type pinRegisters struct {
	Bit  uintptr
	PCR  *volatile.Register32
	PDOR *volatile.BitRegister
	PSOR *volatile.BitRegister
	PCOR *volatile.BitRegister
	PTOR *volatile.BitRegister
	PDIR *volatile.BitRegister
	PDDR *volatile.BitRegister
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	switch config.Mode {
	case PinInput:
		panic("todo")

	case PinOutput:
		p.registers().PDDR.Set()
		p.registers().PCR.SetBits(PortControlRegisterSRE | PortControlRegisterDSE | PortControlRegisterMUX(1))
		p.registers().PCR.ClearBits(PortControlRegisterODE)
	}
}

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p Pin) Set(value bool) {
	if value {
		p.registers().PSOR.Set()
	} else {
		p.registers().PCOR.Set()
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	return p.registers().PDIR.Get()
}
