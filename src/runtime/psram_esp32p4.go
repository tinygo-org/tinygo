//go:build esp32p4 && machine.psram

package runtime

// machineInitPSRAM is provided by package machine.
func machineInitPSRAM(pages uint32) bool

// initPSRAM brings up the PSRAM and maps the __psram_size bytes the target
// declares.
func initPSRAM() string {
	if !machineInitPSRAM(uint32(psramDeclaredSize() / 65536)) {
		return "chip init failed"
	}
	return ""
}
