package runtime

// This file implements stub functions for internal/poll.

//go:linkname poll_runtime_pollServerInit internal/poll.runtime_pollServerInit
func poll_runtime_pollServerInit() {
	// fmt.Printf("poll_runtime_pollServerInit not implemented, skipping panic\n")
}

// //go:linkname poll_runtime_pollOpen internal/poll.runtime_pollOpen
// func poll_runtime_pollOpen(fd uintptr) (uintptr, int) {
// 	// fmt.Printf("poll_runtime_pollOpen not implemented, skipping panic\n")
// 	return 0, 0
// }

//go:linkname poll_runtime_pollClose internal/poll.runtime_pollClose
func poll_runtime_pollClose(ctx uintptr) {
	// fmt.Printf("poll_runtime_pollClose not implemented, skipping panic\n")
}

//go:linkname poll_runtime_pollUnblock internal/poll.runtime_pollUnblock
func poll_runtime_pollUnblock(ctx uintptr) {
	// fmt.Printf("poll_runtime_pollUnblock not implemented, skipping panic\n")
}
