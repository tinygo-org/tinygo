package machine

// ResetCause is why the chip last reset.
//
// Ports that support it provide ResetReason, which returns the cause, and
// ResetReasonRaw, which returns the chip specific reset code.
type ResetCause uint8

const (
	ResetUnknown ResetCause = iota
	ResetPowerOn
	ResetSoftware
	ResetWatchdog
	ResetBrownout
	ResetDeepSleep
	ResetDebug
	ResetExternal
	ResetOther
)

// String returns a short name for the reset cause.
func (c ResetCause) String() string {
	switch c {
	case ResetPowerOn:
		return "power on"
	case ResetSoftware:
		return "software"
	case ResetWatchdog:
		return "watchdog"
	case ResetBrownout:
		return "brownout"
	case ResetDeepSleep:
		return "deep sleep"
	case ResetDebug:
		return "debug"
	case ResetExternal:
		return "external"
	case ResetOther:
		return "other"
	default:
		return "unknown"
	}
}
