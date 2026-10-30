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
	if argv == nil {
		return 0, errors.New("exec: no argv")
	}

	if len(argv) == 0 {
		return 0, errors.New("exec: no argv")
	}

	if attr == nil {
		attr = new(ProcAttr)
	}

	// The pipe closes on exec, so the parent reads EOF when exec succeeds.
	var pipe [2]int
	if err := syscall.Pipe2(pipe[:], syscall.O_CLOEXEC); err != nil {
		return 0, err
	}

	p, err := fork()
	if err != nil {
		syscall.Close(pipe[0])
		syscall.Close(pipe[1])
		return 0, err
	}

	if p == 0 {
		syscall.Close(pipe[0])
		errno, _ := execve(argv0, argv, attr.Env).(syscall.Errno)
		buf := [4]byte{byte(errno), byte(errno >> 8), byte(errno >> 16), byte(errno >> 24)}
		syscall.Write(pipe[1], buf[:])
		syscall.Syscall(syscall.SYS_EXIT_GROUP, 127, 0, 0)
	}

	pid = int(p)
	syscall.Close(pipe[1])
	defer syscall.Close(pipe[0])

	var buf [4]byte
	n, err := syscall.Read(pipe[0], buf[:])
	for err == syscall.EINTR {
		n, err = syscall.Read(pipe[0], buf[:])
	}
	if n == 0 {
		return pid, nil
	}

	// exec failed in the child, so reap it and return its error
	var ws syscall.WaitStatus
	syscall.Wait4(pid, &ws, 0, nil)
	if n == len(buf) {
		return 0, syscall.Errno(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24)
	}
	return 0, syscall.EPIPE
}

// In Golang, the idiomatic way to create a new process is to use the StartProcess function.
// Since the Model of operating system processes in tinygo differs from the one in Golang, we need to implement the StartProcess function differently.
// The startProcess function is a wrapper around the forkExec function, which is a wrapper around the fork and execve system calls.
// The StartProcess function creates a new process by forking the current process and then calling execve to replace the current process with the new process.
// It thereby replaces the newly created process with the specified command and arguments.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	if attr != nil {
		if attr.Dir != "" {
			return nil, ErrNotImplementedDir
		}

		if attr.Sys != nil {
			return nil, ErrNotImplementedSys
		}

		if len(attr.Files) != 0 {
			return nil, ErrNotImplementedFiles
		}
	}

	pid, err := forkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return findProcess(pid)
}
