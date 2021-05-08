// +build stm32,stm32f103

package runtime

import (
	"device/stm32"
	"machine"
)

/*
   timer settings used for tick and sleep.

   note: TICK_TIMER_FREQ and SLEEP_TIMER_FREQ are controlled by PLL / clock
   settings configured in initCLK, so must be kept in sync if the clock settings
   are changed.
*/
const (
	TICK_RATE        = 1000 // 1 KHz
	SLEEP_TIMER_IRQ  = stm32.IRQ_TIM3
	SLEEP_TIMER_FREQ = 72000000 // 72 MHz
	TICK_TIMER_IRQ   = stm32.IRQ_TIM4
	TICK_TIMER_FREQ  = 72000000 // 72 MHz
)

type arrtype = uint32

func init() {
	initCLK()

	initSleepTimer(&timerInfo{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
		Device:         stm32.TIM3,
	})

	machine.UART0.Configure(machine.UARTConfig{})

	initTickTimer(&timerInfo{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM4EN,
		Device:         stm32.TIM4,
	})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

// initCLK sets clock to 72MHz using HSE 8MHz crystal w/ PLL X 9 (8MHz x 9 = 72MHz).
func initCLK() {
	stm32.FLASH.ACR.SetBits(stm32.FLASH_ACR_LATENCY_WS2)                          // Two wait states, per datasheet
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE1_Div2 << stm32.RCC_CFGR_PPRE1_Pos) // prescale PCLK1 = HCLK/2
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE2_Div1 << stm32.RCC_CFGR_PPRE2_Pos) // prescale PCLK2 = HCLK/1
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)                                      // enable HSE clock

	// wait for the HSEREADY flag
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION) // enable HSI clock

	// wait for the HSIREADY flag
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PLLSRC)                                   // set PLL source to HSE
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PLLMUL_Mul9 << stm32.RCC_CFGR_PLLMUL_Pos) // multiply by 9
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)                                        // enable the PLL

	// wait for the PLLRDY flag
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_SW_PLL) // set clock source to pll

	// wait for PLL to be CLK
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_SWS_PLL << stm32.RCC_CFGR_SWS_Pos) {
	}
}
