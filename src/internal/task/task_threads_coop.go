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

// Tasks that the host started after the first one. They share the main thread's
// TLS, so each is found by its stack. No lock is needed.
var adoptedTasks *Task

// Code that runs before task.Init is part of the main thread. A task that the
// host started is the one whose stack has the stack pointer.
func currentFallback() *Task {
	if adoptedTasks != nil {
		sp := uintptr(stacksave())
		for t := adoptedTasks; t != nil; t = t.state.adoptNext {
			if sp < t.state.stackTop && sp >= t.state.stackTop-t.state.stackSize {
				return t
			}
		}
	}
	return &mainTask
}

// mallocEnd is implemented by the runtime.
func mallocEnd(addr uintptr) uintptr

// AdoptThread makes a task of the calling host task. The first one is the main
// task. It must run first in the export, because the GC needs the stack.
func AdoptThread() {
	sp := uintptr(stacksave())
	size := mainStackSize()
	// The task hook of wasi-libc uses the main stack when it is free, and
	// otherwise a malloc block. The stack pointer can be at the block end.
	var top uintptr
	if high := uintptr(unsafe.Pointer(&stackHighSymbol)); sp <= high && sp >= high-size {
		top = high
	} else {
		top = mallocEnd(sp - 1)
		if top < sp || top-sp > size {
			// This is not such a stack.
			top = (sp + 16) &^ 15
		}
	}
	t := &Task{}
	t.state.id = atomic.AddUintptr(&goroutineID, 1)
	t.state.stackTop = top
	t.state.stackSize = size
	activeTaskLock.Lock()
	t.state.QueueNext = activeTasks
	activeTasks = t
	activeTaskCount++
	t.state.adoptNext = adoptedTasks
	adoptedTasks = t
	activeTaskLock.Unlock()
}

// ReleaseThread ends the task that AdoptThread made, when the export returns.
func ReleaseThread() {
	t := Current()
	if t == &mainTask {
		return
	}
	for q := &adoptedTasks; *q != nil; q = &(*q).state.adoptNext {
		if *q == t {
			*q = t.state.adoptNext
			break
		}
	}
	exit(t)
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

	// The GC can start before a stack has a task, such as in the task hook.
	// Current is then the main task, which is suspended and must be scanned.
	sp := uintptr(stacksave())
	inCurrent := current.state.stackTop != 0 && sp <= current.state.stackTop &&
		(current.state.stackSize == 0 || current.state.stackTop-sp <= current.state.stackSize)
	if !inCurrent && current == &mainTask && current.state.stackTop != 0 {
		current = nil
	}

	// Don't allow threads to start or exit during the scan.
	activeTaskLock.Lock()

	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current && t.state.stackTop != 0 {
			// The stack of the main thread starts at the bottom of memory.
			var low uintptr
			if t.state.stackTop > t.state.stackSize {
				low = t.state.stackTop - t.state.stackSize
			}
			// A thread that suspended in a futex knows where its stack ends.
			if sp := t.state.stackBottom; sp > low && sp <= t.state.stackTop {
				low = sp
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
	if current != nil {
		top := current.state.stackTop
		if top == 0 {
			// The GC was started before task.Init.
			top = uintptr(unsafe.Pointer(&stackHighSymbol))
		}
		markRoots(sp&^3, top&^3)
	}

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

//export tinygo_task_set_suspended_sp
func setSuspendedSP(sp uintptr) {
	Current().state.stackBottom = sp
}

// MarkSuspended records the caller's stack pointer before a blocking call, so
// the GC skips the unused stack. Call it with false when the call returns.
//
//go:noinline
func MarkSuspended(suspended bool) {
	if suspended {
		Current().state.stackBottom = uintptr(stacksave())
	} else {
		Current().state.stackBottom = 0
	}
}
