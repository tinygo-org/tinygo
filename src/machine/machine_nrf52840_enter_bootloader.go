//go:build nrf52840
// +build nrf52840

package machine

import (
	"tinygo.org/x/device/arm"
	"tinygo.org/x/device/nrf"
)

const (
	DFU_MAGIC_SERIAL_ONLY_RESET = 0x4e
	DFU_MAGIC_UF2_RESET         = 0x57
	DFU_MAGIC_OTA_RESET         = 0xA8
)

// EnterSerialBootloader resets the chip into the serial bootloader. After
// reset, it can be flashed using serial/nrfutil.
func EnterSerialBootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(DFU_MAGIC_SERIAL_ONLY_RESET)
	arm.SystemReset()
}

// EnterUF2Bootloader resets the chip into the UF2 bootloader. After reset, it
// can be flashed via nrfutil or by copying a UF2 file to the mass storage device
func EnterUF2Bootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(DFU_MAGIC_UF2_RESET)
	arm.SystemReset()
}

// EnterOTABootloader resets the chip into the bootloader so that it can be
// flashed via an OTA update
func EnterOTABootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(DFU_MAGIC_OTA_RESET)
	arm.SystemReset()
}
