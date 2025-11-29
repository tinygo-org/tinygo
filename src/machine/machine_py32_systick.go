//go:build py32

package machine

import (
	"device/arm"
)

var msCounter uint64

// CPUFrequency returns the core clock frequency.
func CPUFrequency() uint32 {
	return 4_000_000
}

func Init() {
	LED4.Configure(PinConfig{Mode: PinOutput})
	arm.SetupSystemTimer(CPUFrequency() / 1000)
}

//export SysTick_Handler
func handleSysTick() {
	msCounter = msCounter + 1
	LED4.Set(!LED4.Get())
}

// Ticks returns the number of milliseconds since boot.
func GetTicks() uint64 {
	return msCounter
}
