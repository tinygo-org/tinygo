//go:build (esp32c3 || esp32c6 || esp32h2 || esp32p4) && !scheduler.cores

package runtime

import "runtime/interrupt"

//go:inline
func lockTicks() interrupt.State {
	return 0
}

//go:inline
func unlockTicks(interrupt.State) {
}
