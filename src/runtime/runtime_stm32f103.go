//go:build stm32 && stm32f103
// +build stm32,stm32f103

package runtime

import (
	"machine"
	"tinygo.org/x/device/stm32"
)

func init() {
	initCLK()

	machine.Serial.Configure(machine.UARTConfig{})

	initTickTimer(&machine.TIM4)
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
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
