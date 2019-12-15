// +build nrf51

package machine

import (
	"device/nrf"
)

func CPUFrequency() uint32 {
	return 16000000
}

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	return nrf.GPIO, uint32(p)
}

func (uart UART) setPins(tx, rx Pin) {
	nrf.UART0.PSELTXD.Set(uint32(tx))
	nrf.UART0.PSELRXD.Set(uint32(rx))
}

func (i2c I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSELSCL.Set(uint32(scl))
	i2c.Bus.PSELSDA.Set(uint32(sda))
}

// SPI
func (spi SPI) setPins(sck, mosi, miso Pin) {
	if sck == 0 {
		sck = SPI0_SCK_PIN
	}
	if mosi == 0 {
		mosi = SPI0_MOSI_PIN
	}
	if miso == 0 {
		miso = SPI0_MISO_PIN
	}
	spi.Bus.PSELSCK.Set(uint32(sck))
	spi.Bus.PSELMOSI.Set(uint32(mosi))
	spi.Bus.PSELMISO.Set(uint32(miso))
}
