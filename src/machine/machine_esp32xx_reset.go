//go:build esp32 || esp32c3 || esp32c6 || esp32s3

package machine

import "runtime/interrupt"

//export software_reset
func romSoftwareReset()

// CPUReset performs a system reset. It resets the CPU and all digital
// peripherals, and keeps the RTC domain.
func CPUReset() {
	interrupt.Disable()
	romSoftwareReset()
	for {
	}
}

// ResetReason returns why the chip last reset.
//
// The chip reports power on and chip level brownout with the same code, so
// both return ResetPowerOn.
func ResetReason() ResetCause {
	// Codes from ESP-IDF soc/<chip>/include/soc/reset_reasons.h.
	switch ResetReasonRaw() {
	case 0x01:
		return ResetPowerOn
	case 0x03, 0x0C:
		return ResetSoftware
	case 0x05:
		return ResetDeepSleep
	case 0x07, 0x08, 0x09, 0x0B, 0x0D, 0x10, 0x11, 0x12:
		return ResetWatchdog
	case 0x0F:
		return ResetBrownout
	case 0x15, 0x16, 0x18:
		return ResetDebug
	default:
		return ResetOther
	}
}
