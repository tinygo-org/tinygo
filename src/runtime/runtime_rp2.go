//go:build rp2040 || rp2350

package runtime

import (
	"device/arm"
	"device/rp"
	"internal/task"
	"machine"
	_ "machine/usb/cdc"
	"math/bits"
	"runtime/volatile"
	"unsafe"
)

// Commands sent to the other core through the SIO FIFO. They start at 1.
const (
	rp2SIOFIFOCommandGC uint32 = iota + 1
	rp2SIOFIFOCommandFlashSafe
)

const numCPU = 2
const numSpinlocks = 32

// machineTicks is provided by package machine.
func machineTicks() uint64

// machineLightSleep is provided by package machine.
func machineLightSleep(uint64)

// ticks returns the number of ticks (microseconds) elapsed since power up.
func ticks() timeUnit {
	t := machineTicks()
	return timeUnit(t)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func sleepTicks(d timeUnit) {
	if hasScheduler {
		// With scheduler, sleepTicks may return early if an interrupt or
		// event fires - so scheduler can schedule any go routines now
		// eligible to run
		machineLightSleep(uint64(d))
		return
	}

	// Busy loop
	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
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
	sleepingCore = uint8(currentCPU())

	// Note: interruptSleepTicksMulticore will be able to interrupt this, since
	// it executes the "sev" instruction which would make sleepTicks return
	// immediately without sleeping. Even if it happens while configuring the
	// sleep operation.

	schedulerLock.Unlock()
	sleepTicks(d)
	schedulerLock.Lock()

	sleepingCore = 0xff
}

// Interrupt an ongoing call to sleepTicksMulticore on another core.
func interruptSleepTicksMulticore(wakeup timeUnit) {
	arm.Asm("sev")
}

// Bitset of cores waiting in schedulerUnlockAndWait.
// Must only be accessed with the scheduler lock held.
var waitingCores uint8

// Set to 1 by schedulerWakeCore to end a wait in schedulerUnlockAndWait.
var schedulerWakePending [2]uint32

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= 1 << core
	schedulerWakePending[core] = 0
	schedulerLock.Unlock()

	// The sev in Unlock also sets the local event register, so one wfe returns at once.
	// See the Arm Cortex-M33 Devices Generic User Guide (100235) section 4.13.11.
	for volatile.LoadUint32(&schedulerWakePending[core]) == 0 {
		arm.Asm("wfe")
	}
	schedulerLock.Lock()
}

// Wake another core, if one is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWake() {
	if waitingCores != 0 {
		schedulerWakeCore(uint32(bits.TrailingZeros8(waitingCores)))
	} else if hasSleepingCore() {
		interruptSleepTicksMulticore(0)
	}
}

// Wake the given core if it is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWakeCore(core uint32) {
	if waitingCores&(1<<core) != 0 {
		waitingCores &^= 1 << core
		volatile.StoreUint32(&schedulerWakePending[core], 1)
		arm.Asm("sev")
	} else if uint32(sleepingCore) == core {
		interruptSleepTicksMulticore(0)
	}
}

// Return the current core number: 0 or 1.
func currentCPU() uint32 {
	return rp.SIO.CPUID.Get()
}

// Start the secondary cores for this chip.
// On the RP2040/RP2350, there is only one other core to start.
func startSecondaryCores() {
	// Start the second core of the RP2040/RP2350.
	// See sections 2.8.2 and 5.3 in the datasheets for RP2040 and RP2350 respectively.
	seq := 0
	for {
		cmd := core1StartSequence[seq]
		if cmd == 0 {
			multicore_fifo_drain()
			arm.Asm("sev")
		}
		multicore_fifo_push_blocking(cmd)
		response := multicore_fifo_pop_blocking()
		if cmd != response {
			seq = 0
			continue
		}
		seq = seq + 1
		if seq >= len(core1StartSequence) {
			break
		}
	}

	// Enable the FIFO interrupt for the GC stop the world phase.
	// We can only do this after we don't need the FIFO anymore for starting the
	// second core.
	enableSIOFifoInterruptCore0()
}

var core1StartSequence = [...]uint32{
	0, 0, 1,
	uint32(uintptr(unsafe.Pointer(&__isr_vector))),
	uint32(uintptr(unsafe.Pointer(&stack1TopSymbol))),
	uint32(exportedFuncPtr(runCore1)),
}

//go:extern __isr_vector
var __isr_vector [0]uint32

//go:extern _stack1_top
var stack1TopSymbol [0]uint32

// The function that is started on the second core.
//
//export tinygo_runCore1
func runCore1() {
	initCore()

	// Clear the sticky bits that were set while starting this core.
	rp.SIO.FIFO_ST.Set(rp.SIO_FIFO_ST_ROE | rp.SIO_FIFO_ST_WOF)

	// Enable the FIFO interrupt, mainly used for the stop-the-world phase of
	// the GC.
	// Use the lowest possible priority (highest priority value), so that other
	// interrupts can still happen while the GC is running.
	enableSIOFifoInterruptCore1()

	// The stop-the-world interrupt must be enabled before the
	// secondary core is allowed to enter the scheduler.
	waitForSecondaryCoresReady()

	// Now start running the scheduler on this core.
	schedulerLock.Lock()
	scheduler(false)
	schedulerLock.Unlock()

	// The main function returned.
	exit(0)
}

// The below multicore_fifo_* functions have been translated from the Raspberry
// Pi Pico SDK.

func multicore_fifo_rvalid() bool {
	return rp.SIO.FIFO_ST.Get()&rp.SIO_FIFO_ST_VLD != 0
}

func multicore_fifo_wready() bool {
	return rp.SIO.FIFO_ST.Get()&rp.SIO_FIFO_ST_RDY != 0
}

func multicore_fifo_drain() {
	for multicore_fifo_rvalid() {
		rp.SIO.FIFO_RD.Get()
	}
}

func multicore_fifo_push_blocking(data uint32) {
	for !multicore_fifo_wready() {
	}
	rp.SIO.FIFO_WR.Set(data)
	arm.Asm("sev")
}

func multicore_fifo_pop_blocking() uint32 {
	for !multicore_fifo_rvalid() {
		arm.Asm("wfe")
	}

	return rp.SIO.FIFO_RD.Get()
}

// Value used to communicate between the GC core and the other (paused) cores.
var gcSignalWait volatile.Register8

// The GC interrupted this core for the stop-the-world phase.
// This function handles that, and only returns after the stop-the-world phase
// ended.
func gcInterruptHandler(hartID uint32) {
	// Let the GC know we're ready.
	gcScanState.Add(1)
	arm.Asm("sev")

	// Wait until we get a signal to start scanning.
	for gcSignalWait.Get() == 0 {
		arm.Asm("wfe")
	}
	gcSignalWait.Set(0)

	// Scan the stack(s) of this core.
	scanCurrentStack()
	if !task.OnSystemStack() {
		// Mark system stack.
		markRoots(task.SystemStack(), coreStackTop(hartID))
	}

	// Signal we've finished scanning.
	gcScanState.Store(1)
	arm.Asm("sev")

	// Wait until we get a signal that the stop-the-world phase has ended.
	for gcSignalWait.Get() == 0 {
		arm.Asm("wfe")
	}
	gcSignalWait.Set(0)

	// Signal we received the signal and are going to exit the interrupt.
	gcScanState.Add(1)
	arm.Asm("sev")
}

// Pause the given core by sending it an interrupt.
func gcPauseCore(core uint32) {
	multicore_fifo_push_blocking(rp2SIOFIFOCommandGC)
}

// Signal the given core that it can resume one step.
// This is called twice after gcPauseCore: the first time to scan the stack of
// the core, and the second time to end the stop-the-world phase.
func gcSignalCore(core uint32) {
	gcSignalWait.Set(1)
	arm.Asm("sev")
}

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

// These spinlocks are needed by the runtime.
var (
	printLock     = spinLock{id: 20}
	schedulerLock = spinLock{id: 21}
	atomicsLock   = spinLock{id: 22}
	futexLock     = spinLock{id: 23}

	// flashSafeLock is used for XIP operations on RP2040 and RP2350.
	flashSafeLock = spinLock{id: 24}
)

func resetSpinLocks() {
	for i := uint8(0); i < numSpinlocks; i++ {
		l := &spinLock{id: i}
		l.spinlock().Set(0)
	}
}

// A hardware spinlock, one of the 32 spinlocks defined in the SIO peripheral.
type spinLock struct {
	id uint8
}

// Return the spinlock register: rp.SIO.SPINLOCKx
func (l *spinLock) spinlock() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&rp.SIO.SPINLOCK0), l.id*4))
}

func (l *spinLock) Lock() {
	// Wait for the lock to be available.
	spinlock := l.spinlock()
	for spinlock.Get() == 0 {
		arm.Asm("wfe")
	}
}

func (l *spinLock) Unlock() {
	l.spinlock().Set(0)
	arm.Asm("sev")
}

// Wait until a signal is received, indicating that it can resume from the
// spinloop.
func spinLoopWait() {
	arm.Asm("wfe")
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

// machineInit is provided by package machine.
func machineInit()

func init() {
	machineInit()

	machine.InitSerial()
}

func prerun() {
	initCore()

	// Reset spinlocks before the full machineInit() so the scheduler doesn't
	// hang waiting for schedulerLock after a soft reset.
	resetSpinLocks()
}

//export Reset_Handler
func main() {
	preinit()
	prerun()
	run()
	exit(0)
}
