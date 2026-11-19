//go:build !machine.psram

package runtime

// initPSRAM is a no-op when PSRAM support is disabled.
func initPSRAM() string { return "" }

const hasPSRAM = false
