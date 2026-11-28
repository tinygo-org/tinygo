//go:build (sam && atsamd51) || (sam && atsame5x)

package machine

import "device/sam"

// ResetReasonRaw returns the RSTC.RCAUSE register.
func ResetReasonRaw() uint32 {
	return uint32(sam.RSTC.RCAUSE.Get())
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// RSTC.RCAUSE, SAM D5x/E5x datasheet DS60001507 section 21.8.1.
	cause := ResetReasonRaw()
	switch {
	case cause&sam.RSTC_RCAUSE_SYST != 0:
		return ResetSoftware
	case cause&sam.RSTC_RCAUSE_WDT != 0:
		return ResetWatchdog
	case cause&sam.RSTC_RCAUSE_EXT != 0:
		return ResetExternal
	case cause&sam.RSTC_RCAUSE_BACKUP != 0:
		return ResetDeepSleep
	case cause&sam.RSTC_RCAUSE_NVM != 0:
		return ResetOther
	case cause&(sam.RSTC_RCAUSE_BODCORE|sam.RSTC_RCAUSE_BODVDD) != 0:
		return ResetBrownout
	case cause&sam.RSTC_RCAUSE_POR != 0:
		return ResetPowerOn
	default:
		return ResetUnknown
	}
}
