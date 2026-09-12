//go:build rp2040 && !scheduler.cores

package runtime

import "runtime/interrupt"

func rp2040EnterFlashSafeSection() interrupt.State {
	return interrupt.Disable()
}

func rp2040ExitFlashSafeSection(state interrupt.State) {
	interrupt.Restore(state)
}

func rp2FlashSafeInterruptHandler(uint32) {
	// No-op on single-core schedulers.
}
