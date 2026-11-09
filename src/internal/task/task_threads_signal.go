//go:build scheduler.threads && !wasip3

package task

import (
	"sync/atomic"
	"unsafe"
)

// scanWaitGroup is used to wait on until all threads have finished the current state transition.
var scanWaitGroup waitGroup

type waitGroup struct {
	f Futex
}

//go:noheap
func (wg *waitGroup) reset(n uint32) {
	wg.f.Store(n)
}

//go:noheap
func (wg *waitGroup) done() {
	if wg.f.Add(^uint32(0)) == 0 {
		wg.f.WakeAll()
	}
}

//go:noheap
func (wg *waitGroup) wait() {
	for {
		val := wg.f.Load()
		if val == 0 {
			return
		}
		wg.f.Wait(val)
	}
}

// gcState is used to track and notify threads when the GC is stopping/resuming.
var gcState Futex

const (
	gcStateResumed = iota
	gcStateStopped
)

// GC scan phase. Because we need to stop the world while scanning, this kinda
// needs to be done in the tasks package.
//
// After calling this function, GCResumeWorld needs to be called once to resume
// all threads again.
//
//go:noheap
func GCStopWorldAndScan() {
	current := Current()

	// NOTE: This does not need to be atomic.
	if gcState.Load() == gcStateResumed {
		// Don't allow new goroutines to be started while pausing/resuming threads
		// in the stop-the-world phase.
		activeTaskLock.Lock()

		// Wait for threads to finish resuming.
		scanWaitGroup.wait()

		// Change the gc state to stopped.
		// NOTE: This does not need to be atomic.
		gcState.Store(gcStateStopped)

		// Set the number of threads to wait for.
		scanWaitGroup.reset(otherTasks(current))

		// Pause all other threads.
		for t := activeTasks; t != nil; t = t.state.QueueNext {
			if t != current {
				tinygo_task_send_gc_signal(t.state.thread)
			}
		}

		// Wait for the threads to finish stopping.
		scanWaitGroup.wait()
	}

	// Scan other thread stacks.
	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current {
			markRoots(t.state.stackBottom, t.state.stackTop)
		}
	}

	// Scan the current stack, and all current registers.
	scanCurrentStack()

	// Scan all globals (implemented in the runtime).
	gcScanGlobals()
}

// After the GC is done scanning, resume all other threads.
//
//go:noheap
func GCResumeWorld() {
	// NOTE: This does not need to be atomic.
	if gcState.Load() == gcStateResumed {
		// This is already resumed.
		return
	}

	// Set the wait group to track resume progress.
	scanWaitGroup.reset(otherTasks(Current()))

	// Set the state to resumed.
	gcState.Store(gcStateResumed)

	// Wake all of the stopped threads.
	gcState.WakeAll()

	// Allow goroutines to start and exit again.
	activeTaskLock.Unlock()
}

var stackScanLock PMutex

//export tinygo_task_gc_pause
func tingyo_task_gc_pause(sig int32) {
	// Write the entrty stack pointer to the state.
	Current().state.stackBottom = uintptr(stacksave())

	// Notify the GC that we are stopped.
	scanWaitGroup.done()

	// Wait for the GC to resume.
	for gcState.Load() == gcStateStopped {
		gcState.Wait(gcStateStopped)
	}

	// Notify the GC that we have resumed.
	scanWaitGroup.done()
}

//go:export tinygo_scanCurrentStack
func scanCurrentStack()

//go:linkname stacksave runtime.stacksave
func stacksave() unsafe.Pointer

// Pause the thread by sending it a signal.
//
//export tinygo_task_send_gc_signal
func tinygo_task_send_gc_signal(threadID)

// Start a new OS thread.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	inheritSynctest(t)
	t.state.id = atomic.AddUintptr(&goroutineID, 1)
	if verbose {
		println("*** start:  ", t.state.id, "from", Current().state.id)
	}

	t.state.stackSize = stackSize

	// Start the new thread, and add it to the list of threads.
	// Do this with a lock so that only started threads are part of the queue
	// and the stop-the-world GC won't see threads that haven't started yet or
	// are not fully started yet.
	activeTaskLock.Lock()
	errCode := tinygo_task_start(fn, args, t, &t.state.thread, &t.state.stackTop, stackSize)
	if errCode != 0 {
		runtimeFatal("could not start thread")
	}
	t.state.QueueNext = activeTasks
	activeTasks = t
	activeTaskCount++
	activeTaskLock.Unlock()
}
func mainStackSize() uintptr {
	return 0
}

func currentFallback() *Task {
	runtimeFatal("unknown current task")
	return nil
}

func mainStackTop(sp uintptr) uintptr {
	return sp
}
