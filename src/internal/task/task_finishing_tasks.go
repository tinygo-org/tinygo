//go:build scheduler.tasks

package task

// MarkFinishing does nothing for scheduler.tasks because it does not use asyncify heap stacks.
func MarkFinishing() {}

// DetachStack does nothing for scheduler.tasks, which keeps no list of stacks.
func DetachStack() {}
