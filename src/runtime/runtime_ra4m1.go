//go:build ra4m1

package runtime

import (
	"device/arm"
	"device/renesas"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

// Prior to loading SysTick at startup, [initClocks] selects the high-speed
// on-chip oscillator.
// It sets ICLK to 48 MHz and starts SysTick with a 1 ms period.
// The MCU OFS1 setting must select the 48 MHz oscillator frequency.
const (
	cpuFrequency              = 48_000_000
	cyclesPerMicrosecond      = cpuFrequency / 1_000_000
	microsecondsPerInterrupt  = 1_000
	nanosecondsPerMicrosecond = 1_000
	sysTickReload             = cyclesPerMicrosecond*microsecondsPerInterrupt - 1
)

// RA4M1 requires BCK bits 16 to 18 to match PCKB.
// See ArduinoCore-renesas 1.6.0 ra4m1/bsp_feature.h.
const (
	bckDivPos    = 16
	clockDivMask = renesas.SYSTEM_SCKDIVCR_FCK_Msk |
		renesas.SYSTEM_SCKDIVCR_ICK_Msk |
		renesas.SYSTEM_SCKDIVCR_PCKA_Msk |
		(0x7 << bckDivPos) |
		renesas.SYSTEM_SCKDIVCR_PCKB_Msk |
		renesas.SYSTEM_SCKDIVCR_PCKC_Msk |
		renesas.SYSTEM_SCKDIVCR_PCKD_Msk
	clockDivSlow = (renesas.SYSTEM_SCKDIVCR_FCK_100 << renesas.SYSTEM_SCKDIVCR_FCK_Pos) |
		(renesas.SYSTEM_SCKDIVCR_ICK_100 << renesas.SYSTEM_SCKDIVCR_ICK_Pos) |
		(renesas.SYSTEM_SCKDIVCR_PCKA_100 << renesas.SYSTEM_SCKDIVCR_PCKA_Pos) |
		(renesas.SYSTEM_SCKDIVCR_PCKB_100 << bckDivPos) |
		(renesas.SYSTEM_SCKDIVCR_PCKB_100 << renesas.SYSTEM_SCKDIVCR_PCKB_Pos) |
		(renesas.SYSTEM_SCKDIVCR_PCKC_100 << renesas.SYSTEM_SCKDIVCR_PCKC_Pos) |
		(renesas.SYSTEM_SCKDIVCR_PCKD_100 << renesas.SYSTEM_SCKDIVCR_PCKD_Pos)
	clockDivFast = (renesas.SYSTEM_SCKDIVCR_FCK_001 << renesas.SYSTEM_SCKDIVCR_FCK_Pos) |
		(renesas.SYSTEM_SCKDIVCR_PCKB_001 << bckDivPos) |
		(renesas.SYSTEM_SCKDIVCR_PCKB_001 << renesas.SYSTEM_SCKDIVCR_PCKB_Pos)
)

var sysTickCount volatile.Register64

// ticks returns the time in microseconds since reset.
func ticks() timeUnit {
	for {
		state := interrupt.Disable()
		milliseconds := sysTickCount.Get()
		current := arm.SYST.SYST_CVR.Get()
		pending := arm.SCB.ICSR.HasBits(arm.SCB_ICSR_PENDSTSET)
		interrupt.Restore(state)

		if pending {
			continue
		}

		elapsedCycles := uint32(sysTickReload) - current
		microseconds := elapsedCycles / cyclesPerMicrosecond
		return timeUnit(
			milliseconds*microsecondsPerInterrupt + uint64(microseconds),
		)
	}
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

func initClocks() {
	const protectKey = renesas.SYSTEM_PRCR_PRKEY_0x5A << renesas.SYSTEM_PRCR_PRKEY_Pos

	renesas.SYSTEM.PRCR.Set(
		protectKey | renesas.SYSTEM_PRCR_PRC0_Msk | renesas.SYSTEM_PRCR_PRC1_Msk,
	)

	renesas.SYSTEM.SCKDIVCR.ReplaceBits(clockDivSlow, clockDivMask, 0)
	renesas.SYSTEM.MEMWAIT.Set(renesas.SYSTEM_MEMWAIT_MEMWAIT_1)
	renesas.SYSTEM.HOCOCR.ClearBits(renesas.SYSTEM_HOCOCR_HCSTP_Msk)

	for !renesas.SYSTEM.OSCSF.HasBits(renesas.SYSTEM_OSCSF_HOCOSF_Msk) {
	}
	for renesas.SYSTEM.OPCCR.HasBits(renesas.SYSTEM_OPCCR_OPCMTSF_Msk) {
	}

	renesas.SYSTEM.OPCCR.ReplaceBits(
		renesas.SYSTEM_OPCCR_OPCM_00,
		renesas.SYSTEM_OPCCR_OPCM_Msk,
		renesas.SYSTEM_OPCCR_OPCM_Pos,
	)

	for renesas.SYSTEM.OPCCR.HasBits(renesas.SYSTEM_OPCCR_OPCMTSF_Msk) {
	}

	renesas.SYSTEM.SCKSCR.ReplaceBits(
		renesas.SYSTEM_SCKSCR_CKSEL_000,
		renesas.SYSTEM_SCKSCR_CKSEL_Msk,
		renesas.SYSTEM_SCKSCR_CKSEL_Pos,
	)
	renesas.SYSTEM.SCKDIVCR.ReplaceBits(clockDivFast, clockDivMask, 0)

	renesas.SYSTEM.PRCR.Set(protectKey)
}

func init() {
	initClocks()
	arm.SetupSystemTimer(sysTickReload)
	arm.EnableInterrupts(0)
}

//go:export SysTick_Handler
func handleSysTick() {
	sysTickCount.Set(sysTickCount.Get() + 1)
}

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}
