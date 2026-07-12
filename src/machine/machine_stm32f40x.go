//go:build stm32f4 && (stm32f405 || stm32f407)

package machine

func CPUFrequency() uint32 {
	pll := PLLParams168MHz()
	return xtalHz / pll.M * pll.N / pll.P
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 42000000 * 2
const APB2_TIM_FREQ = 84000000 * 2
