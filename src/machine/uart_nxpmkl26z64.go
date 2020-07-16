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
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

type UART = *UARTData

type UARTData struct {
	*UARTReg

	Config     PeripheralConfig
	Configured *UARTConfig

	Buffer       RingBuffer // RX Buffer
	TXBuffer     RingBuffer
	Transmitting volatile.Register8
	Interrupt    interrupt.Interrupt
}

type UARTReg struct {
	BDH volatile.Register8 // 0x0
	BDL volatile.Register8 // 0x1
	C1  volatile.Register8 // 0x2
	C2  volatile.Register8 // 0x3
	S1  volatile.Register8 // 0x4
	S2  volatile.Register8 // 0x5
	C3  volatile.Register8 // 0x6
	D   volatile.Register8 // 0x7
}

var UART0 = UARTData{UARTReg: (*UARTReg)(unsafe.Pointer(nxp.UART0)), Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART0, []PinMux{UART0RX0, UART0TX0}}}
var UART1 = UARTData{UARTReg: (*UARTReg)(unsafe.Pointer(nxp.UART1)), Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART1, []PinMux{UART1RX0, UART1TX0}}}
var UART2 = UARTData{UARTReg: (*UARTReg)(unsafe.Pointer(nxp.UART2)), Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART2, []PinMux{UART2RX0, UART2TX0}}}

func init() {
	UART0.Interrupt = interrupt.New(nxp.IRQ_UART0, UART0.isrNoFIFO)
	UART1.Interrupt = interrupt.New(nxp.IRQ_UART1, UART1.isrNoFIFO)
	UART2.Interrupt = interrupt.New(nxp.IRQ_UART2, UART2.isrNoFIFO)
}

func (u UART) setBaudRate(baudRate uint32, canSched bool) {
	// copied from teensy core's BAUD2DIVn macros
	var divisor uint32
	if u == &UART0 {
		divisor = (fPLL>>5 + baudRate>>1) / baudRate
	} else {
		divisor = (fBUS>>4 + baudRate>>1) / baudRate
	}
	if divisor < 1 {
		divisor = 1
	}

	// set the divisor
	u.BDH.Set(uint8((divisor >> 8) & 0x1F))
	u.BDL.Set(uint8(divisor & 0xFF))
}

func (u UART) postConfigure() {
	u.C1.Set(0)
}

//go:inline
func (u UART) poll() {
	u.isrNoFIFO(u.Interrupt)
}

//go:inline
func (u UART) txReady() bool {
	return u.S1.HasBits(nxp.UART_S1_TDRE)
}
