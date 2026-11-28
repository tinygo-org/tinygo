//go:build sam && atsamd21

package machine

import "device/sam"

// ResetReasonRaw returns the PM.RCAUSE register.
func ResetReasonRaw() uint32 {
	return uint32(sam.PM.RCAUSE.Get())
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// PM.RCAUSE, SAM D21 datasheet DS40001882 section 16.8.14.
	cause := ResetReasonRaw()
	switch {
	case cause&sam.PM_RCAUSE_SYST != 0:
		return ResetSoftware
	case cause&sam.PM_RCAUSE_WDT != 0:
		return ResetWatchdog
	case cause&sam.PM_RCAUSE_EXT != 0:
		return ResetExternal
	case cause&(sam.PM_RCAUSE_BOD12|sam.PM_RCAUSE_BOD33) != 0:
		return ResetBrownout
	case cause&sam.PM_RCAUSE_POR != 0:
		return ResetPowerOn
	default:
		return ResetUnknown
	}
}
