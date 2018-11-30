// +build nrf51

package machine

import (
	"device/nrf"
)

// Get peripheral and pin number for this GPIO pin.
func (p GPIO) getPortPin() (*nrf.GPIO_Type, uint8) {
	return nrf.GPIO, p.Pin
}

func (uart UART) setPins(tx, rx uint32) {
	nrf.UART0.PSELTXD = nrf.RegValue(tx)
	nrf.UART0.PSELRXD = nrf.RegValue(rx)
}

//go:export UART0_IRQHandler
func handleUART0() {
	UART0.handleInterrupt()
}

func (i2c I2C) setPins(scl, sda uint8) {
	i2c.Bus.PSELSCL = nrf.RegValue(scl)
	i2c.Bus.PSELSDA = nrf.RegValue(sda)
}

// SPI
func (spi SPI) setPins(sck, mosi, miso uint8) {
	spi.Bus.PSELSCK = nrf.RegValue(sck)
	spi.Bus.PSELMOSI = nrf.RegValue(mosi)
	spi.Bus.PSELMISO = nrf.RegValue(miso)
}
