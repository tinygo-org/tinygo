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

func NumGoroutine() int {
	return int(atomic.LoadUint32(&liveTasks))
}

// Goexit exits the current task because runtime.Goexit was called.
func Goexit() {
	exit(true, nil)
}

// Exit exits the current task after its entry function returns.
func Exit() {
	exit(false, nil)
}

func CoroExit(next *Task) {
	exit(false, next)
}

func exit(goexit bool, next *Task) {
	t := Current()
	if hasReleasableStack {
		t.Exited = true
	}
	if next != nil {
		synctestTaskWake(next)
	}
	exitSynctest(t)
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
	if next != nil {
		scheduleTaskNoWake(next)
	}

	// TODO: explicitly free the stack after switching back to the scheduler.
	Pause()
	runtimeFatal("unreachable")
}
