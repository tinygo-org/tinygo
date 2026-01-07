//go:build stm32g0

package machine

// SPI on STM32G0 uses 16-bit registers

import (
	"device/stm32"
	"runtime/volatile"
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
	// disable SPI interface before any configuration changes
	spi.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)

	// enable clock for SPI
	enableAltFuncClock(unsafe.Pointer(spi.Bus))

	// init pins - use defaults if not specified
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}
	spi.configurePins(config)

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

	// use software CS (GPIO) by default
	conf |= stm32.SPI_CR1_SSM

	// Set CR1 configuration WITHOUT enabling SPE yet
	// (STM32G0 requires CR2 DS bits to be set before SPE is enabled)
	spi.Bus.CR1.Set(uint16(conf))

	// Series-specific configuration to set 8-bit transfer mode (must be done before SPE)
	spi.config8Bits()

	// Now enable SPI
	spi.Bus.SetCR1_SPE(1)

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	// STM32G0 requires 8-bit access to DR for 8-bit transfers
	// Using 16-bit access causes data packing issues
	dr := (*volatile.Register8)(unsafe.Pointer(&spi.Bus.DR))

	// Write data to be transmitted to the SPI data register (8-bit access)
	dr.Set(w)

	// Wait until transmit complete
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_TXE) {
	}

	// Wait until receive complete
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_RXNE) {
	}

	// Wait until SPI is not busy
	for spi.Bus.SR.HasBits(stm32.SPI_SR_BSY) {
	}

	// Return received data from SPI data register (8-bit access)
	return dr.Get(), nil
}
