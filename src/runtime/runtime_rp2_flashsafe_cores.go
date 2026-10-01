//go:build (rp2040 || rp2350) && scheduler.cores

package runtime

import (
	"device/arm"
	"runtime/interrupt"
	"runtime/volatile"
	_ "unsafe" // required for //go:section
)

// Values of rp2FlashSafeState.
const (
	rp2FlashSafeIdle    uint8 = iota // not in a flash-safe section, or the other core has resumed
	rp2FlashSafeLocked               // the other core is waiting in RAM
	rp2FlashSafeRelease              // flash operation complete; the other core may resume
)

// rp2FlashSafeState synchronizes both cores during flash operations.
var rp2FlashSafeState volatile.Register8

// rp2EnterFlashSafeSection enters a section where flash operations may disable XIP.
// With scheduler=cores it must not be called from an interrupt handler or with
// interrupts disabled; the GC stop-the-world path has the same constraint (see #5610).
func rp2EnterFlashSafeSection() (interrupt.State, bool) {
	multicore := secondaryCoresReady.Load() != 0
	if !multicore {
		return interrupt.Disable(), false
	}

	flashSafeLock.Lock()

	// Disable local interrupts before the handshake. A GC interrupt here would
	// block this core while the other core is parked.
	state := interrupt.Disable()

	rp2FlashSafeState.Set(rp2FlashSafeIdle)

	// RP2040 and RP2350 have two cores, so there is exactly one core to pause.
	rp2FlashSafePauseCore()

	for rp2FlashSafeState.Get() != rp2FlashSafeLocked {
		spinLoopWait()
	}

	return state, true
}

func rp2ExitFlashSafeSection(state interrupt.State, multicore bool) {
	if multicore {
		rp2FlashSafeState.Set(rp2FlashSafeRelease)
		arm.Asm("sev")

		for rp2FlashSafeState.Get() != rp2FlashSafeIdle {
			spinLoopWait()
		}

		flashSafeLock.Unlock()
	}

	interrupt.Restore(state)
}

func rp2FlashSafePauseCore() {
	multicore_fifo_push_blocking(rp2SIOFIFOCommandFlashSafe)
}

// rp2FlashSafeInterruptHandler waits in RAM with interrupts disabled
// while XIP is unavailable.

// See RP2040 datasheet section 2.6.3 for XIP access during flash operations.
//
//go:section .ramfuncs
func rp2FlashSafeInterruptHandler() {
	state := interrupt.Disable()

	rp2FlashSafeState.Set(rp2FlashSafeLocked)
	arm.Asm("sev")

	for rp2FlashSafeState.Get() == rp2FlashSafeLocked {
		arm.Asm("wfe")
	}

	// Set Idle before restoring interrupts to avoid deadlocking with
	// a pending GC interrupt.
	rp2FlashSafeState.Set(rp2FlashSafeIdle)
	arm.Asm("sev")

	interrupt.Restore(state)
}
