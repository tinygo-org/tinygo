//go:build rp2350 && machine.psram

package runtime

// machineInitPSRAM is provided by package machine.
func machineInitPSRAM() bool

// initPSRAM maps the PSRAM. It must run after machineInit, which sets up the clocks.
func initPSRAM() string {
	if !machineInitPSRAM() {
		return "PSRAM init failed: no PSRAM found"
	}
	return ""
}
