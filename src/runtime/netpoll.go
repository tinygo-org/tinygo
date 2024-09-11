// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// For debugging purposes this is used for all target architectures, but this is only valid for linux systems.
// TODO: add linux specific build tags
type pollDesc struct {
	runtimeCtx uintptr
}

func (pd *pollDesc) wait(mode int, isFile bool) error {
	return nil
}

const (
	pollNoError        = 0 // no error
	pollErrClosing     = 1 // descriptor is closed
	pollErrTimeout     = 2 // I/O timeout
	pollErrNotPollable = 3 // general error polling descriptor
)

//go:linkname poll_runtime_pollReset internal/poll.runtime_pollReset
func poll_runtime_pollReset(pd *pollDesc, mode int) int {
	// println("poll_runtime_pollReset not implemented", pd, mode)
	return pollNoError
}

//go:linkname poll_runtime_pollWait internal/poll.runtime_pollWait
func poll_runtime_pollWait(pd *pollDesc, mode int) int {
	// println("poll_runtime_pollWait not implemented", pd, mode)
	return pollNoError
}

//go:linkname poll_runtime_pollSetDeadline internal/poll.runtime_pollSetDeadline
func poll_runtime_pollSetDeadline(pd *pollDesc, d int64, mode int) {
	// println("poll_runtime_pollSetDeadline not implemented", pd, d, mode)
}

//go:linkname poll_runtime_pollOpen internal/poll.runtime_pollOpen
func poll_runtime_pollOpen(fd uintptr) (*pollDesc, int) {
	// println("poll_runtime_pollOpen not implemented", fd)
	return &pollDesc{runtimeCtx: 0x13371337}, pollNoError
}
