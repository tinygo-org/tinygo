//go:build !scheduler.cores

package runtime

// Only core 0 runs without the cores scheduler.
func currentCPU() uint32 {
	return 0
}

// No other core can end a sleep without the cores scheduler.
func sleepWoken(core uint32) bool {
	return false
}

// Only the ESP interrupt.Enable calls this, through go:linkname, and only when
// currentCPU() is not 0, which does not occur without the cores scheduler.
func enableInterruptOnCore0(num uint32) {
}

func waitForSecondaryCoresReady() {}
