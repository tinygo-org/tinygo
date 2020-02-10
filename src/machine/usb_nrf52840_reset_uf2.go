// +build nrf52840,nrf52840_reset_uf2

package machine

import (
	"device/arm"
	"device/nrf"
)

const (
	DFU_MAGIC_SERIAL_ONLY_RESET = 0x4e
	DFU_MAGIC_UF2_RESET         = 0x57
	DFU_MAGIC_OTA_RESET         = 0xA8
)

func checkShouldReset() {
	if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
		EnterUF2Bootloader()
	}
}

// EnterSerialBootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware via serial/nrfutil
func EnterSerialBootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Reg = uint32(DFU_MAGIC_SERIAL_ONLY_RESET)
	arm.SystemReset()
}

// EnterUF2Bootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware via MSD/UF2
func EnterUF2Bootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Reg = uint32(DFU_MAGIC_UF2_RESET)
	arm.SystemReset()
}

// EnterOTABootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware via OTA update
func EnterOTABootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Reg = uint32(DFU_MAGIC_OTA_RESET)
	arm.SystemReset()
}
