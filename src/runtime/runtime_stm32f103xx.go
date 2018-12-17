// +build stm32,stm32f103xx

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
)

const tickMicros = 1 // TODO

func init() {
	initCLK()
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

// initCLK sets clock to 72MHz using HSE 8MHz crystal w/ PLL X 9 (8MHz x 9 = 72MHz).
func initCLK() {
	stm32.FLASH.ACR |= stm32.FLASH_ACR_LATENCY_2 // Two wait states, per datasheet
	stm32.RCC.CFGR |= stm32.RCC_CFGR_PPRE1_DIV_2 // prescale AHB1 = HCLK/2
	stm32.RCC.CR |= stm32.RCC_CR_HSEON           // enable HSE clock

	// wait for the HSEREADY flag
	for (stm32.RCC.CR & stm32.RCC_CR_HSERDY) == 0 {
	}

	stm32.RCC.CFGR |= stm32.RCC_CFGR_PLLSRC   // set PLL source to HSE
	stm32.RCC.CFGR |= stm32.RCC_CFGR_PLLMUL_9 // multiply by 9
	stm32.RCC.CR |= stm32.RCC_CR_PLLON        // enable the PLL

	// wait for the PLLRDY flag
	for (stm32.RCC.CR & stm32.RCC_CR_PLLRDY) == 0 {
	}

	stm32.RCC.CFGR |= stm32.RCC_CFGR_SW_PLL // set clock source to pll

	// wait for PLL to be CLK
	for (stm32.RCC.CFGR & stm32.RCC_CFGR_SWS_PLL) == 0 {
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
