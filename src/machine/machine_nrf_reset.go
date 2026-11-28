//go:build nrf

package machine

import "device/nrf"

func readAndClearResetFlags() uint32 {
	flags := nrf.POWER.RESETREAS.Get()
	nrf.POWER.RESETREAS.Set(flags)
	return flags
}

// ResetReason returns why the chip last reset.
//
// The chip reports power on and brownout the same way, so both return
// ResetPowerOn.
func ResetReason() ResetCause {
	// POWER.RESETREAS, nRF52840 Product Specification v1.7 section 5.3.7.11.
	flags := ResetReasonRaw()
	switch {
	case flags == 0:
		return ResetPowerOn
	case flags&nrf.POWER_RESETREAS_DOG != 0:
		return ResetWatchdog
	case flags&nrf.POWER_RESETREAS_SREQ != 0:
		return ResetSoftware
	case flags&nrf.POWER_RESETREAS_LOCKUP != 0:
		return ResetOther
	case flags&nrf.POWER_RESETREAS_DIF != 0:
		return ResetDebug
	case flags&(0x1f<<16) != 0: // OFF, LPCOMP, NFC, VBUS
		return ResetDeepSleep
	case flags&nrf.POWER_RESETREAS_RESETPIN != 0:
		return ResetExternal
	default:
		return ResetOther
	}
}
