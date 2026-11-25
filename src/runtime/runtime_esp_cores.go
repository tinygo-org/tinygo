//go:build (esp32 || esp32s3 || esp32p4) && scheduler.cores

package runtime

import (
	"device"
	"runtime/interrupt"
	"sync/atomic"
	"unsafe"
)

var (
	core1Started   atomic.Uint32
	gcPauseRequest [numCPU]atomic.Uint32
	gcSignalWait   [numCPU]atomic.Uint32
)

func spinLoopWait() {
	device.Asm("nop")
}

var ticksLock spinLock

func lockTicks() interrupt.State {
	mask := interrupt.Disable()
	ticksLock.Lock()
	return mask
}

func unlockTicks(mask interrupt.State) {
	ticksLock.Unlock()
	interrupt.Restore(mask)
}

// The FROM_CPU interrupt can wake a waiting core.
const hasExternalWakeups = true

// End schedulerUnlockAndWait on the given core.
func schedulerEndWait(core uint32) {
	schedulerWakePending[core].Store(1)
	signalCore(core, 1)
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
