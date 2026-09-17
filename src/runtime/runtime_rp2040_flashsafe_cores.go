//go:build rp2040 && scheduler.cores

package runtime

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"runtime/volatile"
	_ "unsafe" // required for //go:section
)

// Values of rp2040FlashSafeState.
const (
	rp2040FlashSafeIdle    uint8 = iota // not in a flash-safe section, or the other core has resumed
	rp2040FlashSafeLocked               // the other core is waiting in RAM
	rp2040FlashSafeRelease              // flash operation complete; the other core may resume
)

// rp2040FlashSafeState synchronizes both cores during flash operations.
var rp2040FlashSafeState volatile.Register8

// rp2040EnterFlashSafeSection enters a section where flash operations may disable XIP.
// It must not be called from an interrupt handler or with interrupts disabled.
func rp2040EnterFlashSafeSection() (interrupt.State, bool) {
	multicore := secondaryCoresStarted
	if !multicore {
		return interrupt.Disable(), false
	}

	flashSafeLock.Lock()

	state := interrupt.Disable()

	core := currentCPU()
	rp2040FlashSafeState.Set(rp2040FlashSafeIdle)

	for i := uint32(0); i < numCPU; i++ {
		if i == core {
			continue
		}
		rp2040FlashSafePauseCore()
	}

	for rp2040FlashSafeState.Get() != rp2040FlashSafeLocked {
		spinLoopWait()
	}

	return state, true
}

func rp2040ExitFlashSafeSection(state interrupt.State, multicore bool) {
	if multicore {
		rp2040FlashSafeState.Set(rp2040FlashSafeRelease)
		arm.Asm("sev")

		for rp2040FlashSafeState.Get() != rp2040FlashSafeIdle {
			spinLoopWait()
		}

		flashSafeLock.Unlock()
	}

	interrupt.Restore(state)
}

func rp2040FlashSafePauseCore() {
	// Values written to FIFO_WR are received by the other core.
	rp.SIO.FIFO_WR.Set(rp2SIOFIFOCommandFlashSafe)
	arm.Asm("sev")
}

// rp2FlashSafeInterruptHandler waits in RAM with interrupts disabled
// while XIP is unavailable.
// See RP2040 datasheet section 2.6.3 for XIP access during flash operations.
//
//go:section .ramfuncs
func rp2FlashSafeInterruptHandler() {
	state := interrupt.Disable()

	rp2040FlashSafeState.Set(rp2040FlashSafeLocked)
	arm.Asm("sev")

	for rp2040FlashSafeState.Get() == rp2040FlashSafeLocked {
		arm.Asm("wfe")
	}

	// Set Idle before restoring interrupts to avoid deadlocking with
	// a pending GC interrupt.
	rp2040FlashSafeState.Set(rp2040FlashSafeIdle)
	arm.Asm("sev")

	interrupt.Restore(state)
}
