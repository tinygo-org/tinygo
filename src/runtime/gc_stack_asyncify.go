//go:build scheduler.asyncify

package runtime

import "internal/task"

// stackReached is called with the address of the header of a goroutine stack.
func stackReached(header uintptr) { task.StackReached(header) }
