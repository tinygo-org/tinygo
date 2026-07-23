//go:build scheduler.tasks

package task

// MarkFinishing is a no-op for the stack-based scheduler. Zeroing a finished
// goroutine's stack to drop the stale pointers its returned frames leave behind
// is only implemented for the asyncify scheduler, whose goroutine stacks are
// heap buffers scanned conservatively (see the asyncify MarkFinishing and
// Resume). deadlock and goexit in the cooperative scheduler call this for both.
func MarkFinishing() {}
