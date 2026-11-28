//go:build stm32 && stm32h7

package machine

import "device/stm32"

func readAndClearResetFlags() uint32 {
	flags := stm32.RCC.RSR.Get()
	stm32.RCC.RSR.SetBits(stm32.RCC_RSR_RMVF)
	return flags
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// RCC_RSR, RM0433 section 8.7.40.
	flags := ResetReasonRaw()
	switch {
	case flags&(stm32.RCC_RSR_IWDG1RSTF|stm32.RCC_RSR_WWDG1RSTF) != 0:
		return ResetWatchdog
	case flags&stm32.RCC_RSR_SFTRSTF != 0:
		return ResetSoftware
	case flags&stm32.RCC_RSR_LPWRRSTF != 0:
		return ResetOther
	case flags&stm32.RCC_RSR_PORRSTF != 0:
		return ResetPowerOn
	case flags&stm32.RCC_RSR_BORRSTF != 0:
		return ResetBrownout
	case flags&stm32.RCC_RSR_PINRSTF != 0:
		return ResetExternal
	default:
		return ResetOther
	}
}
