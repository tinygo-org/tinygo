//go:build esp32s3 && scheduler.cores

package runtime

import (
	"device"
	"device/esp"
	"internal/task"
	"math/bits"
	"runtime/interrupt"
	"runtime/volatile"
	"sync/atomic"
	"unsafe"
)

const numCPU = 2

// Level 1, level triggered line for the FROM_CPU software interrupts.
// See esp-idf components/esp_hw_support/port/esp32s3/esp_cpu_intr.c.
const crossCoreCPUInterrupt = 3

var (
	printLock     spinLock
	schedulerLock spinLock
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
)

// Set (1) or clear (0) the FROM_CPU interrupt of the given core.
func signalCore(core, value uint32) {
	if core == 0 {
		esp.SYSTEM.CPU_INTR_FROM_CPU_0.Set(value)
	} else {
		esp.SYSTEM.CPU_INTR_FROM_CPU_1.Set(value)
	}
}

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
	if flashPauseRequest[core].Swap(0) != 0 {
		state := interrupt.Disable()
		tinygo_flashPark(&flashParkState)
		interrupt.Restore(state)
	}
}

// Values of flashParkState, shared with tinygo_flashPark in esp32s3.S.
const (
	flashParkIdle    = 0
	flashParkLocked  = 1
	flashParkRelease = 2
)

var (
	flashSafeLock     spinLock
	flashPauseRequest [numCPU]atomic.Uint32
	flashParkState    uint32
)

//export tinygo_flashPark
func tinygo_flashPark(state *uint32)

// Stop the other core in IRAM before a flash erase or write.
// Must not be called from an interrupt or with interrupts disabled.
func esp32s3EnterFlashSafeSection() (interrupt.State, bool) {
	if secondaryCoresReady.Load() == 0 {
		return interrupt.Disable(), false
	}

	flashSafeLock.Lock()

	// Disable interrupts first, like rp2EnterFlashSafeSection. A GC pause here
	// would wait for the other core while that core waits in IRAM.
	state := interrupt.Disable()
	other := 1 - currentCPU()
	flashPauseRequest[other].Store(1)
	signalCore(other, 1)
	for volatile.LoadUint32(&flashParkState) != flashParkLocked {
		spinLoopWait()
	}
	return state, true
}

func esp32s3ExitFlashSafeSection(state interrupt.State, multicore bool) {
	if multicore {
		volatile.StoreUint32(&flashParkState, flashParkRelease)
		for volatile.LoadUint32(&flashParkState) != flashParkIdle {
			spinLoopWait()
		}
		flashSafeLock.Unlock()
	}
	interrupt.Restore(state)
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
	esp.INTERRUPT_CORE0.SetCPU_INTR_FROM_CPU_0_MAP(crossCoreCPUInterrupt)
	esp.INTERRUPT_CORE1.SetCPU_INTR_FROM_CPU_1_MAP(crossCoreCPUInterrupt)
	signalCore(0, 0)
	signalCore(1, 0)
	crossCoreInterrupt = interrupt.New(crossCoreCPUInterrupt, crossCoreHandler)
	crossCoreInterrupt.Enable()

	esp.RTC_CNTL.SetOPTIONS0_SW_STALL_APPCPU_C0(0)
	esp.RTC_CNTL.SetSW_CPU_STALL_SW_STALL_APPCPU_C1(0)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_CLKGATE_EN(1)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RUNSTALL(0)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RESETING(1)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RESETING(0)
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

// Bitset of cores waiting in schedulerUnlockAndWait.
// Must only be accessed with the scheduler lock held.
var waitingCores uint8

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= 1 << core

	// waiti lowers PS.INTLEVEL and waits in one step, so a wake sent after the
	// unlock is not lost. See the Xtensa ISA Reference Manual, WAITI.
	state := interrupt.Disable()
	schedulerLock.Unlock()
	for schedulerWakePending[core].Swap(0) == 0 {
		device.Asm("waiti 0")
		interrupt.Disable()
	}
	interrupt.Restore(state)
	schedulerLock.Lock()
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
