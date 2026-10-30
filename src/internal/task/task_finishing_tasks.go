//go:build scheduler.tasks

package task

// MarkFinishing does nothing for scheduler.tasks because it does not use asyncify heap stacks.
func MarkFinishing() {}
