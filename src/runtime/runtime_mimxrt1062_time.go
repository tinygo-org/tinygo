//go:build mimxrt1062
// +build mimxrt1062

package runtime

import (
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

type timeUnit int64

const (
	lastCycle      = SYSTICK_FREQ/1000 - 1
	cyclesPerMicro = CORE_FREQ / 1000000
)

const (
	pitFreq           = OSC_FREQ // PIT/GPT are muxed to 24 MHz OSC
	pitCyclesPerMicro = pitFreq / 1000000
	pitSleepTimer     = 0 // x4 32-bit PIT timers [0..3]
)

var (
	tickCount  volatile.Register64
	cycleCount volatile.Register32
	pitActive  volatile.Register32
	pitTimeout interrupt.Interrupt
)

var (
	// debug exception and monitor control
	DEM_CR     = (*volatile.Register32)(unsafe.Pointer(uintptr(0xe000edfc)))
	DWT_CR     = (*volatile.Register32)(unsafe.Pointer(uintptr(0xe0001000)))
	DWT_CYCCNT = (*volatile.Register32)(unsafe.Pointer(uintptr(0xe0001004)))
)

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func initSysTick() {

	const (
		traceEnable      = 0x01000000 // enable debugging & monitoring blocks
		cycleCountEnable = 0x00000001 // cycle count register
	)

	// disable SysTick if already running
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

	// turn on cycle counter
	DEM_CR.SetBits(traceEnable)
	DWT_CR.SetBits(cycleCountEnable)
	cycleCount.Set(DWT_CYCCNT.Get())

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

func initRTC() {
	if !nxp.SNVS.LPCR.HasBits(nxp.SNVS_LPCR_SRTC_ENV) {
		// if SRTC isn't running, start it with default Jan 1, 2019
		nxp.SNVS.LPSRTCLR.Set(uint32((0x5c2aad80 << 15) & 0xFFFFFFFF))
		nxp.SNVS.LPSRTCMR.Set(uint32(0x5c2aad80 >> 17))
		nxp.SNVS.LPCR.SetBits(nxp.SNVS_LPCR_SRTC_ENV)
	}
}

//go:export SysTick_Handler
func tick() {
	tickCount.Set(tickCount.Get() + 1)
	cycleCount.Set(DWT_CYCCNT.Get())
}

func ticks() timeUnit {
	mask := arm.DisableInterrupts()
	tick := tickCount.Get()
	cycs := cycleCount.Get()
	curr := DWT_CYCCNT.Get()
	arm.EnableInterrupts(mask)
	var diff uint32
	if curr < cycs { // cycle counter overflow/rollover occurred
		diff = (0xFFFFFFFF - cycs) + curr
	} else {
		diff = curr - cycs
	}
	frac := uint64(diff*0xFFFFFFFF/cyclesPerMicro) >> 32
	if frac > 1000 {
		frac = 1000
	}
	return timeUnit(1000*tick + frac)
}

func sleepTicks(duration timeUnit) {
	if duration >= 0 {
		curr := ticks()
		last := curr + duration // 64-bit overflow unlikely
		for curr < last {
			cycles := timeUnit((last - curr) / pitCyclesPerMicro)
			if cycles > 0xFFFFFFFF {
				cycles = 0xFFFFFFFF
			}
			if !timerSleep(uint32(cycles)) {
				return // return early due to interrupt
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
