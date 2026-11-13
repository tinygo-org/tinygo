// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"errors"
	"runtime"
	"syscall"
)

// The only signal values guaranteed to be present in the os package on all
// systems are os.Interrupt (send the process an interrupt) and os.Kill (force
// the process to exit). On Windows, sending os.Interrupt to a process with
// os.Process.Signal is not implemented; it will return an error instead of
// sending a signal.
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

func (p *Process) wait() (*ProcessState, error) {
	// wait4 with a pid of 0 or less waits for any child in a process group, see wait(2).
	if p.Pid <= 0 {
		return nil, errors.New("os: process not initialized")
	}
	var status syscall.WaitStatus
	pid, err := syscall.Wait4(p.Pid, &status, 0, nil)
	for err == syscall.EINTR {
		pid, err = syscall.Wait4(p.Pid, &status, 0, nil)
	}
	if err != nil {
		return nil, NewSyscallError("wait", err)
	}
	return &ProcessState{pid: pid, status: status}, nil
}

// forkExec starts argv0 with posix_spawn.
func forkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	if argv == nil {
		return 0, errors.New("exec: no argv")
	}

	if len(argv) == 0 {
		return 0, errors.New("exec: no argv")
	}

	return spawn(argv0, argv, attr.Env)
}

// startProcess does not support ProcAttr.Dir, Files or Sys yet.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	if attr == nil {
		return nil, errors.New("exec: no ProcAttr")
	}

	if attr.Dir != "" {
		return nil, ErrNotImplementedDir
	}

	if attr.Sys != nil {
		return nil, ErrNotImplementedSys
	}

	if len(attr.Files) != 0 {
		return nil, ErrNotImplementedFiles
	}

	pid, err := forkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return findProcess(pid)
}
