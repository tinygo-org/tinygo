//go:build esp32p4 && !scheduler.cores

package runtime

const numCPU = 1

func currentCPU() uint32 {
	return 0
}

func sleepWoken(core uint32) bool {
	return false
}

func enableInterruptOnCore0(num uint32) {
}
