// +build stm32

package machine

// Peripheral abstraction layer for SPI on the stm32 family

import (
	"device/stm32"
	"unsafe"
)

// SPI on the STM32.
type SPI struct {
	Bus *stm32.SPI_Type
}

// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
// TODO: implement SPI2 and SPI3.
var (
	SPI0 = SPI{
		Bus: stm32.SPI1,
	}
	SPI1 = &SPI0
)

// SPI phase and polarity configs CPOL and CPHA
const (
	SPI_ClkLowValid1stEdge  = 0
	SPI_ClkLowValid2ndEdge  = 1
	SPI_ClkHighValid1stEdge = 2
	SPI_ClkHighValid2ndEdge = 3
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
	EnableAltFuncClock(unsafe.Pointer(spi.Bus))

	var conf uint32 = spi.getBaudRate(config)

	// set bit transfer order
	if config.LSBFirst {
		conf |= stm32.SPI_CR1_LSBFIRST
	}

	// set polarity and phase on the SPI interface
	switch config.Mode {
	case SPI_ClkLowValid1stEdge:
		conf &^= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf &^= (1 << stm32.SPI_CR1_CPHA_Pos)
	case SPI_ClkLowValid2ndEdge:
		conf &^= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf |= (1 << stm32.SPI_CR1_CPHA_Pos)
	case SPI_ClkHighValid1stEdge:
		conf |= (1 << stm32.SPI_CR1_CPOL_Pos)
		conf &^= (1 << stm32.SPI_CR1_CPHA_Pos)
	case SPI_ClkHighValid2ndEdge:
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
	if config.SCK == 0 {
		config.SCK = SPI0_SCK_PIN
	}
	if config.MOSI == 0 {
		config.MOSI = SPI0_MOSI_PIN
	}
	if config.MISO == 0 {
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
