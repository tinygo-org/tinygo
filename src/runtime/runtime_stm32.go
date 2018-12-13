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
	FLASH_ACR_LATENCY_0 = (0)
	FLASH_ACR_LATENCY_1 = (1)
	FLASH_ACR_LATENCY_2 = (2)
	FLASH_ACR_HLFCYA    = (1 << 3)
	FLASH_ACR_PRFTBE    = (1 << 4)
	FLASH_ACR_PRFTBS    = (1 << 5)

	// Reset and Clock Control Control Register flag values.
	RCC_CFGR_SW_HSI = (0 << 0)
	RCC_CFGR_SW_HSE = (1 << 0)
	RCC_CFGR_SW_PLL = (2 << 0)

	RCC_CFGR_PPRE1_DIV_NONE = (0 << 8)
	RCC_CFGR_PPRE1_DIV_2    = ((1 << 10) | (0 << 8))
	RCC_CFGR_PPRE1_DIV_4    = ((1 << 10) | (1 << 8))
	RCC_CFGR_PPRE1_DIV_8    = ((1 << 10) | (2 << 8))
	RCC_CFGR_PPRE1_DIV_16   = ((1 << 10) | (3 << 8))

	RCC_CFGR_PLLMUL_2  = (0 << 18)
	RCC_CFGR_PLLMUL_3  = (1 << 18)
	RCC_CFGR_PLLMUL_4  = (2 << 18)
	RCC_CFGR_PLLMUL_5  = (3 << 18)
	RCC_CFGR_PLLMUL_6  = (4 << 18)
	RCC_CFGR_PLLMUL_7  = (5 << 18)
	RCC_CFGR_PLLMUL_8  = (6 << 18)
	RCC_CFGR_PLLMUL_9  = (7 << 18)
	RCC_CFGR_PLLMUL_10 = (8 << 18)
	RCC_CFGR_PLLMUL_11 = (9 << 18)
	RCC_CFGR_PLLMUL_12 = (10 << 18)
	RCC_CFGR_PLLMUL_13 = (11 << 18)
	RCC_CFGR_PLLMUL_14 = (12 << 18)
	RCC_CFGR_PLLMUL_15 = (13 << 18)
	RCC_CFGR_PLLMUL_16 = (14 << 18)
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
	for (stm32.RCC.CFGR & RCC_CFGR_SW_PLL) == 0 {
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
