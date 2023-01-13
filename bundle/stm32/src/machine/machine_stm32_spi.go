//go:build stm32 && !stm32f7x2 && !stm32l5x2

package machine

// Peripheral abstraction layer for SPI on the stm32 family

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

// Configure is intended to setup the STM32 SPI1 interface.
func (spi SPI) Configure(config SPIConfig) {

	// -- CONFIGURING THE SPI IN MASTER MODE --
	//
	// 1. Select the BR[2:0] bits to define the serial clock baud rate (see
	// 	  SPI_CR1 register).
	// 2. Select the CPOL and CPHA bits to define one of the four relationships
	//    between the data transfer and the serial clock (see Figure 248). This
	//    step is not required when the TI mode is selected.
	// 3. Set the DFF bit to define 8- or 16-bit data frame format
	// 4. Configure the LSBFIRST bit in the SPI_CR1 register to define the frame
	//    format. This step is not required when the TI mode is selected.
	// 5. If the NSS pin is required in input mode, in hardware mode, connect the
	//    NSS pin to a high-level signal during the complete byte transmit
	//    sequence. In NSS software mode, set the SSM and SSI bits in the SPI_CR1
	//    register. If the NSS pin is required in output mode, the SSOE bit only
	//    should be set. This step is not required when the TI mode is selected.
	// 6. Set the FRF bit in SPI_CR2 to select the TI protocol for serial
	//    communications.
	// 7. The MSTR and SPE bits must be set (they remain set only if the NSS pin
	//    is connected to a high-level signal).

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

	// Get SPI baud rate based on the bus speed it's attached to
	var conf uint32 = spi.getBaudRate(config)

	// set bit transfer order
	if config.LSBFirst {
		conf |= stm32.SPI_CR1_LSBFIRST
	}

	// set polarity and phase on the SPI interface
	switch config.Mode {
	case Mode1:
		conf |= stm32.SPI_CR1_CPHA
	case Mode2:
		conf |= stm32.SPI_CR1_CPOL
	case Mode3:
		conf |= stm32.SPI_CR1_CPOL
		conf |= stm32.SPI_CR1_CPHA
	}

	// configure as SPI master
	conf |= stm32.SPI_CR1_MSTR | stm32.SPI_CR1_SSI

	// enable the SPI interface
	conf |= stm32.SPI_CR1_SPE

	// use software CS (GPIO) by default
	conf |= stm32.SPI_CR1_SSM

	// now set the configuration
	spi.Bus.CR1.Set(conf)

	// Series-specific configuration to set 8-bit transfer mode
	spi.config8Bits()

	// enable SPI
	spi.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {

	// 1. Enable the SPI by setting the SPE bit to 1.
	// 2. Write the first data item to be transmitted into the SPI_DR register
	//    (this clears the TXE flag).
	// 3. Wait until TXE=1 and write the second data item to be transmitted. Then
	//    wait until RXNE=1 and read the SPI_DR to get the first received data
	//    item (this clears the RXNE bit). Repeat this operation for each data
	//    item to be transmitted/received until the n–1 received data.
	// 4. Wait until RXNE=1 and read the last received data.
	// 5. Wait until TXE=1 and then wait until BSY=0 before disabling the SPI.

	// put output word (8-bit) in data register (DR), which is parallel-loaded
	// into shift register, and shifted out on MOSI.  Some series have 16-bit
	// register but writes must be strictly 8-bit to output a byte.  Writing
	// 16-bits indicates a packed transfer (2 bytes).
	(*volatile.Register8)(unsafe.Pointer(&spi.Bus.DR.Reg)).Set(w)

	// wait for SPI bus receive buffer not empty bit (RXNE) to be set.
	// warning: blocks forever until this condition is met.
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_RXNE) {
	}

	// copy input word (8-bit) in data register (DR), which was shifted in on MISO
	// and parallel-loaded into register.
	data := byte(spi.Bus.DR.Get())

	// wait for SPI bus transmit buffer empty bit (TXE) to be set.
	// warning: blocks forever until this condition is met.
	for !spi.Bus.SR.HasBits(stm32.SPI_SR_TXE) {
	}

	// wait for SPI bus busy bit (BSY) to be clear to indicate synchronous
	// transfer complete. this will effectively prevent this Transfer() function
	// from being capable of maintaining high-bandwidth communication throughput,
	// but it will help guarantee stability on the bus.
	for spi.Bus.SR.HasBits(stm32.SPI_SR_BSY) {
	}

	// clear the overrun flag (only in full-duplex mode)
	if !spi.Bus.CR1.HasBits(stm32.SPI_CR1_RXONLY | stm32.SPI_CR1_BIDIMODE | stm32.SPI_CR1_BIDIOE) {
		spi.Bus.SR.Get()
	}

	// Return received data from SPI data register
	return data, nil
}
