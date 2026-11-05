//go:build esp32 && numa_psram_qspi

package machine

import (
	"errors"
	_ "unsafe"
)

var errPinUsedByPSRAM = errors.New("machine: pin is used by PSRAM")

//go:linkname psramActive runtime.psramActive
func psramActive() bool

// psramPin reports whether p is a PSRAM data, clock or select pin of the WROVER
// while PSRAM is on. See src/runtime/runtime_esp32_psram.go.
func psramPin(p Pin) bool {
	return (p == 9 || p == 10 || p == 16 || p == 17) && psramActive()
}
