// +build cortexm

package main

import (
	"device/arm"
)

func callTrap() {
	// Run a trap instruction.
	arm.Asm("trap")
}

func abort() {
	// Special call to exit from QEMU.
	arm.SemihostingCall(arm.SemihostingReportException, arm.SemihostingApplicationExit)
}

//export NMI_Handler
func handleNMI() {
	// This is just a test that handlers can be redefined (they are weak
	// symbols).
	arm.Asm("bkpt #20") // check that the NMI_Handler is indeed set
	for {
	}
}
