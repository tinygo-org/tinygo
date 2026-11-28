//go:build rp2350

package machine

import "device/rp"

// ResetReasonRaw returns POWMAN CHIP_RESET with WATCHDOG.REASON in the low bits.
func ResetReasonRaw() uint32 {
	return rp.POWMAN.CHIP_RESET.Get()&^0xff | rp.WATCHDOG.REASON.Get()&0x3
}

func chipResetCause(raw uint32) ResetCause {
	// POWMAN CHIP_RESET, RP2350 datasheet section 6.4.
	switch {
	case raw&(rp.POWMAN_CHIP_RESET_HAD_WATCHDOG_RESET_RSM|rp.POWMAN_CHIP_RESET_HAD_WATCHDOG_RESET_SWCORE|
		rp.POWMAN_CHIP_RESET_HAD_WATCHDOG_RESET_POWMAN|rp.POWMAN_CHIP_RESET_HAD_WATCHDOG_RESET_POWMAN_ASYNC) != 0:
		return ResetWatchdog
	case raw&(rp.POWMAN_CHIP_RESET_HAD_DP_RESET_REQ|rp.POWMAN_CHIP_RESET_HAD_RESCUE|rp.POWMAN_CHIP_RESET_HAD_HZD_SYS_RESET_REQ) != 0:
		return ResetDebug
	case raw&rp.POWMAN_CHIP_RESET_HAD_SWCORE_PD != 0:
		return ResetDeepSleep
	case raw&rp.POWMAN_CHIP_RESET_HAD_GLITCH_DETECT != 0:
		return ResetOther
	case raw&rp.POWMAN_CHIP_RESET_HAD_RUN_LOW != 0:
		return ResetExternal
	case raw&rp.POWMAN_CHIP_RESET_HAD_BOR != 0:
		return ResetBrownout
	case raw&rp.POWMAN_CHIP_RESET_HAD_POR != 0:
		return ResetPowerOn
	default:
		return ResetUnknown
	}
}
