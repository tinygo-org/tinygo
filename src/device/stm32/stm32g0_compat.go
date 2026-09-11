//go:build stm32g0

package stm32

// Compatibility constants for STM32G0 that are named differently
// in the auto-generated SVD-based device file.

// GPIO OTYPER constants (open-drain/push-pull)
const (
	GPIO_OTYPER_OT0_PushPull  = 0x0
	GPIO_OTYPER_OT0_OpenDrain = 0x1
)

// USART CR1 constants (these have FIFO_ENABLED/FIFO_DISABLED variants in G0)
// We use the FIFO_DISABLED variants as defaults for compatibility
const (
	USART_CR1_UE     = USART_CR1_FIFO_DISABLED_UE
	USART_CR1_TE     = USART_CR1_FIFO_DISABLED_TE
	USART_CR1_RE     = USART_CR1_FIFO_DISABLED_RE
	USART_CR1_RXNEIE = USART_CR1_FIFO_DISABLED_RXNEIE
)

// USART ISR constants
const (
	USART_ISR_TXE  = USART_ISR_FIFO_DISABLED_TXE
	USART_ISR_RXNE = USART_ISR_FIFO_DISABLED_RXNE
	USART_ISR_TC   = USART_ISR_FIFO_DISABLED_TC
)

// SPI CR1 BR (baud rate) divisor constants
const (
	SPI_CR1_BR_Div2   = SPI_CR1_BR_B_0x0 << SPI_CR1_BR_Pos // fPCLK/2
	SPI_CR1_BR_Div4   = SPI_CR1_BR_B_0x1 << SPI_CR1_BR_Pos // fPCLK/4
	SPI_CR1_BR_Div8   = SPI_CR1_BR_B_0x2 << SPI_CR1_BR_Pos // fPCLK/8
	SPI_CR1_BR_Div16  = SPI_CR1_BR_B_0x3 << SPI_CR1_BR_Pos // fPCLK/16
	SPI_CR1_BR_Div32  = SPI_CR1_BR_B_0x4 << SPI_CR1_BR_Pos // fPCLK/32
	SPI_CR1_BR_Div64  = SPI_CR1_BR_B_0x5 << SPI_CR1_BR_Pos // fPCLK/64
	SPI_CR1_BR_Div128 = SPI_CR1_BR_B_0x6 << SPI_CR1_BR_Pos // fPCLK/128
	SPI_CR1_BR_Div256 = SPI_CR1_BR_B_0x7 << SPI_CR1_BR_Pos // fPCLK/256
)

// Flash ACR latency constants (wait states)
const (
	Flash_ACR_LATENCY_WS0 = 0x0 // 0 wait states
	Flash_ACR_LATENCY_WS1 = 0x1 // 1 wait state
	Flash_ACR_LATENCY_WS2 = 0x2 // 2 wait states
)

// RCC PLLCFGR PLLSRC values
const (
	RCC_PLLCFGR_PLLSRC_NONE  = 0x0 // No clock
	RCC_PLLCFGR_PLLSRC_HSI16 = 0x2 // HSI16 clock selected as PLL input
	RCC_PLLCFGR_PLLSRC_HSE   = 0x3 // HSE clock selected as PLL input
)

// RCC CFGR SW (System clock switch) values
const (
	RCC_CFGR_SW_HSISYS  = 0x0 // HSISYS selected as system clock
	RCC_CFGR_SW_HSE     = 0x1 // HSE selected as system clock
	RCC_CFGR_SW_PLLRCLK = 0x2 // PLLRCLK selected as system clock
	RCC_CFGR_SW_LSI     = 0x3 // LSI selected as system clock
	RCC_CFGR_SW_LSE     = 0x4 // LSE selected as system clock
)

// RCC CFGR SWS (System clock switch status) values
const (
	RCC_CFGR_SWS_HSISYS  = 0x0 // HSISYS used as system clock
	RCC_CFGR_SWS_HSE     = 0x1 // HSE used as system clock
	RCC_CFGR_SWS_PLLRCLK = 0x2 // PLLRCLK used as system clock
	RCC_CFGR_SWS_LSI     = 0x3 // LSI used as system clock
	RCC_CFGR_SWS_LSE     = 0x4 // LSE used as system clock
)

// RCC CFGR HPRE (AHB prescaler) values
const (
	RCC_CFGR_HPRE_Div1   = 0x0 // SYSCLK not divided
	RCC_CFGR_HPRE_Div2   = 0x8 // SYSCLK divided by 2
	RCC_CFGR_HPRE_Div4   = 0x9 // SYSCLK divided by 4
	RCC_CFGR_HPRE_Div8   = 0xA // SYSCLK divided by 8
	RCC_CFGR_HPRE_Div16  = 0xB // SYSCLK divided by 16
	RCC_CFGR_HPRE_Div64  = 0xC // SYSCLK divided by 64
	RCC_CFGR_HPRE_Div128 = 0xD // SYSCLK divided by 128
	RCC_CFGR_HPRE_Div256 = 0xE // SYSCLK divided by 256
	RCC_CFGR_HPRE_Div512 = 0xF // SYSCLK divided by 512
)

// RCC CFGR PPRE (APB prescaler) values
const (
	RCC_CFGR_PPRE_Div1  = 0x0 // HCLK not divided
	RCC_CFGR_PPRE_Div2  = 0x4 // HCLK divided by 2
	RCC_CFGR_PPRE_Div4  = 0x5 // HCLK divided by 4
	RCC_CFGR_PPRE_Div8  = 0x6 // HCLK divided by 8
	RCC_CFGR_PPRE_Div16 = 0x7 // HCLK divided by 16
)
