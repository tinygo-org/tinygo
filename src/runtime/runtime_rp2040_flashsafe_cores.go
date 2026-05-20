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
    rp2040FlashSafeFIFOCommand = 2

    rp2040FlashSafeIdle    = 0
    rp2040FlashSafeLocked  = 1
    rp2040FlashSafeRelease = 2
)

// rp2040FlashSafeState is used to synchronize the core that performs a flash
// operation with the other core that must stop executing from XIP flash.
var rp2040FlashSafeState volatile.Register8

// flashSafeLock serializes Enter/Exit so that only one core at a time
// owns the flash-safe state machine. The other core can still participate
// as a victim through the FIFO interrupt while spinning on this lock.
//
// id: 24 is reserved here. ids 20-23 are already used by printLock,
// schedulerLock, atomicsLock, futexLock (see runtime_rp2.go).
var flashSafeLock = spinLock{id: 24}

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

    flashSafeLock.Lock()

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

        flashSafeLock.Unlock()
    }

    interrupt.Restore(state)
}

func rp2040FlashSafePauseCore(core uint32) {
    _ = core // RP2040 SIO FIFO writes to the other core.
    rp.SIO.FIFO_WR.Set(rp2040FlashSafeFIFOCommand)
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
