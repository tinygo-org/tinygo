//go:build tinygo.riscv && virt

package runtime

import (
	"device/riscv"
	"internal/task"
	"runtime/volatile"
	"sync/atomic"
	"unsafe"
)

// This file implements the VirtIO RISC-V interface implemented in QEMU, which
// is an interface designed for emulation.

const numCPU = 4

//export main
func main() {
	// Set the interrupt address.
	// Note that this address must be aligned specially, otherwise the MODE bits
	// of MTVEC won't be zero.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))

	// Enable software interrupts. We'll need them to wake up other cores.
	riscv.MIE.SetBits(riscv.MIE_MSIE)

	// If we're not hart 0, wait until we get the signal everything has been set
	// up.
	if hartID := riscv.MHARTID.Get(); hartID != 0 {
		// Wait until we get the signal this hart is ready to start.
		// Note that interrupts are disabled, which means that the interrupt
		// isn't actually taken. But we can still wait for it using wfi.
		// If the cores scheduler is not used, we'll stay in this state forever.
		for riscv.MIP.Get()&riscv.MIP_MSIP == 0 {
			riscv.Asm("wfi")
		}

		// Leave the start interrupt pending. handleInterrupt clears it and then
		// handles a GC pause request that may have arrived with it.
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// The stop-the-world interrupt must be enabled before the
		// secondary core is allowed to enter the scheduler.
		waitForSecondaryCoresReady()

		// Now start running the scheduler on this core.
		schedulerLock.Lock()
		scheduler(false)

		// The scheduler exited, which means main returned and the program
		// should exit immediately.
		// Signal hart 0 to exit.
		exitCodePlusOne.Store(0 + 1) // exit code 0
		signalHart(0)

		// Unlock the scheduler to be sure. Shouldn't be needed.
		schedulerLock.Unlock()

		// Wait until hart 0 actually exits.
		for {
			riscv.Asm("wfi")
		}
	}

	// Enable interrupts, for GC pause and exit requests from other harts.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

	run()
	exit(0)
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint(cause &^ (1 << 31))
	if cause&(1<<31) != 0 {
		// Topmost bit is set, which means that it is an interrupt.
		hartID := currentCPU()
		switch code {
		case riscv.MachineSoftwareInterrupt:
			// Clear the interrupt before checking state so a new request stays pending.
			// See RISC-V Unprivileged ISA, section 2.7.
			aclintMSWI.MSIP[hartID].Set(0)
			riscv.Asm("fence")
			if exitCodePlusOne.Load() != 0 {
				exitNow(exitCodePlusOne.Load() - 1)
			}
			if gcPauseRequest[hartID].Swap(0) != 0 {
				// The GC needs to run.
				gcInterruptHandler(hartID)
			}
			if exitCodePlusOne.Load() != 0 {
				exitNow(exitCodePlusOne.Load() - 1)
			}
		default:
			runtimeFatal("unknown interrupt")
			abort()
		}
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(code)
	}

	// Zero MCAUSE so that it can later be used to see whether we're in an
	// interrupt or not.
	riscv.MCAUSE.Set(0)
}

var (
	// State used to request a GC pause on each hart.
	gcPauseRequest [numCPU]atomic.Uint32

	// State used to signal the next GC phase to each paused hart.
	gcSignalWait [numCPU]atomic.Uint32
)

// The GC interrupted this core for the stop-the-world phase.
// This function handles that, and only returns after the stop-the-world phase
// ended.
func gcInterruptHandler(hartID uint32) {
	// *only* enable the MSIE interrupt
	savedMIE := riscv.MIE.Get()
	riscv.MIE.Set(riscv.MIE_MSIE)

	// Let the GC know we're ready.
	gcScanState.Add(1)

	// Wait until we get a signal to start scanning.
	gcWaitForSignal(hartID)

	// Scan the stack(s) of this core.
	scanCurrentStack()
	if !task.OnSystemStack() {
		// Mark system stack.
		markRoots(task.SystemStack(), coreStackTop(hartID))
	}

	// Signal we've finished scanning.
	gcScanState.Store(1)

	// Wait until we get a signal that the stop-the-world phase has ended.
	gcWaitForSignal(hartID)

	// Restore MIE bits.
	riscv.MIE.Set(savedMIE)

	// Signal we received the signal and are going to exit the interrupt.
	gcScanState.Add(1)
}

func gcWaitForSignal(hartID uint32) {
	for gcSignalWait[hartID].Load() == 0 {
		// Clear unrelated wakeups before checking state to avoid losing a signal.
		// See RISC-V Unprivileged ISA, section 2.7.
		aclintMSWI.MSIP[hartID].Set(0)
		riscv.Asm("fence")
		if hartID == 0 && exitCodePlusOne.Load() != 0 {
			exitNow(exitCodePlusOne.Load() - 1)
		}
		if gcSignalWait[hartID].Load() == 0 {
			riscv.Asm("wfi")
		}
	}
	gcSignalWait[hartID].Store(0)
	aclintMSWI.MSIP[hartID].Set(0)
	riscv.Asm("fence")
}

//go:extern _stack_top
var stack0TopSymbol [0]byte

//go:extern _stack1_top
var stack1TopSymbol [0]byte

//go:extern _stack2_top
var stack2TopSymbol [0]byte

//go:extern _stack3_top
var stack3TopSymbol [0]byte

// Returns the stack top (highest address) of the system stack of the given
// core.
func coreStackTop(core uint32) uintptr {
	switch core {
	case 0:
		return uintptr(unsafe.Pointer(&stack0TopSymbol))
	case 1:
		return uintptr(unsafe.Pointer(&stack1TopSymbol))
	case 2:
		return uintptr(unsafe.Pointer(&stack2TopSymbol))
	case 3:
		return uintptr(unsafe.Pointer(&stack3TopSymbol))
	default:
		runtimeFatal("unexpected core")
		return 0
	}
}

// One tick is 100ns by default in QEMU.
// (This is not a standard, just the default used by QEMU).
func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 100 // one tick is 100ns
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 100) // one tick is 100ns
}

// Set to 1 by interruptSleepTicksMulticore to end a sleep in sleepTicks.
var sleepWake [numCPU]atomic.Uint32

//go:inline
func sleepWoken(core uint32) bool {
	return sleepWake[core].Load() != 0
}

// Currently sleeping core, or 0xff.
// Must only be accessed with the scheduler lock held.
var sleepingCore uint8 = 0xff

// Return whether another core is sleeping.
// May only be called with the scheduler lock held.
func hasSleepingCore() bool {
	return sleepingCore != 0xff
}

// Almost identical to sleepTicks, except that it will unlock/lock the scheduler
// while sleeping and is interruptible by interruptSleepTicksMulticore.
// This may only be called with the scheduler lock held.
func sleepTicksMulticore(d timeUnit) {
	core := currentCPU()
	sleepingCore = uint8(core)
	sleepWake[core].Store(0)

	schedulerLock.Unlock()
	sleepTicks(d)
	schedulerLock.Lock()

	sleepingCore = 0xff
}

// Interrupt an ongoing call to sleepTicksMulticore on another core.
// This may only be called with the scheduler lock held.
func interruptSleepTicksMulticore(wakeup timeUnit) {
	if hasSleepingCore() {
		sleepWake[sleepingCore].Store(1)
	}
}

func ticks() timeUnit {
	// Combining the low bits and the high bits (at a rate of 100ns per tick)
	// yields a time span of over 59930 years without counter rollover.
	highBits := aclintMTIME.high.Get()
	for {
		lowBits := aclintMTIME.low.Get()
		newHighBits := aclintMTIME.high.Get()
		if newHighBits == highBits {
			// High bits stayed the same.
			return timeUnit(lowBits) | (timeUnit(highBits) << 32)
		}
		// Retry, because there was a rollover in the low bits (happening every
		// ~7 days).
		highBits = newHighBits
	}
}

// Memory-mapped I/O as defined by QEMU.
// Source: https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
// Technically this is an implementation detail but hopefully they won't change
// the memory-mapped I/O registers.
var (
	// UART0 output register.
	stdoutWrite = (*volatile.Register8)(unsafe.Pointer(uintptr(0x10000000)))
	// SiFive test finisher
	testFinisher = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))

	// RISC-V Advanced Core Local Interruptor.
	// It is backwards compatible with the SiFive CLINT.
	// https://github.com/riscvarchive/riscv-aclint/blob/main/riscv-aclint.adoc
	aclintMTIME = (*struct {
		low  volatile.Register32
		high volatile.Register32
	})(unsafe.Pointer(uintptr(0x0200_bff8)))
	aclintMSWI = (*struct {
		MSIP [4095]volatile.Register32
	})(unsafe.Pointer(uintptr(0x0200_0000)))
)

func putchar(c byte) {
	stdoutWrite.Set(uint8(c))
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

// Define the various spinlocks needed by the runtime.
var (
	schedulerLock schedulerSpinLock
	futexLock     spinLock
	atomicsLock   spinLock
	printLock     spinLock
)

type spinLock struct {
	atomic.Uint32
}

func (l *spinLock) Lock() {
	// Try to replace 0 with 1. Once we succeed, the lock has been acquired.
	for !l.Uint32.CompareAndSwap(0, 1) {
		spinLoopWait()
	}
}

func (l *spinLock) Unlock() {
	// Safety check: the spinlock should have been locked.
	if schedulerAsserts && l.Uint32.Load() != 1 {
		runtimeFatal("unlock of unlocked spinlock")
	}

	// Unlock the lock. Simply write 0, because we already know it is locked.
	l.Uint32.Store(0)
}

// Hint to the CPU that this core is just waiting, and the core can go into a
// lower energy state.
func spinLoopWait() {
	// This is a no-op in QEMU TCG (but added here for completeness):
	// https://github.com/qemu/qemu/blob/v9.2.3/target/riscv/insn_trans/trans_rvi.c.inc#L856
	riscv.Asm("pause")
}

func currentCPU() uint32 {
	return uint32(riscv.MHARTID.Get())
}

func startSecondaryCores() {
	// Start all the other cores besides hart 0.
	for hart := 1; hart < numCPU; hart++ {
		// Signal the given hart it is ready to start using a software
		// interrupt.
		signalHart(uint32(hart))
	}
}

// Pause the given core by sending it an interrupt.
func gcPauseCore(core uint32) {
	gcPauseRequest[core].Store(1)
	signalHart(core)
}

// Signal the given core that it can resume one step.
// This is called twice after gcPauseCore: the first time to scan the stack of
// the core, and the second time to end the stop-the-world phase.
func gcSignalCore(core uint32) {
	gcSignalWait[core].Store(1)
	signalHart(core)
}

func signalHart(hart uint32) {
	// Order state writes before the interrupt notification.
	// See RISC-V Unprivileged ISA, section 2.7.
	riscv.Asm("fence")
	aclintMSWI.MSIP[hart].Set(1)
}

func abort() {
	exit(1)
}

// Zero in the default state, when non-zero it indicates the exit code plus one.
// So exit(0) will result in 1, exit(1) in 2, etc.
var exitCodePlusOne atomic.Uint32

func exit(code int) {
	// Check for invalid values, to be sure.
	if code < 0 {
		code = 255
	}

	// If we're not on hart 0, we can't exit QEMU.
	// Therefore, send an interrupt to hart 0 instead to request an exit.
	if currentCPU() != 0 {
		// Signal hart 0 to exit.
		exitCodePlusOne.Store(uint32(code) + 1)
		signalHart(0)

		// Wait for the interrupt to happen. This should happen immediately.
		for {
			riscv.Asm("wfi")
		}
	}

	exitNow(uint32(code))
}

// Send an exit signal to the test finisher pseudo-device, without checking
// whether we are on hart 0.
func exitNow(code uint32) {
	// Make sure the QEMU process exits.
	if code == 0 {
		testFinisher.Set(0x5555) // FINISHER_PASS
	} else {
		// Exit code is stored in the upper 16 bits of the 32 bit value.
		testFinisher.Set(code<<16 | 0x3333) // FINISHER_FAIL
	}

	// Lock up forever (as a fallback).
	for {
		riscv.Asm("wfi")
	}
}

// handleException is called from the interrupt handler for any exception.
// Exceptions can be things like illegal instructions, invalid memory
// read/write, and similar issues.
func handleException(code uint) {
	// For a list of exception codes, see:
	// https://content.riscv.org/wp-content/uploads/2019/08/riscv-privileged-20190608-1.pdf#page=49
	print("fatal error: exception with mcause=", code, " pc=", riscv.MEPC.Get(), " mtval=", riscv.MTVAL.Get(), " hart=", uint(riscv.MHARTID.Get()), "\r\n")
	abort()
}
