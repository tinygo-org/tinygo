package main

import (
	"machine"
	"runtime"
	"runtime/interrupt"
)

// The application must establish and stabilize the hardware clock before
// calling clockChanged.
func clockChanged(frequency uint32, uartConfig machine.UARTConfig) error {
	state := interrupt.Disable()

	machine.CPUFrequencyHz = frequency
	runtime.ConfigureSystemTimer()
	err := machine.DefaultUART.Configure(uartConfig)
	interrupt.Restore(state)
	return err
}

func main() {
	_ = clockChanged(machine.CPUFrequency(), machine.UARTConfig{})
}
