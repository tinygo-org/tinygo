//go:build py32

package machine

// CPUFrequencyHz is the current CPU frequency in hertz.
//
// Application code that changes the system clock through the RCC registers
// must update this value after the new clock is stable, then reconfigure
// SysTick and any peripherals derived from the system clock. Assigning this
// variable does not configure the clock hardware.
var CPUFrequencyHz uint32 = defaultCPUFrequencyHz

func CPUFrequency() uint32 {
	return CPUFrequencyHz
}
