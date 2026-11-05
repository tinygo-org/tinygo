//go:build esp32 && !numa_psram_qspi

package machine

var errPinUsedByPSRAM error

func psramPin(p Pin) bool {
	return false
}
