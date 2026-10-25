//go:build sam && atsam3x8e

package runtime

import (
	"device/arm"
	"device/sam"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	cpuFrequency              = 84_000_000
	cyclesPerMicrosecond      = cpuFrequency / 1_000_000
	microsecondsPerInterrupt  = 1_000
	nanosecondsPerMicrosecond = 1_000
	sysTickReload             = cyclesPerMicrosecond*microsecondsPerInterrupt - 1
)

// WDT_MR is write once and the watchdog runs after reset.
// See SAM3X/A datasheet 11057C, section 15.5.2.
var wdtMR = (*volatile.Register32)(unsafe.Pointer(uintptr(0x400e1a54)))

const wdtMRWDDIS = 1 << 15

var sysTickCount volatile.Register64

// ticks returns the time in microseconds since reset.
func ticks() timeUnit {
	state := interrupt.Disable()
	milliseconds := sysTickCount.Get()
	current := arm.SYST.SYST_CVR.Get()
	if arm.SCB.ICSR.HasBits(arm.SCB_ICSR_PENDSTSET) {
		current = arm.SYST.SYST_CVR.Get()
		milliseconds++
	}
	interrupt.Restore(state)

	elapsedCycles := uint32(sysTickReload) - current
	microseconds := elapsedCycles / cyclesPerMicrosecond
	return timeUnit(milliseconds*microsecondsPerInterrupt + uint64(microseconds))
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * nanosecondsPerMicrosecond
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / nanosecondsPerMicrosecond)
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}

	if hasScheduler {
		if d >= microsecondsPerInterrupt {
			waitForEvents()
		}
		return
	}

	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
		waitForEvents()
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

// initClocks runs MCK at 84 MHz from PLLA with the 12 MHz crystal.
// See ArduinoCore-sam system/CMSIS/Device/ATMEL/sam3xa/source/system_sam3xa.c.
func initClocks() {
	const (
		morKey     = sam.PMC_CKGR_MOR_KEY_PASSWD << sam.PMC_CKGR_MOR_KEY_Pos
		morStartup = 8 << sam.PMC_CKGR_MOR_MOSCXTST_Pos
		morXtal    = morKey | morStartup | sam.PMC_CKGR_MOR_MOSCRCEN | sam.PMC_CKGR_MOR_MOSCXTEN
		mckrPres   = sam.PMC_PMC_MCKR_PRES_CLK_2 << sam.PMC_PMC_MCKR_PRES_Pos
	)

	sam.EFC0.FMR.Set(4 << sam.EFC_FMR_FWS_Pos)
	sam.EFC1.FMR.Set(4 << sam.EFC_FMR_FWS_Pos)

	if !sam.PMC.CKGR_MOR.HasBits(sam.PMC_CKGR_MOR_MOSCSEL) {
		sam.PMC.CKGR_MOR.Set(morXtal)
		for !sam.PMC.PMC_SR.HasBits(sam.PMC_PMC_SR_MOSCXTS) {
		}
	}
	sam.PMC.CKGR_MOR.Set(morXtal | sam.PMC_CKGR_MOR_MOSCSEL)
	for !sam.PMC.PMC_SR.HasBits(sam.PMC_PMC_SR_MOSCSELS) {
	}

	sam.PMC.PMC_MCKR.ReplaceBits(sam.PMC_PMC_MCKR_CSS_MAIN_CLK, sam.PMC_PMC_MCKR_CSS_Msk, 0)
	waitMCKRDY()

	// PLLA = 12 MHz * 14 = 168 MHz.
	sam.PMC.CKGR_PLLAR.Set(sam.PMC_CKGR_PLLAR_ONE |
		13<<sam.PMC_CKGR_PLLAR_MULA_Pos |
		0x3f<<sam.PMC_CKGR_PLLAR_PLLACOUNT_Pos |
		1<<sam.PMC_CKGR_PLLAR_DIVA_Pos)
	for !sam.PMC.PMC_SR.HasBits(sam.PMC_PMC_SR_LOCKA) {
	}

	sam.PMC.PMC_MCKR.Set(mckrPres | sam.PMC_PMC_MCKR_CSS_MAIN_CLK)
	waitMCKRDY()
	sam.PMC.PMC_MCKR.Set(mckrPres | sam.PMC_PMC_MCKR_CSS_PLLA_CLK)
	waitMCKRDY()
}

func waitMCKRDY() {
	for !sam.PMC.PMC_SR.HasBits(sam.PMC_PMC_SR_MCKRDY) {
	}
}

func init() {
	initClocks()
	arm.SetupSystemTimer(sysTickReload)
	arm.EnableInterrupts(0)
	machine.InitSerial()
}

//go:export SysTick_Handler
func handleSysTick() {
	sysTickCount.Set(sysTickCount.Get() + 1)
}

//export Reset_Handler
func main() {
	wdtMR.Set(wdtMRWDDIS)
	preinit()
	run()
	exit(0)
}
