//go:build nrf52840 && nrf52840_reset_bossa
// +build nrf52840,nrf52840_reset_bossa

package machine

// EnterBootloader resets the chip into the serial bootloader. After
// reset, it can be flashed using serial/nrfutil.
func EnterBootloader() {
	EnterSerialBootloader()
}
