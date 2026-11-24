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

//go:build nxp && mk66f18

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

func sysTickFrequency() uint32 {
	return machine.CPUFrequency()
}

func initSleepTimer() {
	nxp.SIM.SCGC5.SetBits(nxp.SIM_SCGC5_LPTMR)
	nxp.LPTMR0.CSR.Set(nxp.LPTMR0_CSR_TIE)

	timerInterrupt = interrupt.New(nxp.IRQ_LPTMR0, timerWake)
	timerInterrupt.Enable()
}

// sleepTicks spins for a number of microseconds
func sleepTicks(duration timeUnit) {
	now := ticks()
	end := duration + now
	cyclesPerMicro := machine.ClockFrequency() / 1000000

	if duration <= 0 {
		return
	}

	nxp.LPTMR0.PSR.Set((3 << nxp.LPTMR0_PSR_PCS_Pos) | nxp.LPTMR0_PSR_PBYP) // use 16MHz clock, undivided

	for now < end {
		count := uint32(end-now) / cyclesPerMicro
		if count > 65535 {
			count = 65535
		}

		if !timerSleep(count) {
			// return early due to interrupt
			return
		}

		now = ticks()
	}
}

var timerInterrupt interrupt.Interrupt
var timerActive volatile.Register32

func timerSleep(count uint32) bool {
	timerActive.Set(1)
	nxp.LPTMR0.CMR.Set(count)                  // set count
	nxp.LPTMR0.CSR.SetBits(nxp.LPTMR0_CSR_TEN) // enable

	for {
		arm.Asm("wfi")
		if timerActive.Get() == 0 {
			return true
		}

		if hasScheduler {
			// bail out, as the interrupt may have awoken a goroutine
			break
		}

		// if there is no scheduler, block for the entire count
	}

	timerWake(timerInterrupt)
	return false
}

func timerWake(interrupt.Interrupt) {
	timerActive.Set(0)
	nxp.LPTMR0.CSR.Set(nxp.LPTMR0.CSR.Get()&^nxp.LPTMR0_CSR_TEN | nxp.LPTMR0_CSR_TCF) // clear flag and disable
}
