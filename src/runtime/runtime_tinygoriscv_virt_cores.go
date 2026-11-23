//go:build tinygo.riscv && virt && scheduler.cores

package runtime

// Spin until the time is up, or until another hart ends the sleep. esp32p4
// sleeps the same way.
func sleepTicks(d timeUnit) {
	core := currentCPU()
	target := ticks() + d
	for ticks() < target && !sleepWoken(core) {
	}
}
