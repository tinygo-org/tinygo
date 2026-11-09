//go:build scheduler.threads

package task

import (
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

	// Size of the stack, when known. Used by the cooperative threads variant
	// to scan the stacks of suspended threads.
	stackSize uintptr

	// Next task in the activeTasks queue.
	QueueNext *Task

	// Arguments of the goroutine. In wasip3 the C code that starts the thread
	// keeps this pointer in a WebAssembly local, which the GC cannot see. Keep
	// it reachable through the task instead.
	args unsafe.Pointer

	// Semaphore to pause/resume the thread atomically.
	pauseSem Semaphore
}

// Goroutine counter, starting at 0 for the main goroutine.
var goroutineID uintptr

var numCPU int32

var mainTask Task

// Queue of tasks (see QueueNext) that currently exist in the program.
var activeTasks = &mainTask
var activeTaskCount uint32 = 1
var activeTaskLock PMutex
var mainExitedByGoexit bool

func OnSystemStack() bool {
	runtimeFatal("todo: task.OnSystemStack")
	return false
}

// Initialize the main goroutine state. Must be called by the runtime on
// startup, before starting any other goroutines.
func Init(sp uintptr) {
	mainTask.state.stackTop = mainStackTop(sp)
	mainTask.state.stackSize = mainStackSize()
	tinygo_task_init(&mainTask, &mainTask.state.thread, &numCPU)
}

// Return the task struct for the current thread.
func Current() *Task {
	t := (*Task)(tinygo_task_current())
	if t == nil {
		t = currentFallback()
	}
	return t
}

func NumGoroutine() int {
	activeTaskLock.Lock()
	count := activeTaskCount
	activeTaskLock.Unlock()
	return int(count)
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

//export tinygo_task_exited
func taskExited(t *Task) {
	if verbose {
		println("*** exit:", t.state.id)
	}

	if exit(t) {
		runtimeFatal("all goroutines are asleep - deadlock!")
	}
}

func exit(t *Task) bool {
	exitSynctest(t)

	// Remove from the queue.
	// TODO: this can be made more efficient by using a doubly linked list.
	activeTaskLock.Lock()
	found := false
	for q := &activeTasks; *q != nil; q = &(*q).state.QueueNext {
		if *q == t {
			*q = t.state.QueueNext
			found = true
			activeTaskCount--
			break
		}
	}
	deadlocked := mainExitedByGoexit && activeTaskCount == 0
	activeTaskLock.Unlock()

	// Sanity check.
	if !found {
		runtimeFatal("taskExited failed")
	}
	return deadlocked
}

func otherTasks(current *Task) uint32 {
	if activeTaskCount == 0 {
		return 0
	}
	return activeTaskCount - 1
}

// Goexit exits the current task. If this is the main task and there are no other
// goroutines, it reports a deadlock.
func Goexit() {
	t := Current()
	if t == &mainTask {
		activeTaskLock.Lock()
		noOtherTasks := otherTasks(t) == 0
		if !noOtherTasks {
			mainExitedByGoexit = true
		}
		activeTaskLock.Unlock()
		if noOtherTasks {
			runtimeFatal("all goroutines are asleep - deadlock!")
		}
	}
	if exit(t) {
		runtimeFatal("all goroutines are asleep - deadlock!")
	}
	tinygo_task_exit()
}

func CoroExit(next *Task) {
	t := Current()
	synctestTaskWake(next)
	if exit(t) {
		runtimeFatal("all goroutines are asleep - deadlock!")
	}
	scheduleTaskNoWake(next)
	tinygo_task_exit()
}

// Yield yields the current thread to the OS scheduler.
func Yield() {
	sched_yield()
}

//go:linkname markRoots runtime.markRoots
func markRoots(start, end uintptr)

// Scan globals, implemented in the runtime package.
func gcScanGlobals()

// Return the highest address of the current stack.
func StackTop() uintptr {
	return Current().state.stackTop
}

// Using //go:linkname instead of //export so that we don't tell the compiler
// that the 't' parameter won't escape (because it will).
//
//go:linkname tinygo_task_init tinygo_task_init
func tinygo_task_init(t *Task, thread *threadID, numCPU *int32)

// Here same as for tinygo_task_init.
//
//go:linkname tinygo_task_start tinygo_task_start
func tinygo_task_start(fn uintptr, args unsafe.Pointer, t *Task, thread *threadID, stackTop *uintptr, stackSize uintptr) int32

//go:linkname tinygo_task_exit tinygo_task_exit_thread
func tinygo_task_exit()

//export tinygo_task_current
func tinygo_task_current() unsafe.Pointer

//export sched_yield
func sched_yield() int32

func NumCPU() int {
	return int(numCPU)
}
