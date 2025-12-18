//go:build scheduler.threads

package runtime

import (
	"internal/task"
	"runtime/interrupt"
)

const hasScheduler = false // not using the cooperative scheduler

// We use threads, so yes there is parallelism.
const hasParallelism = true

var (
	timerQueueLock    task.PMutex
	timerQueueStarted bool
	timerFutex        task.Futex
)

// Because we just use OS threads, we don't need to do anything special here. We
// can just initialize everything and run main.main on the main thread.
func run() {
	initRand()
	initHeap()
	task.Init(stackTop)
	initAll()
	callMain()
}

// Pause the current task for a given time.
//
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	sleepTicks(nanosecondsToTicks(duration))
}

func deadlock() {
	// TODO: exit the thread via pthread_exit.
	task.Pause()
}

func scheduleTask(t *task.Task) {
	t.Resume()
}

func Gosched() {
	// Each goroutine runs in a thread, so there's not much we can do here.
	// There is sched_yield but it's only really intended for realtime
	// operation, so is probably best not to use.
}

// NumCPU returns the number of logical CPUs usable by the current process.
func NumCPU() int {
	return task.NumCPU()
}

// Separate goroutine (thread) that runs timer callbacks when they expire.
func timerRunner() {
	for {
		timerQueueLock.Lock()

		if timerQueue == nil {
			// No timer in the queue, so wait until one becomes available.
			val := timerFutex.Load()
			timerQueueLock.Unlock()
			timerFutex.Wait(val)
			continue
		}

		now := ticks()
		if now < timerQueue.whenTicks() {
			// There is a timer in the queue, but we need to wait until it
			// expires.
			// Using a futex, so that the wait is exited early when adding a new
			// (sooner-to-expire) timer.
			val := timerFutex.Load()
			timerQueueLock.Unlock()
			timeout := ticksToNanoseconds(timerQueue.whenTicks() - now)
			timerFutex.WaitUntil(val, uint64(timeout))
			continue
		}

		// Pop timer from queue.
		tn := timerQueue
		timerQueue = tn.next
		tn.next = nil

		timerQueueLock.Unlock()

		// Run the callback stored in this timer node.
		delay := ticksToNanoseconds(now - tn.whenTicks())
		tn.callback(tn, delay)
	}
}

func addTimer(tim *timerNode) {
	timerQueueLock.Lock()

	if !timerQueueStarted {
		timerQueueStarted = true
		go timerRunner()
	}

	timerQueueAdd(tim)

	timerFutex.Add(1)
	timerFutex.Wake()

	timerQueueLock.Unlock()
}

func removeTimer(tim *timer) *timerNode {
	timerQueueLock.Lock()
	n := timerQueueRemove(tim)
	timerQueueLock.Unlock()
	return n
}

func schedulerRunQueue() *task.Queue {
	// This function is not actually used, it is only called when hasScheduler
	// is true. So we can just return nil here.
	return nil
}

func runqueueForGC() *task.Queue {
	// There is only a runqueue when using the cooperative scheduler.
	return nil
}

// Lock to make sure print calls do not interleave.
var printLock task.Mutex

func printlock() {
	printLock.Lock()
}

func printunlock() {
	printLock.Unlock()
}

// The atomics lock isn't used as a lock for actual atomics. It is used inside
// internal/task.Stack and internal/task.Queue to make sure their operations are
// actually atomic. (This might not actually be needed, since the use in
// sync.Cond doesn't need atomicity).

var atomicsLock task.Mutex

func lockAtomics() interrupt.State {
	atomicsLock.Lock()
	return 0
}

func unlockAtomics(mask interrupt.State) {
	atomicsLock.Unlock()
}
