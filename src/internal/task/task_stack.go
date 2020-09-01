// +build scheduler.tasks

package task

import "unsafe"

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(str string)

// Stack canary, to detect a stack overflow. The number is a random number
// generated by random.org. The bit fiddling dance is necessary because
// otherwise Go wouldn't allow the cast to a smaller integer size.
const stackCanary = uintptr(uint64(0x670c1333b83bf575) & uint64(^uintptr(0)))

// state is a structure which holds a reference to the state of the task.
// When the task is suspended, the registers are stored onto the stack and the stack pointer is stored into sp.
type state struct {
	// sp is the stack pointer of the saved state.
	// When the task is inactive, the saved registers are stored at the top of the stack.
	sp uintptr

	// canaryPtr points to the top word of the stack (the lowest address).
	// This is used to detect stack overflows.
	// When initializing the goroutine, the stackCanary constant is stored there.
	// If the stack overflowed, the word will likely no longer equal stackCanary.
	canaryPtr *uintptr
}

// currentTask is the current running task, or nil if currently in the scheduler.
var currentTask *Task

// Current returns the current active task.
func Current() *Task {
	return currentTask
}

// Pause suspends the current task and returns to the scheduler.
// This function may only be called when running on a goroutine stack, not when running on the system stack or in an interrupt.
func Pause() {
	// Check whether the canary (the lowest address of the stack) is still
	// valid. If it is not, a stack overflow has occured.
	if *currentTask.state.canaryPtr != stackCanary {
		runtimePanic("goroutine stack overflow")
	}
	currentTask.state.pause()
}

// Resume the task until it pauses or completes.
// This may only be called from the scheduler.
func (t *Task) Resume() {
	currentTask = t
	t.state.resume()
	currentTask = nil
}

// initialize the state and prepare to call the specified function with the specified argument bundle.
func (s *state) initialize(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	// Create a stack.
	stack := make([]uintptr, stackSize/unsafe.Sizeof(uintptr(0)))

	// Invoke architecture-specific initialization.
	s.archInit(stack, fn, args)
}

//go:linkname runqueuePushBack runtime.runqueuePushBack
func runqueuePushBack(*Task)

// start creates and starts a new goroutine with the given function and arguments.
// The new goroutine is scheduled to run later.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	t.state.initialize(fn, args, stackSize)
	runqueuePushBack(t)
}

// OnSystemStack returns whether the caller is running on the system stack.
func OnSystemStack() bool {
	// If there is not an active goroutine, then this must be running on the system stack.
	return Current() == nil
}
