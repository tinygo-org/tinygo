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

// +build nxp,mkl26z4

package machine

import (
	"device/nxp"
	"runtime/volatile"
	"unsafe"
)

const Kinetis = "L"

type PinMode uint8

const (
	PinInput PinMode = iota
	PinInputPullUp
	PinInputPullDown
	PinOutput
	PinDisable
)

const (
	PA00 Pin = iota + (0 << 5)
	PA01
	PA02
	PA03
	PA04
	PA05
	PA06
	PA07

	PA12 Pin = iota + PA00 + 12
	PA13
	PA14
	PA15
	PA16
	PA17
	PA18
	PA19
	PA20
)

const (
	PB00 Pin = iota + (1 << 5)
	PB01
	PB02
	PB03

	PB07 Pin = iota + PB00 + 7
	PB08
	PB09
	PB10
	PB11

	PB16 Pin = iota + PB00 + 16
	PB17
	PB18
	PB19
	PB20
	PB21
	PB22
	PB23
)

const (
	PC00 Pin = iota + (2 << 5)
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

	PC16 Pin = iota + PC00 + 16
	PC17
	PC18

	PC20 Pin = iota + PC00 + 20
	PC21
	PC22
	PC23
)

const (
	PD00 Pin = iota + (3 << 5)
	PD01
	PD02
	PD03
	PD04
	PD05
	PD06
	PD07
)

const (
	PE00 Pin = iota + (4 << 5)
	PE01
	PE02
	PE03
	PE04
	PE05
	PE06

	PE16 Pin = iota + PC00 + 16
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

	PE29 Pin = iota + PC00 + 29
	PE30
	PE31
)

//go:inline
func (p Pin) reg() (*nxp.FGPIO_Type, *volatile.Register32, uint8) {
	var gpio *nxp.FGPIO_Type
	var pcr *nxp.PORT_Type

	switch p / 32 {
	case 0:
		gpio, pcr = nxp.FGPIOA, nxp.PORTA
	case 1:
		gpio, pcr = nxp.FGPIOB, nxp.PORTB
	case 2:
		gpio, pcr = nxp.FGPIOC, nxp.PORTC
	case 3:
		gpio, pcr = nxp.FGPIOD, nxp.PORTD
	case 5:
		gpio, pcr = nxp.FGPIOE, nxp.PORTE
	default:
		panic("invalid pin number")
	}

	return gpio, &(*[32]volatile.Register32)(unsafe.Pointer(pcr))[p%32], uint8(p % 32)
}

func (p Pin) Control() *volatile.Register32 {
	_, pcr, _ := p.reg()
	return pcr
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	gpio, pcr, pos := p.reg()

	switch config.Mode {
	case PinOutput:
		gpio.PDDR.SetBits(1 << pos)
		pcr.Set((1 << nxp.PORT_PCR0_MUX_Pos) | nxp.PORT_PCR0_SRE | nxp.PORT_PCR0_DSE)

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
