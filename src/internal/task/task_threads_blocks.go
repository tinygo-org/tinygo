//go:build scheduler.threads && (gc.conservative || gc.precise)

package task

// stopWaitGroup is used to wait until all threads have stopped.
var stopWaitGroup waitGroup

// scanWaitGroup is used to wait until all threads have finished scanning their stacks.
var scanWaitGroup waitGroup

// resumeWaitGroup is used to wait until all threads have resumed.
var resumeWaitGroup waitGroup

// GC scan phase. Because we need to stop the world while scanning, this kinda
// needs to be done in the tasks package.
//
// After calling this function, GCResumeWorld needs to be called once to resume
// all threads again.
func GCStopWorldAndScan() {
	// Wait for threads to resume from the previous scan.
	resumeWaitGroup.wait()

	// Don't allow new goroutines to be started while pausing/resuming threads
	// in the stop-the-world phase.
	activeTaskLock.Lock()

	// Set the number of threads to wait for.
	otherGoroutines := otherGoroutines
	stopWaitGroup = initWaitGroup(otherGoroutines)
	scanWaitGroup = initWaitGroup(otherGoroutines + 1)
	resumeWaitGroup = initWaitGroup(otherGoroutines)

	// Pause all other threads.
	current := Current()
	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current {
			tinygo_task_send_gc_signal(t.state.thread)
		}
	}

	// Wait for everything to stop.
	stopWaitGroup.wait()

	// Scan all globals (implemented in the runtime).
	gcScanGlobals()

	// Scan our stack and wait for everything else to complete.
	localScan()
}

//export tinygo_task_gc_pause
func tingyo_task_gc_pause(sig int32) {
	// We have stopped.
	stopWaitGroup.done()

	// Wait for all other threads to stop.
	stopWaitGroup.wait()

	// Scan the local stack.
	localScan()

	// We are resuming.
	resumeWaitGroup.done()
}

func localScan() {
	// Scan the current stack, and all current registers.
	scanCurrentStack()

	// Assist scanning of heap objects.
	finishMark()

	// We are done scanning.
	scanWaitGroup.done()

	// Wait for all other threads to finish scanning.
	scanWaitGroup.wait()
}

//go:linkname finishMark runtime.finishMark
func finishMark()

// GCResumeWorld does not resume anything with the blocks collector.
// The threads will resume as soon as the scan completes.
func GCResumeWorld() {
	// Allow goroutines to start and exit again.
	activeTaskLock.Unlock()
}
