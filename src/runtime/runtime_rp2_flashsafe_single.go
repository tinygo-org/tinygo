//go:build (rp2040 || rp2350) && !scheduler.cores

package runtime

import "runtime/interrupt"

func rp2EnterFlashSafeSection() (interrupt.State, bool) {
	return interrupt.Disable(), false
}

func rp2ExitFlashSafeSection(state interrupt.State, _ bool) {
	interrupt.Restore(state)
}

func rp2FlashSafeInterruptHandler() {
	// No-op on single-core schedulers.
}
