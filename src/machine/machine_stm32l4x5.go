//go:build stm32l4x5

package machine

// Peripheral abstraction layer for the stm32l4x5

func CPUFrequency() uint32 {
	return 120e6
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 120e6 // 120MHz
const APB2_TIM_FREQ = 120e6 // 120MHz

//---------- I2C related code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange(br uint32) uint32 {
	// This is a 'magic' value calculated by STM32CubeMX
	// for 120MHz PCLK1.
	// TODO: Do calculations based on PCLK1
	switch br {
	case 10 * KHz:
		return 0x0 // does this even work? zero is weird here.
	case 100 * KHz:
		return 0x307075B1
	case 400 * KHz:
		return 0x00B03FDB
	case 500 * KHz:
		return 0x005017C7
	default:
		return 0
	}
}
