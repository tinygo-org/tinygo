// +build stm32f405

package machine

// Peripheral abstraction layer for the stm32f405

import (
	"device/stm32"
	"runtime/interrupt"
)

func CPUFrequency() uint32 {
	return 168000000
}

// -- UART ---------------------------------------------------------------------

type UART struct {
	Buffer          *RingBuffer
	Bus             *stm32.USART_Type
	Interrupt       interrupt.Interrupt
	AltFuncSelector stm32.AltFunc
}

func (uart UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.AltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.AltFuncSelector)
}

func (uart UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() / 2 // APB2 Frequency
	case stm32.USART2, stm32.USART3, stm32.UART4, stm32.UART5:
		clock = CPUFrequency() / 4 // APB1 Frequency
	}
	return clock / baudRate
}

// -- SPI ----------------------------------------------------------------------

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector stm32.AltFunc
}

func (spi SPI) configurePins(config SPIConfig)      {}
func (spi SPI) getBaudRate(config SPIConfig) uint32 { return 0 }

// -- I2C ----------------------------------------------------------------------

type I2C struct {
	Bus             *stm32.I2C_Type
	AltFuncSelector stm32.AltFunc
}

func (i2c I2C) configurePins(config I2CConfig) {
	config.SCL.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSCL}, i2c.AltFuncSelector)
	config.SDA.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSDA}, i2c.AltFuncSelector)
}

func (i2c I2C) getFreqRange(config I2CConfig) uint32 {
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

func (i2c I2C) getRiseTime(config I2CConfig) uint32 {
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

func (i2c I2C) getSpeed(config I2CConfig) uint32 {
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
