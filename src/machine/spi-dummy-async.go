//go:build !baremetal || atmega || esp32 || fe310 || k210 || nrf || (nxp && !mk66f18) || atsamd21 || (stm32 && !stm32f7x2 && !stm32l5x2)

package machine

// This is a non-async implementation of the async SPI calls.
// It is useful for devices that don't support DMA on SPI, or for which it
// hasn't been implemented yet.

// IsAsync returns whether the SPI supports async operation (usually DMA).
//
// This SPI does not support async operations.
func (s SPI) IsAsync() bool {
	return false
}

// Start a transfer in the background.
//
// Because this SPI implementation doesn't support async operation, it is an
// alias for Tx.
func (s SPI) StartTx(tx, rx []byte) error {
	return s.Tx(tx, rx)
}

// Wait until all active transactions (started by StartTx) have finished.
//
// This is a no-op on this SPI implementation.
func (s SPI) Wait() error {
	return nil
}
