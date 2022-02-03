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
// +build nxp,mk66f18

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

type timeUnit int64

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

// cyclesPerMilli-1 is used for the systick reset value
//   the systick current value will be decremented on every clock cycle
//   an interrupt is generated when the current value reaches 0
//   a value of freq/1000 generates a tick (irq) every millisecond (1/1000 s)
var cyclesPerMilli = machine.CPUFrequency() / 1000

// number of systick irqs (milliseconds) since boot
var systickCount volatile.Register64

func millisSinceBoot() uint64 {
	return systickCount.Get()
}

func initSysTick() {
	nxp.SysTick.RVR.Set(cyclesPerMilli - 1)
	nxp.SysTick.CVR.Set(0)
	nxp.SysTick.CSR.Set(nxp.SysTick_CSR_CLKSOURCE | nxp.SysTick_CSR_TICKINT | nxp.SysTick_CSR_ENABLE)
	nxp.SystemControl.SHPR3.Set((32 << nxp.SystemControl_SHPR3_PRI_15_Pos) | (32 << nxp.SystemControl_SHPR3_PRI_14_Pos)) // set systick and pendsv priority to 32
}

func initSleepTimer() {
	nxp.SIM.SCGC5.SetBits(nxp.SIM_SCGC5_LPTMR)
	nxp.LPTMR0.CSR.Set(nxp.LPTMR0_CSR_TIE)

	timerInterrupt = interrupt.New(nxp.IRQ_LPTMR0, timerWake)
	timerInterrupt.Enable()
}

//go:export SysTick_Handler
func tick() {
	systickCount.Set(systickCount.Get() + 1)
}

// ticks are in microseconds
func ticks() timeUnit {
	mask := arm.DisableInterrupts()
	current := nxp.SysTick.CVR.Get()        // current value of the systick counter
	count := millisSinceBoot()              // number of milliseconds since boot
	istatus := nxp.SystemControl.ICSR.Get() // interrupt status register
	arm.EnableInterrupts(mask)

	micros := timeUnit(count * 1000) // a tick (1ms) = 1000 us

	// if the systick counter was about to reset and ICSR indicates a pending systick irq, increment count
	if istatus&nxp.SystemControl_ICSR_PENDSTSET != 0 && current > 50 {
		micros += 1000
	} else {
		cycles := cyclesPerMilli - 1 - current // number of cycles since last 1ms tick
		cyclesPerMicro := machine.CPUFrequency() / 1000000
		micros += timeUnit(cycles / cyclesPerMicro)
	}

	return micros
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
