//go:build nrf52840 && nrf52840_reset_uf2
// +build nrf52840,nrf52840_reset_uf2

package machine

// EnterBootloader resets the chip into the UF2 bootloader. After reset, it
// can be flashed via nrfutil or by copying a UF2 file to the mass storage device
func EnterBootloader() {
	EnterUF2Bootloader()
}
