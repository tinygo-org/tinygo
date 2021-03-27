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

//---------- Clock and Oscillator related code

const (
	// +----------------------+
	// |    Clock Settings    |
	// +-------------+--------+
	// | HSE         | 12mhz  |
	// | SYSCLK      | 168mhz |
	// | HCLK        | 168mhz |
	// | APB1(PCLK1) | 42mhz  |
	// | APB2(PCLK2) | 84mhz  |
	// +-------------+--------+
	HCLK_FREQ_HZ  = 168000000
	PCLK1_FREQ_HZ = HCLK_FREQ_HZ / 4
	PCLK2_FREQ_HZ = HCLK_FREQ_HZ / 2
)

const (
	PWR_SCALE1 = 1 << stm32.PWR_CSR_VOSRDY_Pos // max value of HCLK = 168 MHz
	PWR_SCALE2 = 0                             // max value of HCLK = 144 MHz

	PLL_SRC_HSE = 1 << stm32.RCC_PLLCFGR_PLLSRC_Pos // use HSE for PLL and PLLI2S
	PLL_SRC_HSI = 0                                 // use HSI for PLL and PLLI2S

	PLL_DIV_M = 6 << stm32.RCC_PLLCFGR_PLLM_Pos
	PLL_MLT_N = 168 << stm32.RCC_PLLCFGR_PLLN_Pos
	PLL_DIV_P = ((2 >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos
	PLL_DIV_Q = 7 << stm32.RCC_PLLCFGR_PLLQ_Pos

	SYSCLK_SRC_PLL  = stm32.RCC_CFGR_SW_PLL << stm32.RCC_CFGR_SW_Pos
	SYSCLK_STAT_PLL = stm32.RCC_CFGR_SWS_PLL << stm32.RCC_CFGR_SWS_Pos

	RCC_DIV_PCLK1 = stm32.RCC_CFGR_PPRE1_Div4 << stm32.RCC_CFGR_PPRE1_Pos // HCLK / 4
	RCC_DIV_PCLK2 = stm32.RCC_CFGR_PPRE2_Div2 << stm32.RCC_CFGR_PPRE2_Pos // HCLK / 2
	RCC_DIV_HCLK  = stm32.RCC_CFGR_HPRE_Div1 << stm32.RCC_CFGR_HPRE_Pos   // SYSCLK / 1

	CLK_CCM_RAM = 1 << 20
)

const (
	// +-----------------------------------+
	// |    Voltage range = 2.7V - 3.6V    |
	// +----------------+------------------+
	// |   Wait states  |    System Bus    |
	// |  (WS, LATENCY) |    HCLK (MHz)    |
	// +----------------+------------------+
	// | 0 WS, 1 cycle  |   0 < HCLK ≤ 30  |
	// | 1 WS, 2 cycles |  30 < HCLK ≤ 60  |
	// | 2 WS, 3 cycles |  60 < HCLK ≤ 90  |
	// | 3 WS, 4 cycles |  90 < HCLK ≤ 120 |
	// | 4 WS, 5 cycles | 120 < HCLK ≤ 150 |
	// | 5 WS, 6 cycles | 150 < HCLK ≤ 168 |
	// +----------------+------------------+
	FLASH_LATENCY = 5 << stm32.FLASH_ACR_LATENCY_Pos // 5 WS (6 CPU cycles)

	// instruction cache, data cache, and prefetch
	FLASH_OPTIONS = stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | stm32.FLASH_ACR_PRFTEN
)

/*
   timer settings used for tick and sleep.

   note: TICK_TIMER_FREQ and SLEEP_TIMER_FREQ are controlled by PLL / clock
   settings above, so must be kept in sync if the clock settings are changed.
*/
const (
	TICK_RATE        = 1000 // 1 KHz
	SLEEP_TIMER_IRQ  = stm32.IRQ_TIM3
	SLEEP_TIMER_FREQ = PCLK1_FREQ_HZ * 2
	TICK_TIMER_IRQ   = stm32.IRQ_TIM7
	TICK_TIMER_FREQ  = PCLK1_FREQ_HZ * 2
)

type arrtype = uint32

const asyncScheduler = false

func InitializeClocks() {
	initOSC() // configure oscillators
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

func initOSC() {
	// enable voltage regulator
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	stm32.PWR.CR.SetBits(PWR_SCALE1)

	// enable HSE
	stm32.RCC.CR.Set(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Since the main-PLL configuration parameters cannot be changed once PLL is
	// enabled, it is recommended to configure PLL before enabling it (selection
	// of the HSI or HSE oscillator as PLL clock source, and configuration of
	// division factors M, N, P, and Q).

	// disable PLL and wait for it to reset
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// set HSE as PLL source and configure clock divisors
	stm32.RCC.PLLCFGR.Set(PLL_SRC_HSE | PLL_DIV_M | PLL_MLT_N | PLL_DIV_P | PLL_DIV_Q)

	// enable PLL and wait for it to sync
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}
}

func initPLL() {
	// After reset, the CPU clock frequency is 16 MHz and 0 wait state (WS) is
	// configured in the FLASH_ACR register.
	//
	// It is highly recommended to use the following software sequences to tune
	// the number of wait states needed to access the Flash memory with the CPU
	// frequency.
	//
	// 1. Program the new number of wait states to the LATENCY bits in the
	//    FLASH_ACR register
	// 2. Check that the new number of wait states is taken into account to access
	//    the Flash memory by reading the FLASH_ACR register
	// 3. Modify the CPU clock source by writing the SW bits in the RCC_CFGR
	//    register
	// 4. If needed, modify the CPU clock prescaler by writing the HPRE bits in
	//    RCC_CFGR
	// 5. Check that the new CPU clock source or/and the new CPU clock prescaler
	//    value is/are taken into account by reading the clock source status (SWS
	//    bits) or/and the AHB prescaler value (HPRE bits), respectively, in the
	//    RCC_CFGR register.

	// configure instruction/data caching, prefetch, and flash access wait states
	stm32.FLASH.ACR.Set(FLASH_OPTIONS | FLASH_LATENCY)
	for !stm32.FLASH.ACR.HasBits(FLASH_LATENCY) { // verify new wait states
	}

	// After a system reset, the HSI oscillator is selected as the system clock.
	// When a clock source is used directly or through PLL as the system clock, it
	// is not possible to stop it.
	//
	// A switch from one clock source to another occurs only if the target clock
	// source is ready (clock stable after startup delay or PLL locked). If a
	// clock source that is not yet ready is selected, the switch occurs when the
	// clock source is ready. Status bits in the RCC clock control register
	// (RCC_CR) indicate which clock(s) is (are) ready and which clock is
	// currently used as the system clock.

	// set CPU clock source to PLL
	stm32.RCC.CFGR.SetBits(SYSCLK_SRC_PLL)

	// update PCKL1/2 and HCLK divisors
	stm32.RCC.CFGR.SetBits(RCC_DIV_PCLK1 | RCC_DIV_PCLK2 | RCC_DIV_HCLK)

	// verify system clock source is ready
	for !stm32.RCC.CFGR.HasBits(SYSCLK_STAT_PLL) {
	}

	// enable the CCM RAM clock
	stm32.RCC.AHB1ENR.SetBits(CLK_CCM_RAM)
}

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
	uart.rxReg = &uart.Bus.DR
	uart.txReg = &uart.Bus.DR
	uart.statusReg = &uart.Bus.SR
	uart.txEmptyFlag = stm32.USART_SR_TXE
}

// -- SPI ----------------------------------------------------------------------

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
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
