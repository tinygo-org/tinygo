//go:build ra4m1

package machine

import "device/renesas"

// readAndClearResetFlags returns RSTSR0 in bits 0 to 7, RSTSR1 in bits 8 to
// 23 and RSTSR2 in bits 24 to 31.
func readAndClearResetFlags() uint32 {
	sys := renesas.SYSTEM
	flags := uint32(sys.RSTSR0.Get()) | uint32(sys.RSTSR1.Get())<<8 | uint32(sys.RSTSR2.Get())<<24
	sys.RSTSR0.Set(0)
	sys.RSTSR1.Set(0)
	sys.RSTSR2.Set(renesas.SYSTEM_RSTSR2_CWSF)
	return flags
}

// ResetReason returns why the chip last reset.
func ResetReason() ResetCause {
	// RSTSR0, RSTSR1 and RSTSR2, RA4M1 user manual R01UH0887 section 6.2.
	flags := ResetReasonRaw()
	rstsr0 := flags & 0xff
	rstsr1 := (flags >> 8) & 0xffff
	switch {
	case rstsr0&renesas.SYSTEM_RSTSR0_PORF != 0:
		return ResetPowerOn
	case rstsr0&(renesas.SYSTEM_RSTSR0_LVD0RF|renesas.SYSTEM_RSTSR0_LVD1RF|renesas.SYSTEM_RSTSR0_LVD2RF) != 0:
		return ResetBrownout
	case rstsr1&(renesas.SYSTEM_RSTSR1_IWDTRF|renesas.SYSTEM_RSTSR1_WDTRF) != 0:
		return ResetWatchdog
	case rstsr1&renesas.SYSTEM_RSTSR1_SWRF != 0:
		return ResetSoftware
	case rstsr1 != 0:
		return ResetOther
	case flags>>24&renesas.SYSTEM_RSTSR2_CWSF == 0:
		return ResetPowerOn
	default:
		return ResetExternal
	}
}
