//go:build !linux || baremetal || nintendoswitch || tinygo.wasm

package runtime

import (
	_ "unsafe"
)

//go:linkname poll_runtime_pollServerInit internal/poll.runtime_pollServerInit
func poll_runtime_pollServerInit() {
}

//go:linkname poll_runtime_pollOpen internal/poll.runtime_pollOpen
func poll_runtime_pollOpen(fd uintptr) (uintptr, int) {
	return 0, 1
}

//go:linkname poll_runtime_pollClose internal/poll.runtime_pollClose
func poll_runtime_pollClose(ctx uintptr) {
}

//go:linkname poll_runtime_pollWait internal/poll.runtime_pollWait
func poll_runtime_pollWait(ctx uintptr, mode int) int {
	return 0
}

//go:linkname poll_runtime_pollWaitCanceled internal/poll.runtime_pollWaitCanceled
func poll_runtime_pollWaitCanceled(ctx uintptr, mode int) {
}

//go:linkname poll_runtime_pollReset internal/poll.runtime_pollReset
func poll_runtime_pollReset(ctx uintptr, mode int) int {
	return 0
}

//go:linkname poll_runtime_pollSetDeadline internal/poll.runtime_pollSetDeadline
func poll_runtime_pollSetDeadline(ctx uintptr, d int64, mode int) {
}

//go:linkname poll_runtime_pollUnblock internal/poll.runtime_pollUnblock
func poll_runtime_pollUnblock(ctx uintptr) {
}

//go:linkname poll_runtime_isPollServerDescriptor internal/poll.runtime_isPollServerDescriptor
func poll_runtime_isPollServerDescriptor(fd uintptr) bool {
	return false
}
