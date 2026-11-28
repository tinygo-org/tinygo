//go:build nxp && mk66f18

package machine

import "device/nxp"

// ResetReasonRaw returns RCM.SRS1 in bits 8 to 15 and RCM.SRS0 in bits 0 to 7.
func ResetReasonRaw() uint32 {
	return uint32(nxp.RCM.SRS1.Get())<<8 | uint32(nxp.RCM.SRS0.Get())
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// RCM_SRS0 and RCM_SRS1, K66 reference manual K66P144M180SF5RMV2 section 15.2.
	srs0 := uint8(ResetReasonRaw())
	srs1 := uint8(ResetReasonRaw() >> 8)
	switch {
	case srs0&nxp.RCM_SRS0_POR != 0:
		return ResetPowerOn
	case srs0&nxp.RCM_SRS0_WDOG != 0:
		return ResetWatchdog
	case srs1&nxp.RCM_SRS1_SW != 0:
		return ResetSoftware
	case srs1&(nxp.RCM_SRS1_JTAG|nxp.RCM_SRS1_MDM_AP) != 0:
		return ResetDebug
	case srs0&nxp.RCM_SRS0_LVD != 0:
		return ResetBrownout
	case srs0&nxp.RCM_SRS0_WAKEUP != 0:
		return ResetDeepSleep
	case srs0&nxp.RCM_SRS0_PIN != 0:
		return ResetExternal
	case srs0 != 0 || srs1 != 0:
		return ResetOther
	default:
		return ResetUnknown
	}
}
