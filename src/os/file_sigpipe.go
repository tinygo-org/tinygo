//go:build darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch)

package os

import (
	"errors"
	"syscall"
)

// sigpipe is implemented in package runtime.
func sigpipe()

// epipecheck makes a failed write to stdout or stderr end the process with
// SIGPIPE, as in Go. Other descriptors return EPIPE.
func epipecheck(file *File, e error) {
	if (file == Stdout || file == Stderr) && errors.Is(e, syscall.EPIPE) {
		sigpipe()
	}
}
