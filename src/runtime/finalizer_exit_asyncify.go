//go:build (gc.conservative || gc.precise) && scheduler.asyncify

package runtime

import "internal/task"

// exitDrainSupported reports whether this build has anything to drain at exit.
const exitDrainSupported = true

var exitDrained bool

// drainIfExited releases the stacks of goroutines that the stopped program
// will never resume, then collects and runs the finalizers that were waiting
// for those stacks to become unreachable.
func drainIfExited() {
	if !mainExited || exitDrained {
		return
	}
	exitDrained = true

	task.ReleaseStacksAtExit()

	gcLock.Lock()
	runGC()
	gcLock.Unlock()

	drainFinalizers()
}
