//go:build rp2040 || rp2350

package machine

import "device/rp"

// ResetReason returns why the chip last reset.
//
// CPUReset uses SYSRESETREQ, which the chip does not record. After CPUReset
// this returns the cause of the reset before it.
func ResetReason() ResetCause {
	raw := ResetReasonRaw()
	switch {
	case raw&rp.WATCHDOG_REASON_FORCE != 0:
		return ResetSoftware
	case raw&rp.WATCHDOG_REASON_TIMER != 0:
		return ResetWatchdog
	default:
		return chipResetCause(raw)
	}
}
