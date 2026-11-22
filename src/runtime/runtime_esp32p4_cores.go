//go:build esp32p4 && scheduler.cores

package runtime

import (
	"device/esp"
	"device/riscv"
	"internal/task"
	"math/bits"
	"runtime/interrupt"
	"sync/atomic"
	"unsafe"
)

const numCPU = 2

// CPU interrupt line for the FROM_CPU software interrupts, on both cores.
const crossCoreCPUInterrupt = 11

var (
	printLock     spinLock
	schedulerLock schedulerSpinLock
	atomicsLock   spinLock
	futexLock     spinLock
	ticksLock     spinLock
)

type spinLock struct {
	atomic.Uint32
}

func (l *spinLock) Lock() {
	for !l.Uint32.CompareAndSwap(0, 1) {
		spinLoopWait()
	}
}

func (l *spinLock) Unlock() {
	if schedulerAsserts && l.Uint32.Load() != 1 {
		runtimeFatal("unlock of unlocked spinlock")
	}
	l.Uint32.Store(0)
}

func spinLoopWait() {
	riscv.Asm("nop")
}

func currentCPU() uint32 {
	return uint32(riscv.MHARTID.Get())
}

func lockTicks() interrupt.State {
	mask := interrupt.Disable()
	ticksLock.Lock()
	return mask
}

func unlockTicks(mask interrupt.State) {
	ticksLock.Unlock()
	interrupt.Restore(mask)
}

var crossCoreInterrupt interrupt.Interrupt

var (
	core1Started         atomic.Uint32
	schedulerWakePending [numCPU]atomic.Uint32
	sleepWake            [numCPU]atomic.Uint32
	gcPauseRequest       [numCPU]atomic.Uint32
	gcSignalWait         [numCPU]atomic.Uint32
)

// Set (1) or clear (0) the FROM_CPU interrupt of the given core.
func signalCore(core, value uint32) {
	if core == 0 {
		esp.HP_SYS.CPU_INTR_FROM_CPU_0.Set(value)
	} else {
		esp.HP_SYS.CPU_INTR_FROM_CPU_1.Set(value)
	}
}

func crossCoreHandler(interrupt.Interrupt) {
	core := currentCPU()

	// Clear the source before reading the flags, so a new signal makes a new
	// edge.
	signalCore(core, 0)
	if core == 0 {
		if num := core0EnableRequest.Load(); num != 0 {
			interruptEnableLocal(num)
			core0EnableRequest.Store(0)
		}
	}
	if gcPauseRequest[core].Swap(0) != 0 {
		gcInterruptHandler(core)
	}
}

// CPU interrupt line that core 1 asked core 0 to enable, or 0.
var core0EnableRequest atomic.Uint32

// Enable a CPU interrupt line on core 0. Called by interrupt.Enable on core 1,
// which must have interrupts enabled so the GC can still pause it.
func enableInterruptOnCore0(num uint32) {
	for !core0EnableRequest.CompareAndSwap(0, num) {
		spinLoopWait()
	}
	signalCore(0, 1)
	for core0EnableRequest.Load() != 0 {
		spinLoopWait()
	}
}

//go:linkname interruptEnableLocal runtime/interrupt.enableLocal
func interruptEnableLocal(num uint32)

// Route the FROM_CPU interrupt of the calling core and enable it.
func initCrossCoreInterrupt(core uint32) {
	// The interrupt matrix takes the CLIC source number.
	if core == 0 {
		esp.INTERRUPT_CORE0.CPU_INT_FROM_CPU_0_MAP.Set(crossCoreCPUInterrupt + clicExtIntrNumOffset)
	} else {
		esp.INTERRUPT_CORE1.CPU_INT_FROM_CPU_1_MAP.Set(crossCoreCPUInterrupt + clicExtIntrNumOffset)
	}
	signalCore(core, 0)
	if core == 0 {
		// Enable through the handle, so the compiler keeps the handler.
		crossCoreInterrupt.Enable()
	} else {
		interruptEnableLocal(crossCoreCPUInterrupt)
	}
}

func startSecondaryCores() {
	crossCoreInterrupt = interrupt.New(crossCoreCPUInterrupt, crossCoreHandler)
	initCrossCoreInterrupt(0)

	startCore1()

	// The GC can pause core 1 only after it has enabled its interrupt.
	for core1Started.Load() == 0 {
		spinLoopWait()
	}
}

//go:extern call_start_cpu1
var call_start_cpu1 [0]byte

// The function that is started on the second core, from call_start_cpu1.
//
//export tinygo_runCore1
func runCore1() {
	interruptInit()
	initCrossCoreInterrupt(1)
	core1Started.Store(1)

	waitForSecondaryCoresReady()

	schedulerLock.Lock()
	scheduler(false)
	schedulerLock.Unlock()

	exit(0)
}

// Bitset of cores waiting in schedulerUnlockAndWait.
// Must only be accessed with the scheduler lock held.
var waitingCores uint8

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= 1 << core

	// wfi wakes on a pending interrupt even with MSTATUS.MIE clear, so a wake
	// sent after the unlock is not lost. Interrupts are turned on for a moment
	// after each wake, so a GC pause request can run.
	state := schedulerLock.unlockMasked()
	for schedulerWakePending[core].Swap(0) == 0 {
		riscv.Asm("wfi")
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
		riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)
	}
	schedulerLock.lockMasked(state)
}

// Wake another core, if one is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWake() {
	if core := bits.TrailingZeros8(waitingCores); core < numCPU {
		schedulerWakeCore(uint32(core))
	} else if hasSleepingCore() {
		interruptSleepTicksMulticore(0)
	}
}

// Wake the given core if it is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWakeCore(core uint32) {
	if waitingCores&(1<<core) != 0 {
		waitingCores &^= 1 << core
		schedulerWakePending[core].Store(1)
		signalCore(core, 1)
	} else if uint32(sleepingCore) == core {
		interruptSleepTicksMulticore(0)
	}
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
// Must be called with the scheduler lock held.
func interruptSleepTicksMulticore(wakeup timeUnit) {
	if hasSleepingCore() {
		sleepWake[sleepingCore].Store(1)
	}
}

//go:inline
func sleepWoken(core uint32) bool {
	return sleepWake[core].Load() != 0
}

// The GC interrupted this core for the stop-the-world phase.
// This function handles that, and only returns after the stop-the-world phase
// ended.
func gcInterruptHandler(core uint32) {
	// Let the GC know we're ready.
	gcScanState.Add(1)

	// Wait until we get a signal to start scanning.
	gcWaitForSignal(core)

	// Scan the stack(s) of this core.
	scanCurrentStack()
	if !task.OnSystemStack() {
		// Mark system stack.
		markRoots(task.SystemStack(), coreStackTop(core))
	}

	// Signal we've finished scanning.
	gcScanState.Store(1)

	// Wait until we get a signal that the stop-the-world phase has ended.
	gcWaitForSignal(core)

	// Signal we received the signal and are going to exit the interrupt.
	gcScanState.Add(1)
}

func gcWaitForSignal(core uint32) {
	for gcSignalWait[core].Swap(0) == 0 {
		spinLoopWait()
	}
}

// Pause the given core by sending it an interrupt.
func gcPauseCore(core uint32) {
	gcPauseRequest[core].Store(1)
	signalCore(core, 1)
}

// Signal the given core that it can resume one step.
// The paused core polls this flag, so no interrupt is needed.
func gcSignalCore(core uint32) {
	gcSignalWait[core].Store(1)
}

//go:extern _stack1_top
var stack1TopSymbol [0]byte

// Returns the stack top (highest address) of the system stack of the given
// core.
func coreStackTop(core uint32) uintptr {
	switch core {
	case 0:
		return uintptr(unsafe.Pointer(&stackTopSymbol))
	case 1:
		return uintptr(unsafe.Pointer(&stack1TopSymbol))
	default:
		runtimeFatal("unexpected core")
		return 0
	}
}
