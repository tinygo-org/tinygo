//go:build esp32 && !numa_psram_qspi

package machine

var errPinUsedByPSRAM error

// PSRAM returns nil: only the esp32-psram target switches the PSRAM on.
func PSRAM() []byte {
	return nil
}

func psramPin(p Pin) bool {
	return false
}
