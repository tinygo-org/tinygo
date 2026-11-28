//go:build rp2040

package machine

import "device/rp"

// ResetReasonRaw returns CHIP_RESET with WATCHDOG.REASON in the low bits.
func ResetReasonRaw() uint32 {
	return rp.VREG_AND_CHIP_RESET.CHIP_RESET.Get()&^0xff | rp.WATCHDOG.REASON.Get()&0x3
}

func chipResetCause(raw uint32) ResetCause {
	// CHIP_RESET, RP2040 datasheet section 2.10.6.
	switch {
	case raw&rp.VREG_AND_CHIP_RESET_CHIP_RESET_HAD_PSM_RESTART != 0:
		return ResetDebug
	case raw&rp.VREG_AND_CHIP_RESET_CHIP_RESET_HAD_RUN != 0:
		return ResetExternal
	case raw&rp.VREG_AND_CHIP_RESET_CHIP_RESET_HAD_POR != 0:
		return ResetPowerOn
	default:
		return ResetUnknown
	}
}
