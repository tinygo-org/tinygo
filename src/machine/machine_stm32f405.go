// +build stm32f405

package machine

// Peripheral abstraction layer for the stm32f405

import (
	"device/stm32"
	"math/bits"
)

func CPUFrequency() uint32 {
	return 168000000
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 42000000 * 2
const APB2_TIM_FREQ = 84000000 * 2

// Alternative peripheral pin functions
const (
	AF0_SYSTEM                = 0
	AF1_TIM1_2                = 1
	AF2_TIM3_4_5              = 2
	AF3_TIM8_9_10_11          = 3
	AF4_I2C1_2_3              = 4
	AF5_SPI1_SPI2             = 5
	AF6_SPI3                  = 6
	AF7_USART1_2_3            = 7
	AF8_USART4_5_6            = 8
	AF9_CAN1_CAN2_TIM12_13_14 = 9
	AF10_OTG_FS_OTG_HS        = 10
	AF11_ETH                  = 11
	AF12_FSMC_SDIO_OTG_HS_1   = 12
	AF13_DCMI                 = 13
	AF14                      = 14
	AF15_EVENTOUT             = 15
)

// -- UART ---------------------------------------------------------------------

func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() / 2 // APB2 Frequency
	case stm32.USART2, stm32.USART3, stm32.UART4, stm32.UART5:
		clock = CPUFrequency() / 4 // APB1 Frequency
	}
	return clock / baudRate
}

// Register names vary by ST processor, these are for STM F405
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.DR.Register32
	uart.txReg = &uart.Bus.DR.Register32
	uart.statusReg = &uart.Bus.SR.Register32
	uart.txEmptyFlag = stm32.USART_SR_TXE
}

// -- SPI ----------------------------------------------------------------------

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi SPI) config8Bits() {
	// no-op on this series
}

func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var clock uint32
	switch spi.Bus {
	case stm32.SPI1:
		clock = CPUFrequency() / 2
	case stm32.SPI2, stm32.SPI3:
		clock = CPUFrequency() / 4
	}

	// limit requested frequency to bus frequency and min frequency (DIV256)
	freq := config.Frequency
	if min := clock / 256; freq < min {
		freq = min
	} else if freq > clock {
		freq = clock
	}

	// calculate the exact clock divisor (freq=clock/div -> div=clock/freq).
	// truncation is fine, since it produces a less-than-or-equal divisor, and
	// thus a greater-than-or-equal frequency.
	// divisors only come in consecutive powers of 2, so we can use log2 (or,
	// equivalently, bits.Len - 1) to convert to respective enum value.
	div := bits.Len32(clock/freq) - 1

	// but DIV1 (2^0) is not permitted, as the least divisor is DIV2 (2^1), so
	// subtract 1 from the log2 value, keeping a lower bound of 0
	if div < 0 {
		div = 0
	} else if div > 0 {
		div--
	}

	// finally, shift the enumerated value into position for SPI CR1
	return uint32(div) << stm32.SPI_CR1_BR_Pos
}

// -- I2C ----------------------------------------------------------------------

type I2C struct {
	Bus             *stm32.I2C_Type
	AltFuncSelector uint8
}

func (i2c *I2C) configurePins(config I2CConfig) {
	config.SCL.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSCL}, i2c.AltFuncSelector)
	config.SDA.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSDA}, i2c.AltFuncSelector)
}

func (i2c *I2C) getFreqRange(config I2CConfig) uint32 {
	// all I2C interfaces are on APB1 (42 MHz)
	clock := CPUFrequency() / 4
	// convert to MHz
	clock /= 1000000
	// must be between 2 MHz (or 4 MHz for fast mode (Fm)) and 50 MHz, inclusive
	var min, max uint32 = 2, 50
	if config.Frequency > 10000 {
		min = 4 // fast mode (Fm)
	}
	if clock < min {
		clock = min
	} else if clock > max {
		clock = max
	}
	return clock << stm32.I2C_CR2_FREQ_Pos
}

func (i2c *I2C) getRiseTime(config I2CConfig) uint32 {
	// These bits must be programmed with the maximum SCL rise time given in the
	// I2C bus specification, incremented by 1.
	// For instance: in Sm mode, the maximum allowed SCL rise time is 1000 ns.
	// If, in the I2C_CR2 register, the value of FREQ[5:0] bits is equal to 0x08
	// and PCLK1 = 125 ns, therefore the TRISE[5:0] bits must be programmed with
	// 09h (1000 ns / 125 ns = 8 + 1)
	freqRange := i2c.getFreqRange(config)
	if config.Frequency > 100000 {
		// fast mode (Fm) adjustment
		freqRange *= 300
		freqRange /= 1000
	}
	return (freqRange + 1) << stm32.I2C_TRISE_TRISE_Pos
}

func (i2c *I2C) getSpeed(config I2CConfig) uint32 {
	ccr := func(pclk uint32, freq uint32, coeff uint32) uint32 {
		return (((pclk - 1) / (freq * coeff)) + 1) & stm32.I2C_CCR_CCR_Msk
	}
	sm := func(pclk uint32, freq uint32) uint32 { // standard mode (Sm)
		if s := ccr(pclk, freq, 2); s < 4 {
			return 4
		} else {
			return s
		}
	}
	fm := func(pclk uint32, freq uint32, duty uint8) uint32 { // fast mode (Fm)
		if duty == DutyCycle2 {
			return ccr(pclk, freq, 3)
		} else {
			return ccr(pclk, freq, 25) | stm32.I2C_CCR_DUTY
		}
	}
	// all I2C interfaces are on APB1 (42 MHz)
	clock := CPUFrequency() / 4
	if config.Frequency <= 100000 {
		return sm(clock, config.Frequency)
	} else {
		s := fm(clock, config.Frequency, config.DutyCycle)
		if (s & stm32.I2C_CCR_CCR_Msk) == 0 {
			return 1
		} else {
			return s | stm32.I2C_CCR_F_S
		}
	}
}
