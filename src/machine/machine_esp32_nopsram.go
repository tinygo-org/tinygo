//go:build esp32 && !machine.psram

package machine

var errPinUsedByPSRAM error

func psramPin(p Pin) bool {
	return false
}
