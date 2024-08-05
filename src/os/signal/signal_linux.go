// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go: build linux

package signal

import (
	"os"
	"syscall"
)

// these functions are linked to the runtime at src/runtime/runtime_unix.go
func signal_enable(sig uint32)
func signal_disable(sig uint32)
func signal_ignore(sig uint32)
func signal_ignored() []bool

func signum(s os.Signal) int {
	return int(s.(syscall.Signal))
}

// Ignore causes the provided signals to be ignored.
// If they are received by the program, nothing will happen
func Ignore(sig ...os.Signal) {
	for _, s := range sig {
		signal_ignore(uint32(signum(s)))
	}
}

// Return true if the provided signal is being ignored
// The runtime keeps track of the ignored signals in the signalIgnored array
func Ignored(sig os.Signal) bool {
	return signal_ignored()[signum(sig)]
}

// TODO: reset all previously set masks
func Reset(sig ...os.Signal) {}
