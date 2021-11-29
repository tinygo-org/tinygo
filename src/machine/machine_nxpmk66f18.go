// Derivative work of Teensyduino Core Library
// http://www.pjrc.com/teensy/
// Copyright (c) 2017 PJRC.COM, LLC.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// 1. The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// 2. If the Software is incorporated into a build system that allows
// selection among a list of target devices, then similar target
// devices manufactured by PJRC.COM must be included in the list of
// target devices and selectable in the same manner.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
// BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// +build nxp,mk66f18

package machine

import (
	"device/nxp"
	"runtime/volatile"
	"unsafe"
)

const deviceName = nxp.Device

const (
	PinInput PinMode = iota
	PinInputPullUp
	PinInputPullDown
	PinOutput
	PinOutputOpenDrain
	PinDisable
)

const (
	PA00 Pin = iota
	PA01
	PA02
	PA03
	PA04
	PA05
	PA06
	PA07
	PA08
	PA09
	PA10
	PA11
	PA12
	PA13
	PA14
	PA15
	PA16
	PA17
	PA18
	PA19
	PA20
	PA21
	PA22
	PA23
	PA24
	PA25
	PA26
	PA27
	PA28
	PA29
)

const (
	PB00 Pin = iota + 32
	PB01
	PB02
	PB03
	PB04
	PB05
	PB06
	PB07
	PB08
	PB09
	PB10
	PB11
	_
	_
	_
	_
	PB16
	PB17
	PB18
	PB19
	PB20
	PB21
	PB22
	PB23
)

const (
	PC00 Pin = iota + 64
	PC01
	PC02
	PC03
	PC04
	PC05
	PC06
	PC07
	PC08
	PC09
	PC10
	PC11
	PC12
	PC13
	PC14
	PC15
	PC16
	PC17
	PC18
	PC19
)

const (
	PD00 Pin = iota + 96
	PD01
	PD02
	PD03
	PD04
	PD05
	PD06
	PD07
	PD08
	PD09
	PD10
	PD11
	PD12
	PD13
	PD14
	PD15
)

const (
	PE00 Pin = iota + 128
	PE01
	PE02
	PE03
	PE04
	PE05
	PE06
	PE07
	PE08
	PE09
	PE10
	PE11
	PE12
	PE13
	PE14
	PE15
	PE16
	PE17
	PE18
	PE19
	PE20
	PE21
	PE22
	PE23
	PE24
	PE25
	PE26
	PE27
	PE28
)

//go:inline
func (p Pin) reg() (*nxp.GPIO_Type, *volatile.Register32, uint8) {
	var gpio *nxp.GPIO_Type
	var pcr *nxp.PORT_Type

	switch p / 32 {
	case 0:
		gpio, pcr = nxp.GPIOA, nxp.PORTA
	case 1:
		gpio, pcr = nxp.GPIOB, nxp.PORTB
	case 2:
		gpio, pcr = nxp.GPIOC, nxp.PORTC
	case 3:
		gpio, pcr = nxp.GPIOD, nxp.PORTD
	case 5:
		gpio, pcr = nxp.GPIOE, nxp.PORTE
	default:
		panic("invalid pin number")
	}

	return gpio, &(*[32]volatile.Register32)(unsafe.Pointer(pcr))[p%32], uint8(p % 32)
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	gpio, pcr, pos := p.reg()

	switch config.Mode {
	case PinOutput:
		gpio.PDDR.SetBits(1 << pos)
		pcr.Set((1 << nxp.PORT_PCR0_MUX_Pos) | nxp.PORT_PCR0_SRE | nxp.PORT_PCR0_DSE)

	case PinOutputOpenDrain:
		gpio.PDDR.SetBits(1 << pos)
		pcr.Set((1 << nxp.PORT_PCR0_MUX_Pos) | nxp.PORT_PCR0_SRE | nxp.PORT_PCR0_DSE | nxp.PORT_PCR0_ODE)

	case PinInput:
		gpio.PDDR.ClearBits(1 << pos)
		pcr.Set((1 << nxp.PORT_PCR0_MUX_Pos))

	case PinInputPullUp:
		gpio.PDDR.ClearBits(1 << pos)
		pcr.Set((1 << nxp.PORT_PCR0_MUX_Pos) | nxp.PORT_PCR0_PE | nxp.PORT_PCR0_PS)

	case PinInputPullDown:
		gpio.PDDR.ClearBits(1 << pos)
		pcr.Set((1 << nxp.PORT_PCR0_MUX_Pos) | nxp.PORT_PCR0_PE)

	case PinDisable:
		gpio.PDDR.ClearBits(1 << pos)
		pcr.Set((0 << nxp.PORT_PCR0_MUX_Pos))
	}
}

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p Pin) Set(value bool) {
	gpio, _, pos := p.reg()
	if value {
		gpio.PSOR.Set(1 << pos)
	} else {
		gpio.PCOR.Set(1 << pos)
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	gpio, _, pos := p.reg()
	return gpio.PDIR.HasBits(1 << pos)
}

func (p Pin) Control() *volatile.Register32 {
	_, pcr, _ := p.reg()
	return pcr
}

func (p Pin) Fast() FastPin {
	gpio, _, pos := p.reg()
	return FastPin{
		PDOR: gpio.PDOR.Bit(pos),
		PSOR: gpio.PSOR.Bit(pos),
		PCOR: gpio.PCOR.Bit(pos),
		PTOR: gpio.PTOR.Bit(pos),
		PDIR: gpio.PDIR.Bit(pos),
		PDDR: gpio.PDDR.Bit(pos),
	}
}

type FastPin struct {
	PDOR *volatile.BitRegister
	PSOR *volatile.BitRegister
	PCOR *volatile.BitRegister
	PTOR *volatile.BitRegister
	PDIR *volatile.BitRegister
	PDDR *volatile.BitRegister
}

func (p FastPin) Set()         { p.PSOR.Set(true) }
func (p FastPin) Clear()       { p.PCOR.Set(true) }
func (p FastPin) Toggle()      { p.PTOR.Set(true) }
func (p FastPin) Write(v bool) { p.PDOR.Set(v) }
func (p FastPin) Read() bool   { return p.PDIR.Get() }
