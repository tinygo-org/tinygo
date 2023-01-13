//go:build nrf52840 && !nrf52840_reset_uf2 && !nrf52840_reset_bossa

package machine

// EnterBootloader resets the chip into the serial bootloader.
func EnterBootloader() {
	// skip
}
