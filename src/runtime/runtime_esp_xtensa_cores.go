//go:build (esp32 || esp32s3) && scheduler.cores

package runtime

import (
	"device"
	"device/esp"
	"internal/task"
	"runtime/interrupt"
	"sync/atomic"
	"unsafe"
)

const numCPU = 2

// Level 1, level triggered line for the FROM_CPU software interrupts. See
// XCHAL_INT3_LEVEL in esp-idf components/xtensa/<chip>/include/xtensa/config/core-isa.h.
const crossCoreCPUInterrupt = 3

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
	device.Asm("nop")
}

// Return the current core number: 0 or 1. PRID bit 13 is the core ID, see
// xt_utils_get_raw_core_id in esp-idf components/xtensa/include/xt_utils.h.
func currentCPU() uint32 {
	return uint32(device.AsmFull("rsr {}, PRID", nil)>>13) & 1
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
	coreIdle             [numCPU]atomic.Uint32
)

func crossCoreHandler(interrupt.Interrupt) {
	core := currentCPU()

	// The source is level triggered, so clear it before reading the flags.
	signalCore(core, 0)
	if core == 0 {
		if mask := core0EnableRequest.Load(); mask != 0 {
			interruptEnableLocal(mask)
			core0EnableRequest.And(^mask)
		}
	}
	if gcPauseRequest[core].Swap(0) != 0 {
		gcInterruptHandler(core)
	}
	crossCoreChipHandler(core)
}

// CPU interrupt lines that core 1 asked core 0 to enable.
var core0EnableRequest atomic.Uint32

// Enable a CPU interrupt line on core 0. Called by interrupt.Enable on core 1,
// which must have interrupts enabled so the GC can still pause it.
func enableInterruptOnCore0(num uint32) {
	bit := uint32(1) << num
	core0EnableRequest.Or(bit)
	signalCore(0, 1)
	for core0EnableRequest.Load()&bit != 0 {
		spinLoopWait()
	}
}

//go:linkname interruptEnableLocal runtime/interrupt.enableLocal
func interruptEnableLocal(mask uint32)

// Start core 1, following start_other_core in esp-idf
// components/esp_system/port/cpu_start.c.
func startSecondaryCores() {
	initCore1()
	signalCore(0, 0)
	signalCore(1, 0)
	crossCoreInterrupt = interrupt.New(crossCoreCPUInterrupt, crossCoreHandler)
	crossCoreInterrupt.Enable()

	esp.RTC_CNTL.SetOPTIONS0_SW_STALL_APPCPU_C0(0)
	esp.RTC_CNTL.SetSW_CPU_STALL_SW_STALL_APPCPU_C1(0)
	resetCore1()
	ets_set_appcpu_boot_addr(uintptr(unsafe.Pointer(&call_start_cpu1)))

	// The GC can pause core 1 only after it has enabled its interrupt.
	for core1Started.Load() == 0 {
		spinLoopWait()
	}
}

//export ets_set_appcpu_boot_addr
func ets_set_appcpu_boot_addr(addr uintptr)

//go:extern call_start_cpu1
var call_start_cpu1 [0]byte

// The function that is started on the second core, from call_start_cpu1.
//
//export tinygo_runCore1
func runCore1() {
	interruptInit()
	interruptEnableLocal(1 << crossCoreCPUInterrupt)
	core1Started.Store(1)

	waitForSecondaryCoresReady()

	schedulerLock.Lock()
	scheduler(false)
	schedulerLock.Unlock()

	exit(0)
}

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= 1 << core

	// waiti lowers PS.INTLEVEL and waits in one step, so a wake sent after the
	// unlock is not lost. See the Xtensa ISA Reference Manual, WAITI.
	// Interrupts stay off from the unlock until the lock is taken again.
	state := schedulerLock.unlockMasked()
	coreIdle[core].Store(1)
	for schedulerWakePending[core].Swap(0) == 0 {
		device.Asm("waiti 0")
		interrupt.Disable()
	}
	coreIdle[core].Store(0)
	schedulerLock.lockMasked(state)
}

// End schedulerUnlockAndWait on the given core.
func schedulerEndWait(core uint32) {
	schedulerWakePending[core].Store(1)
	signalCore(core, 1)
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
	coreIdle[core].Store(1)
	sleepTicks(d)
	coreIdle[core].Store(0)
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

	// An idle core holds no live pointers, and its stale stack slots and
	// registers would keep exited goroutines and their stacks alive.
	if coreIdle[core].Load() == 0 {
		// Scan the stack(s) of this core.
		scanCurrentStack()
		if !task.OnSystemStack() {
			// Mark system stack.
			markRoots(task.SystemStack(), coreStackTop(core))
		}
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
