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
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
)

type UART = *UARTData

type UARTData struct {
	*nxp.UART_Type

	Config     PeripheralConfig
	Configured *UARTConfig

	Buffer       RingBuffer // RX Buffer
	TXBuffer     RingBuffer
	Transmitting volatile.Register8
	Interrupt    interrupt.Interrupt
}

var UART0 = UARTData{UART_Type: nxp.UART0, Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART0, []PinMux{UART0RX0, UART0TX0}}}
var UART1 = UARTData{UART_Type: nxp.UART1, Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART1, []PinMux{UART1RX0, UART1TX0}}}
var UART2 = UARTData{UART_Type: nxp.UART2, Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART2, []PinMux{UART2RX0, UART2TX0}}}
var UART3 = UARTData{UART_Type: nxp.UART3, Config: PeripheralConfig{&nxp.SIM.SCGC4, nxp.SIM_SCGC4_UART3, []PinMux{UART3RX0, UART3TX0}}}
var UART4 = UARTData{UART_Type: nxp.UART4, Config: PeripheralConfig{&nxp.SIM.SCGC1, nxp.SIM_SCGC1_UART4, []PinMux{UART4RX0, UART4TX0}}}

func init() {
	UART0.Interrupt = interrupt.New(nxp.IRQ_UART0_RX_TX, UART0.isrWithFIFO)
	UART1.Interrupt = interrupt.New(nxp.IRQ_UART1_RX_TX, UART1.isrWithFIFO)
	UART2.Interrupt = interrupt.New(nxp.IRQ_UART2_RX_TX, UART2.isrNoFIFO)
	UART3.Interrupt = interrupt.New(nxp.IRQ_UART3_RX_TX, UART3.isrNoFIFO)
	UART4.Interrupt = interrupt.New(nxp.IRQ_UART4_RX_TX, UART4.isrNoFIFO)
}

func (u UART) setBaudRate(baudRate uint32, canSched bool) {
	// copied from teensy core's BAUD2DIVn macros
	var divisor uint32
	if u == &UART0 || u == &UART1 {
		divisor = (fCPU>>1 + baudRate>>1) / baudRate
	} else {
		divisor = (fBUS>>1 + baudRate>>1) / baudRate
	}
	if divisor < 32 {
		divisor = 32
	}

	// set the divisor
	u.BDH.Set(uint8((divisor >> 13) & 0x1F))
	u.BDL.Set(uint8((divisor >> 5) & 0xFF))
	u.C4.Set(uint8(divisor & 0x1F))
}

func (u UART) postConfigure() {
	if u == &UART0 || u == &UART1 {
		u.C1.Set(nxp.UART_C1_ILT)

		// configure TX and RX watermark
		u.TWFIFO.Set(2) // causes bit TDRE of S1 to set
		u.RWFIFO.Set(4) // causes bit RDRF of S1 to set

		// enable FIFOs
		u.PFIFO.Set(nxp.UART_PFIFO_TXFE | nxp.UART_PFIFO_RXFE)

	} else {
		u.C1.Set(0)
		u.PFIFO.Set(0)
	}
}

//go:inline
func (u UART) poll() {
	if u == &UART0 || u == &UART1 {
		u.isrWithFIFO(u.Interrupt)
	} else {
		u.isrNoFIFO(u.Interrupt)
	}
}

//go:inline
func (u UART) txReady() bool {
	if u == &UART0 || u == &UART1 {
		return u.TCFIFO.Get() == 0
	} else {
		return u.S1.HasBits(nxp.UART_S1_TDRE)
	}
}

func (u UART) isrWithFIFO(interrupt.Interrupt) {
	// from: uart0_status_isr

	// receive
	if u.S1.HasBits(nxp.UART_S1_RDRF | nxp.UART_S1_IDLE) {
		intrs := arm.DisableInterrupts()
		avail := u.RCFIFO.Get()
		if avail == 0 {
			// The only way to clear the IDLE interrupt flag is
			// to read the data register.  But reading with no
			// data causes a FIFO underrun, which causes the
			// FIFO to return corrupted data.  If anyone from
			// Freescale reads this, what a poor design!  There
			// write should be a write-1-to-clear for IDLE.
			u.D.Get()
			// flushing the fifo recovers from the underrun,
			// but there's a possible race condition where a
			// new character could be received between reading
			// RCFIFO == 0 and flushing the FIFO.  To minimize
			// the chance, interrupts are disabled so a higher
			// priority interrupt (hopefully) doesn't delay.
			// TODO: change this to disabling the IDLE interrupt
			// which won't be simple, since we already manage
			// which transmit interrupts are enabled.
			u.CFIFO.Set(nxp.UART_CFIFO_RXFLUSH)
			arm.EnableInterrupts(intrs)

		} else {
			arm.EnableInterrupts(intrs)

			for {
				u.Buffer.Put(u.D.Get())
				avail--
				if avail <= 0 {
					break
				}
			}
		}
	}

	// transmit
	if u.C2.HasBits(nxp.UART_C2_TIE) && u.S1.HasBits(nxp.UART_S1_TDRE) {
		data := make([]byte, 0, uartTXFIFODepth)
		avail := uartTXFIFODepth - u.TCFIFO.Get()

		// get avail bytes from ring buffer
		for len(data) < int(avail) {
			if b, ok := u.TXBuffer.Get(); ok {
				data = append(data, b)
			} else {
				break
			}
		}

		// write data to FIFO
		l := len(data)
		for i, b := range data {
			if i == l-1 {
				// only clear TDRE on last write, per the manual
				u.S1.Get()
			}
			u.D.Set(b)
		}

		// if FIFO still has room, disable TIE, enable TCIE
		if u.S1.HasBits(nxp.UART_S1_TDRE) {
			u.C2.Set(uartC2TXCompleting)
		}
	}

	// transmit complete
	if u.C2.HasBits(nxp.UART_C2_TCIE) && u.S1.HasBits(nxp.UART_S1_TC) {
		u.Transmitting.Set(0)
		u.C2.Set(uartC2TXInactive)
	}
}
