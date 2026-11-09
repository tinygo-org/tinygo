//go:build scheduler.threads && wasip3

package task

import (
	"sync/atomic"
	"unsafe"
)

// Threads are cooperative in wasip3. A thread only stops running at a call that
// blocks or yields, so every other thread is suspended while the GC runs. There
// is no need to signal them. The stack of a suspended thread is scanned
// completely, because its stack pointer is not known.

// A pthread_t is a pointer in wasi-libc.
type threadID unsafe.Pointer

//go:extern __stack_high
var stackHighSymbol [0]byte

//go:extern __stack_low
var stackLowSymbol [0]byte

// Code that runs before task.Init, or in a task that was started by the host
// instead of a thread (such as cabi_realloc), is part of the main thread.
func currentFallback() *Task {
	return &mainTask
}

// mainStackTop returns the highest address of the stack of the main thread. The
// current stack pointer is not enough: the frames of the callers have values
// that must be scanned too.
func mainStackTop(sp uintptr) uintptr {
	high := uintptr(unsafe.Pointer(&stackHighSymbol))
	low := uintptr(unsafe.Pointer(&stackLowSymbol))
	if sp >= low && sp <= high {
		return high
	}
	// The stack was allocated by the task hook, so the size is not known.
	return (sp + 15) &^ 15
}

func mainStackSize() uintptr {
	return uintptr(unsafe.Pointer(&stackHighSymbol)) - uintptr(unsafe.Pointer(&stackLowSymbol))
}

//go:linkname stacksave runtime.stacksave
func stacksave() unsafe.Pointer

// Stop the world and scan all roots. After calling this function,
// GCResumeWorld needs to be called once.
//
//go:noheap
func GCStopWorldAndScan() {
	current := Current()

	// Don't allow threads to start or exit during the scan.
	activeTaskLock.Lock()

	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current && t.state.stackTop != 0 {
			// The stack of the main thread starts at the bottom of memory.
			var low uintptr
			if t.state.stackTop > t.state.stackSize {
				low = t.state.stackTop - t.state.stackSize
			}
			// The range must be aligned to a pointer.
			low = (low + 3) &^ 3
			top := t.state.stackTop &^ 3
			if low < top {
				markRoots(low, top)
			}
		}
	}

	// Scan the stack of the current thread.
	top := current.state.stackTop
	if top == 0 {
		// The GC was started before task.Init.
		top = uintptr(unsafe.Pointer(&stackHighSymbol))
	}
	markRoots(uintptr(stacksave())&^3, top&^3)

	// Scan all globals (implemented in the runtime).
	gcScanGlobals()
}

// After the GC is done scanning, allow threads to start and exit again.
//
//go:noheap
func GCResumeWorld() {
	activeTaskLock.Unlock()
}

// Start a new thread.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	inheritSynctest(t)
	t.state.id = atomic.AddUintptr(&goroutineID, 1)
	t.state.stackSize = stackSize
	t.state.args = args
	if verbose {
		println("*** start:  ", t.state.id, "from", Current().state.id)
	}

	// Add the task to the list before starting the thread. Creating a thread
	// allocates memory, which may start a GC cycle that needs activeTaskLock.
	// The GC skips tasks that did not set their stack yet.
	activeTaskLock.Lock()
	t.state.QueueNext = activeTasks
	activeTasks = t
	activeTaskCount++
	activeTaskLock.Unlock()

	errCode := tinygo_task_start(fn, args, t, &t.state.thread, &t.state.stackTop, stackSize)
	if errCode != 0 {
		runtimeFatal("could not start thread")
	}
}
