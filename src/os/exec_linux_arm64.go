// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && arm64

package os

import (
	"errors"
	"runtime"
	"syscall"
)

var (
	Interrupt Signal = syscall.SIGINT
	Kill      Signal = syscall.SIGKILL
)

// Keep compatible with golang and always succeed and return new proc with pid on Linux.
func findProcess(pid int) (*Process, error) {
	return &Process{Pid: pid}, nil
}

func (p *Process) release() error {
	// NOOP for unix.
	p.Pid = -1
	// no need for a finalizer anymore
	runtime.SetFinalizer(p, nil)
	return nil
}

// On the aarch64 architecture, the fork system call is not available.
// Therefore, the fork function is implemented to return an error.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	return nil, errors.New("fork not yet supported on aarch64")
}
