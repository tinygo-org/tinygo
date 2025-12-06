//go:build scheduler.threads && !(gc.conservative || gc.precise)

package task

// scanWaitGroup is used to wait on until all threads have finished the current state transition.
var scanWaitGroup waitGroup

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
		scanWaitGroup = initWaitGroup(otherGoroutines)

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
func GCResumeWorld() {
	// NOTE: This does not need to be atomic.
	if gcState.Load() == gcStateResumed {
		// This is already resumed.
		return
	}

	// Set the wait group to track resume progress.
	scanWaitGroup = initWaitGroup(otherGoroutines)

	// Set the state to resumed.
	gcState.Store(gcStateResumed)

	// Wake all of the stopped threads.
	gcState.WakeAll()

	// Allow goroutines to start and exit again.
	activeTaskLock.Unlock()
}

//go:linkname markRoots runtime.markRoots
func markRoots(start, end uintptr)

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
