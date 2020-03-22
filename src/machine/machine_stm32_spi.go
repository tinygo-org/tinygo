// +build stm32

package machine

// Peripheral abstraction layer for SPI on the stm32 family

import (
	"device/stm32"
	"unsafe"
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	MOSI      Pin
	MISO      Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the STM32 SPI1 interface.
// Features still TODO:
// - support SPI2 and SPI3
// - allow setting data size to 16 bits?
// - allow setting direction in HW for additional optimization?
// - hardware SS pin?
func (spi SPI) Configure(config SPIConfig) {
	// enable clock for SPI
	enableAltFuncClock(unsafe.Pointer(spi.Bus))

	// Get SPI baud rate based on the bus speed it's attached to
	var conf uint32 = spi.getBaudRate(config)

	// set bit transfer order
	if config.LSBFirst {
		conf |= stm32.SPI_CR1_LSBFIRST
	}

	// set polarity and phase on the SPI interface
	switch config.Mode {
	case Mode0:
		conf &^= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf &^= (1 << stm32.SPI_CR1_CPHA_Pos)
	case Mode1:
		conf &^= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf |= (1 << stm32.SPI_CR1_CPHA_Pos)
	case Mode2:
		conf |= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf &^= (1 << stm32.SPI_CR1_CPHA_Pos)
	case Mode3:
		conf |= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf |= (1 << stm32.SPI_CR1_CPHA_Pos)
	default: // to mode 0
		conf &^= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf &^= (1 << stm32.SPI_CR1_CPHA_Pos)
	}

	// set to SPI master
	conf |= stm32.SPI_CR1_MSTR

	// disable MCU acting as SPI slave
	conf |= stm32.SPI_CR1_SSM | stm32.SPI_CR1_SSI

	// now set the configuration
	spi.Bus.CR1.Set(conf)

	// init pins
	if config.SCK == 0 && config.MOSI == 0 && config.MISO == 0 {
		config.SCK = SPI0_SCK_PIN
		config.MOSI = SPI0_MOSI_PIN
		config.MISO = SPI0_MISO_PIN
	}
	spi.configurePins(config)

	// enable SPI interface
	spi.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	// Write data to be transmitted to the SPI data register
	spi.Bus.DR.Set(uint32(w))

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
