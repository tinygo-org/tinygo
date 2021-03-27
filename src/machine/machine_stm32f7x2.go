// +build stm32f7x2

package machine

// Peripheral abstraction layer for the stm32f407

import (
	"device/stm32"
)

func CPUFrequency() uint32 {
	return 216000000
}

//---------- Clock and Oscillator related code

/*
   clock settings
   +-------------+--------+
   | HSE         | 8mhz   |
   | SYSCLK      | 216mhz |
   | HCLK        | 216mhz |
   | APB1(PCLK1) | 27mhz  |
   | APB2(PCLK2) | 108mhz |
   +-------------+--------+
*/
const (
	HSE_STARTUP_TIMEOUT = 0x0500
	PLL_M               = 4
	PLL_N               = 216
	PLL_P               = 2
	PLL_Q               = 2
)

/*
   timer settings used for tick and sleep.

   note: TICK_TIMER_FREQ and SLEEP_TIMER_FREQ are controlled by PLL / clock
   settings above, so must be kept in sync if the clock settings are changed.
*/
const (
	TICK_RATE        = 1000 // 1 KHz
	SLEEP_TIMER_IRQ  = stm32.IRQ_TIM3
	SLEEP_TIMER_FREQ = 54000000 // 54 MHz (2x APB1)
	TICK_TIMER_IRQ   = stm32.IRQ_TIM7
	TICK_TIMER_FREQ  = 54000000 // 54 MHz (2x APB1)
)

type arrtype = uint32

func InitializeClocks() {
	initPLL()

	initSleepTimer(&timerInfo{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
		Device:         stm32.TIM3,
	})

	initTickTimer(&timerInfo{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM7EN,
		Device:         stm32.TIM7,
	})
}

func initPLL() {
	// PWR_CLK_ENABLE
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	_ = stm32.RCC.APB1ENR.Get()

	// PWR_VOLTAGESCALING_CONFIG
	stm32.PWR.CR1.ReplaceBits(0x3<<stm32.PWR_CR1_VOS_Pos, stm32.PWR_CR1_VOS_Msk, 0)
	_ = stm32.PWR.CR1.Get()

	// Initialize the High-Speed External Oscillator
	initOsc()

	// Set flash wait states (min 7 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & stm32.FLASH_ACR_LATENCY_Msk) < 7 {
		stm32.FLASH.ACR.ReplaceBits(7, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	// HCLK (0x1C00 = DIV_16, 0x0 = RCC_SYSCLK_DIV1) - ensure timers remain
	// within spec as the SYSCLK source changes.
	stm32.RCC.CFGR.ReplaceBits(0x00001C00, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0x00001C00<<3, stm32.RCC_CFGR_PPRE2_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set SYSCLK source and wait
	// (2 = PLLCLK, 3 = RCC_CFGR_SW mask, 3 << 3 = RCC_CFGR_SWS mask)
	stm32.RCC.CFGR.ReplaceBits(2, 3, 0)
	for stm32.RCC.CFGR.Get()&(3<<2) != (2 << 2) {
	}

	// Set flash wait states (max 7 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & stm32.FLASH_ACR_LATENCY_Msk) > 7 {
		stm32.FLASH.ACR.ReplaceBits(7, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	// Set APB1 and APB2 clocks (0x1800 = DIV8, 0x1000 = DIV2)
	stm32.RCC.CFGR.ReplaceBits(0x1800, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0x1000<<3, stm32.RCC_CFGR_PPRE2_Msk, 0)
}

func initOsc() {
	// Enable HSE, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Disable the PLL, wait until disabled
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL
	stm32.RCC.PLLCFGR.Set(0x20000000 |
		(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | // 1 = HSE
		PLL_M |
		(PLL_N << stm32.RCC_PLLCFGR_PLLN_Pos) |
		(((PLL_P >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		(PLL_Q << stm32.RCC_PLLCFGR_PLLQ_Pos))

	// Enable the PLL, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}
}

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32f7x2.go clock init code
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() / 2 // APB2 Frequency
	case stm32.USART2, stm32.USART3, stm32.UART4, stm32.UART5:
		clock = CPUFrequency() / 8 // APB1 Frequency
	}
	return clock / baudRate
}

// Register names vary by ST processor, these are for STM F7x2
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
}

//---------- I2C related code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange() uint32 {
	// This is a 'magic' value calculated by STM32CubeMX
	// for 27MHz PCLK1 (216MHz CPU Freq / 8).
	// TODO: Do calculations based on PCLK1
	return 0x00606A9B
}
