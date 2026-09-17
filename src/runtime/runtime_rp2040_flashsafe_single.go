//go:build rp2040 && !scheduler.cores

package runtime

import "runtime/interrupt"

func rp2040EnterFlashSafeSection() (interrupt.State, bool) {
	return interrupt.Disable(), false
}

func rp2040ExitFlashSafeSection(state interrupt.State, _ bool) {
	interrupt.Restore(state)
}

func rp2FlashSafeInterruptHandler(uint32) {
	// No-op on single-core schedulers.
}
