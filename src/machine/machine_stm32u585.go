//go:build stm32u585

package machine

import (
	"device/stm32"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 160_000_000
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 160e6 // 160MHz (PLL1: MSIS 4MHz × 80 / 1 / 2)
const APB2_TIM_FREQ = 160e6 // 160MHz (PLL1: MSIS 4MHz × 80 / 1 / 2)

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	if uart.isLPUART1() {
		// LPUART1 is on APB3. Explicitly enable its peripheral clock.
		stm32.RCC.APB3ENR.SetBits(stm32.RCC_APB3ENR_LPUART1EN)
		_ = stm32.RCC.APB3ENR.Get() // delay for clock stabilization

		// Select PCLK3 as LPUART1 kernel clock source.
		stm32.RCC.CCIPR3.ReplaceBits(
			stm32.RCC_CCIPR3_LPUART1SEL_PCLK3<<stm32.RCC_CCIPR3_LPUART1SEL_Pos,
			stm32.RCC_CCIPR3_LPUART1SEL_Msk, 0)
	}

	if config.RX.getPort() == stm32.GPIOG || config.TX.getPort() == stm32.GPIOG {
		// Enable VDDIO2 voltage monitoring and wait for ready before
		// declaring VDDIO2 supply valid (matches HAL_PWREx_EnableVddIO2).
		stm32.PWR.SetSVMCR_IO2VMEN(1)
		for stm32.PWR.GetSVMSR_VDDIO2RDY() == 0 {
		}
		stm32.PWR.SetSVMCR_IO2SV(1)
	}

	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// isLPUART1 returns true if this UART is backed by the LPUART1 peripheral.
func (uart *UART) isLPUART1() bool {
	return uintptr(unsafe.Pointer(uart.Bus)) == uintptr(unsafe.Pointer(stm32.LPUART1))
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32u5.go clock init code
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	if uart.isLPUART1() {
		// LPUART uses BRR = 256 * fclk / baud.
		// Use 64-bit arithmetic to avoid overflow: at 160 MHz,
		// 256 * 160_000_000 = 40_960_000_000 which exceeds uint32 max.
		return uint32(uint64(256) * uint64(CPUFrequency()) / uint64(baudRate))
	}
	// USART requires BRR >= 16 for 16x oversampling (OVER8=0).
	// A divisor below 16 is invalid per the STM32 reference manual and causes
	// undefined hardware behaviour — in practice the receiver fires ORE/RXNE
	// interrupts at an impossible rate, completely starving the CPU.
	const minBRR = 16
	divisor := CPUFrequency() / baudRate
	if divisor < minBRR {
		divisor = minBRR
	}
	return divisor
}

// Register names vary by ST processor, these are for STM U5
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
	uart.errClearReg = &uart.Bus.ICR
}

// SetBaudRate overrides the shared implementation for STM32U5. On this
// family the BRR register is read-only while UE=1 (USART enabled), so the
// USART must be briefly disabled to change the baud rate. This matters when
// the servo library (or any code) calls SetBaudRate after Configure has
// already enabled the USART.
func (uart *UART) SetBaudRate(br uint32) {
	cr1 := uart.Bus.CR1.Get()
	if cr1&stm32.USART_CR1_UE != 0 {
		// Disable the USART so BRR becomes writable.
		uart.Bus.CR1.Set(cr1 &^ stm32.USART_CR1_UE)
	}
	uart.Bus.BRR.Set(uart.getBaudRateDivisor(br))
	if cr1&stm32.USART_CR1_UE != 0 {
		// Restore CR1 exactly as it was (re-enables USART, TE, RE, etc.).
		uart.Bus.CR1.Set(cr1)
	}
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
