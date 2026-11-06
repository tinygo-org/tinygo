//go:build esp32s3 && !scheduler.cores

package runtime

import "runtime/interrupt"

const numCPU = 1

//go:inline
func currentCPU() uint32 {
	return 0
}

//go:inline
func lockTicks() interrupt.State {
	return 0
}

//go:inline
func unlockTicks(interrupt.State) {
}

//go:inline
func sleepWoken(core uint32) bool {
	return false
}

func enableInterruptOnCore0(num uint32) {
}

func esp32s3EnterFlashSafeSection() (interrupt.State, bool) {
	return interrupt.Disable(), false
}

func esp32s3ExitFlashSafeSection(state interrupt.State, _ bool) {
	interrupt.Restore(state)
}
