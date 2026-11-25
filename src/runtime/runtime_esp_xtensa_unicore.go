//go:build (esp32 || esp32s3) && !scheduler.cores

package runtime

import "runtime/interrupt"

const numCPU = 1

func currentCPU() uint32 {
	return 0
}

func lockTicks() interrupt.State {
	return 0
}

func unlockTicks(interrupt.State) {
}

func sleepWoken(core uint32) bool {
	return false
}

func enableInterruptOnCore0(num uint32) {
}
