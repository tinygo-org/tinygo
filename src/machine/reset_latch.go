//go:build nrf || stm32 || py32 || mimxrt1062 || ra4m1

package machine

var (
	resetFlags     uint32
	resetFlagsRead bool
)

// ResetReasonRaw returns the chip specific reset flags.
//
// The flags add up over resets until cleared. The first call reads and clears
// them, and later calls return the same value. Call it early in main so that
// the next reset starts from a clean state.
func ResetReasonRaw() uint32 {
	if !resetFlagsRead {
		resetFlags = readAndClearResetFlags()
		resetFlagsRead = true
	}
	return resetFlags
}
