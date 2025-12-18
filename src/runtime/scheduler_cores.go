//go:build scheduler.cores

package runtime

import (
	"internal/task"
	"runtime/interrupt"
	"sync/atomic"
)

const hasScheduler = true

const hasParallelism = true

var mainExited atomic.Uint32

// True after the secondary cores have started.
var secondaryCoresStarted bool

// Which task is running on a given core (or nil if there is no task running on
// the core).
var cpuTasks [numCPU]*task.Task

var (
	sleepQueue *task.Task
	runqueue   task.Queue
)

func deadlock() {
	// Call yield without requesting a wakeup.
	task.Pause()
	trap()
}

// Mark the given task as ready to resume.
// This is allowed even if the task isn't paused yet, but will pause soon.
func scheduleTask(t *task.Task) {
	schedulerLock.Lock()
	switch t.RunState {
	case task.RunStatePaused:
		// Paused, state is saved on the stack.
		// Add it to the runqueue...
		runqueue.Push(t)
		// ...and wake up a sleeping core, if there is one.
		// (If all cores are already busy, this is a no-op).
		schedulerWake()
	case task.RunStateRunning:
		// Not yet paused (probably going to pause very soon), so let the
		// Pause() function know it can resume immediately.
		t.RunState = task.RunStateResuming
	default:
		if schedulerAsserts {
			runtimePanic("scheduler: unknown run state")
		}
	}
	schedulerLock.Unlock()
}

func addSleepTask(t *task.Task, wakeup timeUnit) {
	// Save the timestamp when the task should be woken up.
	t.Data = uint64(wakeup)

	// If another core is currently using the timer, make sure it wakes up at
	// the right time.
	interruptSleepTicksMulticore(wakeup)

	// Find the position where we should insert this task in the queue.
	q := &sleepQueue
	for {
		if *q == nil {
			// Found the end of the time queue. Insert it here, at the end.
			break
		}
		if timeUnit((*q).Data) > timeUnit(t.Data) {
			// Found a task in the queue that has a timeout before the
			// to-be-sleeping task. Insert our task right before.
			break
		}
		q = &(*q).Next
	}

	// Insert the task into the queue (this could be at the end, if *q is nil).
	t.Next = *q
	*q = t
}

func Gosched() {
	schedulerLock.Lock()
	runqueue.Push(task.Current())
	task.PauseLocked()
}

// NumCPU returns the number of CPU cores on this system.
func NumCPU() int {
	return numCPU
}

func addTimer(tn *timerNode) {
	schedulerLock.Lock()
	timerQueueAdd(tn)
	interruptSleepTicksMulticore(tn.whenTicks())
	schedulerLock.Unlock()
}

func removeTimer(t *timer) *timerNode {
	schedulerLock.Lock()
	n := timerQueueRemove(t)
	schedulerLock.Unlock()
	return n
}

func schedulerRunQueue() *task.Queue {
	return &runqueue
}

// Pause the current task for a given time.
//
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	wakeup := ticks() + nanosecondsToTicks(duration)

	// While the scheduler is locked:
	// - add this task to the sleep queue
	// - switch to the scheduler (only allowed while locked)
	// - let the scheduler handle it from there
	schedulerLock.Lock()
	addSleepTask(task.Current(), wakeup)
	task.PauseLocked()
}

// This function is called on the first core in the system. It will wake up the
// other cores when ready.
func run() {
	initHeap()
	initRand()

	go func() {
		// Package initializers are currently run single-threaded.
		// This might help with registering interrupts and such.
		initAll()

		// After package initializers have finished, start all the other cores.
		startSecondaryCores()
		secondaryCoresStarted = true

		// Run main.main.
		callMain()

		// main.main has exited, so the program should exit.
		mainExited.Store(1)
	}()

	// The scheduler must always be entered while the scheduler lock is taken.
	schedulerLock.Lock()
	scheduler(false)
	schedulerLock.Unlock()
}

func scheduler(_ bool) {
	for mainExited.Load() == 0 {
		// Check for ready-to-run tasks.
		if runnable := runqueue.Pop(); runnable != nil {
			// Resume it now.
			setCurrentTask(runnable)
			runnable.RunState = task.RunStateRunning
			schedulerLock.Unlock() // unlock before resuming, Pause() will lock again
			runnable.Resume()
			setCurrentTask(nil)

			continue
		}

		var now timeUnit
		if sleepQueue != nil || timerQueue != nil {
			now = ticks()

			// Check whether the first task in the sleep queue is ready to run.
			if sleepingTask := sleepQueue; sleepingTask != nil && now >= timeUnit(sleepingTask.Data) {
				// It is, pop it from the queue.
				sleepQueue = sleepQueue.Next
				sleepingTask.Next = nil

				// Run it now.
				setCurrentTask(sleepingTask)
				sleepingTask.RunState = task.RunStateRunning
				schedulerLock.Unlock() // unlock before resuming, Pause() will lock again
				sleepingTask.Resume()
				setCurrentTask(nil)
				continue
			}

			// Check whether a timer has expired that needs to be run.
			if timerQueue != nil && now >= timerQueue.whenTicks() {
				delay := ticksToNanoseconds(now - timerQueue.whenTicks())
				// Pop timer from queue.
				tn := timerQueue
				timerQueue = tn.next
				tn.next = nil

				// Run the callback stored in this timer node.
				schedulerLock.Unlock()
				tn.callback(tn, delay)
				schedulerLock.Lock()
				continue
			}
		}

		// At this point, there are no runnable tasks anymore.
		// If another core is using the clock, let it handle the sleep queue.
		if hasSleepingCore() {
			schedulerUnlockAndWait()
			continue
		}

		// The timer is free to use, so check whether there are any future
		// tasks/timers that we can wait for.
		var timeLeft timeUnit
		if sleepingTask := sleepQueue; sleepingTask != nil {
			// We already checked that there is no ready-to-run sleeping task
			// (using the same 'now' value), so timeLeft will always be
			// positive.
			timeLeft = timeUnit(sleepingTask.Data) - now
		}
		if timerQueue != nil {
			// If the timer queue needs to run earlier, reduce the time we are
			// going to sleep.
			// Like with sleepQueue, we already know there is no timer ready to
			// run since we already checked above.
			timeLeftForTimer := timerQueue.whenTicks() - now
			if sleepQueue == nil || timeLeftForTimer < timeLeft {
				timeLeft = timeLeftForTimer
			}
		}

		if timeLeft > 0 {
			// Sleep for a bit until the next task or timer is ready to run.
			sleepTicksMulticore(timeLeft)
			continue
		}

		// No runnable tasks and no sleeping tasks or timers. There's nothing to
		// do.
		// Wait until something happens (like an interrupt).
		schedulerUnlockAndWait()
	}
}

func currentTask() *task.Task {
	return cpuTasks[currentCPU()]
}

func setCurrentTask(task *task.Task) {
	cpuTasks[currentCPU()] = task
}

func lockScheduler() {
	schedulerLock.Lock()
}

func unlockScheduler() {
	schedulerLock.Unlock()
}

func lockFutex() interrupt.State {
	mask := interrupt.Disable()
	futexLock.Lock()
	return mask
}

func unlockFutex(state interrupt.State) {
	futexLock.Unlock()
	interrupt.Restore(state)
}

// Use a single spinlock for atomics. This works fine, since atomics are very
// short sequences of instructions.
func lockAtomics() interrupt.State {
	mask := interrupt.Disable()
	atomicsLock.Lock()
	return mask
}

func unlockAtomics(mask interrupt.State) {
	atomicsLock.Unlock()
	interrupt.Restore(mask)
}

var systemStack [numCPU]uintptr

// Implementation detail of the internal/task package.
// It needs to store the system stack pointer somewhere, and needs to know how
// many cores there are to do so. But it doesn't know the number of cores. Hence
// why this is implemented in the runtime.
func systemStackPtr() *uintptr {
	return &systemStack[currentCPU()]
}

// Color the 'print' and 'println' output according to the current CPU.
// This may be helpful for debugging, but should be disabled otherwise.
const cpuColoredPrint = false

func printlock() {
	// Don't lock the print output inside an interrupt.
	// Locking the print output inside an interrupt can lead to a deadlock: if
	// the interrupt happens while the print lock is held, the interrupt won't
	// be able to take this lock anymore.
	// This isn't great, but the alternative would be to disable interrupts
	// while printing which seems like a worse idea to me.
	if !interrupt.In() {
		printLock.Lock()
	}

	if cpuColoredPrint {
		switch currentCPU() {
		case 1:
			printstring("\x1b[32m") // green
		case 2:
			printstring("\x1b[33m") // yellow
		case 3:
			printstring("\x1b[34m") // blue
		}
	}
}

func printunlock() {
	if cpuColoredPrint {
		if currentCPU() != 0 {
			printstring("\x1b[0m") // reset colored output
		}
	}

	if !interrupt.In() {
		printLock.Unlock()
	}
}
