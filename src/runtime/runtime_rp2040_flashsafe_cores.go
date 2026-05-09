//go:build rp2040 && scheduler.cores

package runtime

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"runtime/volatile"
)

const (
	rp2040FlashSafeFIFOCommand = 2

	rp2040FlashSafeIdle    = 0
	rp2040FlashSafeLocked  = 1
	rp2040FlashSafeRelease = 2
)

// rp2040FlashSafeState is used to synchronize the core that performs a flash
// operation with the other core that must stop executing from XIP flash.
var rp2040FlashSafeState volatile.Register8

// rp2040EnterFlashSafeSection enters a section in which RP2040 flash operations
// may temporarily disable XIP.
//
// The multicore path asks the other core to enter the flash-safe interrupt
// handler and waits until it acknowledges that it is parked. Local interrupts
// are disabled after the other core is parked.
func rp2040EnterFlashSafeSection() interrupt.State {
	if !secondaryCoresStarted {
		return interrupt.Disable()
	}

	core := currentCPU()
	rp2040FlashSafeState.Set(rp2040FlashSafeIdle)

	for i := uint32(0); i < numCPU; i++ {
		if i == core {
			continue
		}
		rp2040FlashSafePauseCore(i)
	}

	for rp2040FlashSafeState.Get() != rp2040FlashSafeLocked {
		spinLoopWait()
	}

	return interrupt.Disable()
}

// rp2040ExitFlashSafeSection exits a section entered by
// rp2040EnterFlashSafeSection.
func rp2040ExitFlashSafeSection(state interrupt.State) {
	if secondaryCoresStarted {
		rp2040FlashSafeState.Set(rp2040FlashSafeRelease)
		arm.Asm("sev")

		for rp2040FlashSafeState.Get() != rp2040FlashSafeIdle {
			spinLoopWait()
		}
	}

	interrupt.Restore(state)
}

func rp2040FlashSafePauseCore(core uint32) {
	_ = core // RP2040 SIO FIFO writes to the other core.
	rp.SIO.FIFO_WR.Set(rp2040FlashSafeFIFOCommand)
	arm.Asm("sev")
}

// rp2FlashSafeInterruptHandler is called from the SIO FIFO interrupt handler.
//
// NOTE: this first draft keeps the lockout handler in Go. The final version
// should ensure that the wait loop runs from RAM while XIP is disabled.
func rp2FlashSafeInterruptHandler(core uint32) {
	_ = core

	state := interrupt.Disable()

	rp2040FlashSafeState.Set(rp2040FlashSafeLocked)
	arm.Asm("sev")

	for rp2040FlashSafeState.Get() == rp2040FlashSafeLocked {
		arm.Asm("wfe")
	}

	interrupt.Restore(state)

	rp2040FlashSafeState.Set(rp2040FlashSafeIdle)
	arm.Asm("sev")
}
