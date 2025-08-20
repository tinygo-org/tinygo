//go:build scheduler.threads

package task

import (
	"sync/atomic"
	"unsafe"
)

// If true, print verbose debug logs.
const verbose = false

// Scheduler-specific state.
type state struct {
	// Goroutine ID. The number here is not really significant and after a while
	// it could wrap around. But it is useful for debugging.
	id uintptr

	// Thread ID, pthread_t or similar (typically implemented as a pointer).
	thread threadID

	// Highest address of the stack. It is stored when the goroutine starts, and
	// is needed to be able to scan the stack.
	stackTop uintptr

	// Next task in the activeTasks queue.
	QueueNext *Task

	// Semaphore to pause/resume the thread atomically.
	pauseSem Semaphore

	// Semaphore used for stack scanning.
	// We can't reuse pauseSem here since the thread might have been paused for
	// other reasons (for example, because it was waiting on a channel).
	gcSem Semaphore
}

// Goroutine counter, starting at 0 for the main goroutine.
var goroutineID uintptr

var numCPU int32

var mainTask Task

// Queue of tasks (see QueueNext) that currently exist in the program.
var activeTasks = &mainTask
var activeTaskLock PMutex

func OnSystemStack() bool {
	runtimePanic("todo: task.OnSystemStack")
	return false
}

// Initialize the main goroutine state. Must be called by the runtime on
// startup, before starting any other goroutines.
func Init(sp uintptr) {
	mainTask.state.stackTop = sp
	tinygo_task_init(&mainTask, &mainTask.state.thread, &numCPU)
}

// Return the task struct for the current thread.
func Current() *Task {
	t := (*Task)(tinygo_task_current())
	if t == nil {
		runtimePanic("unknown current task")
	}
	return t
}

// Pause pauses the current task, until it is resumed by another task.
// It is possible that another task has called Resume() on the task before it
// hits Pause(), in which case the task won't be paused but continues
// immediately.
func Pause() {
	// Wait until resumed
	t := Current()
	if verbose {
		println("*** pause:  ", t.state.id)
	}
	t.state.pauseSem.Wait()
}

// Resume the given task.
// It is legal to resume a task before it gets paused, it means that the next
// call to Pause() won't pause but will continue immediately. This happens in
// practice sometimes in channel operations, where the Resume() might get called
// between the channel unlock and the call to Pause().
func (t *Task) Resume() {
	if verbose {
		println("*** resume: ", t.state.id)
	}
	// Increment the semaphore counter.
	// If the task is currently paused in Wait(), it will resume.
	// If the task is not yet paused, the next call to Wait() will continue
	// immediately.
	t.state.pauseSem.Post()
}

// Start a new OS thread.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	t.state.id = atomic.AddUintptr(&goroutineID, 1)
	if verbose {
		println("*** start:  ", t.state.id, "from", Current().state.id)
	}

	// Start the new thread, and add it to the list of threads.
	// Do this with a lock so that only started threads are part of the queue
	// and the stop-the-world GC won't see threads that haven't started yet or
	// are not fully started yet.
	activeTaskLock.Lock()
	errCode := tinygo_task_start(fn, args, t, &t.state.thread, &t.state.stackTop, stackSize)
	if errCode != 0 {
		runtimePanic("could not start thread")
	}
	t.state.QueueNext = activeTasks
	activeTasks = t
	activeTaskLock.Unlock()
}

//export tinygo_task_exited
func taskExited(t *Task) {
	if verbose {
		println("*** exit:", t.state.id)
	}

	// Remove from the queue.
	// TODO: this can be made more efficient by using a doubly linked list.
	activeTaskLock.Lock()
	found := false
	for q := &activeTasks; *q != nil; q = &(*q).state.QueueNext {
		if *q == t {
			*q = t.state.QueueNext
			found = true
			break
		}
	}
	activeTaskLock.Unlock()

	// Sanity check.
	if !found {
		runtimePanic("taskExited failed")
	}
}

// Futex to wait on until all tasks have finished scanning the stack.
// This is basically a sync.WaitGroup.
var scanDoneFutex Futex

// GC scan phase. Because we need to stop the world while scanning, this kinda
// needs to be done in the tasks package.
//
// After calling this function, GCResumeWorld needs to be called once to resume
// all threads again.
func GCStopWorldAndScan() {
	current := Current()

	// Don't allow new goroutines to be started while pausing/resuming threads
	// in the stop-the-world phase.
	activeTaskLock.Lock()

	// Pause all other threads.
	numOtherThreads := uint32(0)
	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current {
			numOtherThreads++
			tinygo_task_send_gc_signal(t.state.thread)
		}
	}

	// Store the number of threads to wait for in the futex.
	// This is the equivalent of doing an initial wg.Add(numOtherThreads).
	scanDoneFutex.Store(numOtherThreads)

	// Scan the current stack, and all current registers.
	scanCurrentStack()

	// Wake each paused thread for the first time so it will scan the stack.
	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current {
			t.state.gcSem.Post()
		}
	}

	// Wait until all threads have finished scanning their stack.
	// This is the equivalent of wg.Wait()
	for {
		val := scanDoneFutex.Load()
		if val == 0 {
			break
		}
		scanDoneFutex.Wait(val)
	}

	// Scan all globals (implemented in the runtime).
	gcScanGlobals()
}

// After the GC is done scanning, resume all other threads.
//
// This must only be called after a GCStopWorldAndScan call.
func GCResumeWorld() {
	current := Current()

	// Wake each paused thread for the second time, so they will resume normal
	// operation.
	for t := activeTasks; t != nil; t = t.state.QueueNext {
		if t != current {
			t.state.gcSem.Post()
		}
	}

	// Allow goroutines to start and exit again.
	activeTaskLock.Unlock()
}

// Scan globals, implemented in the runtime package.
func gcScanGlobals()

var stackScanLock PMutex

//export tinygo_task_gc_pause
func tingyo_task_gc_pause(sig int32) {
	// Wait until we get the signal to start scanning the stack.
	Current().state.gcSem.Wait()

	// Scan the thread stack.
	// Only scan a single thread stack at a time, because the GC marking phase
	// doesn't support parallelism.
	// TODO: it may be possible to call markRoots directly (without saving
	// registers) since we are in a signal handler that already saved a bunch of
	// registers. This is an optimization left for a future time.
	stackScanLock.Lock()
	scanCurrentStack()
	stackScanLock.Unlock()

	// Equivalent of wg.Done(): subtract one from the futex and if the result is
	// 0 (meaning we were the last in the waitgroup), wake the waiting thread.
	n := uint32(1)
	if scanDoneFutex.Add(-n) == 0 {
		scanDoneFutex.Wake()
	}

	// Wait until we get the signal we can resume normally (after the mark phase
	// has finished).
	Current().state.gcSem.Wait()
}

//go:export tinygo_scanCurrentStack
func scanCurrentStack()

// Return the highest address of the current stack.
func StackTop() uintptr {
	return Current().state.stackTop
}

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(msg string)

// Using //go:linkname instead of //export so that we don't tell the compiler
// that the 't' parameter won't escape (because it will).
//
//go:linkname tinygo_task_init tinygo_task_init
func tinygo_task_init(t *Task, thread *threadID, numCPU *int32)

// Here same as for tinygo_task_init.
//
//go:linkname tinygo_task_start tinygo_task_start
func tinygo_task_start(fn uintptr, args unsafe.Pointer, t *Task, thread *threadID, stackTop *uintptr, stackSize uintptr) int32

// Pause the thread by sending it a signal.
//
//export tinygo_task_send_gc_signal
func tinygo_task_send_gc_signal(threadID)

//export tinygo_task_current
func tinygo_task_current() unsafe.Pointer

func NumCPU() int {
	return int(numCPU)
}
