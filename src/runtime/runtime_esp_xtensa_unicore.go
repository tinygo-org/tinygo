//go:build (esp32 || esp32s3) && !scheduler.cores

package runtime

import "runtime/interrupt"

const numCPU = 1

func lockTicks() interrupt.State {
	return 0
}

func unlockTicks(interrupt.State) {
}
