//go:build tinygo.riscv && virt && qemu

package runtime

import (
	"device/riscv"
	"internal/task"
	"math/bits"
	"runtime/interrupt"
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

		// Clear the software interrupt.
		aclintMSWI.MSIP[hartID].Set(0)

		// Now that we've cleared the software interrupt, we can enable
		// interrupts as was already done on hart 0.
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// Also enable timer interrupts, for sleepTicksMulticore.
		riscv.MIE.SetBits(riscv.MIE_MTIE)

		// Now start running the scheduler on this core.
		schedulerLock.Lock()
		scheduler(false)

		// The scheduler exited, which means main returned and the program
		// should exit immediately.
		// Signal hart 0 to exit.
		exitCodePlusOne.Store(0 + 1) // exit code 0
		aclintMSWI.MSIP[0].Set(1)

		// Unlock the scheduler to be sure. Shouldn't be needed.
		schedulerLock.Unlock()

		// Wait until hart 0 actually exits.
		for {
			riscv.Asm("wfi")
		}
	}

	// Enable global interrupts now that they've been set up.
	// This is currently only for timer interrupts.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

	// Set all MTIMECMP registers to a value that clears the MTIP bit in MIP.
	// If we don't do this, the wfi instruction won't work as expected.
	for i := 0; i < numCPU; i++ {
		aclintMTIMECMP[i].Set(0xffff_ffff_ffff_ffff)
	}

	// Enable timer interrupts on hart 0.
	riscv.MIE.SetBits(riscv.MIE_MTIE)

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
			if exitCodePlusOne.Load() != 0 {
				exitNow(exitCodePlusOne.Load() - 1)
			}
			if gcScanState.Load() != 0 {
				// The GC needs to run.
				gcInterruptHandler(hartID)
			}
			checkpoint := &schedulerWaitCheckpoints[hartID]
			if checkpoint.Saved() {
				aclintMSWI.MSIP[hartID].Set(0)
				riscv.MCAUSE.Set(0)
				checkpoint.Jump()
			}
		case riscv.MachineTimerInterrupt:
			if sleepCheckpoint.Saved() {
				// Set MTIMECMP to a high value so that MTIP goes low.
				aclintMTIMECMP[hartID].Set(0xffff_ffff_ffff_ffff)
				riscv.MCAUSE.Set(0)
				sleepCheckpoint.Jump()
			}
		default:
			runtimePanic("unknown interrupt")
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

// The GC interrupted this core for the stop-the-world phase.
// This function handles that, and only returns after the stop-the-world phase
// ended.
func gcInterruptHandler(hartID uint32) {
	// *only* enable the MSIE interrupt
	savedMIE := riscv.MIE.Get()
	riscv.MIE.Set(riscv.MIE_MSIE)

	// Disable this interrupt (to be enabled again soon).
	aclintMSWI.MSIP[hartID].Set(0)

	// Let the GC know we're ready.
	gcScanState.Add(1)

	// Wait until we get a signal to start scanning.
	for riscv.MIP.Get()&riscv.MIP_MSIP == 0 {
		riscv.Asm("wfi")
	}
	aclintMSWI.MSIP[hartID].Set(0)

	// Scan the stack(s) of this core.
	scanCurrentStack()
	if !task.OnSystemStack() {
		// Mark system stack.
		markRoots(task.SystemStack(), coreStackTop(hartID))
	}

	// Signal we've finished scanning.
	gcScanState.Store(1)

	// Wait until we get a signal that the stop-the-world phase has ended.
	for riscv.MIP.Get()&riscv.MIP_MSIP == 0 {
		riscv.Asm("wfi")
	}
	aclintMSWI.MSIP[hartID].Set(0)

	// Restore MIE bits.
	riscv.MIE.Set(savedMIE)

	// Signal we received the signal and are going to exit the interrupt.
	gcScanState.Add(1)
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
		runtimePanic("unexpected core")
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

var sleepCheckpoint interrupt.Checkpoint

func sleepTicks(d timeUnit) {
	hartID := currentCPU()
	if sleepCheckpoint.Save() {
		// Configure timeout.
		target := uint64(ticks() + d)
		aclintMTIMECMP[hartID].Set(target)

		// Wait for the interrupt to happen.
		for {
			riscv.Asm("wfi")
		}
	}

	// We got awoken.
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
	// Disable interrupts while configuring sleep.
	// This is needed because unlocking the scheduler and setting the timer
	// interrupt need to happen atomically.
	riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)

	hartID := currentCPU()
	if sleepCheckpoint.Save() {
		sleepingCore = uint8(hartID)

		// Configure timeout.
		target := uint64(ticks() + d)
		aclintMTIMECMP[hartID].Set(target)

		// Unlock, now that the timeout has been set (so that
		// interruptSleepTicksMulticore will see the correct wakeup time).
		schedulerLock.Unlock()

		// Sleep has been configured, interrupts may happen again.
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// Wait for the interrupt to happen.
		for {
			riscv.Asm("wfi")
		}
	}
	// We got awoken.

	// Lock again, after we finished sleeping.
	schedulerLock.Lock()
	sleepingCore = 0xff
}

// Interrupt an ongoing call to sleepTicksMulticore on another core.
// This may only be called with the scheduler lock held.
func interruptSleepTicksMulticore(wakeup timeUnit) {
	if sleepingCore != 0xff {
		// Immediately exit the sleep.
		old := aclintMTIMECMP[sleepingCore].Get()
		if uint64(wakeup) < old {
			aclintMTIMECMP[sleepingCore].Set(uint64(wakeup))
		}
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
	aclintMTIMECMP = (*[4095]volatile.Register64)(unsafe.Pointer(uintptr(0x0200_4000)))
	aclintMSWI     = (*struct {
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
	schedulerLock spinLock
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
		runtimePanic("unlock of unlocked spinlock")
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
		aclintMSWI.MSIP[hart].Set(1)
	}
}

// Bitset of harts that are currently sleeping in schedulerUnlockAndWait.
// This supports up to 8 harts.
// This variable may only be accessed with the scheduler lock held.
var sleepingHarts uint8

// Checkpoints for cores waiting for runnable tasks.
var schedulerWaitCheckpoints [numCPU]interrupt.Checkpoint

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	hartID := currentCPU()

	// Mark the current hart as sleeping.
	sleepingHarts |= uint8(1 << hartID)

	// If this is the last core awake and is going to sleep, the scheduler is
	// deadlocked.
	// We can do this check since this is not baremetal: there won't be any
	// external interrupts that might unblock a goroutine.
	if sleepingHarts == (1<<numCPU)-1 {
		runtimePanic("all cores are sleeping - deadlock!")
	}

	// Need to disable interrupts while saving the checkpoint, otherwise if the
	// software interrupt happens earlier for another reason (e.g. a GC cycle)
	// it will see an incomplete checkpoint and the schedulerLock might not be
	// unlocked yet. That will lead to an invalid state.
	riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)
	if schedulerWaitCheckpoints[hartID].Save() {
		schedulerLock.Unlock()
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// Wait until we get awoken :)
		for {
			riscv.Asm("wfi")
		}
	}

	// We got awoken again. We need to lock the scheduler again before
	// returning.
	schedulerLock.Lock()
}

// Wake another core, if one is sleeping. Must be called with the scheduler lock
// held.
func schedulerWake() {
	// Look up the lowest-numbered hart that is sleeping.
	// Returns 8 if there are no sleeping harts.
	hart := bits.TrailingZeros8(sleepingHarts)

	if hart < 8 {
		// There is a sleeping hart. Wake it.
		sleepingHarts &^= 1 << hart  // clear the bit
		aclintMSWI.MSIP[hart].Set(1) // send software interrupt
	}
}

// Pause the given core by sending it an interrupt.
func gcPauseCore(core uint32) {
	aclintMSWI.MSIP[core].Set(1) // send software interrupt
}

// Signal the given core that it can resume one step.
// This is called twice after gcPauseCore: the first time to scan the stack of
// the core, and the second time to end the stop-the-world phase.
func gcSignalCore(core uint32) {
	aclintMSWI.MSIP[core].Set(1) // send software interrupt
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
		aclintMSWI.MSIP[0].Set(1)

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
	print("fatal error: exception with mcause=", code, " pc=", riscv.MEPC.Get(), " hart=", uint(riscv.MHARTID.Get()), "\r\n")
	abort()
}
