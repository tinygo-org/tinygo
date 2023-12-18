//go:build scheduler.threads

package task

import (
	"runtime/spinlock"
	"sync/atomic"
	"unsafe"
)

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(str string)

//go:linkname currentMachineId runtime.currentMachineId
func currentMachineId() int

type state struct {
	tid int

	paused int32
}

var lock spinlock.SpinLock
var tasks map[int]*Task = map[int]*Task{}

// Current returns the current active task.
func Current() *Task {
	tid := currentMachineId()
	lock.Lock()
	task := tasks[tid]
	lock.Unlock()
	return task
}

// Pause suspends the current task and returns to the scheduler.
// This function may only be called when running on a goroutine stack, not when running on the system stack or in an interrupt.
func Pause() {
	task := Current()
	atomic.StoreInt32(&task.state.paused, 1)
	for atomic.LoadInt32(&task.state.paused) == 1 {
		// wait for notify
	}
}

// Resume the task until it pauses or completes.
// This may only be called from the scheduler.
func (t *Task) Resume() {
	atomic.StoreInt32(&t.state.paused, 0)
}

// OnSystemStack returns whether the caller is running on the system stack.
func OnSystemStack() bool {
	// If there is not an active goroutine, then this must be running on the system stack.
	return Current() == nil
}

func SystemStack() uintptr {
	return 0
}

type taskHandoverParam struct {
	t    *Task
	fn   uintptr
	args uintptr
}

var retainObjects = map[unsafe.Pointer]any{}

// start creates and starts a new goroutine with the given function and arguments.
// The new goroutine is scheduled to run later.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	h := &taskHandoverParam{
		t:    t,
		fn:   fn,
		args: uintptr(args),
	}

	lock.Lock()
	retainObjects[unsafe.Pointer(h)] = h
	lock.Unlock()

	createThread(uintptr(unsafe.Pointer(&callMachineEntryFn)), uintptr(unsafe.Pointer(h)), stackSize)
}

//export tinygo_machineEntry
func machineEntry(param *taskHandoverParam) {
	tid := currentMachineId()
	param.t.state.tid = tid

	lock.Lock()
	delete(retainObjects, unsafe.Pointer(param))
	tasks[tid] = param.t
	lock.Unlock()

	handoverTask(param.fn, param.args)

	lock.Lock()
	delete(tasks, tid)
	lock.Unlock()
}
