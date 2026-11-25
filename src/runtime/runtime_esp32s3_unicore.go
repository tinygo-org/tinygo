//go:build esp32s3 && !scheduler.cores

package runtime

import "runtime/interrupt"

func esp32s3EnterFlashSafeSection() (interrupt.State, bool) {
	return interrupt.Disable(), false
}

func esp32s3ExitFlashSafeSection(state interrupt.State, _ bool) {
	interrupt.Restore(state)
}
