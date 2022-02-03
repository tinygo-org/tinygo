//go:build nrf52840 && nrf52840_reset_bossa
// +build nrf52840,nrf52840_reset_bossa

package machine

import (
	"device/arm"
	"device/nrf"
)

const DFU_MAGIC_SERIAL_ONLY_RESET = 0xb0

// checkShouldReset is called by the USB-CDC implementation to check whether to
// reset into the bootloader/OTA and if so, resets the chip appropriately.
func checkShouldReset() {
	if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
		EnterSerialBootloader()
	}
}

// EnterSerialBootloader resets the chip into the serial bootloader. After
// reset, it can be flashed using serial/nrfutil.
func EnterSerialBootloader() {
	arm.DisableInterrupts()
	nrf.POWER.GPREGRET.Set(DFU_MAGIC_SERIAL_ONLY_RESET)
	arm.SystemReset()
}
