// +build stm32,stm32l0

package runtime

import (
	"device/stm32"
	"machine"
)

/*
   timer settings used for tick and sleep.

   note: TICK_TIMER_FREQ and SLEEP_TIMER_FREQ are controlled by PLL / clock
   settings above, so must be kept in sync if the clock settings are changed.
*/
const (
	TICK_RATE        = 1000 // 1 KHz
	TICK_TIMER_IRQ   = stm32.IRQ_TIM7
	TICK_TIMER_FREQ  = 32000000 // 32 MHz
	SLEEP_TIMER_IRQ  = stm32.IRQ_TIM3
	SLEEP_TIMER_FREQ = 32000000 // 32 MHz
)

type arrtype = uint16

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
		EnableFlag:     stm32.RCC_APB1ENR_TIM7EN,
		Device:         stm32.TIM7,
	})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

// initCLK sets clock to 32MHz
// SEE: https://github.com/WRansohoff/STM32x0_timer_example/blob/master/src/main.c

func initCLK() {

	// Set the Flash ACR to use 1 wait-state
	// enable the prefetch buffer and pre-read for performance
	stm32.FLASH.ACR.SetBits(stm32.Flash_ACR_LATENCY | stm32.Flash_ACR_PRFTEN | stm32.Flash_ACR_PRE_READ)

	// Set presaclers so half system clock (PCLKx = HCLK/2)
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE1_Div2 << stm32.RCC_CFGR_PPRE1_Pos)
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE2_Div2 << stm32.RCC_CFGR_PPRE2_Pos)

	// Enable the HSI16 oscillator, since the L0 series boots to the MSI one.
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSI16ON)

	// Wait for HSI16 to be ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSI16RDYF) {
	}

	// Configure the PLL to use HSI16 with a PLLDIV of 2 and PLLMUL of 4.
	stm32.RCC.CFGR.SetBits(0x01<<stm32.RCC_CFGR_PLLDIV_Pos | 0x01<<stm32.RCC_CFGR_PLLMUL_Pos)
	stm32.RCC.CFGR.ClearBits(0x02<<stm32.RCC_CFGR_PLLDIV_Pos | 0x0E<<stm32.RCC_CFGR_PLLMUL_Pos)
	stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_PLLSRC)

	// Enable PLL
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)

	// Wait for PLL to be ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Use PLL As System clock
	stm32.RCC.CFGR.SetBits(0x3)

}

const asyncScheduler = false
