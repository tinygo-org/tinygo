//go:build rp2350

package runtime

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"sync/atomic"
)

const (
	// On RP2040 each core has a different IRQ number: SIO_IRQ_PROC0 and SIO_IRQ_PROC1.
	// On RP2350 both cores share the same irq number (SIO_IRQ_PROC) just with a
	// different SIO interrupt output routed to that IRQ input on each core.
	// https://www.raspberrypi.com/documentation/pico-sdk/high_level.html#group_pico_multicore_1ga1413ebfa65114c6f408f4675897ac5ee
	sioIrqFifoProc0 = rp.IRQ_SIO_IRQ_FIFO
	sioIrqFifoProc1 = rp.IRQ_SIO_IRQ_FIFO
)

var sioFifoInterrupt = interrupt.New(sioIrqFifoProc0, handleSIOFifoInterrupt)

// On RP2350 both cores share IRQ_SIO_IRQ_FIFO, but the NVIC enable and
// priority state is per-core. Each core must enable the shared IRQ on
// its own NVIC, so both Core0 and Core1 call Enable()/SetPriority().
func enableSIOFifoInterruptCore0() {
	sioFifoInterrupt.Enable()
	sioFifoInterrupt.SetPriority(0xff)
}

func enableSIOFifoInterruptCore1() {
	sioFifoInterrupt.Enable()
	sioFifoInterrupt.SetPriority(0xff)
}

func handleSIOFifoInterrupt(intr interrupt.Interrupt) {
	switch rp.SIO.FIFO_RD.Get() {
	case 1:
		gcInterruptHandler(currentCPU())
	}
}

// These spinlocks are needed by the runtime.
//
// Unlike on the RP2040, these are software spinlocks: erratum RP2350-E2 makes
// the hardware SIO spinlocks unreliable (a core can observe a spinlock as
// acquired when it isn't). The pico-sdk likewise replaced them with LL/SC
// based software spinlocks on this chip. The Cortex-M33 supports ldrex/strex,
// so 32-bit atomics compile to native instructions and are multicore-safe.
var (
	printLock     spinLock
	schedulerLock spinLock
	atomicsLock   spinLock
	futexLock     spinLock
)

// Software spinlocks don't survive in locked state across a soft reset: they
// live in .bss, which preinit() has already cleared before this is called.
func resetSpinLocks() {
}

// A software spinlock, implemented using atomic operations.
type spinLock struct {
	atomic.Uint32
}

func (l *spinLock) Lock() {
	// Try to replace 0 with 1. Once we succeed, the lock has been acquired.
	for !l.Uint32.CompareAndSwap(0, 1) {
		// Wait until the current holder calls Unlock, which executes a "sev".
		arm.Asm("wfe")
	}
}

func (l *spinLock) Unlock() {
	// Safety check: the spinlock should have been locked.
	if schedulerAsserts && l.Uint32.Load() != 1 {
		runtimePanic("unlock of unlocked spinlock")
	}

	// Unlock the lock. Simply write 0, because we already know it is locked.
	l.Uint32.Store(0)

	// Wake up cores that are waiting for this lock in Lock() or for some other
	// event (like a runnable task) in schedulerUnlockAndWait. The hardware
	// spinlock version on the RP2040 has the same behavior.
	arm.Asm("sev")
}
