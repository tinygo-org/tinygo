//go:build stm32u585

package machine

import (
	"device/stm32"
)

func CPUFrequency() uint32 {
	return 4_000_000
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 4e6 // 4MHz (MSI default)
const APB2_TIM_FREQ = 4e6 // 4MHz (MSI default)

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	if config.RX.getPort() == stm32.GPIOG || config.TX.getPort() == stm32.GPIOG {
		// Enable VDDIO2 power supply for PGx pins
		stm32.PWR.SetSVMCR_IO2SV(1)
	}

	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32u5.go clock init code
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	return CPUFrequency() / baudRate
}

// Register names vary by ST processor, these are for STM U5
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
}

//---------- SPI related types and code

// SPI on the STM32U5 using the new SPIv2 peripheral
type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi *SPI) config8Bits() {
	// U5 SPI has DSIZE field in CFG1, set to 7 for 8-bit frames (DSIZE = bits-1)
	spi.Bus.CFG1.ReplaceBits(7, 0x1f, 0) // DSIZE[4:0] = 0x7 = 8 bits
}

// Set baud rate for SPI
func (spi *SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	localFrequency := config.Frequency

	// Default
	if localFrequency == 0 {
		localFrequency = 4e6
	}

	// Set frequency dependent on PCLK prescaler
	// MBR field in CFG1 register, bits [30:28]
	switch {
	case localFrequency < 625000:
		conf = 7 // Div256
	case localFrequency < 1250000:
		conf = 6 // Div128
	case localFrequency < 2500000:
		conf = 5 // Div64
	case localFrequency < 5000000:
		conf = 4 // Div32
	case localFrequency < 10000000:
		conf = 3 // Div16
	case localFrequency < 20000000:
		conf = 2 // Div8
	case localFrequency < 40000000:
		conf = 1 // Div4
	case localFrequency < 80000000:
		conf = 0 // Div2
	default:
		conf = 7 // Div256 (safest)
	}

	return conf << 28 // MBR position in CFG1
}

// Configure SPI pins for input output and clock
func (spi *SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

//---------- I2C related code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange(br uint32) uint32 {
	// These are 'magic' values calculated by STM32CubeMX
	// for 160MHz PCLK1.
	// TODO: Do calculations based on PCLK1
	switch br {
	case 10 * KHz:
		return 0xF010F3FE
	case 100 * KHz:
		return 0x30A0A7FB
	case 400 * KHz:
		return 0x10802D9B
	case 500 * KHz:
		return 0x00802172
	default:
		return 0
	}
}
