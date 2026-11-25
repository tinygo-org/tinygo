//go:build esp32 && machine.psram

package machine

import "errors"

var errPinUsedByPSRAM = errors.New("machine: pin is used by PSRAM")

// psramPin reports whether p is a PSRAM data, clock or select pin of the WROVER
// while PSRAM is on. See src/runtime/machine_esp32_psram_quad.go.
func psramPin(p Pin) bool {
	start, _ := runtimePSRAM() // not PSRAM(), which is also nil when .psram uses all of it
	return (p == 9 || p == 10 || p == 16 || p == 17) && start != 0
}
