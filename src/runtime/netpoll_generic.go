// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (js && wasm) || wasip1 || windows

package runtime

// Network poller descriptor.
//
// No heap pointers.
// For linux to call create Fds with a pollDesc, it needs a ctxRuntime pointer, so use the original pollDesc struct.
// On linux we have a heap.
type pollDesc struct{}

//go:linkname poll_runtime_pollReset internal/poll.runtime_pollReset
func poll_runtime_pollReset(pd *pollDesc, mode int) int {
	println("poll_runtime_pollReset not implemented", pd, mode)
	return 1
}

//go:linkname poll_runtime_pollWait internal/poll.runtime_pollWait
func poll_runtime_pollWait(pd *pollDesc, mode int) int {
	println("poll_runtime_pollWait not implemented", pd, mode)
	return 1
}

//go:linkname poll_runtime_pollSetDeadline internal/poll.runtime_pollSetDeadline
func poll_runtime_pollSetDeadline(pd *pollDesc, d int64, mode int) {
	println("poll_runtime_pollSetDeadline not implemented", pd, d, mode)
}
