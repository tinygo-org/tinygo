// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || netbsd || openbsd || solaris || wasip1 || wasip2 || windows

package os

import (
	"runtime"
	"syscall"
	"unsafe"
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

// Combination of fork and exec, careful to be thread safe.

// https://cs.opensource.google/go/go/+/master:src/syscall/exec_unix.go;l=143?q=forkExec&ss=go%2Fgo
// losely inspired by the golang implementation
// This is a hacky fire-and forget implementation without setting any attributes, using pipes or checking for errors
func forkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	// Convert args to C form.
	var (
		ret uintptr
	)

	argv0p, err := syscall.BytePtrFromString(argv0)
	if err != nil {
		return 0, err
	}
	argvp, err := syscall.SlicePtrFromStrings(argv)
	if err != nil {
		return 0, err
	}
	envvp, err := syscall.SlicePtrFromStrings(attr.Env)
	if err != nil {
		return 0, err
	}

	// pid, _, _ = syscall.Syscall6(syscall.SYS_FORK, 0, 0, 0, 0, 0, 0)
	// 1. fork
	ret, _, _ = syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if ret != 0 {
		// parent
		return int(ret), nil
	} else {
		// 2. exec
		ret, _, _ = syscall.Syscall6(syscall.SYS_EXECVE, uintptr(unsafe.Pointer(argv0p)), uintptr(unsafe.Pointer(&argvp[0])), uintptr(unsafe.Pointer(&envvp[0])), 0, 0, 0)
		if ret != 0 {
			// exec failed
			syscall.Exit(1)
		}
		// 3. TODO: use pipes to communicate back child status
		return int(ret), nil
	}
}

// in regular go this is where the forkExec thingy comes in play
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	pid, err := ForkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return findProcess(pid)
}
