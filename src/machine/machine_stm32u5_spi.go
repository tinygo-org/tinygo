//go:build stm32u5

package machine

// SPI on STM32U5 uses the new SPIv2 peripheral with separate TXDR/RXDR registers.

import (
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

// Configure is intended to setup the STM32U5 SPI peripheral.
func (spi *SPI) Configure(config SPIConfig) error {
	// Disable SPI interface before any configuration changes
	spi.Bus.CR1.ClearBits(1) // Clear SPE (bit 0)

	// Enable clock for SPI
	enableAltFuncClock(unsafe.Pointer(spi.Bus))

	// Init pins - use defaults if not specified
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}
	spi.configurePins(config)

	// Configure CFG1: baud rate and data size
	var cfg1 uint32

	// Set baud rate (MBR bits [30:28])
	cfg1 |= spi.getBaudRate(config)

	// Set data size to 8 bits (DSIZE[4:0] = 7 = 8-1)
	cfg1 |= 7

	spi.Bus.CFG1.Set(cfg1)

	// Configure CFG2: master mode, SS output, polarity, phase
	var cfg2 uint32
	const (
		cfg2_MASTER  = 1 << 22 // MASTER bit
		cfg2_SSM     = 1 << 26 // Software SS management
		cfg2_SSOE    = 1 << 29 // SS output enable
		cfg2_AFCNTR  = 1 << 31 // Alternate function GPIO control always active
		cfg2_CPOL    = 1 << 25
		cfg2_CPHA    = 1 << 24
		cfg2_LSBFRST = 1 << 23
	)

	cfg2 |= cfg2_MASTER | cfg2_SSM | cfg2_AFCNTR

	// Set polarity and phase
	switch config.Mode {
	case Mode1:
		cfg2 |= cfg2_CPHA
	case Mode2:
		cfg2 |= cfg2_CPOL
	case Mode3:
		cfg2 |= cfg2_CPOL | cfg2_CPHA
	}

	// Set bit transfer order
	if config.LSBFirst {
		cfg2 |= cfg2_LSBFRST
	}

	spi.Bus.CFG2.Set(cfg2)

	// Enable SPI
	spi.Bus.CR1.SetBits(1) // Set SPE (bit 0)

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	// Set transfer size to 1 frame
	spi.Bus.CR2.Set(1)

	// Start the transfer
	spi.Bus.CR1.SetBits(1 << 9) // CSTART bit

	// Wait until TX FIFO is ready (TXP bit in SR)
	const sr_TXP = 1 << 1
	const sr_RXP = 1 << 0
	const sr_EOT = 1 << 3

	for !spi.Bus.SR.HasBits(sr_TXP) {
	}

	// Write byte to TXDR
	spi.Bus.TXDR.Set(uint32(w))

	// Wait for RX data available (RXP bit in SR)
	for !spi.Bus.SR.HasBits(sr_RXP) {
	}

	// Read received byte
	data := byte(spi.Bus.RXDR.Get())

	// Wait for end of transfer
	for !spi.Bus.SR.HasBits(sr_EOT) {
	}

	// Clear EOT flag
	spi.Bus.IFCR.Set(sr_EOT)

	return data, nil
}
