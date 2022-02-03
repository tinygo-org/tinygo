//go:build nrf52840 && !nrf52840_reset_uf2 && !nrf52840_reset_bossa
// +build nrf52840,!nrf52840_reset_uf2,!nrf52840_reset_bossa

package machine

func checkShouldReset() {
}
