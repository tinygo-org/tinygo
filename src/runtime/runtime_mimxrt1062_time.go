// +build mimxrt1062

package runtime

import (
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
)

type timeUnit int64

const ( // HW divides 24 MHz XTALOSC down to 100 kHz
	lastCycle      = SYSTICK_FREQ/1000 - 1
	microsPerCycle = 1000000 / SYSTICK_FREQ
)

const (
	PIT_FREQ          = PERCLK_FREQ
	pitCyclesPerMicro = PIT_FREQ / 1000000
	pitSleepTimer     = 0 // x4 32-bit PIT timers [0..3]
)

var (
	tickCount  volatile.Register64
	pitActive  volatile.Register32
	pitTimeout interrupt.Interrupt
)

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func initSysTick() {
	// disable if already running
	if arm.SYST.SYST_CSR.HasBits(arm.SYST_CSR_ENABLE_Msk) {
		arm.SYST.SYST_CSR.ClearBits(arm.SYST_CSR_ENABLE_Msk)
	}
	// zeroize the counter
	tickCount.Set(0)
	arm.SYST.SYST_RVR.Set(lastCycle)
	arm.SYST.SYST_CVR.Set(0)
	arm.SYST.SYST_CSR.Set(arm.SYST_CSR_TICKINT | arm.SYST_CSR_ENABLE)
	// set SysTick and PendSV priority to 32
	nxp.SystemControl.SHPR3.Set((0x20 << nxp.SCB_SHPR3_PRI_15_Pos) |
		(0x20 << nxp.SCB_SHPR3_PRI_14_Pos))
	// enable PIT, disable counters
	nxp.PIT.MCR.Set(0)
	for i := range nxp.PIT.TIMER {
		nxp.PIT.TIMER[i].TCTRL.Set(0)
	}
	// register sleep timer interrupt
	pitTimeout = interrupt.New(nxp.IRQ_PIT, timerWake)
	pitTimeout.SetPriority(0x21)
	pitTimeout.Enable()
}

//go:export SysTick_Handler
func tick() {
	tickCount.Set(tickCount.Get() + 1)
}

func ticks() timeUnit {
	mask := arm.DisableInterrupts()
	curr := arm.SYST.SYST_CVR.Get()
	tick := tickCount.Get()
	pend := nxp.SystemControl.ICSR.HasBits(nxp.SCB_ICSR_PENDSTSET_Msk)
	arm.EnableInterrupts(mask)
	mics := timeUnit(tick * 1000)
	// if the systick counter was about to reset and ICSR indicates a pending
	// SysTick IRQ, increment count
	if pend && (curr > 50) {
		mics += 1000
	} else {
		mics += timeUnit((lastCycle - curr) * microsPerCycle)
	}
	return mics
}

func sleepTicks(duration timeUnit) {
	if duration >= 0 {
		curr := ticks()
		last := curr + duration
		for curr < last {
			cycles := timeUnit((last - curr) / pitCyclesPerMicro)
			if cycles > 0xFFFFFFFF {
				cycles = 0xFFFFFFFF
			}
			if !timerSleep(uint32(cycles)) {
				// return early due to interrupt
				return
			}
			curr = ticks()
		}
	}
}

func timerSleep(cycles uint32) bool {
	pitActive.Set(1)
	nxp.PIT.TIMER[pitSleepTimer].LDVAL.Set(cycles)
	nxp.PIT.TIMER[pitSleepTimer].TCTRL.Set(nxp.PIT_TIMER_TCTRL_TIE)     // enable interrupts
	nxp.PIT.TIMER[pitSleepTimer].TCTRL.SetBits(nxp.PIT_TIMER_TCTRL_TEN) // start timer
	for {
		//arm.Asm("wfi") // TODO: causes hardfault! why?
		if pitActive.Get() == 0 {
			return true
		}
		if hasScheduler {
			break // some other interrupt occurred and needs servicing
		}
	}
	timerWake(interrupt.Interrupt{}) // clear and disable timer
	return false
}

func timerWake(interrupt.Interrupt) {
	pitActive.Set(0)
	// TFLGn[TIF] are set to 1 when a timeout occurs on the associated timer, and
	// are cleared to 0 by writing a 1 to the corresponding TFLGn[TIF].
	nxp.PIT.TIMER[pitSleepTimer].TFLG.Set(nxp.PIT_TIMER_TFLG_TIF) // clear interrupt flag
	nxp.PIT.TIMER[pitSleepTimer].TCTRL.Set(0)                     // disable timer/interrupt enable flags
}
