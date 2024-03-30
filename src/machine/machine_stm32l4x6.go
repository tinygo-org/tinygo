//go:build stm32l4x6

package machine

// Peripheral abstraction layer for the stm32l4x6

func CPUFrequency() uint32 {
	return 80e6
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 80e6 // 80MHz
const APB2_TIM_FREQ = 80e6 // 80MHz

//---------- I2C related code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange() uint32 {
	// This is a 'magic' value calculated by STM32CubeMX
	// for 80MHz PCLK1.
	// TODO: Do calculations based on PCLK1
	return 0x10909CEC
}
