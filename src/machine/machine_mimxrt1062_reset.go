//go:build mimxrt1062

package machine

import "device/nxp"

func readAndClearResetFlags() uint32 {
	flags := nxp.SRC.SRSR.Get()
	nxp.SRC.SRSR.Set(flags)
	return flags
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// SRC_SRSR, i.MX RT1060 reference manual section 21.8.3.
	flags := ResetReasonRaw()
	switch {
	case flags&(nxp.SRC_SRSR_WDOG_RST_B|nxp.SRC_SRSR_WDOG3_RST_B) != 0:
		return ResetWatchdog
	case flags&nxp.SRC_SRSR_LOCKUP_SYSRESETREQ != 0:
		return ResetSoftware
	case flags&(nxp.SRC_SRSR_JTAG_RST_B|nxp.SRC_SRSR_JTAG_SW_RST) != 0:
		return ResetDebug
	case flags&(nxp.SRC_SRSR_CSU_RESET_B|nxp.SRC_SRSR_TEMPSENSE_RST_B) != 0:
		return ResetOther
	case flags&nxp.SRC_SRSR_IPP_USER_RESET_B != 0:
		return ResetExternal
	case flags&nxp.SRC_SRSR_IPP_RESET_B != 0:
		return ResetPowerOn
	default:
		return ResetUnknown
	}
}
