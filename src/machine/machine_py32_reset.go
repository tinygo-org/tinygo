//go:build py32

package machine

import "device/py32"

func readAndClearResetFlags() uint32 {
	flags := py32.RCC.CSR.Get()
	py32.RCC.CSR.SetBits(py32.RCC_CSR_RMVF)
	return flags
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// RCC_CSR bits 25 to 31 are the same on all PY32 parts.
	flags := ResetReasonRaw()
	switch {
	case flags&(3<<29) != 0:
		return ResetWatchdog
	case flags&(1<<28) != 0:
		return ResetSoftware
	case flags&(1<<31) != 0:
		return ResetOther
	case flags&(1<<27) != 0:
		return ResetPowerOn
	case flags&(1<<26) != 0:
		return ResetExternal
	case flags&(1<<25) != 0:
		return ResetOther
	default:
		return ResetUnknown
	}
}
