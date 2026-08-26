package main

import (
	"device/arm"
	"machine"
	"runtime/interrupt"
	_ "unsafe"
)

//go:linkname setCPUFrequency machine.setCPUFrequency
func setCPUFrequency(frequency uint32)

// The application must establish and stabilize the hardware clock before
// calling clockChanged.
func clockChanged(frequency uint32, uartConfig machine.UARTConfig) error {
	state := interrupt.Disable()

	setCPUFrequency(frequency)
	if err := arm.SetupSystemTimer(frequency / 1000); err != nil {
		interrupt.Restore(state)
		return err
	}
	err := machine.DefaultUART.Configure(uartConfig)
	interrupt.Restore(state)
	return err
}

func main() {
	_ = clockChanged(machine.CPUFrequency(), machine.UARTConfig{})
}
