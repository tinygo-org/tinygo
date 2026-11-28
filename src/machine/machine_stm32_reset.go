//go:build stm32 && !stm32h7

package machine

import "device/stm32"

func readAndClearResetFlags() uint32 {
	flags := stm32.RCC.CSR.Get()
	stm32.RCC.CSR.SetBits(stm32.RCC_CSR_RMVF)
	return flags
}

// ResetReason returns why the chip last reset.
//
// Except on F4 and F7, the chip reports power on and brownout with the same
// flag, so both return ResetPowerOn.
func ResetReason() ResetCause {
	// RCC_CSR bits 26 to 31 are the same on all families, see for example
	// RM0090 section 7.3.21 and RM0351 section 6.4.29.
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
	case flags&stm32ResetFlagBOR != 0:
		return ResetBrownout
	case flags&(1<<26) != 0:
		return ResetExternal
	case flags&(1<<25) != 0:
		return ResetOther
	default:
		return ResetUnknown
	}
}
