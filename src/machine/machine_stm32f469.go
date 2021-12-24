//go:build stm32f4 && stm32f469
// +build stm32f4,stm32f469

package machine

func CPUFrequency() uint32 {
	return 180000000
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 45000000 * 2
const APB2_TIM_FREQ = 90000000 * 2
