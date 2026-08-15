//go:build stm32f4 && stm32f401

package machine

// CPUFrequency returns the current CPU frequency of the STM32F401.
// The PLL is configured to 84MHz SYSCLK from an external crystal.
func CPUFrequency() uint32 {
	pll := PLLParams84MHz()
	return xtalHz / pll.M * pll.N / pll.P
}

// Internal use: configured speed of the APB1 and APB2 timers.
// STM32F401 at 84MHz: SYSCLK=84MHz, AHB=84MHz, APB1=42MHz, APB2=84MHz.
// APB1 prescaler = 2, so APB1 timer clock = 42MHz × 2 = 84MHz.
// APB2 prescaler = 1, so APB2 timer clock = 84MHz × 1 = 84MHz.
const APB1_TIM_FREQ = 84_000_000
const APB2_TIM_FREQ = 84_000_000
