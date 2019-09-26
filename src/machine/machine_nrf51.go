// +build nrf51

package machine

import (
	"device/nrf"
)

const CPU_FREQUENCY = 16000000

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	return nrf.GPIO, uint32(p)
}

func (uart UART) setPins(tx, rx Pin) {
	nrf.UART0.PSELTXD.Set(uint32(tx))
	nrf.UART0.PSELRXD.Set(uint32(rx))
}

//go:export UART0_IRQHandler
func handleUART0() {
	UART0.handleInterrupt()
}

func (i2c I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSELSCL.Set(uint32(scl))
	i2c.Bus.PSELSDA.Set(uint32(sda))
}

// SPI
func (spi SPI) setPins(sck, mosi, miso Pin) {
	spi.Bus.PSELSCK.Set(uint32(sck))
	spi.Bus.PSELMOSI.Set(uint32(mosi))
	spi.Bus.PSELMISO.Set(uint32(miso))
}
