// +build nrf52840

package machine

import (
	"device/nrf"
)

// Get peripheral and pin number for this GPIO pin.
func (p GPIO) getPortPin() (*nrf.GPIO_Type, uint8) {
	if p.Pin >= 32 {
		return nrf.P1, p.Pin - 32
	} else {
		return nrf.P0, p.Pin
	}
}

func (uart UART) setPins(tx, rx uint32) {
	nrf.UART0.PSEL.TXD = nrf.RegValue(tx)
	nrf.UART0.PSEL.RXD = nrf.RegValue(rx)
}

//go:export UARTE0_UART0_IRQHandler
func handleUART0() {
	UART0.handleInterrupt()
}

func (i2c I2C) setPins(scl, sda uint8) {
	i2c.Bus.PSEL.SCL = nrf.RegValue(scl)
	i2c.Bus.PSEL.SDA = nrf.RegValue(sda)
}

// SPI
func (spi SPI) setPins(sck, mosi, miso uint8) {
	if sck == 0 {
		sck = SPI0_SCK_PIN
	}
	if mosi == 0 {
		mosi = SPI0_MOSI_PIN
	}
	if miso == 0 {
		miso = SPI0_MISO_PIN
	}
	spi.Bus.PSEL.SCK = nrf.RegValue(sck)
	spi.Bus.PSEL.MOSI = nrf.RegValue(mosi)
	spi.Bus.PSEL.MISO = nrf.RegValue(miso)
}
