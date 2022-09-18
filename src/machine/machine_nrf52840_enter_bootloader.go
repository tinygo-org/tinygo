//go:build nrf52840
// +build nrf52840

package machine

import (
	"device/arm"
	"device/nrf"
)

const (
	dfuMagicSerialOnlyReset = 0x4e
	dfuMagicUF2Reset        = 0x57
	dfuMagicOTAReset        = 0xA8
)

// EnterSerialBootloader resets the chip into the serial bootloader. After
// reset, it can be flashed using serial/nrfutil.
func EnterSerialBootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(dfuMagicSerialOnlyReset)
	arm.SystemReset()
}

// EnterUF2Bootloader resets the chip into the UF2 bootloader. After reset, it
// can be flashed via nrfutil or by copying a UF2 file to the mass storage device
func EnterUF2Bootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(dfuMagicUF2Reset)
	arm.SystemReset()
}

// EnterOTABootloader resets the chip into the bootloader so that it can be
// flashed via an OTA update
func EnterOTABootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(dfuMagicOTAReset)
	arm.SystemReset()
}
