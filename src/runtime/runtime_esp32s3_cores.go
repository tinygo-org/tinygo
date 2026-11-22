//go:build esp32s3 && scheduler.cores

package runtime

import (
	"device/esp"
	"runtime/interrupt"
	"runtime/volatile"
	"sync/atomic"
)

// Set (1) or clear (0) the FROM_CPU interrupt of the given core.
func signalCore(core, value uint32) {
	if core == 0 {
		esp.SYSTEM.CPU_INTR_FROM_CPU_0.Set(value)
	} else {
		esp.SYSTEM.CPU_INTR_FROM_CPU_1.Set(value)
	}
}

func crossCoreChipHandler(core uint32) {
	if flashPauseRequest[core].Swap(0) != 0 {
		state := interrupt.Disable()
		tinygo_flashPark(&flashParkState)
		interrupt.Restore(state)
	}
}

func initCore1() {
	esp.INTERRUPT_CORE0.SetCPU_INTR_FROM_CPU_0_MAP(crossCoreCPUInterrupt)
	esp.INTERRUPT_CORE1.SetCPU_INTR_FROM_CPU_1_MAP(crossCoreCPUInterrupt)
}

func resetCore1() {
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_CLKGATE_EN(1)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RUNSTALL(0)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RESETING(1)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RESETING(0)
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
