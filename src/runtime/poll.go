package runtime

// This file implements stub functions for internal/poll.

//go:linkname poll_runtime_pollServerInit internal/poll.runtime_pollServerInit
func poll_runtime_pollServerInit() {
	// TODO the "net" pkg calls this, so panic() isn't an option.  Right
	// now, just ignore the call.
}

//go:linkname poll_runtime_pollOpen internal/poll.runtime_pollOpen
func poll_runtime_pollOpen(fd uintptr) (uintptr, int) {
	// TODO the "net" pkg calls this, so panic() isn't an option.  Right
	// now, just ignore the call.
	return 0, 0
}

//go:linkname poll_runtime_pollClose internal/poll.runtime_pollClose
func poll_runtime_pollClose(ctx uintptr) {
	panic("todo: runtime_pollClose")
}

//go:linkname poll_runtime_pollUnblock internal/poll.runtime_pollUnblock
func poll_runtime_pollUnblock(ctx uintptr) {
	panic("todo: runtime_pollUnblock")
}
