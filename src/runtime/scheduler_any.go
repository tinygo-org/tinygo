//go:build !scheduler.none

package runtime

import "internal/task"

// Pause the current task for a given time.
//
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	addSleepTask(task.Current(), nanosecondsToTicks(duration))
	task.Pause()
}

// run is called by the program entry point to execute the go program.
// With a scheduler, init and the main function are invoked in a goroutine before starting the scheduler.
func run() {
	initHeap()
	go func() {
		initAll()
		callMain()
		schedulerDone = true
	}()
	scheduler()
}

const hasScheduler = true
