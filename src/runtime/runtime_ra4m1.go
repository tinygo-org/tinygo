//go:build ra4m1

package runtime

import (
	"device/arm"
	"device/renesas"
	"machine"
)

// [initClocks] selects the high-speed on-chip oscillator and sets ICLK to 48 MHz.
// The MCU OFS1 setting must select the 48 MHz oscillator frequency.
func sysTickFrequency() uint32 {
	return 48_000_000
}

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

func initClocks() {
	// Renesas FSP bsp/mcu/all/bsp_register_protection.c uses 0xA500 for PRCR.
	const protectKey = 0xA5 << renesas.SYSTEM_PRCR_PRKEY_Pos

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
	initSysTick()
	arm.EnableInterrupts(0)
	machine.InitSerial()
}

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}
