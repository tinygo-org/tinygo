//go:build stm32 && stm32h7

package machine

import (
	"device/arm"
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

var deviceIDAddr = []uintptr{0x1FF1E800, 0x1FF1E804, 0x1FF1E808}

// Default USB identifiers; board files with USB support should override these.
const (
	usb_STRING_PRODUCT      = "STM32H7"
	usb_STRING_MANUFACTURER = "TinyGo"

	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x0001
)

// EnterBootloader resets the chip. Jumping to the H7 system bootloader
// (0x1FF09800) is not implemented; a plain reset restarts the application.
func EnterBootloader() {
	arm.SystemReset()
}

// HSI_KER_FREQ is the fixed-frequency internal RC oscillator (RM0433 §8.2),
// unaffected by the board's HSE crystal.
const HSI_KER_FREQ = 64_000_000

// sysClockFreq returns SYSCLK (PLL1P output) for the board's configured
// xtalHz, per the M/N/P dividers initCLK() programs into PLL1.
func sysClockFreq() uint32 {
	pll := PLLParams400MHz()
	return xtalHz / pll.M * pll.N / pll.P
}

// pll1QFreq returns the PLL1Q output (SPI1/2/3 kernel clock source).
func pll1QFreq() uint32 {
	pll := PLLParams400MHz()
	return xtalHz / pll.M * pll.N / pll.Q
}

// hclkFreq, pclk1Freq..pclk4Freq return the AHB/APBx bus clocks. initCLK()
// hardcodes HPRE/D2PPRE1/D2PPRE2/D1PPRE/D3PPRE to Div2, so these are fixed
// ratios of SYSCLK regardless of xtal.
func hclkFreq() uint32  { return sysClockFreq() / 2 }
func pclk1Freq() uint32 { return hclkFreq() / 2 }
func pclk2Freq() uint32 { return hclkFreq() / 2 }
func pclk3Freq() uint32 { return hclkFreq() / 2 }
func pclk4Freq() uint32 { return hclkFreq() / 2 }

// Peripheral kernel clocks as configured by initCLK().
func spi123KerFreq() uint32       { return pll1QFreq() } // D2CCIP1R.SPI123SEL = PLL1_Q
func spi45KerFreq() uint32        { return pclk2Freq() } // D2CCIP1R.SPI45SEL  = APB
func spi6KerFreq() uint32         { return pclk4Freq() } // D3CCIPR.SPI6SEL    = PCLK4
const I2C_KER_FREQ = HSI_KER_FREQ // D2CCIP2R.I2C123SEL = HSI_KER

func CPUFrequency() uint32 {
	return sysClockFreq()
}

// initRNG gates the AHB2 bus clock for the RNG and enables the peripheral.
// HSI48 is started and selected as the RNG kernel clock in initCLK().
func initRNG() {
	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}

// Alternate function pin selection.
const (
	AF0_SYSTEM                          = 0
	AF1_TIM1_2_16_17_HRTIM              = 1
	AF2_TIM3_4_5_HRTIM                  = 2
	AF3_TIM8_LPTIM1_DFSDM_HRTIM         = 3
	AF4_I2C1_2_3_4_USART1               = 4
	AF5_SPI1_2_3_4_5_6_I2S              = 5
	AF6_SPI2_3_SAI1_I2S_UART4_DFSDM     = 6
	AF7_SPI2_3_USART1_2_3_UART5_SPDIFRX = 7
	AF8_SAI2_UART4_5_8_SPDIFRX_LPUART   = 8
	AF9_FDCAN1_2_TIM13_14_QUADSPI_FMC   = 9
	AF10_OTG_HS_FS_SAI2_QUADSPI_SDMMC2  = 10
	AF11_SDMMC2_ETH_MDIO_UART7_SWPMI    = 11
	AF12_FMC_SDMMC1_MDIOS_OTG_FS_UART7  = 12
	AF13_DCMI_DSI_COMP_LTDC             = 13
	AF14_LTDC                           = 14
	AF15_EVENTOUT                       = 15
)

// Timer clock = 2×PCLK when both HPRE and PPREx prescalers are active (RM0433 §8.5.5).
func apb1TimFreq() uint64 { return 2 * uint64(pclk1Freq()) }
func apb2TimFreq() uint64 { return 2 * uint64(pclk2Freq()) }

//---------- Timer related code

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA8, AF1_TIM1_2_16_17_HRTIM}, {PE9, AF1_TIM1_2_16_17_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA9, AF1_TIM1_2_16_17_HRTIM}, {PE11, AF1_TIM1_2_16_17_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA10, AF1_TIM1_2_16_17_HRTIM}, {PE13, AF1_TIM1_2_16_17_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA11, AF1_TIM1_2_16_17_HRTIM}, {PE14, AF1_TIM1_2_16_17_HRTIM}}},
		},
		busFreq: apb2TimFreq(),
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1LENR,
		EnableFlag:     stm32.RCC_APB1LENR_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF1_TIM1_2_16_17_HRTIM}, {PA5, AF1_TIM1_2_16_17_HRTIM}, {PA15, AF1_TIM1_2_16_17_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF1_TIM1_2_16_17_HRTIM}, {PB3, AF1_TIM1_2_16_17_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF1_TIM1_2_16_17_HRTIM}, {PB10, AF1_TIM1_2_16_17_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF1_TIM1_2_16_17_HRTIM}, {PB11, AF1_TIM1_2_16_17_HRTIM}}},
		},
		busFreq: apb1TimFreq(),
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1LENR,
		EnableFlag:     stm32.RCC_APB1LENR_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: apb1TimFreq(),
	}

	TIM4 = TIM{
		EnableRegister: &stm32.RCC.APB1LENR,
		EnableFlag:     stm32.RCC_APB1LENR_TIM4EN,
		Device:         stm32.TIM4,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB6, AF2_TIM3_4_5_HRTIM}, {PD12, AF2_TIM3_4_5_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PB7, AF2_TIM3_4_5_HRTIM}, {PD13, AF2_TIM3_4_5_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PB8, AF2_TIM3_4_5_HRTIM}, {PD14, AF2_TIM3_4_5_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PB9, AF2_TIM3_4_5_HRTIM}, {PD15, AF2_TIM3_4_5_HRTIM}}},
		},
		busFreq: apb1TimFreq(),
	}

	TIM5 = TIM{
		EnableRegister: &stm32.RCC.APB1LENR,
		EnableFlag:     stm32.RCC_APB1LENR_TIM5EN,
		Device:         stm32.TIM5,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF2_TIM3_4_5_HRTIM}, {PH10, AF2_TIM3_4_5_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF2_TIM3_4_5_HRTIM}, {PH11, AF2_TIM3_4_5_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF2_TIM3_4_5_HRTIM}, {PH12, AF2_TIM3_4_5_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF2_TIM3_4_5_HRTIM}, {PI0, AF2_TIM3_4_5_HRTIM}}},
		},
		busFreq: apb1TimFreq(),
	}

	TIM8 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM8EN,
		Device:         stm32.TIM8,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PC6, AF3_TIM8_LPTIM1_DFSDM_HRTIM}, {PI5, AF3_TIM8_LPTIM1_DFSDM_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PC7, AF3_TIM8_LPTIM1_DFSDM_HRTIM}, {PI6, AF3_TIM8_LPTIM1_DFSDM_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PC8, AF3_TIM8_LPTIM1_DFSDM_HRTIM}, {PI7, AF3_TIM8_LPTIM1_DFSDM_HRTIM}}},
			TimerChannel{Pins: []PinFunction{{PC9, AF3_TIM8_LPTIM1_DFSDM_HRTIM}, {PI2, AF3_TIM8_LPTIM1_DFSDM_HRTIM}}},
		},
		busFreq: apb2TimFreq(),
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_UP, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleUPInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleUPInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_UP_TIM13, TIM8.handleUPInterrupt)
	}
	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM_CC, TIM1.handleOCInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleOCInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleOCInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleOCInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleOCInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_CC, TIM8.handleOCInterrupt)
	}
	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {
	if t.Device == stm32.TIM1 || t.Device == stm32.TIM8 {
		t.Device.BDTR.SetBits(stm32.TIM_BDTR_MOE)
	}
}

type psctype = uint32
type arrtype = uint32
type arrRegType = volatile.Register32

const ARR_MAX = 0x10000
const PSC_MAX = 0x10000

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	// Default USART kernel clock is the peripheral's own APB clock:
	// USART1/6 sit on APB2, the rest on APB1.
	clock := pclk1Freq()
	if uart.Bus == stm32.USART1 || uart.Bus == stm32.USART6 {
		clock = pclk2Freq()
	}
	return clock / baudRate
}

func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
	uart.errClearReg = &uart.Bus.ICR
}

func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.USART1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.USART2):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_USART2EN)
	case unsafe.Pointer(stm32.USART3):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_USART3EN)
	case unsafe.Pointer(stm32.UART4):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_UART4EN)
	case unsafe.Pointer(stm32.UART5):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_UART5EN)
	case unsafe.Pointer(stm32.USART6):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART6EN)
	case unsafe.Pointer(stm32.UART7):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_UART7EN)
	case unsafe.Pointer(stm32.UART8):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_UART8EN)
	case unsafe.Pointer(stm32.LPUART1):
		stm32.RCC.APB4ENR.SetBits(stm32.RCC_APB4ENR_LPUART1EN)
	case unsafe.Pointer(stm32.I2C1):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_I2C1EN)
	case unsafe.Pointer(stm32.I2C2):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_I2C2EN)
	case unsafe.Pointer(stm32.I2C3):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_I2C3EN)
	case unsafe.Pointer(stm32.I2C4):
		stm32.RCC.APB4ENR.SetBits(stm32.RCC_APB4ENR_I2C4EN)
	case unsafe.Pointer(stm32.SPI1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.SPI2):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_SPI2EN)
	case unsafe.Pointer(stm32.SPI3):
		stm32.RCC.APB1LENR.SetBits(stm32.RCC_APB1LENR_SPI3EN)
	case unsafe.Pointer(stm32.SPI4):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI4EN)
	case unsafe.Pointer(stm32.SPI5):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI5EN)
	case unsafe.Pointer(stm32.SPI6):
		stm32.RCC.APB4ENR.SetBits(stm32.RCC_APB4ENR_SPI6EN)
	case unsafe.Pointer(stm32.WWDG):
		stm32.RCC.APB3ENR.SetBits(stm32.RCC_APB3ENR_WWDG1EN)
	}
}

//---------- GPIO related code

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 4:
		return stm32.GPIOE
	case 5:
		return stm32.GPIOF
	case 6:
		return stm32.GPIOG
	case 7:
		return stm32.GPIOH
	case 8:
		return stm32.GPIOI
	case 9:
		return stm32.GPIOJ
	case 10:
		return stm32.GPIOK
	default:
		panic("machine: unknown port")
	}
}

func (p Pin) enableClock() {
	switch p.getPort() {
	case stm32.GPIOA:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOAEN)
	case stm32.GPIOB:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOBEN)
	case stm32.GPIOC:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOCEN)
	case stm32.GPIOD:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIODEN)
	case stm32.GPIOE:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOEEN)
	case stm32.GPIOF:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOFEN)
	case stm32.GPIOG:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOGEN)
	case stm32.GPIOH:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOHEN)
	case stm32.GPIOI:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOIEN)
	case stm32.GPIOJ:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOJEN)
	case stm32.GPIOK:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOKEN)
	}
}

const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15

	PD0  = portD + 0
	PD1  = portD + 1
	PD2  = portD + 2
	PD3  = portD + 3
	PD4  = portD + 4
	PD5  = portD + 5
	PD6  = portD + 6
	PD7  = portD + 7
	PD8  = portD + 8
	PD9  = portD + 9
	PD10 = portD + 10
	PD11 = portD + 11
	PD12 = portD + 12
	PD13 = portD + 13
	PD14 = portD + 14
	PD15 = portD + 15

	PE0  = portE + 0
	PE1  = portE + 1
	PE2  = portE + 2
	PE3  = portE + 3
	PE4  = portE + 4
	PE5  = portE + 5
	PE6  = portE + 6
	PE7  = portE + 7
	PE8  = portE + 8
	PE9  = portE + 9
	PE10 = portE + 10
	PE11 = portE + 11
	PE12 = portE + 12
	PE13 = portE + 13
	PE14 = portE + 14
	PE15 = portE + 15

	PF0  = portF + 0
	PF1  = portF + 1
	PF2  = portF + 2
	PF3  = portF + 3
	PF4  = portF + 4
	PF5  = portF + 5
	PF6  = portF + 6
	PF7  = portF + 7
	PF8  = portF + 8
	PF9  = portF + 9
	PF10 = portF + 10
	PF11 = portF + 11
	PF12 = portF + 12
	PF13 = portF + 13
	PF14 = portF + 14
	PF15 = portF + 15

	PG0  = portG + 0
	PG1  = portG + 1
	PG2  = portG + 2
	PG3  = portG + 3
	PG4  = portG + 4
	PG5  = portG + 5
	PG6  = portG + 6
	PG7  = portG + 7
	PG8  = portG + 8
	PG9  = portG + 9
	PG10 = portG + 10
	PG11 = portG + 11
	PG12 = portG + 12
	PG13 = portG + 13
	PG14 = portG + 14
	PG15 = portG + 15

	PH0  = portH + 0
	PH1  = portH + 1
	PH2  = portH + 2
	PH3  = portH + 3
	PH4  = portH + 4
	PH5  = portH + 5
	PH6  = portH + 6
	PH7  = portH + 7
	PH8  = portH + 8
	PH9  = portH + 9
	PH10 = portH + 10
	PH11 = portH + 11
	PH12 = portH + 12
	PH13 = portH + 13
	PH14 = portH + 14
	PH15 = portH + 15

	PI0  = portI + 0
	PI1  = portI + 1
	PI2  = portI + 2
	PI3  = portI + 3
	PI4  = portI + 4
	PI5  = portI + 5
	PI6  = portI + 6
	PI7  = portI + 7
	PI8  = portI + 8
	PI9  = portI + 9
	PI10 = portI + 10
	PI11 = portI + 11
	PI12 = portI + 12
	PI13 = portI + 13
	PI14 = portI + 14
	PI15 = portI + 15

	PJ0  = portJ + 0
	PJ1  = portJ + 1
	PJ2  = portJ + 2
	PJ3  = portJ + 3
	PJ4  = portJ + 4
	PJ5  = portJ + 5
	PJ6  = portJ + 6
	PJ7  = portJ + 7
	PJ8  = portJ + 8
	PJ9  = portJ + 9
	PJ10 = portJ + 10
	PJ11 = portJ + 11
	PJ12 = portJ + 12
	PJ13 = portJ + 13
	PJ14 = portJ + 14
	PJ15 = portJ + 15

	PK0 = portK + 0
	PK1 = portK + 1
	PK2 = portK + 2
	PK3 = portK + 3
	PK4 = portK + 4
	PK5 = portK + 5
	PK6 = portK + 6
	PK7 = portK + 7
)

//---------- I2C related code

// getFreqRange returns the TIMINGR value for the given I2C frequency.
// Values are for HSI_KER=64MHz (configured in initCLK).
// Derived from ST I2C timing calculator.
func (i2c *I2C) getFreqRange(br uint32) uint32 {
	switch br {
	case 10 * KHz:
		return 0x30E0E7CF
	case 100 * KHz:
		return 0x10B0BFCF
	case 400 * KHz:
		return 0x00901E74
	case 800 * KHz:
		return 0x00401137
	case 1_000 * KHz:
		return 0x00401028
	default:
		return 0x10B0BFCF
	}
}
