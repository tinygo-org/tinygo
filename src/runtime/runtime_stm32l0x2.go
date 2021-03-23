// +build stm32l0x2

package runtime

import (
	"device/stm32"
	"machine"
)

/*
   timer settings used for tick and sleep.

   note: TICK_TIMER_FREQ and SLEEP_TIMER_FREQ are controlled by PLL / clock
   settings, so must be kept in sync if the clock settings are changed.
*/
const (
	TICK_RATE        = 1000 // 1 KHz
	TICK_TIMER_IRQ   = stm32.IRQ_TIM7
	TICK_TIMER_FREQ  = 32000000 // 32 MHz
	SLEEP_TIMER_IRQ  = stm32.IRQ_TIM3
	SLEEP_TIMER_FREQ = 32000000 // 32 MHz
)

const (
	FlashLatency = stm32.Flash_ACR_LATENCY_WS1
)

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
