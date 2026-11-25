//go:build (esp32c3 || esp32c6 || esp32h2 || esp32p4) && !scheduler.cores

package runtime

import "runtime/interrupt"

func lockTicks() interrupt.State {
	return 0
}

func unlockTicks(interrupt.State) {
}
