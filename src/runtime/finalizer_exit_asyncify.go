//go:build (gc.conservative || gc.precise) && scheduler.asyncify

package runtime

import "internal/task"

// releaseStacksAtExit releases the stacks of goroutines that the stopped
// program will never resume, then collects and runs the finalizers that were
// waiting for those stacks to become unreachable.
func releaseStacksAtExit() {
	task.ReleaseStacksAtExit()

	gcLock.Lock()
	runGC()
	gcLock.Unlock()

	drainFinalizers()
}
