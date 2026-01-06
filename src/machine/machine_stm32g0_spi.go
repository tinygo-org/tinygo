//go:build stm32g0

package machine

// SPI on STM32G0 uses 16-bit registers

import (
	"device/stm32"
	"unsafe"
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the STM32 SPI peripheral
func (spi *SPI) Configure(config SPIConfig) error {
	// enable clock for SPI
	enableAltFuncClock(unsafe.Pointer(spi.Bus))

	// Get SPI baud rate divisor
	conf := spi.getBaudRate(config)

	// set polarity and phase on the SPI interface
	switch config.Mode {
	case Mode1:
		conf |= stm32.SPI_CR1_CPHA
	case Mode2:
		conf |= stm32.SPI_CR1_CPOL
	case Mode3:
		conf |= stm32.SPI_CR1_CPOL | stm32.SPI_CR1_CPHA
	}

	// set bit transfer order
	if config.LSBFirst {
		conf |= stm32.SPI_CR1_LSBFIRST
	}

	// set SPI master
	conf |= stm32.SPI_CR1_MSTR | stm32.SPI_CR1_SSI

	// enable the SPI interface
	conf |= stm32.SPI_CR1_SPE

	// use software CS (GPIO) by default
	conf |= stm32.SPI_CR1_SSM

	// now set the configuration (note: STM32G0 uses 16-bit SPI registers)
	spi.Bus.CR1.Set(uint16(conf))

	// Series-specific configuration to set 8-bit transfer mode
	spi.config8Bits()

	// enable SPI
	spi.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	// Write data to be transmitted to the SPI data register
	spi.Bus.DR.Set(uint16(w))

	// Wait until transmit complete
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_TXE) {
	}

	// Wait until receive complete
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_RXNE) {
	}

	// Wait until SPI is not busy
	for spi.Bus.SR.HasBits(stm32.SPI_SR_BSY) {
	}

	// Return received data from SPI data register
	return byte(spi.Bus.DR.Get()), nil
}
