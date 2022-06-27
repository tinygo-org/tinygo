//go:build nrf52840 && nrf52840_reset_bossa
// +build nrf52840,nrf52840_reset_bossa

package machine

import (
	"device/arm"
	"device/nrf"
)

const DFU_MAGIC_SERIAL_ONLY_RESET = 0xb0

// ResetProcessor resets the chip into the serial bootloader. After
// reset, it can be flashed using serial/nrfutil.
func ResetProcessor() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(DFU_MAGIC_SERIAL_ONLY_RESET)
	arm.SystemReset()
}
