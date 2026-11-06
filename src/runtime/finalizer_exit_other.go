//go:build !((gc.conservative || gc.precise) && scheduler.asyncify)

package runtime

// releaseStacksAtExit is a no-op on builds that do not keep asyncify goroutine
// stacks reachable after the program exits.
func releaseStacksAtExit() {}
