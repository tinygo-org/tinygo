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

	// Lowest address of the stack.
	// This is populated when the thread is stopped by the GC.
	stackBottom uintptr

	// Next task in the activeTasks queue.
	QueueNext *Task

	// Semaphore to pause/resume the thread atomically.
	pauseSem Semaphore
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

// otherGoroutines is the total number of live goroutines minus one.
var otherGoroutines uint32

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
	otherGoroutines++
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
	otherGoroutines--
	activeTaskLock.Unlock()

	// Sanity check.
	if !found {
		runtimePanic("taskExited failed")
	}
}

// scanWaitGroup is used to wait on until all threads have finished the current state transition.
var scanWaitGroup waitGroup

type waitGroup struct {
	f Futex
}

func initWaitGroup(n uint32) waitGroup {
	var wg waitGroup
	wg.f.Store(n)
	return wg
}

func (wg *waitGroup) done() {
	if wg.f.Add(^uint32(0)) == 0 {
		wg.f.WakeAll()
	}
}

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

// Scan globals, implemented in the runtime package.
func gcScanGlobals()

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
