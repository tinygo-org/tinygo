//go:build rp2040 && scheduler.cores

package runtime

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"runtime/volatile"
	_ "unsafe" // required for //go:section
)

const (
	rp2040FlashSafeIdle uint8 = iota
	rp2040FlashSafeLocked
	rp2040FlashSafeRelease
)

// rp2040FlashSafeState is used to synchronize the core that performs a flash
// operation with the other core that must stop executing from XIP flash.
var rp2040FlashSafeState volatile.Register8

// rp2040EnterFlashSafeSection enters a section in which RP2040 flash operations
// may temporarily disable XIP.
//
// The multicore path serializes flash-safe initiators, then disables local
// interrupts before asking the other core to park. Keeping local interrupts
// disabled while waiting for the acknowledgement prevents a GC stop-the-world
// interrupt from blocking this core while the other core is parked in the
// flash-safe handler.
func rp2040EnterFlashSafeSection() interrupt.State {
	if !secondaryCoresStarted {
		return interrupt.Disable()
	}

	flashSafeLock.Lock()

	state := interrupt.Disable()

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

	return state
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

		flashSafeLock.Unlock()
	}

	interrupt.Restore(state)
}

func rp2040FlashSafePauseCore(core uint32) {
	// RP2040 SIO FIFO writes to the other core.
	rp.SIO.FIFO_WR.Set(rp2SIOFIFOCommandFlashSafe)
	arm.Asm("sev")
}

// rp2FlashSafeInterruptHandler runs on the other core while this core is
// performing a flash operation that temporarily disables XIP.
//
// This function MUST be placed in RAM (.ramfuncs section). During the
// flash operation the QSPI flash is in non-XIP mode and instruction
// fetches from the 0x10000000 region will fail. The wait loop below
// runs entirely from RAM so that the parked core can keep executing.
//
//go:section .ramfuncs
func rp2FlashSafeInterruptHandler(core uint32) {
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
