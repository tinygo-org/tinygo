package runtime

// This file contains stub implementations for internal/poll.

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	panic("todo: semacquire")
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	panic("todo: semrelease")
}
