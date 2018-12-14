// +build stm32

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
)

type timeUnit int64

const tickMicros = 1 // TODO

const (
	// Flash Access Control Register flag values.
	FLASH_ACR_LATENCY_0 = 0x00000004
	FLASH_ACR_LATENCY_1 = 0x00000002
	FLASH_ACR_LATENCY_2 = 0x00000004
	FLASH_ACR_HLFCYA    = 0x00000008
	FLASH_ACR_PRFTBE    = 0x00000010
	FLASH_ACR_PRFTBS    = 0x00000020

	// Reset and Clock Control Control Register flag values.
	RCC_CFGR_SW_HSI = 0
	RCC_CFGR_SW_HSE = 1
	RCC_CFGR_SW_PLL = 2

	RCC_CFGR_SWS_HSI = 0x00000000
	RCC_CFGR_SWS_HSE = 0x00000004
	RCC_CFGR_SWS_PLL = 0x00000008

	RCC_CFGR_PPRE1_DIV_NONE = 0x00000000
	RCC_CFGR_PPRE1_DIV_2    = 0x00000400
	RCC_CFGR_PPRE1_DIV_4    = 0x00000500
	RCC_CFGR_PPRE1_DIV_8    = 0x00000600
	RCC_CFGR_PPRE1_DIV_16   = 0x00000700

	RCC_CFGR_PLLMUL_2  = 0x00000000
	RCC_CFGR_PLLMUL_3  = 0x00040000
	RCC_CFGR_PLLMUL_4  = 0x00080000
	RCC_CFGR_PLLMUL_5  = 0x000C0000
	RCC_CFGR_PLLMUL_6  = 0x00100000
	RCC_CFGR_PLLMUL_7  = 0x00140000
	RCC_CFGR_PLLMUL_8  = 0x00180000
	RCC_CFGR_PLLMUL_9  = 0x001C0000
	RCC_CFGR_PLLMUL_10 = 0x00200000
	RCC_CFGR_PLLMUL_11 = 0x00240000
	RCC_CFGR_PLLMUL_12 = 0x00280000
	RCC_CFGR_PLLMUL_13 = 0x002C0000
	RCC_CFGR_PLLMUL_14 = 0x00300000
	RCC_CFGR_PLLMUL_15 = 0x00340000
	RCC_CFGR_PLLMUL_16 = 0x00380000
)

//go:export Reset_Handler
func main() {
	preinit()
	initAll()
	mainWrapper()
	abort()
}

func init() {
	initCLK()
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

// initCLK sets clock to 72MHz using HSE 8MHz crystal w/ PLL X 9 (8MHz x 9 = 72MHz).
func initCLK() {
	stm32.FLASH.ACR |= FLASH_ACR_LATENCY_2 // Two wait states, per datasheet
	stm32.RCC.CFGR |= RCC_CFGR_PPRE1_DIV_2 // prescale AHB1 = HCLK/2
	stm32.RCC.CR |= stm32.RCC_CR_HSEON     // enable HSE clock

	// wait for the HSEREADY flag
	for (stm32.RCC.CR & stm32.RCC_CR_HSERDY) == 0 {
	}

	stm32.RCC.CFGR |= stm32.RCC_CFGR_PLLSRC // set PLL source to HSE
	stm32.RCC.CFGR |= RCC_CFGR_PLLMUL_9     // multiply by 9
	stm32.RCC.CR |= stm32.RCC_CR_PLLON      // enable the PLL

	// wait for the PLLRDY flag
	for (stm32.RCC.CR & stm32.RCC_CR_PLLRDY) == 0 {
	}

	stm32.RCC.CFGR |= RCC_CFGR_SW_PLL // set clock source to pll

	// wait for PLL to be CLK
	for (stm32.RCC.CFGR & RCC_CFGR_SWS_PLL) == 0 {
	}
}

func sleepTicks(d timeUnit) {
	// TODO: use a real timer here
	for i := 0; i < int(d/535); i++ {
		arm.Asm("")
	}
}

func ticks() timeUnit {
	return 0 // TODO
}
