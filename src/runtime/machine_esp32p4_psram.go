//go:build esp32p4 && machine.psram

package runtime

import "unsafe"

// machineInitPSRAM is provided by package machine.
func machineInitPSRAM(pages uint32) bool

// initPSRAM brings up the PSRAM and maps the __psram_size bytes the target
// declares.
func initPSRAM() string {
	size := uintptr(unsafe.Pointer(&_psram_end)) - uintptr(unsafe.Pointer(&_psram_start))
	if !machineInitPSRAM(uint32(size / 65536)) {
		return "PSRAM init failed"
	}
	return ""
}
