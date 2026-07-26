//go:build scheduler.tasks || scheduler.asyncify || scheduler.cores

package task

import "sync/atomic"

var (
	mainTask           *Task
	liveTasks          uint32
	mainExitedByGoexit uint32
)

func addLiveTask(t *Task) {
	if mainTask == nil {
		mainTask = t
	}
	atomic.AddUint32(&liveTasks, 1)
}

// Exit exits the current task because runtime.Goexit was called.
func Exit() {
	exit(true)
}

func exit(goexit bool) {
	t := Current()
	remaining := atomic.AddUint32(&liveTasks, ^uint32(0))
	if t == mainTask {
		if goexit {
			if remaining == 0 {
				runtimeFatal("all goroutines are asleep - deadlock!")
			}
			atomic.StoreUint32(&mainExitedByGoexit, 1)
		}
	} else if atomic.LoadUint32(&mainExitedByGoexit) != 0 && remaining == 0 {
		runtimeFatal("all goroutines are asleep - deadlock!")
	}

	// TODO: explicitly free the stack after switching back to the scheduler.
	Pause()
	runtimeFatal("unreachable")
}
