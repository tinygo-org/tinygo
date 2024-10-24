// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && !arm64

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

// This function is a wrapper around the forkExec function, which is a wrapper around the fork and execve system calls.
// The StartProcess function creates a new process by forking the current process and then calling execve to replace the current process with the new process.
// It thereby replaces the newly created process with the specified command and arguments.
// Differences to upstream golang implementation (https://cs.opensource.google/go/go/+/master:src/syscall/exec_unix.go;l=143):
// * No setting of Process Attributes
// * Ignoring Ctty
// * No ForkLocking (might be introduced by #4273)
// * No parent-child communication via pipes (TODO)
// * No waiting for crashes child processes to prohibit zombie process accumulation / Wait status checking (TODO)
func forkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	if len(argv) == 0 {
		return 0, errors.New("exec: no argv")
	}

	if attr == nil {
		attr = new(ProcAttr)
	}

	pid, err = fork()
	if err != nil {
		return 0, err
	} else {
		// else code runs in child, which then should exec the new process
		err = execve(argv0, argv, attr.Env)
		if err != nil {
			// exec failed
			return 0, err
		}
		// 3. TODO: use pipes to communicate back child status
		return pid, nil
	}
}

// In Golang, the idiomatic way to create a new process is to use the StartProcess function.
// Since the Model of operating system processes in tinygo differs from the one in Golang, we need to implement the StartProcess function differently.
// The startProcess function is a wrapper around the forkExec function, which is a wrapper around the fork and execve system calls.
// The StartProcess function creates a new process by forking the current process and then calling execve to replace the current process with the new process.
// It thereby replaces the newly created process with the specified command and arguments.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	pid, err := forkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return findProcess(pid)
}
