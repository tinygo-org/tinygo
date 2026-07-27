//go:build stm32 && stm32h7

package machine

// Peripheral abstraction layer for SPI on the stm32h7 family

import (
	"device/stm32"
	"errors"
	"math/bits"
	"runtime/volatile"
	"unsafe"
)

var errSPIOverrun = errors.New("SPI overrun or mode fault")

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the STM32 SPI interface.
func (spi *SPI) Configure(config SPIConfig) error {
	// disable SPI interface before any configuration changes
	spi.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)

	// enable clock for SPI
	enableAltFuncClock(unsafe.Pointer(spi.Bus))

	// init pins
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}
	spi.configurePins(config)

	// CFG1 configuration: MBR and DSIZE (8-bit)
	cfg1 := spi.getBaudRate(config)
	cfg1 |= (8 - 1) << stm32.SPI_CFG1_DSIZE_Pos // 8-bit data size
	spi.Bus.CFG1.Set(cfg1)

	// CFG2 configuration: CPOL, CPHA, MASTER, SSM, COMM
	var cfg2 uint32 = stm32.SPI_CFG2_MASTER // bit mask, not field value
	cfg2 |= stm32.SPI_CFG2_SSM              // software NSS; bit mask, not field value

	if config.LSBFirst {
		cfg2 |= 1 << 23 // LSBFRST is bit 23 in CFG2
	}

	// set polarity and phase
	switch config.Mode {
	case Mode1:
		cfg2 |= stm32.SPI_CFG2_CPHA_SecondEdge << stm32.SPI_CFG2_CPHA_Pos
	case Mode2:
		cfg2 |= stm32.SPI_CFG2_CPOL_IdleHigh << stm32.SPI_CFG2_CPOL_Pos
	case Mode3:
		cfg2 |= stm32.SPI_CFG2_CPOL_IdleHigh << stm32.SPI_CFG2_CPOL_Pos
		cfg2 |= stm32.SPI_CFG2_CPHA_SecondEdge << stm32.SPI_CFG2_CPHA_Pos
	}
	spi.Bus.CFG2.Set(cfg2)

	// CR2: TSIZE = 0 (Endless mode)
	spi.Bus.CR2.Set(0)

	// CR1: SPE and SSI (use bit masks, not field values)
	spi.Bus.CR1.Set(stm32.SPI_CR1_SSI | stm32.SPI_CR1_SPE)

	return nil
}

func (spi *SPI) config8Bits() {
	// Already handled in Configure via DSIZE
}

func (spi *SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

func (spi *SPI) getBaudRate(config SPIConfig) uint32 {
	clock := spi45KerFreq()
	if spi.Bus == stm32.SPI1 || spi.Bus == stm32.SPI2 || spi.Bus == stm32.SPI3 {
		clock = spi123KerFreq()
	} else if spi.Bus == stm32.SPI6 {
		clock = spi6KerFreq()
	}

	if config.Frequency == 0 {
		config.Frequency = clock / 2
	}

	// limit requested frequency to bus frequency and min frequency (DIV256)
	freq := config.Frequency
	if min := clock / 256; freq < min {
		freq = min
	} else if freq > clock/2 {
		freq = clock / 2
	}

	// Round up to the next power-of-two divisor so output never exceeds freq.
	// MBR encodes actual divider as 2^(MBR+1), so MBR = ceil_log2(ratio) - 1.
	div := bits.Len32(clock/freq-1) - 1
	if div > 7 {
		div = 7
	}

	return uint32(div) << stm32.SPI_CFG1_MBR_Pos
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	// RM0433 §50.4.9: set CSTART before writing TXDR.
	spi.Bus.CR1.SetBits(stm32.SPI_CR1_CSTART)

	// Wait for TXP (Transmit packet space available)
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_TXP) {
	}

	// Write to TXDR as 8-bit access to push exactly one byte into the FIFO.
	(*volatile.Register8)(unsafe.Pointer(&spi.Bus.TXDR.Reg)).Set(w)

	// Wait for RXP (Receive packet available)
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_RXP) {
	}

	// Check for overrun or mode fault before reading, to avoid returning stale data.
	if sr := spi.Bus.SR.Get(); sr&(stm32.SPI_SR_OVR|stm32.SPI_SR_MODF) != 0 {
		spi.Bus.IFCR.SetBits(stm32.SPI_IFCR_OVRC | stm32.SPI_IFCR_MODFC)
		return 0, errSPIOverrun
	}

	// Read from RXDR
	data := byte(spi.Bus.RXDR.Get())

	return data, nil
}
