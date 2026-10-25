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

//go:build nxp && mk20d7 && teensy32

package machine

func CPUFrequency() uint32 { return 72000000 }

// BusFrequency returns the peripheral bus clock (OUTDIV2).
func BusFrequency() uint32 { return 36000000 }

func ClockFrequency() uint32 { return 16000000 }

const (
	D00 = PB16
	D01 = PB17
	D02 = PD00
	D03 = PA12
	D04 = PA13
	D05 = PD07
	D06 = PD04
	D07 = PD02
	D08 = PD03
	D09 = PC03
	D10 = PC04
	D11 = PC06
	D12 = PC07
	D13 = PC05
	D14 = PD01
	D15 = PC00
	D16 = PB00
	D17 = PB01
	D18 = PB03
	D19 = PB02
	D20 = PD05
	D21 = PD06
	D22 = PC01
	D23 = PC02
	D24 = PA05
	D25 = PB19
	D26 = PE01
	D27 = PC09
	D28 = PC08
	D29 = PC10
	D30 = PC11
	D31 = PE00
	D32 = PB18
	D33 = PA04
)

const LED = PC05

var (
	TeensyUART1 = UART0
	TeensyUART2 = UART1
	TeensyUART3 = UART2
)

var DefaultUART = UART0

const (
	defaultUART0RX = D00
	defaultUART0TX = D01
	defaultUART1RX = D09
	defaultUART1TX = D10
	defaultUART2RX = D07
	defaultUART2TX = D08
)

const (
	I2C0_SDA_PIN = D18
	I2C0_SCL_PIN = D19
	I2C1_SDA_PIN = D30
	I2C1_SCL_PIN = D29

	SDA_PIN = I2C0_SDA_PIN
	SCL_PIN = I2C0_SCL_PIN
)
