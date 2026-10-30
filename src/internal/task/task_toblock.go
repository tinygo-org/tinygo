//go:build !scheduler.asyncify

package task

// CurrentToBlock returns the goroutine that the caller is about to block.
func CurrentToBlock() *Task {
	return Current()
}
