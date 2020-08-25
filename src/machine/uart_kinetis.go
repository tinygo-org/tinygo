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

// +build nxp,kinetis

package machine

import (
	"device/nxp"
	"errors"
	"runtime/interrupt"

	_ "unsafe" // for go:linkname
)

const (
	uartC2Enable       = nxp.UART_C2_TE | nxp.UART_C2_RE | nxp.UART_C2_RIE | nxp.UART_C2_ILIE
	uartC2TXActive     = uartC2Enable | nxp.UART_C2_TIE
	uartC2TXCompleting = uartC2Enable | nxp.UART_C2_TCIE
	uartC2TXInactive   = uartC2Enable

	uartIRQPriority = 64

	// determined from UARTx_PFIFO
	uartRXFIFODepth = 8
	uartTXFIFODepth = 8
)

var ErrNotConfigured = errors.New("device has not been configured")

type UARTConfig struct {
	BaudRate uint32
	TX       PinMux
	RX       PinMux
}

//go:linkname gosched runtime.Gosched
func gosched()

// PutcharUART writes a byte to the UART synchronously, without using interrupts
// or calling the scheduler
func PutcharUART(u UART, c byte) {
	// ensure the UART has been configured
	if !u.Config.Enabled() {
		u.configure(UARTConfig{}, false)
	}

	for !u.txReady() {
		// busy wait
	}

	u.D.Set(c)
	u.C2.Set(uartC2TXActive)
}

// PollUART manually checks a UART status and calls the ISR. This should only be
// called by runtime.abort.
func PollUART(u UART) {
	if !u.Config.Enabled() {
		return
	}

	u.poll()
}

// Configure the UART
func (u UART) Configure(config UARTConfig) {
	u.configure(config, true)
}

func (u UART) configure(config UARTConfig, canSched bool) {
	// from: serial_begin

	if config.RX.Mux == 0 {
		config.RX = u.Config.Pins[0]
	}
	if config.TX.Mux == 0 {
		config.TX = u.Config.Pins[1]
	}
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	wasConfigured := u.Configured != nil
	u.Configured = &config

	if !wasConfigured {
		u.Transmitting.Set(0)

		// turn on the clock
		u.Config.Enable()

		// configure pins
		config.RX.EnableWith(nxp.PORT_PCR0_PE | nxp.PORT_PCR0_PS | nxp.PORT_PCR0_PFE)
		config.TX.EnableWith(nxp.PORT_PCR0_DSE | nxp.PORT_PCR0_SRE)
	}

	if wasConfigured {
		// don't change baud rate mid transmit
		if canSched {
			u.Flush()
		} else {
			for u.Transmitting.Get() != 0 {
				// busy wait flush
			}
		}
	}

	u.setBaudRate(config.BaudRate, canSched)

	if !wasConfigured {
		u.postConfigure()

		// setup interrupts
		u.C2.Set(uartC2TXInactive)
		u.Interrupt.SetPriority(uartIRQPriority)
		u.Interrupt.Enable()
	}
}

func (u UART) Disable() {
	// from: serial_end

	// check if the device has been enabled already
	if u.Configured == nil {
		return
	}

	// flush transmit queue
	u.Flush()

	// disable interrupt
	u.Interrupt.Disable()

	// disable UART
	u.C2.Set(0)

	// disable pins
	u.Configured.RX.Pin.Configure(PinConfig{Mode: PinDisable})
	u.Configured.TX.Pin.Configure(PinConfig{Mode: PinDisable})

	// clear flags
	u.S1.Get()
	u.D.Get()
	u.Buffer.Clear()

	// disable clock gate
	u.Config.Disable()
}

func (u UART) Flush() {
	for u.Transmitting.Get() != 0 {
		gosched()
	}
}

// WriteByte writes a byte of data to the UART.
func (u UART) WriteByte(c byte) error {
	if u.Configured == nil {
		return ErrNotConfigured
	}

	for !u.TXBuffer.Put(c) {
		gosched()
	}

	u.Transmitting.Set(1)
	u.C2.Set(uartC2TXActive)
	return nil
}

func (u UART) isrNoFIFO(interrupt.Interrupt) {
	// from: uart0_status_isr

	// receive
	if u.S1.HasBits(nxp.UART_S1_RDRF | nxp.UART_S1_IDLE) {
		u.Buffer.Put(u.D.Get())
	}

	// transmit
	if u.C2.HasBits(nxp.UART_C2_TIE) && u.S1.HasBits(nxp.UART_S1_TDRE) {
		if b, ok := u.TXBuffer.Get(); ok {
			u.D.Set(b)
		} else {
			u.C2.Set(uartC2TXCompleting)
		}
	}

	// transmit complete
	if u.C2.HasBits(nxp.UART_C2_TCIE) && u.S1.HasBits(nxp.UART_S1_TC) {
		u.Transmitting.Set(0)
		u.C2.Set(uartC2TXInactive)
	}
}
